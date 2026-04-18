import axios from "axios";

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080",
  timeout: 10000,
});

export type Article = {
  id: number;
  title: string;
  url: string;
  source_id: number;
  source_name: string;
  published_at: string | null;
  fetched_at: string;
  excerpt: string;
  category: string;
  tags: string[];
  is_read: boolean;
  is_favorite: boolean;
  created_at: string;
  updated_at: string;
};

export type ListArticlesResponse = {
  items: Article[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
};

export type Source = {
  id: number;
  name: string;
  type: string;
  fetch_url: string;
  fetch_method: string;
  interval_minutes: number;
  default_category: string;
  is_enabled: boolean;
  last_fetched_at: string | null;
  last_fetch_status: string | null;
  last_error_message: string | null;
  created_at: string;
  updated_at: string;
};

export type SourceInput = {
  name: string;
  type: string;
  fetch_url: string;
  fetch_method: string;
  interval_minutes: number;
  default_category: string;
  is_enabled: boolean;
};

export type FetchJob = {
  id: number;
  source_id: number;
  started_at: string;
  finished_at: string | null;
  status: "running" | "success" | "failed";
  fetched_count: number;
  inserted_count: number;
  duplicated_count: number;
  error_message: string | null;
};

export type ListFetchJobsResponse = {
  items: FetchJob[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
};

export type Notification = {
  id: number;
  event_id: string;
  event_type: "article.ingested" | "collector.fetch.failed";
  level: "info" | "error";
  title: string;
  body: string;
  source_id: number | null;
  fetch_job_id: number | null;
  is_read: boolean;
  created_at: string;
  read_at: string | null;
};

export type ListNotificationsResponse = {
  items: Notification[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
};

export type ArticleQueryParams = {
  q?: string;
  category?: string;
  source_id?: number;
  is_read?: boolean;
  is_favorite?: boolean;
  from?: string;
  to?: string;
  sort?: string;
  order?: string;
  page?: number;
};

export async function fetchArticles(params: ArticleQueryParams) {
  const response = await api.get<ListArticlesResponse>("/api/v1/articles", { params: buildArticleQueryParams(params) });
  return response.data;
}

export function buildArticleExportUrl(params: ArticleQueryParams) {
  const searchParams = new URLSearchParams();
  for (const [key, value] of Object.entries(buildArticleQueryParams(params))) {
    if (value === undefined) {
      continue;
    }
    searchParams.set(key, String(value));
  }
  const query = searchParams.toString();
  const baseURL = typeof api.defaults.baseURL === "string" ? api.defaults.baseURL.replace(/\/$/, "") : "";
  return query ? `${baseURL}/api/v1/exports/articles.csv?${query}` : `${baseURL}/api/v1/exports/articles.csv`;
}

export async function fetchArticle(id: string) {
  const response = await api.get<Article>(`/api/v1/articles/${id}`);
  return response.data;
}

export async function updateArticleReadStatus(id: number, isRead: boolean) {
  const response = await api.patch<Article>(`/api/v1/articles/${id}/read-status`, { is_read: isRead });
  return response.data;
}

export async function updateArticleFavoriteStatus(id: number, isFavorite: boolean) {
  const response = await api.patch<Article>(`/api/v1/articles/${id}/favorite-status`, { is_favorite: isFavorite });
  return response.data;
}

export async function fetchSources() {
  const response = await api.get<{ items: Source[] }>("/api/v1/sources");
  return response.data.items;
}

export async function fetchSource(id: string | number) {
  const response = await api.get<Source>(`/api/v1/sources/${id}`);
  return response.data;
}

export async function fetchFetchJobs(params: {
  source_id: number;
  status?: "running" | "success" | "failed";
  page?: number;
  page_size?: number;
}) {
  const response = await api.get<ListFetchJobsResponse>("/api/v1/fetch-jobs", { params });
  return response.data;
}

export async function fetchNotifications(params: {
  is_read?: boolean;
  page?: number;
  page_size?: number;
}) {
  const response = await api.get<ListNotificationsResponse>("/api/v1/notifications", { params });
  return response.data;
}

export async function updateNotificationReadStatus(id: number, isRead: boolean) {
  const response = await api.patch<Notification>(`/api/v1/notifications/${id}/read-status`, { is_read: isRead });
  return response.data;
}

export async function createSource(input: SourceInput) {
  const response = await api.post<Source>("/api/v1/sources", input);
  return response.data;
}

export async function updateSource(id: number, input: SourceInput) {
  const response = await api.patch<Source>(`/api/v1/sources/${id}`, input);
  return response.data;
}

export async function deleteSource(id: number) {
  await api.delete(`/api/v1/sources/${id}`);
}

export function buildArticleQueryParams(params: ArticleQueryParams) {
  return {
    q: params.q || undefined,
    category: params.category || undefined,
    source_id: params.source_id,
    is_read: params.is_read,
    is_favorite: params.is_favorite,
    from: normalizeDateTimeInput(params.from),
    to: normalizeDateTimeInput(params.to),
    sort: params.sort || undefined,
    order: params.order || undefined,
    page: params.page,
  };
}

function normalizeDateTimeInput(value?: string) {
  if (!value) {
    return undefined;
  }
  return new Date(value).toISOString();
}
