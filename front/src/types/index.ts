export type NodeState = 'Red' | 'Green' | 'Blue';

export interface NodePosition {
  x: number;
  y: number;
}

export interface NodeData {
  id: string;
  state: NodeState;
  position: NodePosition;
  peers: string[];
  lastUpdated: number; // Round number when last updated
  value?: string; // Optional: Go実装のValueに相当
}

export interface EdgeData {
  source: string;
  target: string;
  active?: boolean; // メッセージ送信中かどうか
}

export interface GossipMessage {
  from: string;
  to: string;
  state: NodeState;
  round: number;
  timestamp: number;
}

export interface SimulationConfig {
  nodeCount: number;
  maxRounds: number;
  speed: number; // rounds per second (1-100)
}

export interface SimulationState {
  nodes: NodeData[];
  edges: EdgeData[];
  messages: GossipMessage[]; // 現在飛んでいるメッセージ
  currentRound: number;
  isRunning: boolean;
  config: SimulationConfig;
  stats: SimulationStats;
}

export interface SimulationStats {
  totalRounds: number;
  convergedRounds?: number; // 収束したRound数
  nodeStates: {
    [key in NodeState]: number; // 各状態のノード数
  };
  messagesPerRound: number[]; // Round毎のメッセージ数
  convergenceRate: number; // 収束率（%）
}