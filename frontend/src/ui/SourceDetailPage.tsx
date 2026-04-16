import { useQuery } from "@tanstack/react-query";
import { AxiosError } from "axios";
import { useState, type ReactNode } from "react";
import { Link, useParams } from "react-router-dom";
import { fetchFetchJobs, fetchSource, type FetchJob, type Source } from "../lib/api";

const pageSize = 10;

export function SourceDetailPage() {
  const params = useParams();
  const [status, setStatus] = useState<"all" | "success" | "failed" | "running">("all");
  const [page, setPage] = useState(1);

  const sourceId = Number(params.id);
  const isValidSourceID = Number.isInteger(sourceId) && sourceId > 0;

  const sourceQuery = useQuery({
    queryKey: ["source", sourceId],
    queryFn: () => fetchSource(sourceId),
    enabled: isValidSourceID,
  });

  const jobsQuery = useQuery({
    queryKey: ["fetch-jobs", sourceId, status, page],
    queryFn: () =>
      fetchFetchJobs({
        source_id: sourceId,
        status: status === "all" ? undefined : status,
        page,
        page_size: pageSize,
      }),
    enabled: isValidSourceID,
  });

  if (!isValidSourceID) {
    return <ErrorMessage message="Invalid source id" />;
  }

  return (
    <div className="space-y-8">
      <section className="rounded-3xl border border-white/10 bg-white/5 p-6 shadow-2xl shadow-slate-950/40">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="space-y-2">
            <p className="text-sm uppercase tracking-[0.35em] text-cyan-300">Source Detail</p>
            <h1 className="text-4xl font-semibold tracking-tight text-white">
              {sourceQuery.data?.name ?? "Loading source"}
            </h1>
            <p className="max-w-3xl text-sm leading-6 text-slate-300">
              最終取得結果とジョブ履歴を確認します。失敗の原因はここで先に把握し、collector のログ確認は補助的に使います。
            </p>
          </div>
          <Link
            to="/sources"
            className="rounded-full border border-cyan-300/40 px-4 py-2 text-sm text-cyan-200 transition hover:border-cyan-200 hover:text-white"
          >
            Back to Sources
          </Link>
        </div>
      </section>

      {sourceQuery.isLoading ? <PanelMessage>Loading source...</PanelMessage> : null}
      {sourceQuery.isError ? <ErrorMessage message={getErrorMessage(sourceQuery.error)} /> : null}

      {sourceQuery.data ? <SourceOverview source={sourceQuery.data} /> : null}

      <section className="rounded-3xl border border-white/10 bg-slate-950/60 p-6">
        <div className="mb-5 flex flex-wrap items-center justify-between gap-4">
          <div>
            <h2 className="text-xl font-semibold text-white">Fetch Job History</h2>
            <p className="mt-2 text-sm text-slate-400">
              {jobsQuery.data?.total ?? 0} jobs / page {jobsQuery.data?.page ?? page}
            </p>
          </div>
          <label className="flex items-center gap-3 text-sm text-slate-300">
            <span>Status</span>
            <select
              value={status}
              onChange={(event) => {
                setStatus(event.target.value as "all" | "success" | "failed" | "running");
                setPage(1);
              }}
              className="rounded-2xl border border-white/10 bg-slate-900 px-4 py-3 text-sm text-white outline-none"
            >
              <option value="all">all</option>
              <option value="success">success</option>
              <option value="failed">failed</option>
              <option value="running">running</option>
            </select>
          </label>
        </div>

        {jobsQuery.isLoading ? <PanelMessage>Loading fetch jobs...</PanelMessage> : null}
        {jobsQuery.isError ? <ErrorMessage message={getErrorMessage(jobsQuery.error)} /> : null}
        {!jobsQuery.isLoading && jobsQuery.data?.items.length === 0 ? (
          <PanelMessage>No fetch jobs matched the current filter.</PanelMessage>
        ) : null}

        {jobsQuery.data?.items.length ? (
          <div className="space-y-4">
            {jobsQuery.data.items.map((job) => (
              <FetchJobCard key={job.id} job={job} />
            ))}

            <div className="flex flex-wrap items-center justify-between gap-3 border-t border-white/10 pt-4">
              <p className="text-sm text-slate-400">
                Page {jobsQuery.data.page} / {Math.max(jobsQuery.data.total_pages, 1)}
              </p>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setPage((current) => Math.max(current - 1, 1))}
                  disabled={page <= 1}
                  className="rounded-2xl border border-white/10 px-4 py-2 text-sm text-slate-200 transition hover:border-white/20 hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
                >
                  Previous
                </button>
                <button
                  type="button"
                  onClick={() => setPage((current) => current + 1)}
                  disabled={jobsQuery.data.page >= jobsQuery.data.total_pages}
                  className="rounded-2xl border border-white/10 px-4 py-2 text-sm text-slate-200 transition hover:border-white/20 hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            </div>
          </div>
        ) : null}
      </section>
    </div>
  );
}

