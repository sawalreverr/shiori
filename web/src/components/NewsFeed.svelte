<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { newsStore, type MarketData } from '../lib/news-store';
  import { formatDistanceToNow, parseISO } from 'date-fns';
  import { id } from 'date-fns/locale';

  // Source config map
  const sourceMap: Record<string, { name: string; icon: string }> = {
    detik: { name: "Detik", icon: "/icons/detik.png" },
    liputan6: { name: "Liputan6", icon: "/icons/liputan6.png" },
    bloombergtechnoz: { name: "BloombergTechnoz", icon: "/icons/bloombergtechnoz.png" },
    bisnisindonesia: { name: "Bisnis Indonesia", icon: "/icons/bisnis-indonesia.ico" },
    cgsi: { name: "CGSI", icon: "/icons/cgsi.png" },
    cnbc: { name: "CNBC Indonesia", icon: "/icons/cnbc.png" },
    idnfinansials: { name: "IDNFinansials", icon: "/icons/idnfinancial.ico" },
    idxchannel: { name: "IDXChannel", icon: "/icons/idxchannel.png" },
    investorid: { name: "Investor ID", icon: "/icons/investorid.png" },
    kabarbursa: { name: "KabarBursa", icon: "/icons/kabarbursa.webp" },
    katadata: { name: "KataData", icon: "/icons/katadata.ico" },
    kontan: { name: "Kontan", icon: "/icons/kontan.ico" },
    republika: { name: "Republika", icon: "/icons/republika.png" },
    stockwatch: { name: "Stockwatch", icon: "/icons/stockwatch.png" },
    ajaib: { name: "Ajaib", icon: "/icons/ajaib.png" },
    sindonews: { name: "SindoNews", icon: "/icons/sindonews.ico" },
  };

  let storeState = $state<{ data: MarketData | null; fetching: boolean; lastUpdated: Date | null }>({
    data: null,
    fetching: false,
    lastUpdated: null
  });

  const unsubscribe = newsStore.subscribe((value) => {
    storeState = value;
  });

  let isPulling = $state(false);
  let pullStartY = 0;

  function getSourceConfig(groupId: string) {
    const normalizedId = groupId.toLowerCase().replace(/[^a-z0-9]/g, "");
    let config = sourceMap[normalizedId];
    if (!config) {
      const foundKey = Object.keys(sourceMap).find((key) => normalizedId.includes(key));
      config = foundKey ? sourceMap[foundKey] : { name: groupId.toUpperCase(), icon: "" };
    }
    return config;
  }

  function formatTime(dateStr?: string) {
    if (!dateStr) return "";
    try {
      return formatDistanceToNow(parseISO(dateStr), { addSuffix: true, locale: id });
    } catch {
      return "";
    }
  }

  function formatLastUpdated(date: Date | null) {
    if (!date) return "";
    return formatDistanceToNow(date, { addSuffix: true, locale: id });
  }

  function cleanTitle(title: string) {
    return title.replace(/<[^>]*>?/gm, "");
  }

  function retry() {
    newsStore.fetchNews();
  }

  function handleTouchStart(e: TouchEvent) {
    pullStartY = e.touches[0].clientY;
  }

  function handleTouchMove(e: TouchEvent) {
    const currentY = e.touches[0].clientY;
    const diff = currentY - pullStartY;
    if (diff > 0 && window.scrollY === 0 && !isPulling) {
      isPulling = true;
    }
  }

  async function handleTouchEnd() {
    if (isPulling) {
      isPulling = false;
      await newsStore.fetchNews();
    }
  }

  onMount(() => {
    newsStore.startPolling(60000);
  });

  onDestroy(() => {
    unsubscribe();
    newsStore.stopPolling();
  });

  let data = $derived(storeState.data);
  let fetching = $derived(storeState.fetching);
  let lastUpdated = $derived(storeState.lastUpdated);
  let loading = $derived(data === null && fetching);
  let error = $derived(data === null && !fetching && lastUpdated === null);
  let validGroups = $derived(data?.items?.filter((g) => g.news?.length > 0) ?? []);
</script>

<svelte:window 
  on:touchstart={handleTouchStart}
  on:touchmove={handleTouchMove}
  on:touchend={handleTouchEnd}
/>

