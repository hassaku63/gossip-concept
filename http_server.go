package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func startHTTPServer(node *Node) {
	mux := http.NewServeMux()

	// ゴシップメッセージ受信エンドポイント
	mux.HandleFunc("/gossip", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var msg GossipMessage
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// ゴシップ処理
		node.HandleGossipMessage(msg)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "received"})
	})

	// 手動ゴシップトリガー
	mux.HandleFunc("/trigger", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		target, err := node.SendGossip()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"status": "sent",
			"target": target,
		}
		json.NewEncoder(w).Encode(response)
	})

	// ノード状態確認
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := node.GetStatus()
		json.NewEncoder(w).Encode(status)
	})

	// 値設定エンドポイント（テスト用）
	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		value := r.URL.Query().Get("value")
		if value == "" {
			http.Error(w, "value parameter required", http.StatusBadRequest)
			return
		}

		node.SetValue(value)
		json.NewEncoder(w).Encode(map[string]string{"status": "updated", "value": value})
	})

	// サーバー起動
	log.Printf("[%s] HTTP server starting on %s", node.ID, node.Address)
	log.Fatal(http.ListenAndServe(node.Address, mux))
}
