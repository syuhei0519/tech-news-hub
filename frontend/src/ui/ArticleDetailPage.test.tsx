import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { ArticleDetailPage } from "./ArticleDetailPage";
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
  is_favorite: false,
  created_at: "2026-04-16T01:00:00Z",
  updated_at: "2026-04-16T01:00:00Z",
};

describe("ArticleDetailPage", () => {
  it("renders article details after a successful fetch", async () => {
    server.use(http.get("http://localhost:8080/api/v1/articles/1", () => HttpResponse.json(article)));

    renderWithProviders(<ArticleDetailPage />, { path: "/articles/:id", route: "/articles/1" });

    expect(await screen.findByText("Kubernetes Upgrade Guide")).toBeInTheDocument();
    expect(screen.getByText("Tech Blog")).toBeInTheDocument();
    expect(screen.getByText("未読")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Open original article" })).toHaveAttribute(
      "href",
      "https://example.com/articles/1",
    );
  });

  it("shows an error when article fetch fails", async () => {
    server.use(http.get("http://localhost:8080/api/v1/articles/1", () => new HttpResponse(null, { status: 500 })));

    renderWithProviders(<ArticleDetailPage />, { path: "/articles/:id", route: "/articles/1" });

    expect(await screen.findByText("記事詳細の取得に失敗しました。")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Back" })).toHaveAttribute("href", "/");
  });

  it("updates read and favorite state and syncs the cached article list", async () => {
    const user = userEvent.setup();
    const updatedReadArticle = { ...article, is_read: true };
    const updatedFavoriteArticle = { ...updatedReadArticle, is_favorite: true };

    server.use(
      http.get("http://localhost:8080/api/v1/articles/1", () => HttpResponse.json(article)),
      http.patch("http://localhost:8080/api/v1/articles/1/read-status", async () => HttpResponse.json(updatedReadArticle)),
      http.patch("http://localhost:8080/api/v1/articles/1/favorite-status", async () =>
        HttpResponse.json(updatedFavoriteArticle),
      ),
    );

    const { queryClient } = renderWithProviders(<ArticleDetailPage />, {
      path: "/articles/:id",
      route: "/articles/1",
    });

    // 詳細画面の mutation 成功後、一覧キャッシュも同じ記事状態へ同期されることを確認する。
    queryClient.setQueryData(["articles", "", "", "", false, false, "", ""], {
      items: [article],
      total: 1,
      page: 1,
      page_size: 20,
      total_pages: 1,
    });

    expect(await screen.findByText("Kubernetes Upgrade Guide")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "既読にする" }));
    expect(await screen.findByText("既読")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "未読に戻す" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "お気に入りに追加" }));
    expect(await screen.findByText("お気に入り")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "お気に入り解除" })).toBeInTheDocument();

    await waitFor(() => {
      expect(
        queryClient.getQueryData<{
          items: Array<{ is_read: boolean; is_favorite: boolean }>;
        }>(["articles", "", "", "", false, false, "", ""])?.items[0],
      ).toMatchObject({
        is_read: true,
        is_favorite: true,
      });
    });
  });

  it("shows an error when status update fails", async () => {
    const user = userEvent.setup();

    server.use(
      http.get("http://localhost:8080/api/v1/articles/1", () => HttpResponse.json(article)),
      http.patch("http://localhost:8080/api/v1/articles/1/read-status", () => new HttpResponse(null, { status: 500 })),
    );

    renderWithProviders(<ArticleDetailPage />, { path: "/articles/:id", route: "/articles/1" });

    expect(await screen.findByText("Kubernetes Upgrade Guide")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "既読にする" }));

    expect(await screen.findByText("状態の更新に失敗しました。")).toBeInTheDocument();
  });
});
