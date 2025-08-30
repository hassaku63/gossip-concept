package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// ClusterInfo represents the cluster configuration
type ClusterInfo struct {
	NodeCount int    `json:"node_count"`
	BasePort  int    `json:"base_port"`
	AdminPort int    `json:"admin_port"`
	Topology  string `json:"topology"`
	StartedAt int64  `json:"started_at"`
}

// NodeInfo represents a node in the cluster for admin API
type NodeInfo struct {
	ID        string `json:"id"`
	Port      int    `json:"port"`
	Address   string `json:"address"`
	Value     string `json:"value"`
	PeerCount int    `json:"peer_count"`
	LastSeen  int64  `json:"last_seen"`
}

// HealthStatus represents the health of a node
type HealthStatus struct {
	ID      string `json:"id"`
	Port    int    `json:"port"`
	Healthy bool   `json:"healthy"`
	Error   string `json:"error,omitempty"`
}

var clusterStartTime = time.Now().Unix()

func startAdminServer(adminPort, nodeCount, basePort int) {
	mux := http.NewServeMux()

	// クラスター情報エンドポイント
	mux.HandleFunc("/cluster", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		info := ClusterInfo{
			NodeCount: nodeCount,
			BasePort:  basePort,
			AdminPort: adminPort,
			Topology:  "full-mesh",
			StartedAt: clusterStartTime,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})

	// 全ノード情報エンドポイント
	mux.HandleFunc("/nodes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		nodes := make([]NodeInfo, len(allNodes))
		for i, node := range allNodes {
			nodes[i] = NodeInfo{
				ID:        node.ID,
				Port:      basePort + i,
				Address:   node.Address,
				Value:     node.GetValue(),
				PeerCount: len(node.Peers),
				LastSeen:  node.LastSeen,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nodes)
	})

	// ヘルスチェックエンドポイント
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		health := make([]HealthStatus, nodeCount)
		allHealthy := true

		for i := 0; i < nodeCount; i++ {
			port := basePort + i
			health[i] = HealthStatus{
				ID:   fmt.Sprintf("node-%d", i),
				Port: port,
			}

			// 各ノードのステータスエンドポイントをチェック
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/status", port))
			if err != nil {
				health[i].Healthy = false
				health[i].Error = err.Error()
				allHealthy = false
			} else {
				resp.Body.Close()
				health[i].Healthy = resp.StatusCode == http.StatusOK
				if !health[i].Healthy {
					health[i].Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
					allHealthy = false
				}
			}
		}

		response := map[string]interface{}{
			"all_healthy": allHealthy,
			"nodes":       health,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// ルートエンドポイント（管理サービスの情報）
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		info := map[string]interface{}{
			"service": "gossip-cluster-admin",
			"version": "phase0",
			"endpoints": []string{
				"/cluster - Cluster configuration",
				"/nodes - All node information",
				"/health - Health check for all nodes",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})

	log.Printf("Admin server starting on port %d (foreground)", adminPort)
	log.Printf("Press Ctrl+C to stop all services")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", adminPort), mux))
}
