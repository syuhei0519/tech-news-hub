import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { fetchArticle } from "../lib/api";

export function ArticleDetailPage() {
  const { id = "" } = useParams();
  const articleQuery = useQuery({
    queryKey: ["article", id],
    queryFn: () => fetchArticle(id),
  });

  if (articleQuery.isLoading) {
    return <div className="rounded-2xl border border-white/10 bg-white/5 p-6">Loading article...</div>;
  }

  if (articleQuery.isError || !articleQuery.data) {
    return (
      <div className="space-y-4">
        <Link to="/" className="text-sm text-cyan-300">
          Back
        </Link>
        <div className="rounded-2xl border border-rose-400/30 bg-rose-950/40 p-6">
          記事詳細の取得に失敗しました。
        </div>
      </div>
    );
  }

  const article = articleQuery.data;

  return (
    <article className="space-y-6 rounded-3xl border border-white/10 bg-slate-900/70 p-8">
      <Link to="/" className="text-sm text-cyan-300">
        Back to articles
      </Link>
      <div className="space-y-3">
        <div className="flex flex-wrap gap-2 text-xs text-slate-400">
          <span className="rounded-full bg-amber-400/10 px-3 py-1 text-amber-200">{article.category}</span>
          <span>{article.source_name}</span>
        </div>
        <h1 className="text-3xl font-semibold text-white">{article.title}</h1>
        <p className="text-sm text-slate-400">
          Published: {article.published_at ? formatDate(article.published_at) : "unknown"}
        </p>
      </div>
      <p className="text-base leading-8 text-slate-200">{article.excerpt || "概要は未取得です。"}</p>
      <a
        href={article.url}
        target="_blank"
        rel="noreferrer"
        className="inline-flex rounded-full bg-cyan-300 px-5 py-3 text-sm font-medium text-slate-950"
      >
        Open original article
      </a>
    </article>
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
