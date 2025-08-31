import { describe, it, expect } from 'vitest';
import { Node } from './Node';

describe('Node', () => {
  describe('初期化', () => {
    it('IDと初期状態を設定できる', () => {
      const node = new Node('node-1');
      expect(node.id).toBe('node-1');
      expect(node.state).toBe('Red');
    });
    
    it('カスタム初期状態を設定できる', () => {
      const node = new Node('node-1', 'Green');
      expect(node.state).toBe('Green');
    });
    
    it('位置情報を持つ', () => {
      const node = new Node('node-1', 'Red', { x: 10, y: 20 });
      expect(node.position).toEqual({ x: 10, y: 20 });
    });
    
    it('デフォルトの位置は原点である', () => {
      const node = new Node('node-1');
      expect(node.position).toEqual({ x: 0, y: 0 });
    });
    
    it('空のピアリストで初期化される', () => {
      const node = new Node('node-1');
      expect(node.peers).toEqual([]);
    });
    
    it('初期のlastUpdatedは0である', () => {
      const node = new Node('node-1');
      expect(node.lastUpdated).toBe(0);
    });
  });
  
  describe('状態変更', () => {
    it('有効な状態に変更できる', () => {
      const node = new Node('node-1');
      node.setState('Green');
      expect(node.state).toBe('Green');
      
      node.setState('Blue');
      expect(node.state).toBe('Blue');
      
      node.setState('Red');
      expect(node.state).toBe('Red');
    });
    
    it('状態変更時にlastUpdatedが更新される', () => {
      const node = new Node('node-1');
      const initialLastUpdated = node.lastUpdated;
      
      node.setState('Green', 5);
      expect(node.lastUpdated).toBe(5);
      expect(node.lastUpdated).not.toBe(initialLastUpdated);
    });
    
    it('無効な状態への変更は拒否される', () => {
      const node = new Node('node-1');
      expect(() => {
        // @ts-expect-error Testing invalid state
        node.setState('Yellow');
      }).toThrow('Invalid state: Yellow');
    });
  });
  
  describe('ピア管理', () => {
    it('ピアを追加できる', () => {
      const node = new Node('node-1');
      node.addPeer('node-2');
      expect(node.peers).toContain('node-2');
    });
    
    it('重複したピアは追加されない', () => {
      const node = new Node('node-1');
      node.addPeer('node-2');
      node.addPeer('node-2');
      expect(node.peers.filter(p => p === 'node-2')).toHaveLength(1);
    });
    
    it('複数のピアを設定できる', () => {
      const node = new Node('node-1');
      node.setPeers(['node-2', 'node-3', 'node-4']);
      expect(node.peers).toEqual(['node-2', 'node-3', 'node-4']);
    });
  });
  
  describe('位置情報', () => {
    it('位置を更新できる', () => {
      const node = new Node('node-1');
      node.setPosition({ x: 100, y: 200 });
      expect(node.position).toEqual({ x: 100, y: 200 });
    });
    
    it('位置情報は参照ではなくコピーが返される', () => {
      const node = new Node('node-1', 'Red', { x: 10, y: 20 });
      const position = node.position;
      position.x = 999;
      
      // 元のノードの位置は変更されない
      expect(node.position).toEqual({ x: 10, y: 20 });
    });
  });
});