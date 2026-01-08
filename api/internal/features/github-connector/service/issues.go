package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// CreateIssue creates a new GitHub issue in the specified repository
func (c *GithubConnectorService) CreateIssue(userID string, req *types.CreateIssueRequest) (*types.IssueResponse, error) {
	accessToken, err := c.getAccessTokenForUser(userID, req.ConnectorID)
	if err != nil {
		return nil, err
	}

	repoFullName := fmt.Sprintf("%s/%s", req.RepositoryOwner, req.RepositoryName)
	url := fmt.Sprintf("%s/repos/%s/issues", githubAPIBaseURL, repoFullName)

	requestBody := map[string]interface{}{
		"title": req.Title,
		"body":  req.Body,
	}
	if len(req.Labels) > 0 {
		requestBody["labels"] = req.Labels
	}
	if len(req.Assignees) > 0 {
		requestBody["assignees"] = req.Assignees
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to marshal request body: %s", err.Error()), "")
		return nil, err
	}

	client := &http.Client{}
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	httpReq.Header.Set("Accept", "application/vnd.github.v3+json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "nixopus")

	resp, err := client.Do(httpReq)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		c.logger.Log(logger.Error, fmt.Sprintf("GitHub API error: %s - %s", resp.Status, string(bodyBytes)), "")
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var issue types.IssueResponse
	if err := json.Unmarshal(bodyBytes, &issue); err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to unmarshal response: %s", err.Error()), "")
		return nil, err
	}

	return &issue, nil
}

// UpdateIssue updates an existing GitHub issue
func (c *GithubConnectorService) UpdateIssue(userID string, req *types.UpdateIssueRequest) (*types.IssueResponse, error) {
	accessToken, err := c.getAccessTokenForUser(userID, req.ConnectorID)
	if err != nil {
		return nil, err
	}

	repoFullName := fmt.Sprintf("%s/%s", req.RepositoryOwner, req.RepositoryName)
	url := fmt.Sprintf("%s/repos/%s/issues/%d", githubAPIBaseURL, repoFullName, req.IssueNumber)

	// Prepare request body with only provided fields
	requestBody := make(map[string]interface{})
	if req.Title != nil {
		requestBody["title"] = *req.Title
	}
	if req.Body != nil {
		requestBody["body"] = *req.Body
	}
	if req.State != nil {
		requestBody["state"] = *req.State
	}
	if len(req.Labels) > 0 {
		requestBody["labels"] = req.Labels
	}
	if len(req.Assignees) > 0 {
		requestBody["assignees"] = req.Assignees
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to marshal request body: %s", err.Error()), "")
		return nil, err
	}

	client := &http.Client{}
	httpReq, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	httpReq.Header.Set("Accept", "application/vnd.github.v3+json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "nixopus")

	resp, err := client.Do(httpReq)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		c.logger.Log(logger.Error, fmt.Sprintf("GitHub API error: %s - %s", resp.Status, string(bodyBytes)), "")
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var issue types.IssueResponse
	if err := json.Unmarshal(bodyBytes, &issue); err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to unmarshal response: %s", err.Error()), "")
		return nil, err
	}

	return &issue, nil
}

// CommentOnIssue adds a comment to an existing GitHub issue
func (c *GithubConnectorService) CommentOnIssue(userID string, req *types.CommentOnIssueRequest) (*types.IssueCommentResponse, error) {
	accessToken, err := c.getAccessTokenForUser(userID, req.ConnectorID)
	if err != nil {
		return nil, err
	}

	repoFullName := fmt.Sprintf("%s/%s", req.RepositoryOwner, req.RepositoryName)
	url := fmt.Sprintf("%s/repos/%s/issues/%d/comments", githubAPIBaseURL, repoFullName, req.IssueNumber)

	requestBody := map[string]string{
		"body": req.Body,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to marshal request body: %s", err.Error()), "")
		return nil, err
	}

	client := &http.Client{}
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	httpReq.Header.Set("Accept", "application/vnd.github.v3+json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "nixopus")

	resp, err := client.Do(httpReq)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		c.logger.Log(logger.Error, fmt.Sprintf("GitHub API error: %s - %s", resp.Status, string(bodyBytes)), "")
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var comment types.IssueCommentResponse
	if err := json.Unmarshal(bodyBytes, &comment); err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to unmarshal response: %s", err.Error()), "")
		return nil, err
	}

	return &comment, nil
}

// ListIssues lists GitHub issues for the specified repository
func (c *GithubConnectorService) ListIssues(userID string, req *types.ListIssuesRequest) ([]types.IssueResponse, error) {
	accessToken, err := c.getAccessTokenForUser(userID, req.ConnectorID)
	if err != nil {
		return nil, err
	}

	repoFullName := fmt.Sprintf("%s/%s", req.RepositoryOwner, req.RepositoryName)
	apiURL := fmt.Sprintf("%s/repos/%s/issues", githubAPIBaseURL, repoFullName)

	// Build query parameters
	queryParams := url.Values{}
	if req.State != "" {
		queryParams.Set("state", req.State)
	}
	if len(req.Labels) > 0 {
		queryParams.Set("labels", strings.Join(req.Labels, ","))
	}
	if req.Assignee != "" {
		queryParams.Set("assignee", req.Assignee)
	}
	if req.Creator != "" {
		queryParams.Set("creator", req.Creator)
	}
	if req.Page > 0 {
		queryParams.Set("page", fmt.Sprintf("%d", req.Page))
	}
	if req.PerPage > 0 {
		queryParams.Set("per_page", fmt.Sprintf("%d", req.PerPage))
	}

	if len(queryParams) > 0 {
		apiURL += "?" + queryParams.Encode()
	}

	client := &http.Client{}
	httpReq, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	httpReq.Header.Set("Accept", "application/vnd.github.v3+json")
	httpReq.Header.Set("User-Agent", "nixopus")

	resp, err := client.Do(httpReq)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		c.logger.Log(logger.Error, fmt.Sprintf("GitHub API error: %s - %s", resp.Status, string(bodyBytes)), "")
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var issues []types.IssueResponse
	if err := json.Unmarshal(bodyBytes, &issues); err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to unmarshal response: %s", err.Error()), "")
		return nil, err
	}

	return issues, nil
}
