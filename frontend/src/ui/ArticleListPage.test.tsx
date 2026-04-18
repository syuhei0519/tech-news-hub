import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { ArticleListPage } from "./ArticleListPage";
import { renderWithProviders } from "../test/render";
import { server } from "../test/setup";

const article = {
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
  is_favorite: true,
  created_at: "2026-04-16T01:00:00Z",
  updated_at: "2026-04-16T01:00:00Z",
};

const sources = [
  {
    id: 3,
    name: "Tech Blog",
    type: "rss",
    fetch_url: "https://example.com/feed.xml",
    fetch_method: "rss",
    interval_minutes: 60,
    default_category: "kubernetes",
    is_enabled: true,
    last_fetched_at: null,
    last_fetch_status: null,
    last_error_message: null,
    created_at: "2026-04-16T00:00:00Z",
    updated_at: "2026-04-16T00:00:00Z",
  },
];

describe("ArticleListPage", () => {
  it("renders articles and sources on initial load", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/sources", () => HttpResponse.json({ items: sources })),
      http.get("http://localhost:8080/api/v1/articles", () =>
        HttpResponse.json({ items: [article], total: 1, page: 1, page_size: 20, total_pages: 1 }),
      ),
    );

    renderWithProviders(<ArticleListPage />);

    expect(screen.getByText("Loading articles...")).toBeInTheDocument();
    expect(await screen.findByText("Kubernetes Upgrade Guide")).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Tech Blog" })).toBeInTheDocument();
    expect(screen.getByText("Kubernetes Upgrade Guide").closest("a")).toHaveAttribute("href", "/articles/1");
  });

  it("shows an error when article loading fails", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/sources", () => HttpResponse.json({ items: sources })),
      http.get("http://localhost:8080/api/v1/articles", () => new HttpResponse(null, { status: 500 })),
    );

    renderWithProviders(<ArticleListPage />);

    expect(await screen.findByText("API 取得に失敗しました。")).toBeInTheDocument();
  });

  it("shows an error when source loading fails", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/sources", () => new HttpResponse(null, { status: 500 })),
      http.get("http://localhost:8080/api/v1/articles", () =>
        HttpResponse.json({ items: [article], total: 1, page: 1, page_size: 20, total_pages: 1 }),
      ),
    );

    renderWithProviders(<ArticleListPage />);

    expect(await screen.findByText("source 一覧の取得に失敗しました。")).toBeInTheDocument();
  });

  it("submits filters and reflects them in the request and CSV link", async () => {
    const user = userEvent.setup();
    let latestQuery = "";

    server.use(
      http.get("http://localhost:8080/api/v1/sources", () => HttpResponse.json({ items: sources })),
      http.get("http://localhost:8080/api/v1/articles", ({ request }) => {
        latestQuery = new URL(request.url).search;
        return HttpResponse.json({ items: [article], total: 1, page: 1, page_size: 20, total_pages: 1 });
      }),
    );

    renderWithProviders(<ArticleListPage />);

    await screen.findByText("Kubernetes Upgrade Guide");

    await user.type(screen.getByPlaceholderText("タイトル・概要で検索"), "platform");
    await user.type(screen.getByPlaceholderText("category 例: kubernetes"), "kubernetes");
    await user.selectOptions(screen.getByRole("combobox"), "3");

    const dateInputs = screen.getAllByDisplayValue("");
    await user.type(dateInputs[0], "2026-04-16T09:30");
    await user.type(dateInputs[1], "2026-04-17T18:45");
    await user.click(screen.getByLabelText("未読のみ"));
    await user.click(screen.getByLabelText("お気に入りのみ"));
    await user.click(screen.getByRole("button", { name: "Search" }));

    // 一覧 API と CSV export の両方で同じ filter 条件が使われることを守る。
    await waitFor(() => expect(latestQuery).toContain("q=platform"));
    const queryParams = new URLSearchParams(latestQuery);
    expect(queryParams.get("category")).toBe("kubernetes");
    expect(queryParams.get("source_id")).toBe("3");
    expect(queryParams.get("is_read")).toBe("false");
    expect(queryParams.get("is_favorite")).toBe("true");
    expect(queryParams.get("from")).toBe(new Date("2026-04-16T09:30").toISOString());
    expect(queryParams.get("to")).toBe(new Date("2026-04-17T18:45").toISOString());

    const exportUrl = new URL(screen.getByRole("link", { name: "CSV Download" }).getAttribute("href")!);
    expect(exportUrl.pathname).toBe("/api/v1/exports/articles.csv");
    expect(exportUrl.searchParams.get("q")).toBe("platform");
    expect(exportUrl.searchParams.get("category")).toBe("kubernetes");
    expect(exportUrl.searchParams.get("source_id")).toBe("3");
    expect(exportUrl.searchParams.get("is_read")).toBe("false");
    expect(exportUrl.searchParams.get("is_favorite")).toBe("true");
    expect(exportUrl.searchParams.get("from")).toBe(new Date("2026-04-16T09:30").toISOString());
    expect(exportUrl.searchParams.get("to")).toBe(new Date("2026-04-17T18:45").toISOString());
  });
});
