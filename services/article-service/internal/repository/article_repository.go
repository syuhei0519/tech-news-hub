package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"tech-feed-hub/article-service/internal/domain"
)

type ArticleRepository struct {
	db *sql.DB
}

func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) List(ctx context.Context, params domain.ListArticlesParams) (domain.ListArticlesResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}
	sortColumn := "COALESCE(a.published_at, a.fetched_at)"
	if params.Sort == "created_at" {
		sortColumn = "a.created_at"
	}
	order := "DESC"
	if strings.EqualFold(params.Order, "asc") {
		order = "ASC"
	}

	where := []string{"1=1"}
	args := make([]any, 0)

	if params.Query != "" {
		where = append(where, "(a.title LIKE ? OR a.excerpt LIKE ?)")
		q := "%" + params.Query + "%"
		args = append(args, q, q)
	}
	if params.Category != "" {
		where = append(where, "a.category = ?")
		args = append(args, params.Category)
	}
	if params.SourceID > 0 {
		where = append(where, "a.source_id = ?")
		args = append(args, params.SourceID)
	}
	if params.IsRead != nil {
		where = append(where, "a.is_read = ?")
		args = append(args, *params.IsRead)
	}
	if params.IsFavorite != nil {
		where = append(where, "a.is_favorite = ?")
		args = append(args, *params.IsFavorite)
	}

	whereClause := strings.Join(where, " AND ")
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM articles a WHERE %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return domain.ListArticlesResult{}, fmt.Errorf("count articles: %w", err)
	}

	offset := (params.Page - 1) * params.PageSize
	listQuery := fmt.Sprintf(`
		SELECT
			a.id, a.title, a.url, a.source_id, s.name, a.published_at, a.fetched_at,
			a.excerpt, a.category, a.tags, a.is_read, a.is_favorite, a.created_at, a.updated_at
		FROM articles a
		INNER JOIN sources s ON s.id = a.source_id
		WHERE %s
		ORDER BY %s %s
		LIMIT ? OFFSET ?`, whereClause, sortColumn, order)

	listArgs := append(append([]any{}, args...), params.PageSize, offset)
	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return domain.ListArticlesResult{}, fmt.Errorf("list articles: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Article, 0)
	for rows.Next() {
		article, err := scanArticle(rows)
		if err != nil {
			return domain.ListArticlesResult{}, err
		}
		items = append(items, article)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(params.PageSize)))
	}

	return domain.ListArticlesResult{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *ArticleRepository) GetByID(ctx context.Context, id int64) (*domain.Article, error) {
	query := `
		SELECT
			a.id, a.title, a.url, a.source_id, s.name, a.published_at, a.fetched_at,
			a.excerpt, a.category, a.tags, a.is_read, a.is_favorite, a.created_at, a.updated_at
		FROM articles a
		INNER JOIN sources s ON s.id = a.source_id
		WHERE a.id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	article, err := scanArticle(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &article, nil
}

func (r *ArticleRepository) UpdateReadStatus(ctx context.Context, id int64, isRead bool) (*domain.Article, error) {
	result, err := r.db.ExecContext(ctx, "UPDATE articles SET is_read = ? WHERE id = ?", isRead, id)
	if err != nil {
		return nil, fmt.Errorf("update read status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, nil
	}

	return r.GetByID(ctx, id)
}

func (r *ArticleRepository) UpdateFavoriteStatus(ctx context.Context, id int64, isFavorite bool) (*domain.Article, error) {
	result, err := r.db.ExecContext(ctx, "UPDATE articles SET is_favorite = ? WHERE id = ?", isFavorite, id)
	if err != nil {
		return nil, fmt.Errorf("update favorite status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("favorite rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, nil
	}

	return r.GetByID(ctx, id)
}

func (r *ArticleRepository) BulkUpsert(ctx context.Context, sourceID int64, articles []domain.Article) (inserted int, duplicated int, err error) {
	query := `
		INSERT INTO articles
			(title, url, source_id, published_at, fetched_at, excerpt, category, tags, dedupe_key)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			title = VALUES(title),
			excerpt = VALUES(excerpt),
			category = VALUES(category),
			tags = VALUES(tags),
			updated_at = CURRENT_TIMESTAMP`

	for _, article := range articles {
		tagsJSON, err := json.Marshal(article.Tags)
		if err != nil {
			return inserted, duplicated, fmt.Errorf("marshal tags: %w", err)
		}

		result, execErr := r.db.ExecContext(
			ctx,
			query,
			article.Title,
			article.URL,
			sourceID,
			article.PublishedAt,
			article.FetchedAt,
			article.Excerpt,
			article.Category,
			tagsJSON,
			article.DedupeKey,
		)
		if execErr != nil {
			return inserted, duplicated, fmt.Errorf("insert article: %w", execErr)
		}

		affected, _ := result.RowsAffected()
		switch affected {
		case 1:
			inserted++
		default:
			duplicated++
		}
	}

	return inserted, duplicated, nil
}

func scanArticle(scanner interface {
	Scan(dest ...any) error
}) (domain.Article, error) {
	var article domain.Article
	var publishedAt sql.NullTime
	var tagsRaw []byte
	if err := scanner.Scan(
		&article.ID,
		&article.Title,
		&article.URL,
		&article.SourceID,
		&article.SourceName,
		&publishedAt,
		&article.FetchedAt,
		&article.Excerpt,
		&article.Category,
		&tagsRaw,
		&article.IsRead,
		&article.IsFavorite,
		&article.CreatedAt,
		&article.UpdatedAt,
	); err != nil {
		return domain.Article{}, err
	}

	if publishedAt.Valid {
		article.PublishedAt = &publishedAt.Time
	}
	if len(tagsRaw) > 0 {
		_ = json.Unmarshal(tagsRaw, &article.Tags)
	}
	if article.Tags == nil {
		article.Tags = []string{}
	}

	return article, nil
}

func RFC3339OrNow(raw string) time.Time {
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Now().UTC()
	}
	return parsed
}
