export interface IndexSelector {
  selectIndex(arrayLength: number): number | null;
}

export class RandomIndexSelector implements IndexSelector {
  selectIndex(arrayLength: number): number | null {
    if (arrayLength <= 0) return null;
    return Math.floor(Math.random() * arrayLength);
  }
}

export class MockIndexSelector implements IndexSelector {
  private indices: number[] = [];
  private currentIndex = 0;
  
  constructor(indices?: number[]) {
    this.indices = indices || [0];
  }
  
  selectIndex(arrayLength: number): number | null {
    if (arrayLength <= 0) return null;
    
    const index = this.indices[this.currentIndex % this.indices.length];
    this.currentIndex++;
    
    // 範囲チェック: 配列範囲外なら最初のインデックスを返す
    return index < arrayLength ? index : 0;
  }
  
  setIndices(indices: number[]): void {
    this.indices = indices;
    this.currentIndex = 0;
  }
}
