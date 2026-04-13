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

export async function fetchArticles(params: {
  q?: string;
  category?: string;
  page?: number;
}) {
  const response = await api.get<ListArticlesResponse>("/api/v1/articles", { params });
  return response.data;
}

export async function fetchArticle(id: string) {
  const response = await api.get<Article>(`/api/v1/articles/${id}`);
  return response.data;
}