function SourceOverview(props: { source: Source }) {
  const { source } = props;

  return (
    <section className="grid gap-6 lg:grid-cols-[1.1fr,0.9fr]">
      <article className="rounded-3xl border border-white/10 bg-slate-900/70 p-6">
        <h2 className="mb-4 text-xl font-semibold text-white">Source Overview</h2>
        <dl className="grid gap-4 text-sm text-slate-300">
          <DetailRow label="Name" value={source.name} />
          <DetailRow label="Type" value={`${source.type} / ${source.fetch_method}`} />
          <DetailRow label="Interval" value={`every ${source.interval_minutes} minutes`} />
          <DetailRow label="Category" value={source.default_category} />
          <DetailRow label="Status" value={source.is_enabled ? "enabled" : "disabled"} />
          <DetailRow label="Fetch URL" value={source.fetch_url} />
        </dl>
      </article>

      <article className="rounded-3xl border border-white/10 bg-slate-900/70 p-6">
        <h2 className="mb-4 text-xl font-semibold text-white">Latest Result</h2>
        <div className="space-y-4 text-sm text-slate-300">
          <div className="flex flex-wrap gap-2">
            <span className={`rounded-full px-3 py-1 text-xs font-medium ${statusBadgeClass(source.last_fetch_status)}`}>
              {source.last_fetch_status ?? "never"}
            </span>
            <span className="rounded-full bg-white/5 px-3 py-1 text-xs text-slate-300">
              Last fetch: {formatDate(source.last_fetched_at)}
            </span>
          </div>
          <p className="leading-6 text-slate-300">
            {source.last_error_message ?? "最新のエラーメッセージはありません。"}
          </p>
        </div>
      </article>
    </section>
  );
}

function FetchJobCard(props: { job: FetchJob }) {
  const { job } = props;

  return (
    <article className="rounded-3xl border border-white/10 bg-slate-900/70 p-5">
      <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-3">
          <span className={`rounded-full px-3 py-1 text-xs font-medium ${statusBadgeClass(job.status)}`}>{job.status}</span>
          <span className="text-sm text-slate-400">Job #{job.id}</span>
        </div>
        <p className="text-sm text-slate-400">
          {formatDate(job.started_at)} - {formatDate(job.finished_at)}
        </p>
      </div>

      <div className="mb-3 grid gap-3 text-sm text-slate-300 md:grid-cols-3">
        <StatBox label="Fetched" value={String(job.fetched_count)} />
        <StatBox label="Inserted" value={String(job.inserted_count)} />
        <StatBox label="Duplicated" value={String(job.duplicated_count)} />
      </div>

      {job.error_message ? (
        <div className="rounded-2xl border border-rose-400/30 bg-rose-950/30 p-3 text-sm text-rose-200">
          {job.error_message}
        </div>
      ) : null}
    </article>
  );
}

function DetailRow(props: { label: string; value: string }) {
  return (
    <div className="grid gap-1">
      <dt className="text-xs uppercase tracking-[0.25em] text-slate-500">{props.label}</dt>
      <dd className="break-all text-sm text-slate-200">{props.value}</dd>
    </div>
  );
}

function StatBox(props: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
      <p className="text-xs uppercase tracking-[0.25em] text-slate-500">{props.label}</p>
      <p className="mt-2 text-lg font-semibold text-white">{props.value}</p>
    </div>
  );
}

function PanelMessage(props: { children: ReactNode }) {
  return <div className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-slate-300">{props.children}</div>;
}

function ErrorMessage(props: { message: string }) {
  return <div className="rounded-2xl border border-rose-400/30 bg-rose-950/40 p-4 text-sm text-rose-200">{props.message}</div>;
}

function formatDate(value: string | null) {
  if (!value) {
    return "not finished";
  }
  return new Intl.DateTimeFormat("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

function statusBadgeClass(status: string | null) {
  switch (status) {
    case "success":
      return "bg-emerald-400/15 text-emerald-200";
    case "failed":
      return "bg-rose-400/15 text-rose-200";
    case "running":
      return "bg-amber-400/15 text-amber-200";
    default:
      return "bg-slate-700 text-slate-200";
  }
}

function getErrorMessage(error: unknown) {
  if (error instanceof AxiosError) {
    return error.response?.data?.error ?? error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "Unknown error";
}
