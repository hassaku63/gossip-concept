import { useState } from 'react';
import { GossipProtocol } from '../../models/GossipProtocol';

export default function App() {
  const [protocol] = useState(() => new GossipProtocol());
  const [isInitialized, setIsInitialized] = useState(false);
  const [nodeCount, setNodeCount] = useState(10);
  const [currentRound, setCurrentRound] = useState(0);
  const [, setRefreshKey] = useState(0); // 強制再レンダリング用
  
  const handleInitialize = () => {
    protocol.initialize(nodeCount);
    setIsInitialized(true);
    setCurrentRound(0);
    setRefreshKey(prev => prev + 1);
  };
  
  const handleNodeCountChange = (newCount: number) => {
    setNodeCount(newCount);
    // ノード数が変更されたら自動で再初期化
    if (isInitialized) {
      protocol.initialize(newCount);
      setCurrentRound(0);
      setRefreshKey(prev => prev + 1);
    }
  };
  
  const handleInfectFirst = () => {
    const firstNode = protocol.findNode('node-0');
    if (firstNode) {
      firstNode.setState('Green', 0);
      setCurrentRound(0);
      setRefreshKey(prev => prev + 1);
    }
  };
  
  const handleExecuteRound = () => {
    const messages = protocol.executeRound();
    setCurrentRound(protocol.currentRound);
    setRefreshKey(prev => prev + 1);
    console.log(`Round ${protocol.currentRound}: ${messages.length} messages sent`);
    messages.forEach(msg => {
      console.log(`  ${msg.from} → ${msg.to} (${msg.state})`);
    });
  };
  
  const handleReset = () => {
    protocol.reset();
    setCurrentRound(0);
    // 状態の再評価をトリガーするため、強制的に再レンダリング
    setRefreshKey(prev => prev + 1);
  };
  
  return (
    <div style={{ padding: '20px', fontFamily: 'Arial, sans-serif' }}>
      <h1>Gossip Protocol Visualizer</h1>
      
      <div style={{ marginBottom: '20px' }}>
        <label>
          ノード数: 
          <input 
            type="number" 
            value={nodeCount} 
            onChange={(e) => handleNodeCountChange(Math.max(1, Math.min(1000, parseInt(e.target.value) || 10)))}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !isInitialized) {
                handleInitialize();
              }
            }}
            min="1" 
            max="1000"
          />
        </label>
        <button onClick={handleInitialize} style={{ marginLeft: '10px' }}>
          初期化
        </button>
      </div>
      
      {isInitialized && (
        <div>
          <div style={{ marginBottom: '20px' }}>
            <p>Round: {currentRound}</p>
            <p>ノード数: {protocol.nodes.length}</p>
            <p>状態分布: {
              ['Red', 'Green', 'Blue'].map(state => {
                const count = protocol.nodes.filter(n => n.state === state).length;
                return `${state}: ${count}`;
              }).join(', ')
            }</p>
            <p>収束状態: {protocol.isConverged() ? '収束済み' : '未収束'}</p>
          </div>
          
          <div style={{ marginBottom: '20px' }}>
            <button onClick={handleInfectFirst} style={{ marginRight: '10px' }}>
              感染開始 (node-0をGreenに)
            </button>
            <button onClick={handleExecuteRound} style={{ marginRight: '10px' }}>
              1Round実行
            </button>
            <button onClick={handleReset} style={{ marginRight: '10px' }}>
              リセット
            </button>
          </div>
          
          <div>
            <h3>ノード一覧</h3>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(80px, 1fr))', gap: '4px' }}>
              {protocol.nodes.map(node => (
                <div 
                  key={node.id} 
                  style={{ 
                    padding: '4px', 
                    border: '1px solid #ccc',
                    borderRadius: '2px',
                    fontSize: '10px',
                    textAlign: 'center',
                    backgroundColor: node.state === 'Red' ? '#ffebee' : 
                                   node.state === 'Green' ? '#e8f5e8' : '#e3f2fd'
                  }}
                >
                  <div style={{ fontWeight: 'bold' }}>{node.id}</div>
                  <div>{node.state}</div>
                  <div>R{node.lastUpdated}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}