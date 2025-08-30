#!/bin/bash
#
# ゴシップクラスターを起動するヘルパースクリプト
#

set -e

# デフォルト値
NODES=${1:-10}
BASE_PORT=${2:-18000}

# カラー出力用
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Starting Gossip Cluster ===${NC}"
echo "  Nodes: $NODES"
echo "  Base Port: $BASE_PORT"
echo ""

# ビルド確認
if [ ! -f ./gossip-concept ]; then
    echo -e "${YELLOW}Building gossip-concept...${NC}"
    go build .
fi

# 既存プロセスをクリーンアップ
echo -e "${YELLOW}Cleaning up existing processes...${NC}"
pkill -f "gossip-concept" 2>/dev/null || true
sleep 1

# クラスター起動
echo -e "${GREEN}Starting cluster...${NC}"
./gossip-concept --nodes=$NODES --base-port=$BASE_PORT &
PID=$!
echo "Process PID: $PID"

# 起動待ち
echo -e "${YELLOW}Waiting for startup...${NC}"
sleep 2

# ヘルスチェック
echo -e "${GREEN}Health check:${NC}"
MAX_PORT=$((BASE_PORT + NODES - 1))
healthy=0
for port in $(seq $BASE_PORT $MAX_PORT); do
    node_id=$((port - BASE_PORT))
    if curl -s "localhost:${port}/status" > /dev/null 2>&1; then
        VALUE=$(curl -s "localhost:${port}/status" | jq -r .value)
        echo -e "  node-$node_id (port ${port}): ${GREEN}✓${NC} value='${VALUE}'"
        healthy=$((healthy + 1))
    else
        echo -e "  node-$node_id (port ${port}): ${RED}✗${NC}"
    fi
done

echo ""
if [ $healthy -eq $NODES ]; then
    echo -e "${GREEN}Cluster ready! All $NODES nodes are healthy.${NC}"
else
    echo -e "${YELLOW}Warning: Only $healthy/$NODES nodes are healthy.${NC}"
fi

echo ""
echo "Commands:"
echo "  Check status:     curl localhost:$BASE_PORT/status | jq"
echo "  Trigger gossip:   curl -X POST localhost:$BASE_PORT/trigger"
echo "  Set value:        curl -X POST 'localhost:$BASE_PORT/set?value=hello'"
echo "  Stop cluster:     pkill -f gossip-concept"
echo ""

# バックグラウンドプロセスをフォアグラウンドに復帰
echo -e "${GREEN}Bringing cluster to foreground...${NC}"
echo "Press Ctrl+C to stop the cluster"
echo ""
wait $PID
