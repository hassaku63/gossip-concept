# Gossip Protocol Go Implementation

Riak CoreのErlang実装をリファレンスとして、ゴシッププロトコルのコア機能を段階的にGo言語で実装するプロジェクト。

## プロジェクト構成

```
gossip_concept/
├── README.md                    # このファイル
├── docs/                        # 設計・実装ドキュメント
│   └── phase0/                  # Phase 0: Proof of Concept
├── phase0/                      # Phase 0 実装（予定）
├── phase0.5/                    # Phase 0.5 実装（予定）
└── phase1/                      # Phase 1 実装（予定）
```

## 実装フェーズ

### Phase 0: Proof of Concept (最小限の状態伝搬)
**目標**: ゴシップの本質（ランダム選択による確率的伝搬）を体験  
**期間**: 半日  
**規模**: 約180行、10ノード構成  

**ドキュメント**: [`docs/phase0/`](./docs/phase0/)
- 詳細設計書
- 実装手順書  
- テスト戦略
- 10ノード用観察ツール

### Phase 0.5: 自動化とスケーラビリティ (予定)
**目標**: 定期実行と任意ノード数対応  
**期間**: 半日

### Phase 1: Basic Gossip Infrastructure (予定)  
**目標**: 実用的なゴシップの基礎を作る  
**期間**: 1-2日

## 学習目標

このプロジェクトを通じて体験できること：

1. **ゴシップの本質**: ランダム選択→確率的伝搬→最終収束
2. **Riak Core理解**: Erlang実装との対応関係
3. **分散システム**: スケーラビリティと障害耐性の原理
4. **実装技術**: Go言語での分散システム実装

## 理論的背景

### Demers et al. (1987)による分類
- **Direct Mail**: 更新を即座に全サイトに送信
- **Anti-entropy**: 定期的なランダムサイトとの全データ同期  
- **Rumor Mongering**: 新更新のみを確率的伝播

### Riak Core実装との対応
| 理論 | Riak Core実装 | Go実装での学習ポイント |
|------|---------------|----------------------|
| Random Gossip | `random_gossip/1` | ランダム選択アルゴリズム |
| Recursive Gossip | `recursive_gossip/1` | バイナリツリー構造での拡散 |
| Rate Limiting | トークンバケット | DoS防止メカニズム |

## 始め方

### Phase 0から開始
```bash
cd docs/phase0
cat README.md
```

Phase 0の詳細な実装計画、設計書、テスト戦略が含まれています。

### 実装の進め方
1. [`docs/phase0/implementation-steps.md`](./docs/phase0/implementation-steps.md)に従って実装
2. [`docs/phase0/test-strategy.md`](./docs/phase0/test-strategy.md)でテストを実行
3. 観察ツールでゴシップの動作を体験

## 参考文献

- **Riak Core**: [`riak_core_gossip.erl`](../../src/riak_core_gossip.erl)
- **理論**: Demers et al. (1987) "Epidemic algorithms for replicated database maintenance"
- **研究**: [`research-docs/`](../../research-docs/)の詳細な分析

## 貢献

このプロジェクトは学習目的のため、以下の原則に従っています：

- **段階的実装**: 複雑さを段階的に追加
- **観察可能性**: 動作が目に見えるよう設計  
- **理論との対応**: 学術的背景を重視
- **実用性**: 実際のシステムで使える品質

Phase毎に完全なドキュメントと動作する実装を提供し、分散システムの理解を深めることを目指しています。