import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { NotificationListPage } from "./NotificationListPage";
import { renderWithProviders } from "../test/render";
import { server } from "../test/setup";

type NotificationItem = {
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

function buildNotification(id: number, isRead: boolean): NotificationItem {
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
  };
}

describe("NotificationListPage", () => {
  it("loads notifications, applies filters, and paginates", async () => {
    const user = userEvent.setup();
    const requestLog: string[] = [];
    const notifications = Array.from({ length: 11 }, (_, index) => buildNotification(11 - index, (11 - index) % 2 === 0));

    server.use(
      http.get("http://localhost:8080/api/v1/notifications", ({ request }) => {
        const url = new URL(request.url);
        const isRead = url.searchParams.get("is_read");
        const page = Number(url.searchParams.get("page") ?? "1");
        const pageSize = Number(url.searchParams.get("page_size") ?? "10");
        requestLog.push(url.search);

        const filtered =
          isRead == null ? notifications : notifications.filter((item) => item.is_read === (isRead === "true"));
        const start = (page - 1) * pageSize;
        const items = filtered.slice(start, start + pageSize);

        return HttpResponse.json({
          items,
          total: filtered.length,
          page,
          page_size: pageSize,
          total_pages: Math.max(Math.ceil(filtered.length / pageSize), 1),
        });
      }),
    );

    renderWithProviders(<NotificationListPage />);

    expect(await screen.findByText("Notification 11")).toBeInTheDocument();
    expect(requestLog[0]).toContain("page=1");
    expect(requestLog[0]).toContain("page_size=10");
    expect(screen.getByText("11 notifications / page 1")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Next" }));
    expect(await screen.findByText("Notification 1")).toBeInTheDocument();
    expect(requestLog.at(-1)).toContain("page=2");

    await user.selectOptions(screen.getByLabelText("Status"), "unread");
    expect(await screen.findByText("Notification 11")).toBeInTheDocument();

    // page 2 から filter を変えたとき page 1 に戻ることを query で確認する。
    await waitFor(() => expect(requestLog.at(-1)).toContain("is_read=false"));
    expect(requestLog.at(-1)).toContain("page=1");
  });

  it("updates read status and makes the item visible under the read filter", async () => {
    const user = userEvent.setup();
    const requestLog: string[] = [];
    let notifications = [buildNotification(1, false)];

    server.use(
      http.get("http://localhost:8080/api/v1/notifications", ({ request }) => {
        const url = new URL(request.url);
        requestLog.push(url.search);
        const isRead = url.searchParams.get("is_read");
        const filtered =
          isRead == null ? notifications : notifications.filter((item) => item.is_read === (isRead === "true"));
        return HttpResponse.json({
          items: filtered,
          total: filtered.length,
          page: 1,
          page_size: 10,
          total_pages: 1,
        });
      }),
      http.patch("http://localhost:8080/api/v1/notifications/1/read-status", async ({ request }) => {
        const body = (await request.json()) as { is_read: boolean };
        notifications = notifications.map((item) =>
          item.id === 1
            ? {
                ...item,
                is_read: body.is_read,
                read_at: body.is_read ? new Date(Date.UTC(2026, 3, 18, 12, 0, 0)).toISOString() : null,
              }
            : item,
        );
        return HttpResponse.json(notifications[0]);
      }),
    );

    renderWithProviders(<NotificationListPage />);

    expect(await screen.findByText("Notification 1")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "既読にする" }));

    expect(await screen.findByText("既読")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "未読に戻す" })).toBeInTheDocument();

    // 一覧内の局所更新だけでなく、read filter 側にも移動できる状態になっていることを確認する。
    await user.selectOptions(screen.getByLabelText("Status"), "read");
    expect(await screen.findByText("Notification 1")).toBeInTheDocument();
    expect(requestLog.at(-1)).toContain("is_read=true");
  });

  it("shows an API error message", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/notifications", () =>
        HttpResponse.json({ error: "notification api failed" }, { status: 500 }),
      ),
    );

    renderWithProviders(<NotificationListPage />);

    expect(await screen.findByText("notification api failed")).toBeInTheDocument();
  });

  it("shows the empty state when no notifications match the filter", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/notifications", () =>
        HttpResponse.json({
          items: [],
          total: 0,
          page: 1,
          page_size: 10,
          total_pages: 1,
        }),
      ),
    );

    renderWithProviders(<NotificationListPage />);

    expect(await screen.findByText("No notifications matched the current filter.")).toBeInTheDocument();
  });
});
