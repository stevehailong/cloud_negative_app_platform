package helm

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Client Helm客户端
type Client struct {
	kubeconfig string
}

// NewClient 创建Helm客户端
func NewClient(kubeconfig string) *Client {
	return &Client{
		kubeconfig: kubeconfig,
	}
}

// ReleaseInfo Helm Release信息
type ReleaseInfo struct {
	Name      string
	Namespace string
	Status    string
	Revision  int
	Chart     string
	Version   string
}

// InstallOrUpgrade 安装或升级Helm Release
func (c *Client) InstallOrUpgrade(ctx context.Context, releaseName, namespace, chartPath string, values map[string]interface{}) error {
	// 将values转换为YAML
	valuesYAML, err := c.valuesToYAML(values)
	if err != nil {
		return fmt.Errorf("failed to convert values to YAML: %v", err)
	}

	// 检查release是否已存在
	exists, err := c.ReleaseExists(ctx, releaseName, namespace)
	if err != nil {
		return fmt.Errorf("failed to check release existence: %v", err)
	}

	var cmd *exec.Cmd
	if exists {
		// 升级
		log.Printf("[Helm] Upgrading release %s in namespace %s", releaseName, namespace)
		cmd = exec.CommandContext(ctx, "helm", "upgrade", releaseName, chartPath,
			"--namespace", namespace,
			"--values", "-",
			"--kubeconfig", c.kubeconfig,
		)
	} else {
		// 安装
		log.Printf("[Helm] Installing release %s in namespace %s", releaseName, namespace)
		cmd = exec.CommandContext(ctx, "helm", "install", releaseName, chartPath,
			"--namespace", namespace,
			"--create-namespace",
			"--values", "-",
			"--kubeconfig", c.kubeconfig,
		)
	}

	// 通过stdin传入values
	cmd.Stdin = strings.NewReader(valuesYAML)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helm command failed: %v, stderr: %s", err, stderr.String())
	}

	log.Printf("[Helm] Successfully deployed %s in namespace %s", releaseName, namespace)
	return nil
}

// Uninstall 卸载Helm Release
func (c *Client) Uninstall(ctx context.Context, releaseName, namespace string) error {
	log.Printf("[Helm] Uninstalling release %s from namespace %s", releaseName, namespace)

	cmd := exec.CommandContext(ctx, "helm", "uninstall", releaseName,
		"--namespace", namespace,
		"--kubeconfig", c.kubeconfig,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helm uninstall failed: %v, stderr: %s", err, stderr.String())
	}

	log.Printf("[Helm] Successfully uninstalled %s from namespace %s", releaseName, namespace)
	return nil
}

// GetRelease 获取Release信息
func (c *Client) GetRelease(ctx context.Context, releaseName, namespace string) (*ReleaseInfo, error) {
	cmd := exec.CommandContext(ctx, "helm", "status", releaseName,
		"--namespace", namespace,
		"--kubeconfig", c.kubeconfig,
		"--output", "json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("helm status failed: %v", err)
	}

	// 解析JSON输出（简化处理，实际应使用json.Unmarshal）
	output := stdout.String()
	info := &ReleaseInfo{
		Name:      releaseName,
		Namespace: namespace,
	}

	// 提取状态信息（简化版）
	if strings.Contains(output, "deployed") {
		info.Status = "deployed"
	} else if strings.Contains(output, "failed") {
		info.Status = "failed"
	}

	return info, nil
}

// ReleaseExists 检查Release是否存在
func (c *Client) ReleaseExists(ctx context.Context, releaseName, namespace string) (bool, error) {
	cmd := exec.CommandContext(ctx, "helm", "status", releaseName,
		"--namespace", namespace,
		"--kubeconfig", c.kubeconfig,
	)

	err := cmd.Run()
	if err != nil {
		// helm status 失败通常意味着release不存在
		return false, nil
	}
	return true, nil
}

// WaitForRelease 等待Release部署完成
func (c *Client) WaitForRelease(ctx context.Context, releaseName, namespace string, timeout time.Duration) error {
	start := time.Now()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			info, err := c.GetRelease(ctx, releaseName, namespace)
			if err != nil {
				continue
			}

			if info.Status == "deployed" {
				return nil
			}

			if info.Status == "failed" {
				return fmt.Errorf("release %s failed to deploy", releaseName)
			}

			if time.Since(start) > timeout {
				return fmt.Errorf("timeout waiting for release %s to be deployed", releaseName)
			}
		}
	}
}

// valuesToYAML 将 values map 转换为标准 YAML
func (c *Client) valuesToYAML(values map[string]interface{}) (string, error) {
	data, err := yaml.Marshal(values)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
