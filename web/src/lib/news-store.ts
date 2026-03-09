import { writable } from 'svelte/store';
import { API_NEWS } from './config';

export interface MarketNews {
  title: string;
  url: string;
  published_at?: string;
}

export interface SourceGroup {
  id: string;
  news: MarketNews[];
}

export interface MarketData {
  status: string;
  source_count: number;
  total_news: number;
  last_scraped_at?: string;
  items: SourceGroup[];
}

interface StoreState {
  data: MarketData | null;
  fetching: boolean;
  lastUpdated: Date | null;
}

function createNewsStore() {
  const initialState: StoreState = {
    data: null,
    fetching: false,
    lastUpdated: null
  };
  
  const { subscribe, set, update } = writable<StoreState>(initialState);
  let pollInterval: ReturnType<typeof setInterval> | null = null;

  async function fetchNews() {
    update(s => ({ ...s, fetching: true }));
    
    try {
      const response = await fetch(API_NEWS);
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const data = await response.json();
      update(s => ({ 
        data, 
        fetching: false, 
        lastUpdated: new Date() 
      }));
    } catch (err) {
      console.error('Fetch error:', err);
      update(s => ({ ...s, fetching: false }));
    }
  }

  function startPolling(intervalMs = 60000) {
    fetchNews(); // Initial fetch
    if (pollInterval) clearInterval(pollInterval);
    pollInterval = setInterval(fetchNews, intervalMs);
  }

  function stopPolling() {
    if (pollInterval) {
      clearInterval(pollInterval);
      pollInterval = null;
    }
  }

  return {
    subscribe,
    fetchNews,
    startPolling,
    stopPolling
  };
}

export const newsStore = createNewsStore();