package main

import (
	"flag"
	"fmt"
	"log"
)

// グローバル変数でノード管理
var allNodes []*Node

func main() {
	nodeCount := flag.Int("nodes", 10, "Number of nodes")
	basePort := flag.Int("base-port", 18000, "Base port number")
	adminPort := flag.Int("admin-port", 17999, "Admin service port")
	flag.Parse()

	log.Printf("Starting %d nodes...", *nodeCount)

	// ノードインスタンスを作成
	allNodes = make([]*Node, *nodeCount)

	// 全ノードを並行起動（バックグラウンド）
	for i := 0; i < *nodeCount; i++ {
		node := createNode(i, *basePort, *nodeCount)
		allNodes[i] = node
		go startHTTPServer(node)
	}

	log.Printf("All %d nodes started successfully", *nodeCount)
	log.Printf("")
	log.Printf("Node interaction:")
	log.Printf("  Status:  curl localhost:%d/status", *basePort)
	log.Printf("  Gossip:  curl -X POST localhost:%d/trigger", *basePort)
	log.Printf("  Set:     curl -X POST 'localhost:%d/set?value=hello'", *basePort)
	log.Printf("")
	log.Printf("Admin service:")
	log.Printf("  Cluster info: curl localhost:%d/cluster", *adminPort)
	log.Printf("  Node list:    curl localhost:%d/nodes", *adminPort)
	log.Printf("  Health check: curl localhost:%d/health", *adminPort)
	log.Printf("")

	// 管理サービスをメイン実行（フォアグラウンド）
	// Ctrl+Cで全体が終了する
	startAdminServer(*adminPort, *nodeCount, *basePort)
}

func createNode(nodeIndex, basePort, totalNodes int) *Node {
	nodeID := fmt.Sprintf("node-%d", nodeIndex)
	address := fmt.Sprintf("localhost:%d", basePort+nodeIndex)

	// フルメッシュのピアリストを動的生成
	var peers []string
	for i := 0; i < totalNodes; i++ {
		if i != nodeIndex { // 自分以外
			peers = append(peers, fmt.Sprintf("localhost:%d", basePort+i))
		}
	}

	node := &Node{
		ID:       nodeID,
		Address:  address,
		Peers:    peers,
		Value:    "initial-state",
		LastSeen: 0,
	}

	log.Printf("Starting node %s on %s", node.ID, node.Address)
	return node
}
