import { describe, it, expect, beforeEach } from 'vitest';
import { GossipProtocol } from './GossipProtocol';
import { Node } from './Node';
import { MockIndexSelector } from '../services/IndexSelector';

describe('GossipProtocol', () => {
  let protocol: GossipProtocol;
  let mockIndexSelector: MockIndexSelector;
  
  beforeEach(() => {
    mockIndexSelector = new MockIndexSelector();
    protocol = new GossipProtocol(mockIndexSelector);
  });
  
  describe('初期化', () => {
    it('指定された数のノードを生成する', () => {
      protocol.initialize(5);
      expect(protocol.nodes).toHaveLength(5);
    });
    
    it('ノードのIDが適切に設定される', () => {
      protocol.initialize(3);
      const nodes = protocol.nodes;
      expect(nodes[0].id).toBe('node-0');
      expect(nodes[1].id).toBe('node-1');
      expect(nodes[2].id).toBe('node-2');
    });
    
    it('すべてのノードが初期状態Redである', () => {
      protocol.initialize(5);
      protocol.nodes.forEach(node => {
        expect(node.state).toBe('Red');
      });
    });
    
    it('各ノードが他の全ノードをピアとして持つ', () => {
      protocol.initialize(3);
      const node0 = protocol.findNode('node-0')!;
      expect(node0.peers).toEqual(['node-1', 'node-2']);
      
      const node1 = protocol.findNode('node-1')!;
      expect(node1.peers).toEqual(['node-0', 'node-2']);
    });
    
    it('初期Round数が0である', () => {
      protocol.initialize(3);
      expect(protocol.currentRound).toBe(0);
    });
  });
  
  describe('ランダムピア選択', () => {
    beforeEach(() => {
      protocol.initialize(3);
    });
    
    it('ピアがない場合はnullを返す', () => {
      const node = new Node('isolated');
      const peer = protocol.selectRandomPeer(node);
      expect(peer).toBeNull();
    });
    
    it('1つのピアがある場合はそのピアを返す', () => {
      const node = new Node('node-1');
      node.setPeers(['node-2']);
      const peer = protocol.selectRandomPeer(node);
      expect(peer).toBe('node-2');
    });
    
    it('複数のピアからランダムに選択する', () => {
      const node = protocol.findNode('node-0')!;
      
      // インデックス0を選択 → 最初のピア
      mockIndexSelector.setIndices([0]);
      expect(protocol.selectRandomPeer(node)).toBe('node-1');
      
      // インデックス1を選択 → 最後のピア
      mockIndexSelector.setIndices([1]);
      expect(protocol.selectRandomPeer(node)).toBe('node-2');
    });
  });
  
  describe('Round実行', () => {
    beforeEach(() => {
      protocol.initialize(3);
      // node-0をGreenに変更して感染源とする
      protocol.findNode('node-0')!.setState('Green', 0);
    });
    
    it('Round実行後にcurrentRoundが増加する', () => {
      const initialRound = protocol.currentRound;
      protocol.executeRound();
      expect(protocol.currentRound).toBe(initialRound + 1);
    });
    
    it('感染ノードからメッセージが送信される', () => {
      // node-0がGreenなので、メッセージを送信する
      mockIndexSelector.setIndices([0]); // インデックス0 → node-1を選択
      
      const messages = protocol.executeRound();
      expect(messages).toHaveLength(1);
      expect(messages[0].from).toBe('node-0');
      expect(messages[0].to).toBe('node-1');
      expect(messages[0].state).toBe('Green');
    });
    
    it('メッセージ受信により状態が伝播する', () => {
      mockIndexSelector.setIndices([0]); // インデックス0 → node-0 → node-1
      
      const node1 = protocol.findNode('node-1')!;
      expect(node1.state).toBe('Red'); // 初期状態
      
      protocol.executeRound();
      
      expect(node1.state).toBe('Green'); // 感染した
      expect(node1.lastUpdated).toBe(1); // Round 1で更新
    });
    
    it('Red状態のノードは何もしない', () => {
      // すべてのノードをRedに戻す
      protocol.findNode('node-0')!.setState('Red', 0);
      
      const messages = protocol.executeRound();
      expect(messages).toHaveLength(0);
    });
  });
  
  describe('収束判定', () => {
    beforeEach(() => {
      protocol.initialize(3);
    });
    
    it('すべてのノードが同じ状態の場合は収束している', () => {
      // すべてのノードをGreenに設定
      protocol.nodes.forEach(node => node.setState('Green', 1));
      
      expect(protocol.isConverged()).toBe(true);
    });
    
    it('異なる状態のノードがある場合は収束していない', () => {
      protocol.findNode('node-0')!.setState('Green', 1);
      // node-1, node-2はRed
      
      expect(protocol.isConverged()).toBe(false);
    });
    
    it('すべてのノードがRedの場合も収束している', () => {
      // 初期状態（すべてRed）
      expect(protocol.isConverged()).toBe(true);
    });
  });
  
  describe('リセット', () => {
    beforeEach(() => {
      protocol.initialize(3);
      protocol.findNode('node-0')!.setState('Green', 1);
      protocol.executeRound();
    });
    
    it('すべてのノードがRed状態に戻る', () => {
      protocol.reset();
      
      protocol.nodes.forEach(node => {
        expect(node.state).toBe('Red');
      });
    });
    
    it('currentRoundが0にリセットされる', () => {
      expect(protocol.currentRound).toBeGreaterThan(0);
      
      protocol.reset();
      
      expect(protocol.currentRound).toBe(0);
    });
    
    it('すべてのノードのlastUpdatedが0にリセットされる', () => {
      protocol.reset();
      
      protocol.nodes.forEach(node => {
        expect(node.lastUpdated).toBe(0);
      });
    });
  });
  
  describe('ユーティリティ', () => {
    beforeEach(() => {
      protocol.initialize(3);
    });
    
    it('IDによるノード検索ができる', () => {
      const node = protocol.findNode('node-1');
      expect(node).not.toBeNull();
      expect(node!.id).toBe('node-1');
    });
    
    it('存在しないIDの場合はnullを返す', () => {
      const node = protocol.findNode('non-existent');
      expect(node).toBeNull();
    });
    
    it('アクティブなノード（Red以外）を取得できる', () => {
      protocol.findNode('node-0')!.setState('Green', 1);
      protocol.findNode('node-2')!.setState('Blue', 1);
      
      const activeNodes = protocol.getActiveNodes();
      expect(activeNodes).toHaveLength(2);
      expect(activeNodes.map(n => n.id)).toEqual(['node-0', 'node-2']);
    });
  });
});