package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/hassaku63/gossip-concept/internal/client"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

func main() {
	var (
		maxRounds = flag.Int("max-rounds", 100, "Maximum rounds to wait for convergence")
		adminPort = flag.Int("admin-port", 17999, "Admin service port")
		basePort  = flag.Int("base-port", 0, "Base port (auto-detect from admin API if 0)")
		nodeCount = flag.Int("nodes", 0, "Number of nodes (auto-detect from admin API if 0)")
	)
	flag.Parse()

	fmt.Println("=== Gossip Convergence Observer ===")

	// Create clients
	adminClient := client.NewAdminClient(*adminPort)
	gossipClient := client.NewGossipClient()

	// Get cluster configuration
	var actualBasePort, actualNodeCount int
	if *basePort == 0 || *nodeCount == 0 {
		fmt.Printf("Fetching cluster configuration from admin API (port %d)...\n", *adminPort)
		clusterInfo, err := adminClient.GetClusterInfo()
		if err != nil {
			if *basePort == 0 || *nodeCount == 0 {
				log.Fatalf("Failed to get cluster info and no manual override: %v", err)
			}
			fmt.Printf("Warning: Failed to get cluster info, using manual values: %v\n", err)
			actualBasePort = *basePort
			actualNodeCount = *nodeCount
		} else {
			actualBasePort = clusterInfo.BasePort
			actualNodeCount = clusterInfo.NodeCount
			if *basePort != 0 {
				actualBasePort = *basePort
			}
			if *nodeCount != 0 {
				actualNodeCount = *nodeCount
			}
		}
	} else {
		actualBasePort = *basePort
		actualNodeCount = *nodeCount
	}

	fmt.Printf("Parameters:\n")
	fmt.Printf("  Base Port: %d\n", actualBasePort)
	fmt.Printf("  Nodes: %d\n", actualNodeCount)
	fmt.Printf("  Max Rounds: %d\n", *maxRounds)
	fmt.Printf("\n")

	// Set new value on node-0
	newValue := fmt.Sprintf("converged-%d", time.Now().Unix())
	fmt.Printf("Setting new value on node-0: '%s'\n", newValue)

	if err := gossipClient.SetValue(actualBasePort, newValue); err != nil {
		log.Fatalf("Failed to set value on node-0: %v", err)
	}

	fmt.Println("Starting gossip propagation...")
	fmt.Println()

	// Track node states
	nodeUpdated := make([]bool, actualNodeCount)

	// Get initial state
	fmt.Println("Initial state:")
	for i := 0; i < actualNodeCount; i++ {
		port := actualBasePort + i
		status, err := gossipClient.GetStatus(port)
		if err != nil {
			fmt.Printf("  node-%d: ✗ (error: %v)\n", i, err)
			nodeUpdated[i] = false
		} else if status.Value == newValue {
			fmt.Printf("  node-%d: ✓ (source)\n", i)
			nodeUpdated[i] = true
		} else {
			fmt.Printf("  node-%d: ✗ '%s'\n", i, status.Value)
			nodeUpdated[i] = false
		}
	}

	fmt.Println()
	fmt.Println("Propagating...")

	// Execute gossip rounds until convergence
	converged := false
	rounds := 0

	for !converged && rounds < *maxRounds {
		rounds++

		// Find updated nodes
		var updatedNodes []int
		for i := 0; i < actualNodeCount; i++ {
			if nodeUpdated[i] {
				updatedNodes = append(updatedNodes, i)
			}
		}

		// Trigger gossip from a random updated node
		if len(updatedNodes) > 0 {
			senderIndex := updatedNodes[rand.Intn(len(updatedNodes))]
			senderPort := actualBasePort + senderIndex

			trigger, err := gossipClient.TriggerGossip(senderPort)
			if err != nil {
				fmt.Printf("Round %d: node-%d → error (%v)", rounds, senderIndex, err)
			} else if trigger.Status == "sent" && trigger.Target != "" {
				// Extract target port from target address
				targetPort := extractPortFromAddress(trigger.Target)
				if targetPort > 0 {
					targetNode := targetPort - actualBasePort
					fmt.Printf("Round %d: node-%d → %d ", rounds, senderIndex, targetNode)
				} else {
					fmt.Printf("Round %d: node-%d → %s ", rounds, senderIndex, trigger.Target)
				}
			}
		}

		// Short delay
		time.Sleep(100 * time.Millisecond)

		// Check all nodes for convergence
		convergedCount := 0
		var newlyUpdated []int

		for i := 0; i < actualNodeCount; i++ {
			port := actualBasePort + i
			status, err := gossipClient.GetStatus(port)
			if err == nil && status.Value == newValue {
				if !nodeUpdated[i] {
					newlyUpdated = append(newlyUpdated, i)
					nodeUpdated[i] = true
				}
				convergedCount++
			}
		}

		// Show newly updated nodes
		if len(newlyUpdated) > 0 {
			fmt.Printf(" [NEW:")
			for _, nodeID := range newlyUpdated {
				fmt.Printf(" node-%d", nodeID)
			}
			fmt.Printf("]")
		}

		fmt.Printf(" (%d/%d converged)\n", convergedCount, actualNodeCount)

		// Check if fully converged
		if convergedCount == actualNodeCount {
			converged = true
		}

		// Progress indicator every 5 rounds
		if rounds%5 == 0 && !converged {
			fmt.Printf("  Progress: %d/%d nodes updated after %d rounds\n", convergedCount, actualNodeCount, rounds)
		}
	}

	fmt.Println()

	// Show results
	showResults(converged, rounds, actualNodeCount, newValue, actualBasePort, gossipClient)
}

