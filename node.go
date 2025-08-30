package main

import (
	"log"
	"sync"
	"time"
)

type Node struct {
	mu       sync.RWMutex
	ID       string
	Address  string
	Value    string
	Peers    []string
	LastSeen int64
}

// NewNode関数は不要になったため削除

// thread-safeな値の取得
func (n *Node) GetValue() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Value
}

// thread-safeな値の更新
func (n *Node) SetValue(value string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.Value != value {
		log.Printf("[%s] Value updated: '%s' -> '%s'", n.ID, n.Value, value)
		n.Value = value
		n.LastSeen = time.Now().Unix()
	}
}

// ステータス情報取得
func (n *Node) GetStatus() map[string]interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return map[string]interface{}{
		"id":        n.ID,
		"value":     n.Value,
		"peers":     n.Peers,
		"last_seen": n.LastSeen,
	}
}
