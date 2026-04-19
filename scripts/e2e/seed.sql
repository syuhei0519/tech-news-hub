-- Playwright smoke 用の最小 seed。
-- article-flow.spec.ts と notification-flow.spec.ts が前提にする状態だけを作る。
-- 外部 RSS、RabbitMQ 到着待ち、現在時刻依存には寄せない。

DELETE FROM notifications;
DELETE FROM articles;
DELETE FROM fetch_jobs;
DELETE FROM sources;

ALTER TABLE notifications AUTO_INCREMENT = 1;
ALTER TABLE articles AUTO_INCREMENT = 1;
ALTER TABLE fetch_jobs AUTO_INCREMENT = 1;
ALTER TABLE sources AUTO_INCREMENT = 1;

INSERT INTO sources (
  id,
  name,
  type,
  fetch_url,
  fetch_method,
  interval_minutes,
  default_category,
  is_enabled,
  last_fetched_at,
  last_fetch_status,
  last_error_message,
  created_at,
  updated_at
) VALUES
  (
    101,
    'E2E Tech Blog',
    'rss',
    'https://example.com/e2e-tech-blog.xml',
    'rss',
    60,
    'kubernetes',
    TRUE,
    '2026-04-18 09:00:00',
    'success',
    NULL,
    '2026-04-18 08:00:00',
    '2026-04-18 09:00:00'
  ),
  (
    102,
    'Ops Journal',
    'rss',
    'https://example.com/ops-journal.xml',
    'rss',
    120,
    'operations',
    TRUE,
    '2026-04-18 09:00:00',
    'success',
    NULL,
    '2026-04-18 08:00:00',
    '2026-04-18 09:00:00'
  );

INSERT INTO articles (
  id,
  title,
  url,
  source_id,
  published_at,
  fetched_at,
  excerpt,
  category,
  tags,
  is_read,
  is_favorite,
  dedupe_key,
  created_at,
  updated_at
) VALUES
  (
    1001,
    'Platform Weekly Digest',
    'https://example.com/platform-weekly-digest',
    101,
    '2026-04-18 09:30:00',
    '2026-04-18 10:00:00',
    'Platform updates for kubernetes teams.',
    'kubernetes',
    JSON_ARRAY('kubernetes', 'platform'),
    FALSE,
    TRUE,
    'e2e-platform-weekly-digest',
    '2026-04-18 10:00:00',
    '2026-04-18 10:00:00'
  ),
  (
    1002,
    'SRE Incident Review',
    'https://example.com/sre-incident-review',
    102,
    '2026-04-17 18:00:00',
    '2026-04-18 10:00:00',
    'Lessons learned from a recent incident review.',
    'operations',
    JSON_ARRAY('operations', 'sre'),
    TRUE,
    FALSE,
    'e2e-sre-incident-review',
    '2026-04-18 10:00:00',
    '2026-04-18 10:00:00'
  );

INSERT INTO fetch_jobs (
  id,
  source_id,
  started_at,
  finished_at,
  status,
  fetched_count,
  inserted_count,
  duplicated_count,
  error_message,
  created_at
) VALUES
  (
    501,
    101,
    '2026-04-18 09:00:00',
    '2026-04-18 09:03:00',
    'success',
    2,
    1,
    1,
    NULL,
    '2026-04-18 09:00:00'
  ),
  (
    502,
    102,
    '2026-04-18 09:05:00',
    '2026-04-18 09:06:00',
    'failed',
    0,
    0,
    0,
    'fetch rss status=502',
    '2026-04-18 09:05:00'
  );

INSERT INTO notifications (
  id,
  event_id,
  event_type,
  level,
  title,
  body,
  source_id,
  fetch_job_id,
  is_read,
  created_at,
  read_at
) VALUES
  (
    701,
    'evt-e2e-article-ingested',
    'article.ingested',
    'info',
    'Platform Weekly Digest を取り込みました',
    'Latest article: Platform Weekly Digest',
    101,
    501,
    TRUE,
    '2026-04-18 10:01:00',
    '2026-04-18 10:05:00'
  ),
  (
    702,
    'evt-e2e-collector-failed',
    'collector.fetch.failed',
    'error',
    'Collector failed for E2E feed',
    'fetch rss status=502',
    102,
    502,
    FALSE,
    '2026-04-18 10:02:00',
    NULL
  );
