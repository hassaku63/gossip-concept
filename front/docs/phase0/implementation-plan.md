# Phase 0 実装計画

## 概要

Phase 0では、Gossipプロトコルシミュレータの基本機能を実装します。**バックエンドAPIサーバーは使用せず、すべての処理をブラウザ上で完結させます。** t-wadaスタイルのTDDアプローチに従い、Red→Green→Refactorサイクルで開発を進めます。

## 実装スコープ

### 機能要件
- シミュレーションの開始(init/start)と終了(reset)制御
- ノード数の設定（デフォルト10、最大300）
- ノード状態の3色表現（Red/Green/Blue）
- Round数の設定（デフォルト1000）
- シミュレーション速度の調整（1-100 Round/秒、デフォルト5）

### 技術要件
- React + TypeScript（クライアントサイドのみ）
- react-force-graphによるグラフ可視化
- Viteによるビルド環境
- VitestによるTDD環境
- **APIサーバー不要** - すべてブラウザ内で完結

## 実装順序とタスク

### Step 1: プロジェクトセットアップ
- [ ] Viteプロジェクトの初期化
- [ ] TypeScript設定
- [ ] ESLint/Prettier設定
- [ ] Vitest環境構築
- [ ] 基本的なフォルダ構造の作成

### Step 2: ドメインモデルの実装（TDD）
- [ ] Nodeモデルのテスト作成
- [ ] Nodeモデルの実装
- [ ] GossipProtocolモデルのテスト作成
- [ ] GossipProtocolモデルの実装
- [ ] Roundロジックのテスト作成
- [ ] Roundロジックの実装

### Step 3: UIコンポーネントの実装（TDD）
- [ ] ControlPanelコンポーネントのテスト作成
- [ ] ControlPanelコンポーネントの実装
- [ ] GraphVisualizerコンポーネントのテスト作成
- [ ] GraphVisualizerコンポーネントの実装
- [ ] Appコンポーネントの統合テスト作成
- [ ] Appコンポーネントの実装

### Step 4: シミュレーションロジックの実装（TDD）
- [ ] ノード間通信ロジックのテスト作成
- [ ] ノード間通信ロジックの実装
- [ ] 状態伝播アルゴリズムのテスト作成
- [ ] 状態伝播アルゴリズムの実装
- [ ] タイマー制御のテスト作成
- [ ] タイマー制御の実装

### Step 5: インテグレーション
- [ ] 全体の統合テスト作成
- [ ] パフォーマンステスト（300ノード対応確認）
- [ ] UIの最終調整

## ディレクトリ構造

```
front/
├── src/
│   ├── models/           # ドメインモデル
│   │   ├── Node.ts
│   │   ├── Node.test.ts
│   │   ├── GossipProtocol.ts
│   │   └── GossipProtocol.test.ts
│   ├── components/        # UIコンポーネント
│   │   ├── ControlPanel/
│   │   │   ├── ControlPanel.tsx
│   │   │   └── ControlPanel.test.tsx
│   │   ├── GraphVisualizer/
│   │   │   ├── GraphVisualizer.tsx
│   │   │   └── GraphVisualizer.test.tsx
│   │   └── App/
│   │       ├── App.tsx
│   │       └── App.test.tsx
│   ├── services/          # ビジネスロジック
│   │   ├── SimulationService.ts
│   │   └── SimulationService.test.ts
│   └── types/            # 型定義
│       └── index.ts
├── docs/
│   └── phase0/
│       ├── idea.md
│       ├── implementation-plan.md
│       └── test-strategy.md
└── tests/
    └── integration/      # 統合テスト
        └── simulation.test.ts
```

## タイムライン

### Week 1: 基盤構築
- Day 1-2: プロジェクトセットアップ
- Day 3-5: ドメインモデルの実装

### Week 2: UI実装
- Day 6-8: UIコンポーネントの実装
- Day 9-10: シミュレーションロジックの実装

### Week 3: 統合と最適化
- Day 11-12: インテグレーション
- Day 13-14: パフォーマンス最適化とリファクタリング
- Day 15: ドキュメント整備

## 成功基準

### 必須要件
- [ ] 全ユニットテストがパス（カバレッジ80%以上）
- [ ] 10ノードでの基本動作確認
- [ ] シミュレーション制御の正常動作（開始/停止/リセット）
- [ ] Gossipプロトコルの基本的な状態伝播

### 品質基準
- [ ] TypeScript型チェックエラーなし
- [ ] ESLintエラーなし
- [ ] 基本機能のデモが可能

## リスクと対策

### リスク1: TDD学習コスト
- **リスク**: TDDアプローチの習得に時間がかかる
- **対策**: 小さなユニットから始める、ペアプログラミングの活用

### リスク2: ライブラリの互換性
- **リスク**: react-force-graphの設定や使い方の理解不足
- **対策**: 基本的なサンプルから始める、ドキュメント精読

## 次のステップ

Phase 0完了後、以下の機能拡張を検討：
- Phase 1: 詳細な統計情報表示、大規模ノード対応
- Phase 2: 異なるGossipアルゴリズムの実装
- Phase 3: ネットワーク障害シミュレーション