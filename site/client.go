package site

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Intro-Shlok/AutoMate/core"
)

const DefaultBaseURL = "https://intro-shlok.github.io/AutoTest/api/v1"

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchCommands downloads commands.json from the site
func (c *Client) FetchCommands() ([]core.ToolDefinition, error) {
	url := c.BaseURL + "/commands.json"
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch commands: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch commands: HTTP %d", resp.StatusCode)
	}

	var tools []core.ToolDefinition
	if err := json.NewDecoder(resp.Body).Decode(&tools); err != nil {
		return nil, fmt.Errorf("decode commands: %w", err)
	}

	return tools, nil
}

// Sync downloads tool definitions and caches them locally
func Sync(client *Client, cache *core.Cache) (int, error) {
	fmt.Println("Syncing tool definitions from AutoTest site...")

	tools, err := client.FetchCommands()
	if err != nil {
		return 0, fmt.Errorf("fetch from site: %w", err)
	}

	if err := cache.StoreTools(tools); err != nil {
		return 0, fmt.Errorf("cache tools: %w", err)
	}

	if err := cache.SetLastSync(time.Now()); err != nil {
		return 0, fmt.Errorf("update sync timestamp: %w", err)
	}

	return len(tools), nil
}
