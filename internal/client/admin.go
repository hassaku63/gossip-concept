package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ClusterInfo represents cluster configuration from admin API
type ClusterInfo struct {
	NodeCount int    `json:"node_count"`
	BasePort  int    `json:"base_port"`
	AdminPort int    `json:"admin_port"`
	Topology  string `json:"topology"`
	StartedAt int64  `json:"started_at"`
}

// NodeInfo represents node information from admin API
type NodeInfo struct {
	ID        string `json:"id"`
	Port      int    `json:"port"`
	Address   string `json:"address"`
	Value     string `json:"value"`
	PeerCount int    `json:"peer_count"`
	LastSeen  int64  `json:"last_seen"`
}

// AdminClient provides access to the gossip cluster admin API
type AdminClient struct {
	BaseURL string
	Client  *http.Client
}

// NewAdminClient creates a new admin API client
func NewAdminClient(adminPort int) *AdminClient {
	return &AdminClient{
		BaseURL: fmt.Sprintf("http://localhost:%d", adminPort),
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetClusterInfo retrieves cluster configuration
func (c *AdminClient) GetClusterInfo() (*ClusterInfo, error) {
	resp, err := c.Client.Get(c.BaseURL + "/cluster")
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("admin API returned status %d", resp.StatusCode)
	}

	var info ClusterInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode cluster info: %w", err)
	}

	return &info, nil
}

// GetNodes retrieves information about all nodes
func (c *AdminClient) GetNodes() ([]NodeInfo, error) {
	resp, err := c.Client.Get(c.BaseURL + "/nodes")
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("admin API returned status %d", resp.StatusCode)
	}

	var nodes []NodeInfo
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, fmt.Errorf("failed to decode nodes info: %w", err)
	}

	return nodes, nil
}
