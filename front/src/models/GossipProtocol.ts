import { Node } from './Node';
import { GossipMessage } from '../types';
import { IndexSelector, RandomIndexSelector } from '../services/IndexSelector';

export class GossipProtocol {
  private _nodes: Node[] = [];
  private _currentRound: number = 0;
  private indexSelector: IndexSelector;
  
  constructor(indexSelector?: IndexSelector) {
    this.indexSelector = indexSelector || new RandomIndexSelector();
  }
  
  get nodes(): Node[] {
    return [...this._nodes];
  }
  
  get currentRound(): number {
    return this._currentRound;
  }
  
  initialize(nodeCount: number): void {
    this._nodes = [];
    this._currentRound = 0;
    
    // ノードを作成
    for (let i = 0; i < nodeCount; i++) {
      const node = new Node(`node-${i}`);
      this._nodes.push(node);
    }
    
    // 各ノードに他の全ノードをピアとして設定（完全グラフ）
    this._nodes.forEach(node => {
      const peers = this._nodes
        .filter(peer => peer.id !== node.id)
        .map(peer => peer.id);
      node.setPeers(peers);
    });
  }
  
  selectRandomPeer(node: Node): string | null {
    const index = this.indexSelector.selectIndex(node.peers.length);
    return index !== null ? node.peers[index] : null;
  }
  
  executeRound(): GossipMessage[] {
    const messages: GossipMessage[] = [];
    const activeNodes = this.getActiveNodes();
    
    // 各アクティブノード（Red以外）からランダムにゴシップ
    for (const node of activeNodes) {
      const targetId = this.selectRandomPeer(node);
      if (targetId) {
        const message: GossipMessage = {
          from: node.id,
          to: targetId,
          state: node.state,
          round: this._currentRound + 1,
          timestamp: Date.now()
        };
        messages.push(message);
      }
    }
    
    // メッセージを受信処理
    for (const message of messages) {
      const targetNode = this.findNode(message.to);
      if (targetNode) {
        this.receiveMessage(targetNode, message);
      }
    }
    
    this._currentRound++;
    return messages;
  }
  
  private receiveMessage(targetNode: Node, message: GossipMessage): void {
    // 状態の更新（感染モデル）
    if (targetNode.state !== message.state) {
      targetNode.setState(message.state, message.round);
    }
  }
  
  isConverged(): boolean {
    if (this._nodes.length === 0) return true;
    
    const firstNodeState = this._nodes[0].state;
    return this._nodes.every(node => node.state === firstNodeState);
  }
  
  reset(): void {
    this._currentRound = 0;
    this._nodes.forEach(node => {
      node.setState('Red', 0);
    });
  }
  
  findNode(id: string): Node | null {
    return this._nodes.find(node => node.id === id) || null;
  }
  
  getActiveNodes(): Node[] {
    return this._nodes.filter(node => node.state !== 'Red');
  }
}