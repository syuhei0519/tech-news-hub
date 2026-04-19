import { http, HttpResponse } from "msw";
import {
  buildArticlesResponse,
  buildFetchJobsResponse,
  buildNotificationsResponse,
} from "./fixtures";
import type {
  Article,
  FetchJob,
  Notification,
  Source,
} from "../lib/api";

const apiBaseUrl = "http://localhost:8080";

export function sourcesHandler(items: Source[] | (() => Source[])) {
  return http.get(`${apiBaseUrl}/api/v1/sources`, () =>
    HttpResponse.json({ items: typeof items === "function" ? items() : items }),
  );
}

export function sourceDetailHandler(source: Source) {
  return http.get(`${apiBaseUrl}/api/v1/sources/:id`, () => HttpResponse.json(source));
}

export function articlesHandler(items: Article[]) {
  return http.get(`${apiBaseUrl}/api/v1/articles`, () => HttpResponse.json(buildArticlesResponse(items)));
}

export function articleDetailHandler(article: Article) {
  return http.get(`${apiBaseUrl}/api/v1/articles/:id`, () => HttpResponse.json(article));
}

export function notificationsHandler(items: Notification[]) {
  return http.get(`${apiBaseUrl}/api/v1/notifications`, () =>
    HttpResponse.json(buildNotificationsResponse(items)),
  );
}

export function fetchJobsHandler(items: FetchJob[]) {
  return http.get(`${apiBaseUrl}/api/v1/fetch-jobs`, () =>
    HttpResponse.json(buildFetchJobsResponse(items)),
  );
}
