# idea

Gossip プロトコルのシミュレーションを行う。

1画面で構成され、以下の要件を持つ。

- シミュレーションの開始 (init/start) と終了 (reset) を制御できる
- ノード数はシミュレーションごとに指定可能。デフォルトは10ノードで、最大300ノードまで対応
- ノードのステートは Red/Green/Blue の3色で表現。内部的には string で管理
- シミュレーションの開始時に、最大 Round 数を指定可能 (デフォルトは1000)
- Rounds を進める速度は指定可能
    - デフォルトは 5 Round / 秒
    - min = 1 Round / 秒
    - max = 100 Round / 秒

## dependencies

以下のライブラリの使用を想定している。ただし、今回の用途に対してより適切な選定候補がある場合、または更新頻度の問題からライブラリの利用にリスクがある場合は、代替案を検討すること。

- [React](https://www.npmjs.com/package/react)
- [react-force-graph](https://github.com/vasturiano/react-force-graph)
