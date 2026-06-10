package slackapp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type SlackClient interface {
	PostMessage(ctx context.Context, channel string, blocks []map[string]interface{}, text string, threadTS string) (string, error)
	Reactions(ctx context.Context, channel string, timestamp string) ([]Reaction, error)
	UserName(ctx context.Context, userID string) (string, error)
}

type Reaction struct {
	Name  string
	Users []string
}

type RateLimitError struct {
	RetryAfter time.Duration
	Message    string
}

func (e *RateLimitError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "slack api rate limited"
}

type WebAPIClient struct {
	token      string
	httpClient *http.Client
}

func NewWebAPIClient(token string) *WebAPIClient {
	return &WebAPIClient{
		token:      token,
		httpClient: http.DefaultClient,
	}
}

func (c *WebAPIClient) PostMessage(ctx context.Context, channel string, blocks []map[string]interface{}, text string, threadTS string) (string, error) {
	payload := map[string]interface{}{
		"channel": channel,
		"text":    text,
	}
	if len(blocks) > 0 {
		payload["blocks"] = blocks
	}
	if threadTS != "" {
		payload["thread_ts"] = threadTS
	}

	var response struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
		TS    string `json:"ts"`
	}
	if err := c.post(ctx, "https://slack.com/api/chat.postMessage", payload, &response); err != nil {
		return "", err
	}
	if !response.OK {
		return "", errors.Errorf("slack chat.postMessage failed: %s", response.Error)
	}
	return response.TS, nil
}

func (c *WebAPIClient) Reactions(ctx context.Context, channel string, timestamp string) ([]Reaction, error) {
	var response struct {
		OK      bool   `json:"ok"`
		Error   string `json:"error"`
		Message struct {
			Reactions []Reaction `json:"reactions"`
		} `json:"message"`
	}
	query := url.Values{}
	query.Set("channel", channel)
	query.Set("timestamp", timestamp)
	if err := c.get(ctx, "https://slack.com/api/reactions.get", query, &response); err != nil {
		return nil, err
	}
	if !response.OK {
		return nil, errors.Errorf("slack reactions.get failed: %s", response.Error)
	}
	return response.Message.Reactions, nil
}

func (c *WebAPIClient) UserName(ctx context.Context, userID string) (string, error) {
	var response struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
		User  struct {
			Name    string `json:"name"`
			Profile struct {
				RealName string `json:"real_name"`
			} `json:"profile"`
		} `json:"user"`
	}
	query := url.Values{}
	query.Set("user", userID)
	if err := c.get(ctx, "https://slack.com/api/users.info", query, &response); err != nil {
		return "", err
	}
	if !response.OK {
		return "", errors.Errorf("slack users.info failed: %s", response.Error)
	}
	if response.User.Profile.RealName != "" {
		return response.User.Profile.RealName, nil
	}
	if response.User.Name != "" {
		return response.User.Name, nil
	}
	return userID, nil
}

func (c *WebAPIClient) get(ctx context.Context, rawURL string, query url.Values, response interface{}) error {
	reqURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		if resp.StatusCode == http.StatusTooManyRequests {
			return rateLimitError(resp)
		}
		return errors.Errorf("slack api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return json.Unmarshal(respBody, response)
}

func (c *WebAPIClient) post(ctx context.Context, url string, payload interface{}, response interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		if resp.StatusCode == http.StatusTooManyRequests {
			return rateLimitError(resp)
		}
		return errors.Errorf("slack api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return json.Unmarshal(respBody, response)
}

func rateLimitError(resp *http.Response) error {
	retryAfter := time.Minute
	if header := resp.Header.Get("Retry-After"); header != "" {
		if seconds, err := strconv.Atoi(header); err == nil && seconds > 0 {
			retryAfter = time.Duration(seconds) * time.Second
		}
	}
	return &RateLimitError{
		RetryAfter: retryAfter,
		Message:    "slack api rate limited",
	}
}
