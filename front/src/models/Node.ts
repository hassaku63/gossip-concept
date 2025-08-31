import { NodeState, NodePosition } from '../types';

export class Node {
  private _state: NodeState;
  private _position: NodePosition;
  private _peers: string[];
  private _lastUpdated: number;
  
  constructor(
    public readonly id: string,
    initialState: NodeState = 'Red',
    position?: NodePosition
  ) {
    this._state = initialState;
    this._position = position ? { ...position } : { x: 0, y: 0 };
    this._peers = [];
    this._lastUpdated = 0;
  }
  
  get state(): NodeState {
    return this._state;
  }
  
  get position(): NodePosition {
    return { ...this._position };
  }
  
  get peers(): string[] {
    return [...this._peers];
  }
  
  get lastUpdated(): number {
    return this._lastUpdated;
  }
  
  setState(newState: NodeState, round?: number): void {
    if (!this.isValidState(newState)) {
      throw new Error(`Invalid state: ${newState}`);
    }
    this._state = newState;
    if (round !== undefined) {
      this._lastUpdated = round;
    }
  }
  
  setPosition(newPosition: NodePosition): void {
    this._position = { ...newPosition };
  }
  
  addPeer(peerId: string): void {
    if (!this._peers.includes(peerId)) {
      this._peers.push(peerId);
    }
  }
  
  setPeers(peerIds: string[]): void {
    this._peers = [...peerIds];
  }
  
  private isValidState(state: string): state is NodeState {
    return ['Red', 'Green', 'Blue'].includes(state);
  }
}