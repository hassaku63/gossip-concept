# Phase 0 実装実験ガイド

## 概要

Phase 0の実装が完了し、**1つのコマンドで10ノードのゴシップクラスター**を起動できるようになりました。さらに**管理サービス**により、クラスター全体の状態を簡単に監視できます。本ガイドでは、ゴシッププロトコルの本質を体験するための実験手順を説明します。

## 前提条件

- Go 1.24+ がインストール済み
- `gossip_concept/` ディレクトリで作業
- `curl` と `jq` が使用可能

## クイックスタート

### 1. ビルドと起動（設定ファイル不要）
```bash
cd gossip_concept/
go build .
./gossip-concept

# または起動スクリプトを使用
./start-cluster.sh
```

**期待される出力**:
```
2025/08/30 21:57:34 Starting 10 nodes...
2025/08/30 21:57:34 All 10 nodes started successfully  
2025/08/30 21:57:34 Use curl to interact with nodes:
2025/08/30 21:57:34   Status:  curl localhost:18000/status
2025/08/30 21:57:34   Gossip:  curl -X POST localhost:18000/trigger
2025/08/30 21:57:34   Set:     curl -X POST 'localhost:18000/set?value=hello'
[各ノードのHTTPサーバー起動ログ...]
```

### カスタム設定
```bash
# 5ノードで起動
./gossip-concept --nodes=5

# 異なるポート範囲で起動（19000-19009）
./gossip-concept --base-port=19000

# 管理ポートも変更
./gossip-concept --base-port=19000 --admin-port=18999

# ヘルプ表示
./gossip-concept --help
```

### 管理サービスの利用
```bash
# クラスター設定情報を取得
curl localhost:17999/cluster | jq

# 全ノードの情報を取得
curl localhost:17999/nodes | jq

# ヘルスチェック実行
curl localhost:17999/health | jq

# 管理サービス自体の情報
curl localhost:17999/ | jq
```

### 2. 全ノード稼働確認
```bash
for port in {18000..18009}; do 
  echo "Node port $port:"; 
  curl -s "localhost:$port/status" | jq -r '.id + ": " + .value'
done
```

**期待される結果**: 全ノードが`initial-state`で応答

**カスタムノード数での確認**:
```bash
# 5ノードの場合（--nodes=5で起動時）
for port in {18000..18004}; do 
  curl -s "localhost:$port/status" | jq -r '.id + ": " + .value'
done
```

## 実験1: ランダム性の観察

### 目的
ゴシップの本質である**ランダム選択**を観察し、9つの選択肢への分散を確認します。

### 手順

#### 1.1 基本的なランダム選択テスト
```bash
echo "Testing 20 random gossip selections:"
for i in {1..20}; do 
  echo -n "Round $i: "
  curl -s -X POST localhost:18000/trigger | jq -r '.target'
  sleep 0.1
done

# またはスクリプトを使用
./observe-randomness.sh 20
```

#### 1.2 統計的分散の分析
```bash
# スクリプトを使用（管理APIから設定を自動取得）
./observe-randomness.sh 100

# または手動で実行
echo "Statistical distribution test (100 rounds):"
declare -A targets
for i in {1..100}; do
  target=$(curl -s -X POST localhost:18000/trigger | jq -r '.target' 2>/dev/null)
  if [ "$target" != "null" ] && [ ! -z "$target" ]; then
    targets[$target]=$((${targets[$target]:-0} + 1))
  fi
  echo -n "."
  sleep 0.05
done
```

#### 1.3 期待される結果
- **理想的分散**: 各ノードに約11回（100÷9）
- **許容範囲**: 6-16回程度
- **統計的偏り**: カイ二乗検定で p > 0.05

### 1.4 観察ポイント
- [ ] 送信先がランダムに変わること
- [ ] すべてのノード（18001-18009）が選ばれること  
- [ ] 極端な偏りがないこと
- [ ] node-0（18000）は選ばれないこと（自分自身）

## 実験2: 状態伝搬と収束性

### 目的
**確率的伝搬**により、最終的にすべてのノードが同じ状態に収束することを観察します。

### 手順

#### 2.1 初期状態の確認
```bash
echo "Initial state of all nodes:"
for port in {18000..18009}; do 
  value=$(curl -s "localhost:$port/status" | jq -r '.value')
  echo "node-$((port-18000)): $value"
done
```

#### 2.2 値の変更と伝搬開始
```bash
# タイムスタンプ付きの一意な値を設定
test_value="experiment-$(date +%s)"
echo "Setting test value: $test_value"
curl -s -X POST "localhost:18000/set?value=$test_value"
```

#### 2.3 段階的伝搬の観察
```bash
# スクリプトを使用（自動で収束を追跡）
./observe-convergence.sh

# または手動で実行
echo "Triggering gossip rounds:"
for round in {1..20}; do
  # ゴシップ実行
  target=$(curl -s -X POST localhost:18000/trigger | jq -r '.target')
  echo "Round $round: sent to $target"
  sleep 0.2
  
  # 5回おきに伝搬状況を確認
  if [ $((round % 5)) -eq 0 ]; then
    echo "  Progress check:"
    converged=0
    for port in {18000..18009}; do
      value=$(curl -s "localhost:$port/status" | jq -r '.value')
      if [ "$value" = "$test_value" ]; then
        converged=$((converged + 1))
      fi
    done
    echo "  → $converged/10 nodes converged"
  fi
done
```

