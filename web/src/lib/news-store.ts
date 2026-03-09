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

function createNewsStore() {
  const { subscribe, set, update } = writable<MarketData | null>(null);
  let pollInterval: ReturnType<typeof setInterval> | null = null;

  async function fetchNews() {
    try {
      const response = await fetch(API_NEWS);
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const data = await response.json();
      set(data);
    } catch (err) {
      console.error('Fetch error:', err);
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