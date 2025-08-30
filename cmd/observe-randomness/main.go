package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hassaku63/gossip-concept/internal/client"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

type targetCount struct {
	nodeID int
	count  int
}

func main() {
	var (
		rounds    = flag.Int("rounds", 100, "Number of gossip rounds to execute")
		adminPort = flag.Int("admin-port", 17999, "Admin service port")
		basePort  = flag.Int("base-port", 0, "Base port (auto-detect from admin API if 0)")
		nodeCount = flag.Int("nodes", 0, "Number of nodes (auto-detect from admin API if 0)")
	)
	flag.Parse()

	fmt.Println("=== Gossip Randomness Observer ===")

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

	expectedPeers := actualNodeCount - 1
	expected := *rounds / expectedPeers

	fmt.Printf("Parameters:\n")
	fmt.Printf("  Rounds: %d\n", *rounds)
	fmt.Printf("  Base Port: %d\n", actualBasePort)
	fmt.Printf("  Nodes: %d\n", actualNodeCount)
	fmt.Printf("  Expected per node: ~%d times (for %d peers)\n", expected, expectedPeers)
	fmt.Printf("\n")

	// Check if source node is healthy
	if _, err := gossipClient.GetStatus(actualBasePort); err != nil {
		log.Fatalf("Source node (port %d) is not responding: %v", actualBasePort, err)
	}

	fmt.Printf("Executing %d gossip rounds from node-0...\n", *rounds)

	// Track targets
	targetCounts := make(map[int]int)

	// Execute gossip rounds
	for i := 1; i <= *rounds; i++ {
		trigger, err := gossipClient.TriggerGossip(actualBasePort)
		if err != nil {
			fmt.Print("!")
		} else if trigger.Status == "sent" && trigger.Target != "" {
			// Extract port number from target (localhost:18001 -> 18001)
			parts := strings.Split(trigger.Target, ":")
			if len(parts) == 2 {
				if targetPort, err := strconv.Atoi(parts[1]); err == nil {
					nodeID := targetPort - actualBasePort
					if nodeID >= 1 && nodeID < actualNodeCount {
						targetCounts[nodeID]++
					}
				}
			}
			fmt.Print(".")
		} else {
			fmt.Print("!")
		}

		// Progress indicator
		if i%100 == 0 {
			fmt.Printf(" [%d/%d]\n", i, *rounds)
		}

		// Small delay to avoid overwhelming the server
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("\n\n")

	// Analyze results
	analyzeResults(targetCounts, actualNodeCount, *rounds, expected)
}

func analyzeResults(targetCounts map[int]int, nodeCount, totalRounds, expected int) {
	fmt.Println("=== Distribution Analysis ===")
	fmt.Println()

	fmt.Println("Actual distribution:")

	// Convert to slice for sorting
	var results []targetCount
	totalReceived := 0
	for i := 1; i < nodeCount; i++ {
		count := targetCounts[i]
		results = append(results, targetCount{nodeID: i, count: count})
		totalReceived += count
	}

	// Sort by node ID
	sort.Slice(results, func(i, j int) bool {
		return results[i].nodeID < results[j].nodeID
	})

	// Display results
	minCount, maxCount := math.MaxInt32, 0
	for _, result := range results {
		if result.count < minCount {
			minCount = result.count
		}
		if result.count > maxCount {
			maxCount = result.count
		}

		// Determine color based on deviation from expected
		deviation := int(math.Abs(float64(result.count - expected)))
		color := colorGreen
		if expected > 0 && deviation > expected/3 {
			color = colorRed
		}

		// Create bar graph (scale appropriately for large clusters)
		barScale := 2
		if nodeCount > 20 {
			barScale = 4
		}
		barLength := result.count / barScale
		bar := strings.Repeat("█", barLength)

		// Only show nodes that received gossip or if cluster is small
		if nodeCount <= 20 || result.count > 0 {
			fmt.Printf("  node-%d: %s%3d times%s %s\n",
				result.nodeID, color, result.count, colorReset, bar)
		}
	}

	fmt.Println()

	// Statistical summary
	fmt.Println("=== Statistical Summary ===")
	fmt.Printf("Total gossips: %d\n", totalReceived)
	fmt.Printf("Min: %d, Max: %d\n", minCount, maxCount)
	fmt.Printf("Range: %d\n", maxCount-minCount)
	fmt.Printf("Standard expectation: %d\n", expected)

	// Evaluate randomness
	acceptableRange := expected / 2
	if acceptableRange < 1 {
		acceptableRange = 1
	}

	rangeValue := maxCount - minCount
	if rangeValue <= acceptableRange {
		fmt.Printf("%s✓ Good randomness (well distributed)%s\n", colorGreen, colorReset)
	} else if rangeValue <= acceptableRange*2 {
		fmt.Printf("%s~ Acceptable randomness%s\n", colorYellow, colorReset)
	} else {
		fmt.Printf("%s✗ Poor randomness (highly skewed)%s\n", colorRed, colorReset)
	}

	fmt.Println()
}