#### 2.4 最終収束状態の確認
```bash
echo ""
echo "Final convergence state:"
for port in {18000..18009}; do
  value=$(curl -s "localhost:$port/status" | jq -r '.value')
  if [ "$value" = "$test_value" ]; then
    echo "✅ node-$((port-18000)): $value"
  else
    echo "❌ node-$((port-18000)): $value (not converged)"
  fi
done
```

#### 2.5 期待される結果
- **初期段階**: 少数のノードのみ更新
- **中間段階**: 指数的に拡散（2→4→8ノード）
- **最終段階**: 20-30ラウンドで全ノード収束
- **理論値**: 10ノードなら平均25ラウンド程度

## 実験3: 二次伝搬（チェーン伝搬）

### 目的
更新されたノードからさらに他のノードへの伝搬（二次伝搬）を観察します。

### 手順

#### 3.1 値の設定と一次伝搬
```bash
test_value="chain-$(date +%s)"
curl -s -X POST "localhost:18000/set?value=$test_value"

# 最初の伝搬先を特定
first_target=$(curl -s -X POST localhost:18000/trigger | jq -r '.target')
first_node=$(echo $first_target | sed 's/localhost:1800//')
echo "First propagation: node-0 → node-$first_node"
sleep 0.5
```

#### 3.2 二次伝搬の実行
```bash
# 更新されたノードからさらにゴシップ
echo "Second propagation from node-$first_node:"
port="1800$first_node"
second_target=$(curl -s -X POST "localhost:$port/trigger" | jq -r '.target')
second_node=$(echo $second_target | sed 's/localhost:1800//')
echo "Second propagation: node-$first_node → node-$second_node"
```

#### 3.3 チェーン伝搬の確認
```bash
echo "Propagation chain results:"
for port in {18000..18009}; do
  value=$(curl -s "localhost:$port/status" | jq -r '.value')
  node_id=$((port-18000))
  if [ "$value" = "$test_value" ]; then
    echo "✅ node-$node_id: updated"
  else
    echo "   node-$node_id: not updated"
  fi
done
```

## 実験4: 障害耐性の基本確認

### 目的
一部のノードが利用不可でも、利用可能なノードには正常に伝搬することを確認します。

### 手順

#### 4.1 疑似的な障害ノード
```bash
# Note: この実験では実際にノードを停止させず、
# ランダム選択で到達不可ノードが選ばれた場合の動作を観察
echo "Testing resilience to unreachable nodes:"
echo "(Some gossip attempts may fail - this is expected behavior)"

test_value="resilience-$(date +%s)"
curl -s -X POST "localhost:18000/set?value=$test_value"

# 複数回試行（一部は失敗する可能性）
success_count=0
for i in {1..10}; do
  result=$(curl -s -X POST localhost:18000/trigger)
  if echo "$result" | jq -e '.status == "sent"' > /dev/null; then
    target=$(echo "$result" | jq -r '.target')
    echo "Round $i: ✅ sent to $target"
    success_count=$((success_count + 1))
  else
    echo "Round $i: ❌ failed (target unreachable)"
  fi
  sleep 0.2
done

echo "Success rate: $success_count/10"
```

## パフォーマンス測定

### レスポンス時間の測定
```bash
echo "Measuring API response times:"
for i in {1..5}; do
  echo -n "Round $i: "
  time curl -s -X POST localhost:18000/trigger > /dev/null
done
```

### 同時接続テスト
```bash
echo "Testing concurrent gossip operations:"
for i in {0..4}; do
  port="1800$i"
  curl -s -X POST "localhost:$port/trigger" > /dev/null &
done
wait
echo "All concurrent operations completed"
```

## トラブルシューティング

### 一般的な問題と解決策

#### ❌ ポート使用中エラー
```bash
# 問題: listen tcp 127.0.0.1:18000: bind: address already in use
# 解決策:
pkill -f gossip-concept
sleep 2
./gossip-concept
```

#### ❌ JSON解析エラー
```bash
# 問題: parse error: Invalid literal
# 原因: APIエラーレスポンス
# 解決策: HTTPステータスを確認
curl -i localhost:18000/status
```

#### ❌ ノード間通信エラー
```bash
# 問題: dial tcp 127.0.0.1:18008: connect: connection refused  
# 原因: 対象ノードが未起動
# 確認方法:
netstat -tulpn | grep 180
```

### ログの確認方法
```bash
# バックグラウンド実行時のログ確認
./gossip-concept > gossip.log 2>&1 &
tail -f gossip.log

# 特定ノードのログをフィルタ
tail -f gossip.log | grep '\[node-0\]'
```

## 実験結果の解釈

### 成功基準
- [ ] **ランダム性**: 100回中、各ノードに8-14回の範囲で分散
- [ ] **収束性**: 30回以内のゴシップで全ノード更新
- [ ] **チェーン伝搬**: 二次伝搬が正常に動作
- [ ] **レスポンス**: 全APIが100ms以内で応答

### 学習ポイント
1. **確率的性質**: 完璧でない通信でも最終的に収束
2. **スケーラビリティ**: 各ノードは全体を知らなくても動作
3. **自己組織化**: 中央制御なしで情報が拡散
4. **障害耐性**: 一部の失敗があっても全体は機能

## 次のステップ

Phase 0の実験が成功したら：
1. **Phase 0.5**: 自動タイマー機能の追加
2. **Phase 1**: レート制限とkey-value状態管理
3. **独自実験**: アルゴリズムの改良や性能測定

この実験により、Riak Coreの`riak_core_gossip.erl`で実装されているゴシッププロトコルの本質を実体験で理解できます。