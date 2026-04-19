import type {
  Article,
  FetchJob,
  ListArticlesResponse,
  ListFetchJobsResponse,
  ListNotificationsResponse,
  Notification,
  Source,
  SourceInput,
} from "../lib/api";

export function buildArticle(overrides: Partial<Article> = {}): Article {
  return {
    id: 1,
    title: "Kubernetes Upgrade Guide",
    url: "https://example.com/articles/1",
    source_id: 3,
    source_name: "Tech Blog",
    published_at: "2026-04-16T00:00:00Z",
    fetched_at: "2026-04-16T01:00:00Z",
    excerpt: "Upgrade notes for production clusters.",
    category: "kubernetes",
    tags: ["kubernetes"],
    is_read: false,
    is_favorite: false,
    created_at: "2026-04-16T01:00:00Z",
    updated_at: "2026-04-16T01:00:00Z",
    ...overrides,
  };
}

export function buildArticlesResponse(
  items: Article[],
  overrides: Partial<ListArticlesResponse> = {},
): ListArticlesResponse {
  return {
    items,
    total: items.length,
    page: 1,
    page_size: 20,
    total_pages: Math.max(items.length === 0 ? 0 : 1, 1),
    ...overrides,
  };
}

export function buildSource(overrides: Partial<Source> = {}): Source {
  return {
    id: 1,
    name: "Kubernetes Blog",
    type: "rss",
    fetch_url: "https://example.com/kubernetes.xml",
    fetch_method: "rss",
    interval_minutes: 60,
    default_category: "kubernetes",
    is_enabled: true,
    last_fetched_at: "2026-04-18T00:00:00Z",
    last_fetch_status: "success",
    last_error_message: null,
    created_at: "2026-04-18T00:00:00Z",
    updated_at: "2026-04-18T00:00:00Z",
    ...overrides,
  };
}

export function buildSourceInput(overrides: Partial<SourceInput> = {}): SourceInput {
  return {
    name: "Kubernetes Blog",
    type: "rss",
    fetch_url: "https://example.com/kubernetes.xml",
    fetch_method: "rss",
    interval_minutes: 60,
    default_category: "kubernetes",
    is_enabled: true,
    ...overrides,
  };
}

export function buildNotification(overrides: Partial<Notification> = {}): Notification {
  const id = overrides.id ?? 1;
  const isRead = overrides.is_read ?? false;

  return {
    id,
    event_id: `evt-${id}`,
    event_type: id % 2 === 0 ? "collector.fetch.failed" : "article.ingested",
    level: id % 2 === 0 ? "error" : "info",
    title: `Notification ${id}`,
    body: `Body ${id}`,
    source_id: 3,
    fetch_job_id: 10 + id,
    is_read: isRead,
    created_at: new Date(Date.UTC(2026, 3, id, 0, 0, 0)).toISOString(),
    read_at: isRead ? new Date(Date.UTC(2026, 3, id, 1, 0, 0)).toISOString() : null,
    ...overrides,
  };
}

export function buildNotificationsResponse(
  items: Notification[],
  overrides: Partial<ListNotificationsResponse> = {},
): ListNotificationsResponse {
  return {
    items,
    total: items.length,
    page: 1,
    page_size: 10,
    total_pages: Math.max(items.length === 0 ? 0 : 1, 1),
    ...overrides,
  };
}

export function buildFetchJob(
  overrides: Partial<FetchJob> = {},
): FetchJob {
  const id = overrides.id ?? 1;

  return {
    id,
    source_id: 1,
    started_at: new Date(Date.UTC(2026, 3, id, 0, 0, 0)).toISOString(),
    finished_at: new Date(Date.UTC(2026, 3, id, 1, 0, 0)).toISOString(),
    status: "success",
    fetched_count: 10,
    inserted_count: 7,
    duplicated_count: 3,
    error_message: null,
    ...overrides,
  };
}

export function buildFetchJobsResponse(
  items: FetchJob[],
  overrides: Partial<ListFetchJobsResponse> = {},
): ListFetchJobsResponse {
  return {
    items,
    total: items.length,
    page: 1,
    page_size: 10,
    total_pages: Math.max(items.length === 0 ? 0 : 1, 1),
    ...overrides,
  };
}
