CREATE TABLE IF NOT EXISTS sources (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  type VARCHAR(64) NOT NULL,
  fetch_url TEXT NOT NULL,
  fetch_method VARCHAR(64) NOT NULL,
  interval_minutes INT NOT NULL DEFAULT 60,
  default_category VARCHAR(128) NOT NULL,
  is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  last_fetched_at DATETIME NULL,
  last_fetch_status VARCHAR(32) NULL,
  last_error_message TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_sources_name (name)
);

CREATE TABLE IF NOT EXISTS articles (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(512) NOT NULL,
  url TEXT NOT NULL,
  source_id BIGINT NOT NULL,
  published_at DATETIME NULL,
  fetched_at DATETIME NOT NULL,
  excerpt TEXT NULL,
  category VARCHAR(128) NOT NULL,
  tags JSON NULL,
  is_read BOOLEAN NOT NULL DEFAULT FALSE,
  is_favorite BOOLEAN NOT NULL DEFAULT FALSE,
  dedupe_key VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_articles_dedupe_key (dedupe_key),
  INDEX idx_articles_source_id (source_id),
  INDEX idx_articles_published_at (published_at),
  FULLTEXT KEY ft_articles_title_excerpt (title, excerpt),
  CONSTRAINT fk_articles_source_id FOREIGN KEY (source_id) REFERENCES sources(id)
);

CREATE TABLE IF NOT EXISTS fetch_jobs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  source_id BIGINT NOT NULL,
  started_at DATETIME NOT NULL,
  finished_at DATETIME NULL,
  status VARCHAR(32) NOT NULL,
  fetched_count INT NOT NULL DEFAULT 0,
  inserted_count INT NOT NULL DEFAULT 0,
  duplicated_count INT NOT NULL DEFAULT 0,
  error_message TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_fetch_jobs_source_id FOREIGN KEY (source_id) REFERENCES sources(id)
);
