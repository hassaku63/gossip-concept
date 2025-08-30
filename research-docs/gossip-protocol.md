# Riak Core Gossip Protocol

## 概要

Riak CoreのゴシッププロトコルはクラスターのRing状態情報を全ノード間で同期するための仕組みです。各ノードが定期的に他のノードと状態情報を交換し、最終的にクラスター全体で一貫したRing状態に収束させます。

## 実装箇所

### メインモジュール

- **`riak_core_gossip.erl`** - ゴシッププロトコルのコアロジック
  - gen_serverとして実装
  - Ring情報の送受信とreconciliation（調整）を担当

### 関連モジュール

- **`riak_core_ring_manager.erl`** - Ring状態の管理とゴシップトリガー
  - Ring更新時に自動的にゴシップを開始（recursive_gossip）
- **`riak_core_node_watcher.erl`** - 定期的なゴシップのスケジューリング
  - `gossip_interval`設定に基づいて定期ゴシップを実行
- **`riak_core_sup.erl:78` - プロセス階層での配置
  - `riak_core_gossip`がワーカープロセスとして起動

## 他機能との関係

### 直接的な依存関係

1. **Ring Manager** - 双方向依存
   - Ring更新時にgossipが自動実行される
   - Gossip受信時にring_transでRing Managerのreconcile関数が呼び出される

2. **Node Watcher** - ゴシップのトリガー役
   - 定期的（デフォルト60秒）にgossip_ringメッセージを送信
   - ノードのup/down監視と連携

3. **Ring Module** - Ring状態の操作
   - Ring比較、マージ、バージョン管理を提供
   - `reconcile/2`でRing状態の統合を実行

4. **Claimant** - クラスター状態管理
   - メンバーシップ変更時の調整処理
   - ゴシップで受信した状態変更への対応

### 間接的な影響

- **VNode** - Ring変更の影響を受ける
- **Broadcast** - 類似の分散通信メカニズム（独立実装）
- **Statistics** - ゴシップ統計情報の収集

## 動作メカニズム

### 1. 初期化
- `riak_core_gossip`プロセスが起動時にトークンバケット（デフォルト45メッセージ/10秒）を初期化
- 定期リセットタイマーを設定

### 2. 定期ゴシップ
```
Node Watcher (60秒間隔) → gossip_ring cast → random_gossip(Ring)
```
- ランダムな他ノードを選択してRing情報を送信

### 3. Ring更新時のゴシップ
```
Ring Manager → recursive_gossip(Ring)
```
- バイナリツリー構造でファンアウト
- 対数的な拡散によりクラスター全体に効率的に伝播

### 4. Reconciliation（調整）プロセス
```
{reconcile_ring, RingIn} → reconcile/2 → Ring Manager
```

受信したRing情報の処理フロー：
1. Ring情報をアップグレード（下位互換性対応）
2. クラスター名の一致確認
3. メンバーステータスの検証
4. Ring状態のマージ実行
5. 変更があった場合は新Ring状態をRing Managerに反映

### 5. エラー処理と制御
- **レート制限**: トークンバケットによるDoS防止
- **クラスター分離**: 異なるクラスター名のゴシップを無視
- **無効ノード**: downまたはinvalidステータスのノードからのゴシップを無視
- **Rejoin処理**: downノードに対するクラスター復帰支援

## 設定項目

- `gossip_interval` (60000ms) - 定期ゴシップの間隔
- `gossip_limit` ({45, 10000}) - レート制限（メッセージ数/期間）

## 実装の特徴

### 直交性のない設計
ゴシッププロトコルはRing Managerと強く結合しており、完全に直交した設計ではありません：

- Ring更新時の自動ゴシップ
- ゴシップ受信時のRing Manager呼び出し
- 共有状態（Ring）への相互依存

### 効率性の考慮
- **Random Gossip**: 単純なランダム選択によるbaseline伝播
- **Recursive Gossip**: バイナリツリー構造による効率的なファンアウト
- **Rate Limiting**: DoS攻撃やメッセージ嵐の防止

この設計により、クラスターサイズに対して対数的なメッセージ複雑度でRing状態の同期を実現しています。