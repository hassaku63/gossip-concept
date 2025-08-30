#!/usr/bin/env python3
"""
10ノード用の設定ファイルを生成するスクリプト
"""

import json

def generate_config(num_nodes=10, base_port=18000):
    """10ノード用のフルメッシュ設定を生成"""
    
    nodes = {}
    
    for i in range(num_nodes):
        node_id = f"node-{i}"
        node_port = base_port + i
        node_address = f"localhost:{node_port}"
        
        # 自分以外のすべてのノードをpeersに追加
        peers = []
        for j in range(num_nodes):
            if i != j:
                peer_port = base_port + j
                peers.append(f"localhost:{peer_port}")
        
        nodes[node_id] = {
            "address": node_address,
            "peers": peers
        }
    
    config = {
        "nodes": nodes,
        "initial_value": "initial-state"
    }
    
    return config

def main():
    # 10ノード用の設定を生成
    config = generate_config(10, 18000)
    
    # config-10nodes.jsonとして保存
    with open("config-10nodes.json", "w") as f:
        json.dump(config, f, indent=2)
    
    print("Generated config-10nodes.json")
    
    # 統計情報を表示
    print(f"Total nodes: {len(config['nodes'])}")
    for node_id, node_config in config['nodes'].items():
        print(f"  {node_id}: {node_config['address']} -> {len(node_config['peers'])} peers")

if __name__ == "__main__":
    main()