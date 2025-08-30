package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// NodeStatus represents the status of a gossip node
type NodeStatus struct {
	ID       string   `json:"id"`
	Value    string   `json:"value"`
	Peers    []string `json:"peers"`
	LastSeen int64    `json:"last_seen"`
}

// TriggerResponse represents the response from a gossip trigger
type TriggerResponse struct {
	Status string `json:"status"`
	Target string `json:"target"`
}

// GossipClient provides access to gossip node APIs
type GossipClient struct {
	Client *http.Client
}

// NewGossipClient creates a new gossip API client
func NewGossipClient() *GossipClient {
	return &GossipClient{
		Client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

// GetStatus retrieves the status of a specific node
func (c *GossipClient) GetStatus(port int) (*NodeStatus, error) {
	url := fmt.Sprintf("http://localhost:%d/status", port)
	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get status from port %d: %w", port, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("node at port %d returned status %d", port, resp.StatusCode)
	}

	var status NodeStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode status from port %d: %w", port, err)
	}

	return &status, nil
}

// TriggerGossip triggers a gossip round on the specified node
func (c *GossipClient) TriggerGossip(port int) (*TriggerResponse, error) {
	url := fmt.Sprintf("http://localhost:%d/trigger", port)
	resp, err := c.Client.Post(url, "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger gossip on port %d: %w", port, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("node at port %d returned status %d", port, resp.StatusCode)
	}

	var trigger TriggerResponse
	if err := json.NewDecoder(resp.Body).Decode(&trigger); err != nil {
		return nil, fmt.Errorf("failed to decode trigger response from port %d: %w", port, err)
	}

	return &trigger, nil
}

// SetValue sets a new value on the specified node
func (c *GossipClient) SetValue(port int, value string) error {
	baseURL := fmt.Sprintf("http://localhost:%d/set", port)
	params := url.Values{}
	params.Add("value", value)
	url := baseURL + "?" + params.Encode()

	resp, err := c.Client.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to set value on port %d: %w", port, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("node at port %d returned status %d", port, resp.StatusCode)
	}

	return nil
}

// CheckAllNodesHealthy checks if all nodes in the given port range are responding
func (c *GossipClient) CheckAllNodesHealthy(basePort, nodeCount int) (int, error) {
	healthy := 0
	for i := 0; i < nodeCount; i++ {
		port := basePort + i
		if _, err := c.GetStatus(port); err == nil {
			healthy++
		}
	}
	return healthy, nil
}
