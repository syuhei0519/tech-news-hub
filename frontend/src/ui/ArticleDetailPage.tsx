import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import {
  type Article,
  fetchArticle,
  type ListArticlesResponse,
  updateArticleFavoriteStatus,
  updateArticleReadStatus,
} from "../lib/api";

export function ArticleDetailPage() {
  const { id = "" } = useParams();
  const queryClient = useQueryClient();
  const articleQuery = useQuery({
    queryKey: ["article", id],
    queryFn: () => fetchArticle(id),
  });

  const syncArticle = (article: Article) => {
    queryClient.setQueryData(["article", id], article);
    queryClient.setQueriesData({ queryKey: ["articles"] }, (current: ListArticlesResponse | undefined) => {
      if (!current) {
        return current;
      }
      return {
        ...current,
        items: current.items.map((item) => (item.id === article.id ? article : item)),
      };
    });
  };

  const readStatusMutation = useMutation({
    mutationFn: (isRead: boolean) => updateArticleReadStatus(Number(id), isRead),
    onSuccess: syncArticle,
  });

  const favoriteStatusMutation = useMutation({
    mutationFn: (isFavorite: boolean) => updateArticleFavoriteStatus(Number(id), isFavorite),
    onSuccess: syncArticle,
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
  const isUpdating = readStatusMutation.isPending || favoriteStatusMutation.isPending;

  return (
    <article className="space-y-6 rounded-3xl border border-white/10 bg-slate-900/70 p-8">
      <Link to="/" className="text-sm text-cyan-300">
        Back to articles
      </Link>
      <div className="space-y-3">
        <div className="flex flex-wrap gap-2 text-xs text-slate-400">
          <span className="rounded-full bg-amber-400/10 px-3 py-1 text-amber-200">{article.category}</span>
          <span>{article.source_name}</span>
          <span className={article.is_read ? "rounded-full bg-slate-800 px-3 py-1 text-slate-300" : "rounded-full bg-emerald-400/10 px-3 py-1 text-emerald-200"}>
            {article.is_read ? "既読" : "未読"}
          </span>
          {article.is_favorite ? (
            <span className="rounded-full bg-amber-300/10 px-3 py-1 text-amber-100">お気に入り</span>
          ) : null}
        </div>
        <h1 className="text-3xl font-semibold text-white">{article.title}</h1>
        <p className="text-sm text-slate-400">
          Published: {article.published_at ? formatDate(article.published_at) : "unknown"}
        </p>
      </div>
      <div className="flex flex-wrap gap-3">
        <button
          type="button"
          onClick={() => readStatusMutation.mutate(!article.is_read)}
          disabled={isUpdating}
          className="rounded-full border border-white/10 px-5 py-3 text-sm font-medium text-slate-100 transition hover:border-emerald-300/50 hover:text-emerald-200 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {article.is_read ? "未読に戻す" : "既読にする"}
        </button>
        <button
          type="button"
          onClick={() => favoriteStatusMutation.mutate(!article.is_favorite)}
          disabled={isUpdating}
          className="rounded-full border border-white/10 px-5 py-3 text-sm font-medium text-slate-100 transition hover:border-amber-300/50 hover:text-amber-200 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {article.is_favorite ? "お気に入り解除" : "お気に入りに追加"}
        </button>
      </div>
      {readStatusMutation.isError || favoriteStatusMutation.isError ? (
        <div className="rounded-2xl border border-rose-400/30 bg-rose-950/40 p-4 text-sm text-rose-200">
          状態の更新に失敗しました。
        </div>
      ) : null}
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
