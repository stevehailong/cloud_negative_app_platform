package kubecost

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Client Kubecost API client
type Client struct {
	baseURL string
}

// NewClient creates a Kubecost API client
func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

// AllocationResponse Kubecost allocation API response
type AllocationResponse struct {
	Code int `json:"code"`
	Data []struct {
		Name       string            `json:"name"`
		Properties map[string]string `json:"properties"`
		Window     struct {
			Start time.Time `json:"start"`
			End   time.Time `json:"end"`
		} `json:"window"`
		Start           time.Time `json:"start"`
		End             time.Time `json:"end"`
		CPUCost         float64   `json:"cpuCost"`
		GPUCost         float64   `json:"gpuCost"`
		RAMCost         float64   `json:"ramCost"`
		PVCost          float64   `json:"pvCost"`
		NetworkCost     float64   `json:"networkCost"`
		TotalCost       float64   `json:"totalCost"`
		CPUCoreHours    float64   `json:"cpuCoreHours"`
		RAMGBHours      float64   `json:"ramGBHours"`
	} `json:"data"`
}

// NamespaceCost per-namespace cost breakdown
type NamespaceCost struct {
	Namespace    string
	CPUCost      float64
	MemoryCost   float64
	StorageCost  float64
	NetworkCost  float64
	TotalCost    float64
}

// GetAllocation queries Kubecost for current day's allocation
func (c *Client) GetAllocation() ([]NamespaceCost, error) {
	window := "1d"
	url := fmt.Sprintf("%s/model/allocation?window=%s&aggregate=namespace", c.baseURL, window)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("kubecost query failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("kubecost returned %d", resp.StatusCode)
	}

	var result []struct {
		Name       string            `json:"name"`
		Properties map[string]string `json:"properties"`
		CPUCost    float64           `json:"cpuCost"`
		RAMCost    float64           `json:"ramCost"`
		PVCost     float64           `json:"pvCost"`
		NetworkCost float64          `json:"networkCost"`
		TotalCost  float64           `json:"totalCost"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode kubecost response: %w", err)
	}

	var costs []NamespaceCost
	for _, item := range result {
		costs = append(costs, NamespaceCost{
			Namespace:   item.Name,
			CPUCost:     item.CPUCost,
			MemoryCost:  item.RAMCost,
			StorageCost: item.PVCost,
			NetworkCost: item.NetworkCost,
			TotalCost:   item.TotalCost,
		})
	}

	return costs, nil
}

// Ping checks if Kubecost is reachable
func (c *Client) Ping() error {
	url := fmt.Sprintf("%s/model/allocation?window=10m", c.baseURL)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("kubecost returned %d", resp.StatusCode)
	}
	log.Printf("[Kubecost] Connected successfully to %s", c.baseURL)
	return nil
}
