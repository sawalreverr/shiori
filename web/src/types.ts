export interface MarketNewsResponse {
    title: string;
    url: string;
    published_at?: string;
}

export interface SourceGroupResponse {
    id: string;
    news: MarketNewsResponse[];
}

export interface MarketResponse {
    status: string;
    source_count: number;
    total_news: number;
    last_scraped_at?: string;
    items: SourceGroupResponse[];
}
