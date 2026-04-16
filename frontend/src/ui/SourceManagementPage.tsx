import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AxiosError } from "axios";
import { useEffect, useState } from "react";
import type { ReactNode } from "react";
import { useForm } from "react-hook-form";
import { Link } from "react-router-dom";
import { z } from "zod";
import { createSource, fetchSources, Source, SourceInput, updateSource } from "../lib/api";

const schema = z.object({
  name: z.string().trim().min(1, "name is required").max(255, "name must be 255 characters or fewer"),
  type: z.literal("rss"),
  fetch_url: z.string().trim().url("fetch_url must be a valid URL").max(2048, "fetch_url is too long"),
  fetch_method: z.literal("rss"),
  interval_minutes: z.coerce.number().int().min(1, "interval_minutes must be at least 1").max(10080, "interval_minutes must be 10080 or fewer"),
  default_category: z.string().trim().min(1, "default_category is required").max(128, "default_category must be 128 characters or fewer"),
  is_enabled: z.boolean(),
});

type FormValues = z.infer<typeof schema>;

const defaultValues: FormValues = {
  name: "",
  type: "rss",
  fetch_url: "",
  fetch_method: "rss",
  interval_minutes: 60,
  default_category: "",
  is_enabled: true,
};

export function SourceManagementPage() {
  const queryClient = useQueryClient();
  const [selectedSourceId, setSelectedSourceId] = useState<number | null>(null);

  const sourcesQuery = useQuery({
    queryKey: ["sources"],
    queryFn: fetchSources,
  });

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues,
  });

  const selectedSource = sourcesQuery.data?.find((source) => source.id === selectedSourceId) ?? null;

  useEffect(() => {
    // 一覧で選択した source を編集フォームに同期し、新規作成時は初期値へ戻す。
    if (selectedSource) {
      form.reset(toFormValues(selectedSource));
      return;
    }
    form.reset(defaultValues);
  }, [form, selectedSource]);

  const saveMutation = useMutation({
    mutationFn: async (values: FormValues) => {
      const payload: SourceInput = {
        ...values,
        name: values.name.trim(),
        fetch_url: values.fetch_url.trim(),
        default_category: values.default_category.trim(),
      };
      if (selectedSourceId == null) {
        return createSource(payload);
      }
      return updateSource(selectedSourceId, payload);
    },
    onSuccess: async (savedSource) => {
      await queryClient.invalidateQueries({ queryKey: ["sources"] });
      setSelectedSourceId(savedSource.id);
    },
  });

  const toggleMutation = useMutation({
    mutationFn: (source: Source) =>
      // toggle 専用 API はまだ持たず、現状の完全更新 API を使って状態だけ切り替える。
      updateSource(source.id, {
        name: source.name,
        type: source.type,
        fetch_url: source.fetch_url,
        fetch_method: source.fetch_method,
        interval_minutes: source.interval_minutes,
        default_category: source.default_category,
        is_enabled: !source.is_enabled,
      }),
    onSuccess: async (updatedSource) => {
      await queryClient.invalidateQueries({ queryKey: ["sources"] });
      if (selectedSourceId === updatedSource.id) {
        setSelectedSourceId(updatedSource.id);
      }
    },
  });

  const onSubmit = form.handleSubmit(async (values) => {
    await saveMutation.mutateAsync(values);
  });

  return (
    <div className="space-y-8">
      <section className="rounded-3xl border border-white/10 bg-white/5 p-6 shadow-2xl shadow-slate-950/40">
        <div className="mb-4 flex flex-col gap-2">
          <p className="text-sm uppercase tracking-[0.35em] text-cyan-300">Phase 2</p>
          <h1 className="text-4xl font-semibold tracking-tight text-white">収集ソース管理</h1>
          <p className="max-w-3xl text-sm leading-6 text-slate-300">
            ソース一覧、作成、編集、有効/無効切り替えを行います。現在の collector-service は静的設定のままなので、
            ここでの変更は収集対象へはまだ自動反映されません。
          </p>
        </div>
      </section>

      <div className="grid gap-6 lg:grid-cols-[1.2fr,0.8fr]">
        <section className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-xl font-semibold text-white">Source List</h2>
              <p className="text-sm text-slate-400">{sourcesQuery.data?.length ?? 0} sources</p>
            </div>
            <button
              type="button"
              onClick={() => setSelectedSourceId(null)}
              className="rounded-full border border-cyan-300/40 px-4 py-2 text-sm text-cyan-200 transition hover:border-cyan-200 hover:text-white"
            >
              New Source
            </button>
          </div>

          {sourcesQuery.isLoading ? <PanelMessage>Loading sources...</PanelMessage> : null}
          {sourcesQuery.isError ? <ErrorMessage message={getErrorMessage(sourcesQuery.error)} /> : null}

          <div className="grid gap-4">
            {sourcesQuery.data?.map((source) => {
              const isSelected = source.id === selectedSourceId;
              return (
                <article
                  key={source.id}
                  className={`rounded-3xl border p-5 transition ${
                    isSelected
                      ? "border-cyan-300/60 bg-slate-900"
                      : "border-white/10 bg-slate-900/70 hover:border-white/20"
                  }`}
                >
                  <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <h3 className="text-lg font-medium text-white">{source.name}</h3>
                      <p className="text-sm text-slate-400">
                        {source.type} / {source.fetch_method} / every {source.interval_minutes} min
                      </p>
                    </div>
                    <button
                      type="button"
                      onClick={() => toggleMutation.mutate(source)}
                      disabled={toggleMutation.isPending}
                      className={`rounded-full px-4 py-2 text-sm font-medium transition ${
                        source.is_enabled
                          ? "bg-emerald-400/15 text-emerald-200 hover:bg-emerald-400/25"
                          : "bg-slate-700 text-slate-200 hover:bg-slate-600"
                      }`}
                    >
                      {source.is_enabled ? "Enabled" : "Disabled"}
                    </button>
                  </div>

                  <div className="mb-4 flex flex-wrap gap-2 text-xs text-slate-400">
                    <span className="rounded-full bg-cyan-400/10 px-3 py-1 text-cyan-200">
                      {source.default_category}
                    </span>
                    <span className="rounded-full bg-white/5 px-3 py-1">{formatFetchStatus(source)}</span>
                    <span className="rounded-full bg-white/5 px-3 py-1">
                      Last fetch: {formatDate(source.last_fetched_at)}
                    </span>
                  </div>

                  <p className="mb-3 break-all text-sm leading-6 text-slate-300">{source.fetch_url}</p>

                  {source.last_error_message ? (
                    <div className="mb-4 rounded-2xl border border-rose-400/30 bg-rose-950/30 p-3 text-sm text-rose-200">
                      {source.last_error_message}
                    </div>
                  ) : null}

                  <button
                    type="button"
                    onClick={() => setSelectedSourceId(source.id)}
                    className="text-sm font-medium text-cyan-300 transition hover:text-cyan-100"
                  >
                    Edit Source
                  </button>
                  <Link
                    to={`/sources/${source.id}`}
                    className="ml-4 text-sm font-medium text-amber-300 transition hover:text-amber-100"
                  >
                    View History
                  </Link>
                </article>
              );
            })}
          </div>
        </section>

        <section className="rounded-3xl border border-white/10 bg-slate-950/60 p-6">
          <div className="mb-5">
            <h2 className="text-xl font-semibold text-white">
              {selectedSource ? `Edit: ${selectedSource.name}` : "Create Source"}
            </h2>
            <p className="mt-2 text-sm leading-6 text-slate-400">
              {/* 削除は誤操作コストが高いため、まずは API のみ提供して UI からは外す。 */}
              Delete API は提供しますが、この画面からは実行しません。
            </p>
          </div>

          <form onSubmit={onSubmit} className="space-y-4">
            <FormField label="Name" error={form.formState.errors.name?.message}>
              <input
                {...form.register("name")}
                className="w-full rounded-2xl border border-white/10 bg-slate-900 px-4 py-3 text-sm text-white outline-none placeholder:text-slate-500"
                placeholder="Official Kubernetes Blog"
              />
            </FormField>

            <div className="grid gap-4 md:grid-cols-2">
              <FormField label="Type" error={form.formState.errors.type?.message}>
                <select
                  {...form.register("type")}
                  className="w-full rounded-2xl border border-white/10 bg-slate-900 px-4 py-3 text-sm text-white outline-none"
                >
                  <option value="rss">rss</option>
                </select>
              </FormField>

              <FormField label="Fetch Method" error={form.formState.errors.fetch_method?.message}>
                <select
                  {...form.register("fetch_method")}
                  className="w-full rounded-2xl border border-white/10 bg-slate-900 px-4 py-3 text-sm text-white outline-none"
                >
                  <option value="rss">rss</option>
                </select>
              </FormField>
            </div>

            <FormField label="Fetch URL" error={form.formState.errors.fetch_url?.message}>
              <input
                {...form.register("fetch_url")}
                className="w-full rounded-2xl border border-white/10 bg-slate-900 px-4 py-3 text-sm text-white outline-none placeholder:text-slate-500"
                placeholder="https://kubernetes.io/feed.xml"
              />
            </FormField>

            <div className="grid gap-4 md:grid-cols-2">
              <FormField label="Interval Minutes" error={form.formState.errors.interval_minutes?.message}>
                <input
                  {...form.register("interval_minutes", { valueAsNumber: true })}
                  type="number"
                  min={1}
                  max={10080}
                  className="w-full rounded-2xl border border-white/10 bg-slate-900 px-4 py-3 text-sm text-white outline-none"
                />
              </FormField>

              <FormField label="Default Category" error={form.formState.errors.default_category?.message}>
                <input
                  {...form.register("default_category")}
                  className="w-full rounded-2xl border border-white/10 bg-slate-900 px-4 py-3 text-sm text-white outline-none placeholder:text-slate-500"
                  placeholder="kubernetes"
                />
              </FormField>
            </div>

            <label className="flex items-center gap-3 rounded-2xl border border-white/10 bg-slate-900 px-4 py-3 text-sm text-slate-200">
              <input type="checkbox" {...form.register("is_enabled")} className="h-4 w-4 rounded border-white/20" />
              Source is enabled
            </label>

            {saveMutation.isError ? <ErrorMessage message={getErrorMessage(saveMutation.error)} /> : null}
            {toggleMutation.isError ? <ErrorMessage message={getErrorMessage(toggleMutation.error)} /> : null}

            <div className="flex flex-wrap gap-3">
              <button
                type="submit"
                disabled={saveMutation.isPending}
                className="rounded-2xl bg-amber-400 px-5 py-3 text-sm font-medium text-slate-950 transition hover:bg-amber-300 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {saveMutation.isPending ? "Saving..." : selectedSource ? "Save Changes" : "Create Source"}
              </button>
              <button
                type="button"
                onClick={() => {
                  setSelectedSourceId(null);
                  form.reset(defaultValues);
                }}
                className="rounded-2xl border border-white/10 px-5 py-3 text-sm font-medium text-slate-200 transition hover:border-white/20 hover:text-white"
              >
                Reset
              </button>
            </div>
          </form>
        </section>
      </div>
    </div>
  );
}

