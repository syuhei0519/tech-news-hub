import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { SourceDetailPage } from "./SourceDetailPage";
import { renderWithProviders } from "../test/render";
import { server } from "../test/setup";
import { buildFetchJob, buildFetchJobsResponse, buildSource } from "../test/fixtures";
import { sourceDetailHandler } from "../test/handlers";

describe("SourceDetailPage", () => {
  it("shows an error for an invalid source id", () => {
    renderWithProviders(<SourceDetailPage />, { path: "/sources/:id", route: "/sources/abc" });
    expect(screen.getByText("Invalid source id")).toBeInTheDocument();
  });

  it("loads source details, applies job filters, and paginates", async () => {
    const user = userEvent.setup();
    const requestLog: string[] = [];
    const source = buildSource();
    const jobs = [
      buildFetchJob({ id: 11, status: "running", finished_at: null }),
      ...Array.from({ length: 10 }, (_, index) =>
        buildFetchJob({
          id: 10 - index,
          status: index % 2 === 0 ? "failed" : "success",
          error_message: index % 2 === 0 ? "collector failed" : null,
        }),
      ),
    ];

    server.use(
      sourceDetailHandler(source),
      http.get("http://localhost:8080/api/v1/fetch-jobs", ({ request }) => {
        const url = new URL(request.url);
        const status = url.searchParams.get("status");
        const page = Number(url.searchParams.get("page") ?? "1");
        const pageSize = Number(url.searchParams.get("page_size") ?? "10");
        requestLog.push(url.search);

        const filtered = status == null ? jobs : jobs.filter((job) => job.status === status);
        const start = (page - 1) * pageSize;
        const items = filtered.slice(start, start + pageSize);

        return HttpResponse.json(
          buildFetchJobsResponse(items, {
            total: filtered.length,
            page,
            page_size: pageSize,
            total_pages: Math.max(Math.ceil(filtered.length / pageSize), 1),
          }),
        );
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
      sourceDetailHandler(buildSource()),
      http.get("http://localhost:8080/api/v1/fetch-jobs", () =>
        HttpResponse.json({ error: "fetch jobs failed" }, { status: 500 }),
      ),
    );

    renderWithProviders(<SourceDetailPage />, { path: "/sources/:id", route: "/sources/1" });

    expect(await screen.findByText("fetch jobs failed")).toBeInTheDocument();
  });

  it("shows the empty job state", async () => {
    server.use(
      sourceDetailHandler(buildSource()),
      http.get("http://localhost:8080/api/v1/fetch-jobs", () =>
        HttpResponse.json({ items: [], total: 0, page: 1, page_size: 10, total_pages: 1 }),
      ),
    );

    renderWithProviders(<SourceDetailPage />, { path: "/sources/:id", route: "/sources/1" });

    expect(await screen.findByText("No fetch jobs matched the current filter.")).toBeInTheDocument();
  });
});
