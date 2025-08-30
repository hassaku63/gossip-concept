package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type GossipMessage struct {
	From      string `json:"from"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

// ★ ゴシップの本質：ランダム選択
func (n *Node) selectRandomPeer() string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if len(n.Peers) == 0 {
		return ""
	}

	index := rand.Intn(len(n.Peers))
	return n.Peers[index]
}

// ゴシップ送信実行
func (n *Node) SendGossip() (string, error) {
	target := n.selectRandomPeer()
	if target == "" {
		return "", fmt.Errorf("no peers available")
	}

	message := GossipMessage{
		From:      n.ID,
		Value:     n.GetValue(),
		Timestamp: time.Now().Unix(),
	}

	err := n.sendHTTPMessage(target, message)
	if err != nil {
		return target, fmt.Errorf("failed to send to %s: %v", target, err)
	}

	log.Printf("[%s] Sent gossip to %s: value='%s'", n.ID, target, message.Value)
	return target, nil
}

// HTTP経由でメッセージ送信
func (n *Node) sendHTTPMessage(targetAddr string, msg GossipMessage) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/gossip", targetAddr)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return nil
}

// ゴシップメッセージ受信処理
func (n *Node) HandleGossipMessage(msg GossipMessage) {
	log.Printf("[%s] Received gossip from %s: value='%s'",
		n.ID, msg.From, msg.Value)
	n.SetValue(msg.Value)
}
