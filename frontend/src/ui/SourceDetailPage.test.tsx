import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { SourceDetailPage } from "./SourceDetailPage";
import { renderWithProviders } from "../test/render";
import { server } from "../test/setup";

type SourceItem = {
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

type FetchJobItem = {
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

function buildSource(): SourceItem {
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
  };
}

function buildJob(id: number, status: "running" | "success" | "failed", finishedAt: string | null, errorMessage: string | null): FetchJobItem {
  return {
    id,
    source_id: 1,
    started_at: new Date(Date.UTC(2026, 3, id, 0, 0, 0)).toISOString(),
    finished_at: finishedAt,
    status,
    fetched_count: 10,
    inserted_count: 7,
    duplicated_count: 3,
    error_message: errorMessage,
  };
}

describe("SourceDetailPage", () => {
  it("shows an error for an invalid source id", () => {
    renderWithProviders(<SourceDetailPage />, { path: "/sources/:id", route: "/sources/abc" });
    expect(screen.getByText("Invalid source id")).toBeInTheDocument();
  });

  it("loads source details, applies job filters, and paginates", async () => {
    const user = userEvent.setup();
    const requestLog: string[] = [];
    const source = buildSource();
    const jobs = [buildJob(11, "running", null, null), ...Array.from({ length: 10 }, (_, index) => buildJob(10 - index, index % 2 === 0 ? "failed" : "success", new Date(Date.UTC(2026, 3, 10 - index, 1, 0, 0)).toISOString(), index % 2 === 0 ? "collector failed" : null))];

    server.use(
      http.get("http://localhost:8080/api/v1/sources/1", () => HttpResponse.json(source)),
      http.get("http://localhost:8080/api/v1/fetch-jobs", ({ request }) => {
        const url = new URL(request.url);
        const status = url.searchParams.get("status");
        const page = Number(url.searchParams.get("page") ?? "1");
        const pageSize = Number(url.searchParams.get("page_size") ?? "10");
        requestLog.push(url.search);

        const filtered = status == null ? jobs : jobs.filter((job) => job.status === status);
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

    renderWithProviders(<SourceDetailPage />, { path: "/sources/:id", route: "/sources/1" });

    expect(await screen.findByRole("heading", { name: "Kubernetes Blog" })).toBeInTheDocument();
    expect(screen.getByText("Source Overview")).toBeInTheDocument();
    expect(screen.getByText("Fetch Job History")).toBeInTheDocument();
    const runningJobCard = screen.getByText("Job #11").closest("article")!;
    expect(within(runningJobCard).getByText(/not finished/)).toBeInTheDocument();
    expect(requestLog[0]).toContain("source_id=1");
    expect(requestLog[0]).toContain("page=1");
    expect(requestLog[0]).toContain("page_size=10");

    await user.click(screen.getByRole("button", { name: "Next" }));
    expect(await screen.findByText("Job #1")).toBeInTheDocument();

    await user.selectOptions(screen.getByLabelText("Status"), "failed");
    // failed filter では複数件が残るため、単一カードではなく filter 後の一覧全体を確認する。
    expect((await screen.findAllByText("collector failed")).length).toBeGreaterThan(0);

    // page 2 から status filter を変えたとき page 1 に戻ることを request query で固定する。
    await waitFor(() => expect(requestLog.at(-1)).toContain("status=failed"));
    expect(requestLog.at(-1)).toContain("page=1");
  });

  it("shows a source API error", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/sources/1", () =>
        HttpResponse.json({ error: "source detail failed" }, { status: 500 }),
      ),
      http.get("http://localhost:8080/api/v1/fetch-jobs", () =>
        HttpResponse.json({ items: [], total: 0, page: 1, page_size: 10, total_pages: 1 }),
      ),
    );

    renderWithProviders(<SourceDetailPage />, { path: "/sources/:id", route: "/sources/1" });

    expect(await screen.findByText("source detail failed")).toBeInTheDocument();
  });

  it("shows a fetch jobs API error", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/sources/1", () => HttpResponse.json(buildSource())),
      http.get("http://localhost:8080/api/v1/fetch-jobs", () =>
        HttpResponse.json({ error: "fetch jobs failed" }, { status: 500 }),
      ),
    );

    renderWithProviders(<SourceDetailPage />, { path: "/sources/:id", route: "/sources/1" });

    expect(await screen.findByText("fetch jobs failed")).toBeInTheDocument();
  });

  it("shows the empty job state", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/sources/1", () => HttpResponse.json(buildSource())),
      http.get("http://localhost:8080/api/v1/fetch-jobs", () =>
        HttpResponse.json({ items: [], total: 0, page: 1, page_size: 10, total_pages: 1 }),
      ),
    );

    renderWithProviders(<SourceDetailPage />, { path: "/sources/:id", route: "/sources/1" });

    expect(await screen.findByText("No fetch jobs matched the current filter.")).toBeInTheDocument();
  });
});
