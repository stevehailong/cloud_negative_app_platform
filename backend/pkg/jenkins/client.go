package jenkins

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// Client wraps the Jenkins HTTP API
type Client struct {
	baseURL  string
	username string
	token    string
	client   *http.Client
	crumb    string
	crumbField string
}

// NewClient creates a new Jenkins client
func NewClient(baseURL, username, token string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		token:    token,
		client:   &http.Client{Timeout: 60 * time.Second, Jar: jar},
	}
}

// getCrumb fetches the CSRF crumb from Jenkins
func (c *Client) getCrumb() error {
	reqURL := fmt.Sprintf("%s/crumbIssuer/api/json", c.baseURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.username, c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// CSRF might be disabled, clear crumb
		c.crumb = ""
		c.crumbField = ""
		return nil
	}

	var result struct {
		Crumb             string `json:"crumb"`
		CrumbRequestField string `json:"crumbRequestField"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	c.crumb = result.Crumb
	c.crumbField = result.CrumbRequestField
	return nil
}

// addCrumb adds the CSRF crumb header to a request.
// Must be called AFTER getCrumb() so the cookie jar has the session cookie.
func (c *Client) addCrumb(req *http.Request) {
	if c.crumb != "" && c.crumbField != "" {
		req.Header.Set(c.crumbField, c.crumb)
	}
}

// BuildInfo represents a Jenkins build
type BuildInfo struct {
	Number    int    `json:"number"`
	URL       string `json:"url"`
	Result    string `json:"result"`
	Building  bool   `json:"building"`
	Duration  int64  `json:"duration"`
	Timestamp int64  `json:"timestamp"`
}

// QueueItem represents a queued build
type QueueItem struct {
	ID         int    `json:"id"`
	URL        string `json:"url"`
	Executable struct {
		Number int    `json:"number"`
		URL    string `json:"url"`
	} `json:"executable"`
}

// TriggerBuild triggers a Jenkins job build with parameters
func (c *Client) TriggerBuild(jobName string, params map[string]string) (int, error) {
	var reqURL string
	var formData string

	if len(params) > 0 {
		reqURL = fmt.Sprintf("%s/job/%s/buildWithParameters", c.baseURL, url.PathEscape(jobName))
		form := url.Values{}
		for k, v := range params {
			form.Set(k, v)
		}
		formData = form.Encode()
	} else {
		reqURL = fmt.Sprintf("%s/job/%s/build", c.baseURL, url.PathEscape(jobName))
	}

	// Ensure we have a fresh crumb
	c.crumb = ""
	c.getCrumb()

	var body io.Reader
	if formData != "" {
		body = strings.NewReader(formData)
	}

	req, err := http.NewRequest("POST", reqURL, body)
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(c.username, c.token)
	c.addCrumb(req)
	if formData != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to trigger build: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		truncated := string(bodyBytes)
		if len(truncated) > 200 {
			truncated = truncated[:200]
		}
		return 0, fmt.Errorf("jenkins returned status %d: %s", resp.StatusCode, truncated)
	}

	// Extract queue location to get build number later
	location := resp.Header.Get("Location")
	if location != "" {
		// Wait for queue item to start and get build number
		queueID := extractQueueID(location)
		if queueID > 0 {
			return c.waitForBuildNumber(queueID)
		}
	}

	// Fallback: get last build number
	build, err := c.GetLastBuild(jobName)
	if err != nil {
		return 0, nil
	}
	return build.Number, nil
}

// GetBuild gets build info by number
func (c *Client) GetBuild(jobName string, buildNumber int) (*BuildInfo, error) {
	reqURL := fmt.Sprintf("%s/job/%s/%d/api/json", c.baseURL, url.PathEscape(jobName), buildNumber)
	return c.getBuildInfo(reqURL)
}

// GetLastBuild gets the last build info
func (c *Client) GetLastBuild(jobName string) (*BuildInfo, error) {
	reqURL := fmt.Sprintf("%s/job/%s/lastBuild/api/json", c.baseURL, url.PathEscape(jobName))
	return c.getBuildInfo(reqURL)
}

// GetBuildLog gets the console output of a build
func (c *Client) GetBuildLog(jobName string, buildNumber int) (string, error) {
	reqURL := fmt.Sprintf("%s/job/%s/%d/consoleText", c.baseURL, url.PathEscape(jobName), buildNumber)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.username, c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

// WaitForBuildComplete polls until the build is done
func (c *Client) WaitForBuildComplete(jobName string, buildNumber int, timeout time.Duration) (*BuildInfo, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		build, err := c.GetBuild(jobName, buildNumber)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}
		if !build.Building {
			return build, nil
		}
		time.Sleep(5 * time.Second)
	}
	return nil, fmt.Errorf("build %d timed out after %v", buildNumber, timeout)
}

// Ping checks if Jenkins is reachable
func (c *Client) Ping() error {
	reqURL := fmt.Sprintf("%s/api/json", c.baseURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.username, c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("jenkins unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jenkins returned status %d", resp.StatusCode)
	}
	return nil
}

// CreateJob creates a Jenkins job from XML config
func (c *Client) CreateJob(jobName, configXML string) error {
	reqURL := fmt.Sprintf("%s/createItem?name=%s", c.baseURL, url.QueryEscape(jobName))

	// Ensure we have a fresh crumb
	c.crumb = ""
	c.getCrumb()

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(configXML))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.username, c.token)
	c.addCrumb(req)
	req.Header.Set("Content-Type", "text/xml")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create job failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// UpdateJob updates an existing Jenkins job config
func (c *Client) UpdateJob(jobName, configXML string) error {
	reqURL := fmt.Sprintf("%s/job/%s/config.xml", c.baseURL, url.PathEscape(jobName))

	c.crumb = ""
	c.getCrumb()

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(configXML))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.username, c.token)
	c.addCrumb(req)
	req.Header.Set("Content-Type", "text/xml")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update job failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}


func (c *Client) JobExists(jobName string) bool {
	reqURL := fmt.Sprintf("%s/job/%s/api/json", c.baseURL, url.PathEscape(jobName))
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.SetBasicAuth(c.username, c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (c *Client) getBuildInfo(reqURL string) (*BuildInfo, error) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.username, c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jenkins returned status %d", resp.StatusCode)
	}

	var build BuildInfo
	if err := json.NewDecoder(resp.Body).Decode(&build); err != nil {
		return nil, err
	}
	return &build, nil
}

func (c *Client) waitForBuildNumber(queueID int) (int, error) {
	reqURL := fmt.Sprintf("%s/queue/item/%d/api/json", c.baseURL, queueID)
	deadline := time.Now().Add(60 * time.Second)

	for time.Now().Before(deadline) {
		req, _ := http.NewRequest("GET", reqURL, nil)
		req.SetBasicAuth(c.username, c.token)

		resp, err := c.client.Do(req)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		var item QueueItem
		json.NewDecoder(resp.Body).Decode(&item)
		resp.Body.Close()

		if item.Executable.Number > 0 {
			return item.Executable.Number, nil
		}
		time.Sleep(2 * time.Second)
	}
	return 0, fmt.Errorf("queue item %d did not start within timeout", queueID)
}

func extractQueueID(location string) int {
	// Location format: http://jenkins:8080/queue/item/123/
	parts := strings.Split(strings.TrimRight(location, "/"), "/")
	if len(parts) > 0 {
		var id int
		fmt.Sscanf(parts[len(parts)-1], "%d", &id)
		return id
	}
	return 0
}
