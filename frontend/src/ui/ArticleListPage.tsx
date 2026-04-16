import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { Link } from "react-router-dom";
import { z } from "zod";
import { buildArticleExportUrl, fetchArticles, fetchSources } from "../lib/api";
import { useSearchStore } from "../store/searchStore";

const schema = z.object({
  q: z.string().max(100).default(""),
  category: z.string().max(50).default(""),
  sourceId: z.string().default(""),
  isReadOnly: z.boolean().default(false),
  isFavoriteOnly: z.boolean().default(false),
  from: z.string().default(""),
  to: z.string().default(""),
});

type FormValues = z.infer<typeof schema>;

export function ArticleListPage() {
  const { q, category, sourceId, isReadOnly, isFavoriteOnly, from, to, setFilters } = useSearchStore();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { q, category, sourceId, isReadOnly, isFavoriteOnly, from, to },
  });
  const sourcesQuery = useQuery({
    queryKey: ["sources", "article-filters"],
    queryFn: fetchSources,
  });

  const articlesQuery = useQuery({
    queryKey: ["articles", q, category, sourceId, isReadOnly, isFavoriteOnly, from, to],
    queryFn: () =>
      fetchArticles({
        q,
        category,
        source_id: sourceId ? Number(sourceId) : undefined,
        is_read: isReadOnly ? false : undefined,
        is_favorite: isFavoriteOnly ? true : undefined,
        from: from || undefined,
        to: to || undefined,
      }),
  });
  const exportURL = buildArticleExportUrl({
    q,
    category,
    source_id: sourceId ? Number(sourceId) : undefined,
    is_read: isReadOnly ? false : undefined,
    is_favorite: isFavoriteOnly ? true : undefined,
    from: from || undefined,
    to: to || undefined,
  });

  const onSubmit = form.handleSubmit((values) => {
    setFilters(values);
  });

  return (
    <div className="space-y-8">
      <section className="rounded-3xl border border-white/10 bg-white/5 p-6 shadow-2xl shadow-slate-950/40">
        <div className="mb-4 flex flex-col gap-2">
          <p className="text-sm uppercase tracking-[0.35em] text-cyan-300">Phase 1</p>
          <h1 className="text-4xl font-semibold tracking-tight text-white">技術情報の一覧・検索</h1>
          <p className="max-w-3xl text-sm leading-6 text-slate-300">
            article-service から取得した記事を一覧表示します。検索条件は CSV ダウンロードにもそのまま反映されます。
          </p>
        </div>

        <form onSubmit={onSubmit} className="grid gap-4 md:grid-cols-2 xl:grid-cols-[2fr,1fr,1fr,1fr,auto,auto,auto]">
          <input
            {...form.register("q")}
            placeholder="タイトル・概要で検索"
            className="rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 text-sm text-white outline-none ring-0 placeholder:text-slate-500"
          />
          <input
            {...form.register("category")}
            placeholder="category 例: kubernetes"
            className="rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 text-sm text-white outline-none placeholder:text-slate-500"
          />
          <select
            {...form.register("sourceId")}
            className="rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 text-sm text-white outline-none"
          >
            <option value="">all sources</option>
            {sourcesQuery.data?.map((source) => (
              <option key={source.id} value={String(source.id)}>
                {source.name}
              </option>
            ))}
          </select>
          <input
            type="datetime-local"
            {...form.register("from")}
            className="rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 text-sm text-white outline-none placeholder:text-slate-500"
          />
          <input
            type="datetime-local"
            {...form.register("to")}
            className="rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 text-sm text-white outline-none placeholder:text-slate-500"
          />
          <button
            type="submit"
            className="rounded-2xl bg-amber-400 px-5 py-3 text-sm font-medium text-slate-950 transition hover:bg-amber-300"
          >
            Search
          </button>
          <label className="flex items-center gap-3 rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 text-sm text-slate-200">
            <input type="checkbox" {...form.register("isReadOnly")} className="size-4 rounded border-white/20" />
            未読のみ
          </label>
          <label className="flex items-center gap-3 rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 text-sm text-slate-200">
            <input type="checkbox" {...form.register("isFavoriteOnly")} className="size-4 rounded border-white/20" />
            お気に入りのみ
          </label>
        </form>
      </section>

      <section className="space-y-4">
        <div className="flex items-center justify-between text-sm text-slate-400">
          <span>{articlesQuery.data?.total ?? 0} articles</span>
          <div className="flex items-center gap-4">
            <span>公開日範囲・source フィルタ対応</span>
            <a
              href={exportURL}
              className="rounded-full border border-cyan-300/40 px-4 py-2 text-cyan-200 transition hover:border-cyan-200 hover:text-white"
            >
              CSV Download
            </a>
          </div>
        </div>

        {articlesQuery.isLoading ? <LoadingState /> : null}
        {articlesQuery.isError ? (
          <div className="rounded-2xl border border-rose-400/30 bg-rose-950/40 p-4 text-sm text-rose-200">
            API 取得に失敗しました。
          </div>
        ) : null}
        {sourcesQuery.isError ? (
          <div className="rounded-2xl border border-rose-400/30 bg-rose-950/40 p-4 text-sm text-rose-200">
            source 一覧の取得に失敗しました。
          </div>
        ) : null}

        <div className="grid gap-4">
          {articlesQuery.data?.items.map((article) => (
            <Link
              key={article.id}
              to={`/articles/${article.id}`}
              className="group rounded-3xl border border-white/10 bg-slate-900/70 p-5 transition hover:border-cyan-300/50 hover:bg-slate-900"
            >
              <div className="mb-3 flex flex-wrap gap-2 text-xs text-slate-400">
                <span className="rounded-full bg-cyan-400/10 px-3 py-1 text-cyan-200">{article.category}</span>
                <span>{article.source_name}</span>
                <span>{formatDate(article.published_at ?? article.fetched_at)}</span>
                <span
                  className={
                    article.is_read
                      ? "rounded-full bg-slate-800 px-3 py-1 text-slate-300"
                      : "rounded-full bg-emerald-400/10 px-3 py-1 text-emerald-200"
                  }
                >
                  {article.is_read ? "既読" : "未読"}
                </span>
                {article.is_favorite ? (
                  <span className="rounded-full bg-amber-400/10 px-3 py-1 text-amber-200">お気に入り</span>
                ) : null}
              </div>
              <h2 className="mb-2 text-xl font-medium text-white group-hover:text-cyan-200">{article.title}</h2>
              <p className="line-clamp-3 text-sm leading-6 text-slate-300">{article.excerpt || "概要は未取得です。"}</p>
            </Link>
          ))}
        </div>
      </section>
    </div>
  );
}

function LoadingState() {
  return (
    <div className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-slate-300">
      Loading articles...
    </div>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
