package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client wraps the GitLab REST API
type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

// Project represents a GitLab project
type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	NameWithNamespace string `json:"name_with_namespace"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
	Description       string `json:"description"`
	WebURL            string `json:"web_url"`
	HTTPURLToRepo     string `json:"http_url_to_repo"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	DefaultBranch     string `json:"default_branch"`
	Visibility        string `json:"visibility"`
	LastActivityAt    string `json:"last_activity_at"`
}

// Branch represents a GitLab branch
type Branch struct {
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
	Default   bool   `json:"default"`
	Commit    Commit `json:"commit"`
}

// Commit represents a GitLab commit
type Commit struct {
	ID             string `json:"id"`
	ShortID        string `json:"short_id"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	AuthorName     string `json:"author_name"`
	AuthorEmail    string `json:"author_email"`
	CommittedDate  string `json:"committed_date"`
}

// Hook represents a GitLab project webhook
type Hook struct {
	ID                int    `json:"id"`
	URL               string `json:"url"`
	PushEvents        bool   `json:"push_events"`
	MergeRequestsEvents bool `json:"merge_requests_events"`
	TagPushEvents     bool   `json:"tag_push_events"`
}

// PushEvent represents a GitLab push webhook payload
type PushEvent struct {
	ObjectKind string `json:"object_kind"`
	Ref        string `json:"ref"`
	Before     string `json:"before"`
	After      string `json:"after"`
	ProjectID  int    `json:"project_id"`
	Project    struct {
		Name              string `json:"name"`
		PathWithNamespace string `json:"path_with_namespace"`
		WebURL            string `json:"web_url"`
	} `json:"project"`
	Commits []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commits"`
}

// NewClient creates a new GitLab API client
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetToken returns the GitLab access token
func (c *Client) GetToken() string {
	return c.token
}

// Ping tests the GitLab connection
func (c *Client) Ping() error {
	_, err := c.doRequest("GET", "/api/v4/version", nil)
	return err
}

// GetCurrentUser returns the authenticated user info
func (c *Client) GetCurrentUser() (map[string]interface{}, error) {
	body, err := c.doRequest("GET", "/api/v4/user", nil)
	if err != nil {
		return nil, err
	}
	var user map[string]interface{}
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}
	return user, nil
}

// ListProjects lists projects accessible by the authenticated user
func (c *Client) ListProjects(search string, page, perPage int) ([]*Project, error) {
	if perPage == 0 {
		perPage = 20
	}
	if page == 0 {
		page = 1
	}

	path := fmt.Sprintf("/api/v4/projects?membership=true&order_by=last_activity_at&sort=desc&page=%d&per_page=%d", page, perPage)
	if search != "" {
		path += "&search=" + search
	}

	body, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var projects []*Project
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, fmt.Errorf("parse projects: %w", err)
	}
	return projects, nil
}

// GetProject gets a single project by ID or path
func (c *Client) GetProject(projectID string) (*Project, error) {
	path := fmt.Sprintf("/api/v4/projects/%s", projectID)
	body, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(body, &project); err != nil {
		return nil, fmt.Errorf("parse project: %w", err)
	}
	return &project, nil
}

// ListBranches lists branches of a project
func (c *Client) ListBranches(projectID string, search string) ([]*Branch, error) {
	path := fmt.Sprintf("/api/v4/projects/%s/repository/branches?per_page=50", projectID)
	if search != "" {
		path += "&search=" + search
	}

	body, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var branches []*Branch
	if err := json.Unmarshal(body, &branches); err != nil {
		return nil, fmt.Errorf("parse branches: %w", err)
	}
	return branches, nil
}

// GetLatestCommit gets the latest commit on a branch
func (c *Client) GetLatestCommit(projectID, branch string) (*Commit, error) {
	path := fmt.Sprintf("/api/v4/projects/%s/repository/commits?ref_name=%s&per_page=1", projectID, branch)
	body, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var commits []*Commit
	if err := json.Unmarshal(body, &commits); err != nil {
		return nil, fmt.Errorf("parse commits: %w", err)
	}
	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits on branch %s", branch)
	}
	return commits[0], nil
}

// CreateWebhook creates a webhook for a project
func (c *Client) CreateWebhook(projectID string, hookURL string, pushEvents, mrEvents bool) (*Hook, error) {
	path := fmt.Sprintf("/api/v4/projects/%s/hooks", projectID)
	payload := fmt.Sprintf(`{"url":"%s","push_events":%t,"merge_requests_events":%t,"token":"my-cloud-webhook-secret"}`,
		hookURL, pushEvents, mrEvents)

	body, err := c.doRequest("POST", path, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	var hook Hook
	if err := json.Unmarshal(body, &hook); err != nil {
		return nil, fmt.Errorf("parse hook: %w", err)
	}
	return &hook, nil
}

// ListWebhooks lists webhooks for a project
func (c *Client) ListWebhooks(projectID string) ([]*Hook, error) {
	path := fmt.Sprintf("/api/v4/projects/%s/hooks", projectID)
	body, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var hooks []*Hook
	if err := json.Unmarshal(body, &hooks); err != nil {
		return nil, fmt.Errorf("parse hooks: %w", err)
	}
	return hooks, nil
}

// SetCommitStatus sets the build status for a commit
func (c *Client) SetCommitStatus(projectID, commitSHA, state, name, targetURL, description string) error {
	path := fmt.Sprintf("/api/v4/projects/%s/statuses/%s", projectID, commitSHA)
	payload := fmt.Sprintf(`{"state":"%s","name":"%s","target_url":"%s","description":"%s"}`,
		state, name, targetURL, description)

	_, err := c.doRequest("POST", path, strings.NewReader(payload))
	return err
}

// doRequest executes an HTTP request against the GitLab API
func (c *Client) doRequest(method, path string, body io.Reader) ([]byte, error) {
	url := c.baseURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gitlab request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("gitlab API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
