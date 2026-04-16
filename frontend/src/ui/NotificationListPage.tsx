import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AxiosError } from "axios";
import { useState } from "react";
import { fetchNotifications, updateNotificationReadStatus, type ListNotificationsResponse, type Notification } from "../lib/api";

const pageSize = 10;

export function NotificationListPage() {
  const [status, setStatus] = useState<"all" | "unread" | "read">("all");
  const [page, setPage] = useState(1);
  const queryClient = useQueryClient();

  const notificationsQuery = useQuery({
    queryKey: ["notifications", status, page],
    queryFn: () =>
      fetchNotifications({
        is_read: status === "all" ? undefined : status === "read",
        page,
        page_size: pageSize,
      }),
  });

  const readStatusMutation = useMutation({
    mutationFn: ({ id, isRead }: { id: number; isRead: boolean }) => updateNotificationReadStatus(id, isRead),
    onSuccess: async (updated) => {
      queryClient.setQueryData<ListNotificationsResponse | undefined>(["notifications", status, page], (current) => {
        if (!current) {
          return current;
        }
        return {
          ...current,
          items: current.items.map((item) => (item.id === updated.id ? updated : item)),
        };
      });
      await queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  return (
    <div className="space-y-8">
      <section className="rounded-3xl border border-white/10 bg-white/5 p-6 shadow-2xl shadow-slate-950/40">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="space-y-2">
            <p className="text-sm uppercase tracking-[0.35em] text-cyan-300">Phase 3</p>
            <h1 className="text-4xl font-semibold tracking-tight text-white">通知一覧</h1>
            <p className="max-w-3xl text-sm leading-6 text-slate-300">
              RabbitMQ 経由で生成された新着通知と取得失敗通知を確認します。通知の既読化はここで行います。
            </p>
          </div>
          <label className="flex items-center gap-3 text-sm text-slate-300">
            <span>Status</span>
            <select
              value={status}
              onChange={(event) => {
                setStatus(event.target.value as "all" | "unread" | "read");
                setPage(1);
              }}
              className="rounded-2xl border border-white/10 bg-slate-900 px-4 py-3 text-sm text-white outline-none"
            >
              <option value="all">all</option>
              <option value="unread">unread</option>
              <option value="read">read</option>
            </select>
          </label>
        </div>
      </section>

      <section className="rounded-3xl border border-white/10 bg-slate-950/60 p-6">
        <div className="mb-5 flex flex-wrap items-center justify-between gap-4">
          <div>
            <h2 className="text-xl font-semibold text-white">Notification Feed</h2>
            <p className="mt-2 text-sm text-slate-400">
              {notificationsQuery.data?.total ?? 0} notifications / page {notificationsQuery.data?.page ?? page}
            </p>
          </div>
        </div>

        {notificationsQuery.isLoading ? <PanelMessage>Loading notifications...</PanelMessage> : null}
        {notificationsQuery.isError ? <ErrorMessage message={getErrorMessage(notificationsQuery.error)} /> : null}
        {!notificationsQuery.isLoading && notificationsQuery.data?.items.length === 0 ? (
          <PanelMessage>No notifications matched the current filter.</PanelMessage>
        ) : null}

        {notificationsQuery.data?.items.length ? (
          <div className="space-y-4">
            {notificationsQuery.data.items.map((notification) => (
              <NotificationCard
                key={notification.id}
                notification={notification}
                isPending={readStatusMutation.isPending}
                onToggleRead={(isRead) => readStatusMutation.mutate({ id: notification.id, isRead })}
              />
            ))}

            <div className="flex flex-wrap items-center justify-between gap-3 border-t border-white/10 pt-4">
              <p className="text-sm text-slate-400">
                Page {notificationsQuery.data.page} / {Math.max(notificationsQuery.data.total_pages, 1)}
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
                  disabled={notificationsQuery.data.page >= notificationsQuery.data.total_pages}
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

function NotificationCard(props: {
  notification: Notification;
  isPending: boolean;
  onToggleRead: (isRead: boolean) => void;
}) {
  const { notification, isPending, onToggleRead } = props;

  return (
    <article
      className={`rounded-3xl border p-5 transition ${
        notification.is_read ? "border-white/10 bg-slate-900/55" : "border-cyan-300/25 bg-slate-900/85"
      }`}
    >
      <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-3">
          <span className={`rounded-full px-3 py-1 text-xs font-medium ${levelBadgeClass(notification.level)}`}>
            {notification.level}
          </span>
          <span className={`rounded-full px-3 py-1 text-xs ${notification.is_read ? "bg-slate-800 text-slate-300" : "bg-emerald-400/10 text-emerald-200"}`}>
            {notification.is_read ? "既読" : "未読"}
          </span>
          {notification.fetch_job_id ? <span className="text-sm text-slate-400">Job #{notification.fetch_job_id}</span> : null}
        </div>
        <p className="text-sm text-slate-400">{formatDate(notification.created_at)}</p>
      </div>

      <h2 className="mb-2 text-xl font-medium text-white">{notification.title}</h2>
      <p className="mb-4 text-sm leading-6 text-slate-300">{notification.body}</p>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap gap-2 text-xs text-slate-400">
          <span className="rounded-full bg-white/5 px-3 py-1">{notification.event_type}</span>
          {notification.source_id ? <span className="rounded-full bg-white/5 px-3 py-1">Source #{notification.source_id}</span> : null}
        </div>
        <button
          type="button"
          disabled={isPending}
          onClick={() => onToggleRead(!notification.is_read)}
          className="rounded-2xl border border-white/10 px-4 py-2 text-sm text-slate-200 transition hover:border-white/20 hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
        >
          {notification.is_read ? "未読に戻す" : "既読にする"}
        </button>
      </div>
    </article>
  );
}

function PanelMessage(props: { children: React.ReactNode }) {
  return <div className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-slate-300">{props.children}</div>;
}

function ErrorMessage(props: { message: string }) {
  return <div className="rounded-2xl border border-rose-400/30 bg-rose-950/40 p-4 text-sm text-rose-200">{props.message}</div>;
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

function levelBadgeClass(level: Notification["level"]) {
  switch (level) {
    case "error":
      return "bg-rose-400/15 text-rose-200";
    default:
      return "bg-cyan-400/15 text-cyan-200";
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
