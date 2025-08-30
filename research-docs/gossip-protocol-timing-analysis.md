# Riak Core ゴシッププロトコル: いつ何をするか

## 理論的背景

### Demers et al. (1987) による分類

Alan Demers らが1987年のPODC論文「Epidemic algorithms for replicated database maintenance」で定義した3つの基本アプローチ：

1. **Direct Mail**: 更新を即座に全サイトに送信
2. **Anti-entropy**: 定期的にランダムな他サイトと全データセットを比較・同期
3. **Rumor Mongering**: 新しい更新のみを確率的に伝播、「興味を失う」メカニズム付き

## Riak Core実装における動作タイミング

### 1. 定期ゴシップ (Anti-entropy相当)

**トリガー**: `riak_core_node_watcher`からの定期タイマー
**間隔**: `gossip_interval` (デフォルト60秒)
**動作フロー**:
```
Node Watcher Timer → gossip_ring cast → random_gossip(Ring) → 
ランダムな他ノード選択 → send_ring/2
```

**実装箇所**:
- `riak_core_node_watcher.erl:408-410` - タイマー設定
- `riak_core_gossip.erl:225-230` - `handle_cast(gossip_ring, State)`
- `riak_core_gossip.erl:85-91` - `random_gossip/1`

### 2. Ring更新時ゴシップ (Rumor Mongering + Anti-entropy のハイブリッド)

**トリガー**: Ring状態の変更
**条件**: `ring_trans/2`で`{new_ring, NewRing}`または`{reconciled_ring, NewRing}`が返された時

**実装箇所**:
- `riak_core_ring_manager.erl:423` - `random_recursive_gossip(NewRing)`
- `riak_core_ring_manager.erl:430` - `recursive_gossip(NewRing)`

**動作の違い**:
- `new_ring`: 新しいRing → `random_recursive_gossip` (ランダム起点)
- `reconciled_ring`: 調整されたRing → `recursive_gossip` (自ノード起点)

### 3. レート制限メカニズム

**実装**: トークンバケットアルゴリズム
**設定**: `gossip_limit` (デフォルト {45, 10000} = 45メッセージ/10秒)

**動作フロー**:
```
init/1 → トークン初期化 → schedule_next_reset() →
10秒後 → reset_tokens → トークン補充 → 次回タイマー設定
```

**実装箇所**:
- `riak_core_gossip.erl:131-134` - 初期化
- `riak_core_gossip.erl:193-206` - トークンチェック
- `riak_core_gossip.erl:257-261` - トークンリセット

## 理論と実装の比較分析

### Demers理論との対応関係

| Demers (1987) | Riak Core実装 | 対応する機能 |
|---------------|---------------|-------------|
| Anti-entropy | 定期ゴシップ | `random_gossip` (60秒間隔) |
| Rumor Mongering | Recursive Gossip | Ring更新時の`recursive_gossip` |
| Direct Mail | 未実装 | Ring分散時の`distribute_ring`のみ部分的 |

### ハイブリッドアプローチの採用

**理論**: Demers論文では「rumor-mongeringで高速伝播、anti-entropyで完全性保証」の組み合わせを推奨

**Riak Core実装**: 
1. **Rumor-like**: Ring更新時のrecursive gossip（バイナリツリー構造でファンアウト）
2. **Anti-entropy-like**: 定期的なrandom gossip（完全な状態同期）

### 実装における改良点

**空間的分散考慮**: Demers論文で提案された「ネットワーク topology を考慮した分散」
- Riak Core: `riak_core_util:build_tree/3`でバイナリツリー構築
- 対数的な拡散パターンでメッセージ数を削減

**Reconciliation プロセス**: 
- 理論: 単純な状態マージ
- 実装: 複雑なRing調整ロジック（クラスター名、メンバーステータス、バージョン管理）

## 具体的な動作条件

### 1. 定期ゴシップが実行される条件
- Node Watcherタイマーが発火 (60秒間隔)
- 対象ノードが存在 (`riak_core_ring:random_other_active_node/1`で取得)
- 単一ノードクラスターの場合はスキップ

### 2. Ring更新ゴシップが実行される条件
- `ring_trans/2`での状態変更
- `{new_ring, NewRing}`: 新規Ring作成時
- `{reconciled_ring, NewRing}`: 他ノードからの情報でRingを調整した時
- `{set_only, NewRing}`: ゴシップなしでのローカル更新のみ

### 3. メッセージ送信が拒否される条件
- ゴシップトークンが枯渇 (`gossip_tokens=0`)
- 送信先が自ノード (`send_ring(Node, Node) -> ok`)
- 無効なクラスター状態

### 4. Reconciliationが無視される条件
- 異なるクラスター名
- 送信元ノードのステータスが`invalid`または`down`
- Ring情報にタイント（破損）がある場合

## 実装の特徴

### Push型ゴシップのみ
**理論**: Demers論文ではpush/pull/push-pullを比較
**実装**: Push型のみ（自分のRing情報を他ノードに送信）
- Pull型アンチエントロピーは実装されていない

### 確率的停止メカニズムの簡略化
**理論**: Rumor mongeringでは「既に知っているノードに送信した回数」で停止判定
**実装**: Recursive gossipでは単純なツリー構造での1回限りの伝播

この設計により、理論的な基盤を保ちつつ、Riak Coreの分散ハッシュリング同期という特定用途に最適化された実装となっている。