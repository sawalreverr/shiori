export interface NewsItem {
    title: string;
    url: string;
    category: string;
    published_at?: string;
}

export interface SourceGroup {
    id: string;
    news: NewsItem[];
}

export interface ApiResponse {
    status: string;
    items: SourceGroup[];
}
