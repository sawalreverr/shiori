// Client-side news poller for auto-refresh functionality
class NewsPoller {
  constructor() {
    this.isPolling = false;
    this.pollInterval = null;
    this.retryButton = document.getElementById('retry-button');
    this.refreshIndicator = document.getElementById('refresh-indicator');
    
    if (this.retryButton) {
      this.retryButton.addEventListener('click', () => this.fetchNews());
    }
  }

  async fetchNews() {
    try {
      const response = await fetch('http://localhost:8080/api/news');
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const data = await response.json();
      
      // Store the last fetched data timestamp
      localStorage.setItem('lastFetchedAt', new Date().toISOString());
      
      // Show refresh indicator
      if (this.refreshIndicator) {
        this.refreshIndicator.classList.remove('hidden');
      }
      
      // Hide retry button if it was shown due to error
      if (this.retryButton) {
        this.retryButton.classList.add('hidden');
      }
    } catch (error) {
      console.error('Polling error:', error);
      // Show retry button on error
      if (this.retryButton) {
        this.retryButton.classList.remove('hidden');
      }
    }
  }

  start() {
    if (this.isPolling) return;
    this.isPolling = true;
    
    // Initial fetch after a short delay
    setTimeout(() => {
      this.fetchNews();
      // Start polling every 60 seconds
      this.pollInterval = setInterval(() => this.fetchNews(), 60000);
    }, 2000);
  }

  stop() {
    this.isPolling = false;
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
      this.pollInterval = null;
    }
  }
}

// Initialize polling when page loads
if (typeof document !== 'undefined') {
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
      const poller = new NewsPoller();
      poller.start();
    });
  } else {
    const poller = new NewsPoller();
    poller.start();
  }
}