{#if error}
  <div class="flex flex-col items-center justify-center py-20 text-gray-500">
    <div class="text-6xl mb-4">⚡</div>
    <div class="font-mono text-lg">Connection to server failed.</div>
    <p class="text-sm mt-2 opacity-50">Is the backend running on port 8080?</p>
    <button onclick={retry} class="mt-4 px-4 py-2 bg-gray-900 text-white font-mono text-sm hover:bg-gray-700 transition-colors">
      Retry
    </button>
  </div>
{:else if loading}
  <div class="max-w-7xl mx-auto px-4 py-4">
    <div class="flex gap-4 text-sm font-mono text-gray-400 border-b border-gray-200 pb-4 mb-6">
      <span>sources: <span class="bg-gray-200 animate-pulse h-4 w-8 inline-block rounded"></span></span>
      <span>news: <span class="bg-gray-200 animate-pulse h-4 w-12 inline-block rounded"></span></span>
    </div>
  </div>
  <div class="w-full max-w-400 mx-auto px-4 py-8">
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
      {#each Array(8) as _, i}
        <div class="bg-white border border-gray-200 shadow-sm animate-pulse" style="animation-delay: {i * 50}ms">
          <div class="p-4 border-b border-gray-100 flex items-center gap-3">
            <div class="w-8 h-8 bg-gray-200 rounded-sm"></div>
            <div>
              <div class="h-5 w-24 bg-gray-200 rounded mb-2"></div>
              <div class="h-3 w-16 bg-gray-200 rounded"></div>
            </div>
          </div>
          <div class="px-4 pt-2 pb-4 space-y-3">
            {#each Array(5) as _}
              <div class="h-4 bg-gray-100 rounded w-full"></div>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  </div>
{:else if data}
  {#if isPulling}
    <div class="fixed top-0 left-0 right-0 z-50 flex justify-center py-2 bg-gray-900 text-white text-sm font-mono">
      Pull to refresh...
    </div>
  {/if}

  <div class="max-w-7xl mx-auto px-4 py-4">
    <div class="flex flex-wrap gap-4 text-sm font-mono text-gray-500 border-b border-gray-200 pb-4 mb-6 items-center">
      <span>sources: <strong>{data.source_count}</strong></span>
      <span>news: <strong>{data.total_news}</strong></span>
      <span class="ml-auto flex items-center gap-2">
        {#if fetching}
          <span class="inline-block w-3 h-3 border-2 border-gray-400 border-t-transparent rounded-full animate-spin"></span>
          <span class="text-gray-400">Refreshing...</span>
        {:else if lastUpdated}
          <span class="text-gray-400">Updated {formatLastUpdated(lastUpdated)}</span>
        {/if}
      </span>
    </div>
  </div>

  {#if validGroups.length === 0}
    <div class="flex flex-col items-center justify-center py-20 text-gray-400">
      <svg class="w-24 h-24 mb-4 opacity-50" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
      </svg>
      <div class="font-mono text-lg">No news available</div>
      <p class="text-sm mt-2 opacity-50">Check back later for updates</p>
    </div>
  {:else}
    <div class="w-full max-w-400 mx-auto px-4 py-8">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        {#each validGroups as group, i (group.id)}
          {@const sourceConfig = getSourceConfig(group.id)}
          <article class="flex flex-col bg-white border border-gray-200 shadow-[2px_2px_0px_rgba(0,0,0,0.05)] hover:shadow-[4px_4px_0px_rgba(0,0,0,0.1)] transition-shadow duration-300" style="animation-delay: {i * 30}ms">
            <div class="p-4 border-b border-gray-100 flex items-center justify-between bg-gray-50/50">
              <div class="flex items-center gap-3">
                {#if sourceConfig.icon}
                  <img src={sourceConfig.icon} alt="{sourceConfig.name} logo" class="w-8 h-8 object-contain rounded-sm mix-blend-multiply" loading="lazy" />
                {:else}
                  <div class="w-8 h-8 bg-gray-200 rounded-sm flex items-center justify-center text-xs font-bold text-gray-500">
                    {sourceConfig.name.substring(0, 2)}
                  </div>
                {/if}
                <div>
                  <h2 class="font-bold text-lg tracking-tight text-gray-900 leading-none">{sourceConfig.name}</h2>
                  {#if group.news[0]?.published_at}
                    <span class="text-[10px] font-mono text-gray-400 uppercase tracking-wide">{formatTime(group.news[0].published_at)}</span>
                  {/if}
                </div>
              </div>
            </div>
            <div class="px-4 pt-2 pb-4">
              <div class="flex flex-col max-h-96 overflow-y-auto pr-2 scrollbar-thin scrollbar-thumb-gray-200 scrollbar-track-transparent">
                {#each group.news as news}
                  <a href={news.url} target="_blank" rel="noopener noreferrer" class="group block py-3 border-b border-gray-100 last:border-0 hover:bg-gray-50 -mx-2 px-2 transition-colors duration-200 focus:outline-none focus:bg-gray-100">
                    <h3 class="font-medium text-gray-900 group-hover:text-black leading-snug text-sm line-clamp-2">{cleanTitle(news.title)}</h3>
                  </a>
                {/each}
              </div>
            </div>
          </article>
        {/each}
      </div>
    </div>
  {/if}
{/if}

<style>
  .scrollbar-thin::-webkit-scrollbar { width: 4px; }
  .scrollbar-thin::-webkit-scrollbar-track { background: transparent; }
  .scrollbar-thin::-webkit-scrollbar-thumb { background-color: #e5e7eb; border-radius: 20px; }
</style>