function FormField(props: { label: string; error?: string; children: ReactNode }) {
  return (
    <label className="block space-y-2">
      <span className="text-sm font-medium text-slate-200">{props.label}</span>
      {props.children}
      {props.error ? <p className="text-sm text-rose-300">{props.error}</p> : null}
    </label>
  );
}

function PanelMessage(props: { children: ReactNode }) {
  return <div className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-slate-300">{props.children}</div>;
}

function ErrorMessage(props: { message: string }) {
  return <div className="rounded-2xl border border-rose-400/30 bg-rose-950/40 p-4 text-sm text-rose-200">{props.message}</div>;
}

function formatFetchStatus(source: Source) {
  if (!source.last_fetch_status) {
    return "fetch status: never";
  }
  return `fetch status: ${source.last_fetch_status}`;
}

function formatDate(value: string | null) {
  if (!value) {
    return "never";
  }
  return new Intl.DateTimeFormat("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

function toFormValues(source: Source): FormValues {
  return {
    name: source.name,
    type: source.type === "rss" ? "rss" : "rss",
    fetch_url: source.fetch_url,
    fetch_method: source.fetch_method === "rss" ? "rss" : "rss",
    interval_minutes: source.interval_minutes,
    default_category: source.default_category,
    is_enabled: source.is_enabled,
  };
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