func extractPortFromAddress(address string) int {
	// Extract port from "localhost:18001" format
	if len(address) > 10 && address[:9] == "localhost" {
		portStr := address[10:]
		if port := parseInt(portStr); port > 0 {
			return port
		}
	}
	return 0
}

func parseInt(s string) int {
	result := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			return 0
		}
	}
	return result
}

func showResults(converged bool, rounds, nodeCount int, expectedValue string, basePort int, gossipClient *client.GossipClient) {
	fmt.Println("=== Results ===")
	fmt.Println()

	if converged {
		fmt.Printf("%s✓ Converged successfully in %d rounds!%s\n", colorGreen, rounds, colorReset)

		fmt.Println()
		fmt.Println("Analysis:")
		fmt.Printf("  Nodes: %d\n", nodeCount)
		fmt.Printf("  Rounds: %d\n", rounds)

		// Calculate efficiency (avoiding floating point)
		efficiency := (rounds * 10) / nodeCount
		fmt.Printf("  Efficiency: %d.%d rounds per node\n", efficiency/10, efficiency%10)

		// Evaluate performance
		expectedMax := nodeCount * 3
		if rounds <= nodeCount*2 {
			fmt.Printf("  %sExcellent - Very efficient propagation%s\n", colorGreen, colorReset)
		} else if rounds <= expectedMax {
			fmt.Printf("  %sGood - Expected range for %d nodes%s\n", colorGreen, nodeCount, colorReset)
		} else if rounds <= expectedMax*2 {
			fmt.Printf("  %sAcceptable - Slightly slow but normal%s\n", colorYellow, colorReset)
		} else {
			fmt.Printf("  %sPoor - Unusually slow propagation%s\n", colorRed, colorReset)
		}
	} else {
		fmt.Printf("%s✗ Failed to converge within %d rounds%s\n", colorRed, rounds, colorReset)

		fmt.Println()
		fmt.Println("Final state:")
		for i := 0; i < nodeCount; i++ {
			port := basePort + i
			status, err := gossipClient.GetStatus(port)
			if err != nil {
				fmt.Printf("  node-%d: ✗ (error: %v)\n", i, err)
			} else if status.Value == expectedValue {
				fmt.Printf("  node-%d: ✓\n", i)
			} else {
				fmt.Printf("  node-%d: ✗ (still has '%s')\n", i, status.Value)
			}
		}
	}

	fmt.Println()
}
