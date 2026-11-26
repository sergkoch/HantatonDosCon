package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres persists scans and findings. It is used when DATABASE_URL is set.
type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, databaseURL string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) Create(ctx context.Context, scan Scan) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	statsJSON, err := json.Marshal(scan.Stats)
	if err != nil {
		return err
	}

	var changesJSON []byte
	if scan.Changes != nil {
		changesJSON, err = json.Marshal(scan.Changes)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO scans (
			id, owner_id, url, depth, status, lang, version, previous_id, score,
			pages_count, checked_links, unverified_links, error,
			stats, changes, created_at, finished_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		scan.ID, scan.Owner, scan.URL, scan.Depth, string(scan.Status), scan.Lang, scan.Version,
		nullableUUID(scan.PreviousID), scan.Score, scan.Pages,
		scan.CheckedLinks, scan.UnverifiedLinks, scan.Error,
		statsJSON, changesJSON, scan.CreatedAt, finishedAt(scan),
	)
	if err != nil {
		return err
	}

	if err := insertFindings(ctx, tx, scan.ID, scan.Findings, scan.CreatedAt); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (p *Postgres) Get(ctx context.Context, id string) (Scan, error) {
	scan, err := p.scanBy(ctx, `WHERE id = $1`, id)
	if err != nil {
		return Scan{}, err
	}

	findings, err := p.findingsBy(ctx, id)
	if err != nil {
		return Scan{}, err
	}
	scan.Findings = findings

	return scan, nil
}

func (p *Postgres) Update(ctx context.Context, scan Scan) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	statsJSON, err := json.Marshal(scan.Stats)
	if err != nil {
		return err
	}

	var changesJSON []byte
	if scan.Changes != nil {
		changesJSON, err = json.Marshal(scan.Changes)
		if err != nil {
			return err
		}
	}

	tag, err := tx.Exec(ctx, `
		UPDATE scans SET
			status = $2, version = $3, previous_id = $4, score = $5,
			pages_count = $6, checked_links = $7, unverified_links = $8,
			error = $9, stats = $10, changes = $11, finished_at = $12
		WHERE id = $1`,
		scan.ID, string(scan.Status), scan.Version, nullableUUID(scan.PreviousID),
		scan.Score, scan.Pages, scan.CheckedLinks, scan.UnverifiedLinks,
		scan.Error, statsJSON, changesJSON, finishedAt(scan),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	if _, err := tx.Exec(ctx, `DELETE FROM findings WHERE scan_id = $1`, scan.ID); err != nil {
		return err
	}
	if err := insertFindings(ctx, tx, scan.ID, scan.Findings, scan.CreatedAt); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (p *Postgres) List(ctx context.Context, targetURL, owner string, limit int) ([]Scan, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := p.pool.Query(ctx, `
		SELECT id::text, owner_id, url, depth, status, lang, version, COALESCE(previous_id::text, ''),
		       score, pages_count, checked_links, unverified_links, error,
		       stats, changes, created_at
		FROM scans
		WHERE ($1 = '' OR url = $1) AND ($2 = '' OR owner_id = $2)
		ORDER BY created_at DESC
		LIMIT $3`, targetURL, owner, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scans := make([]Scan, 0, limit)
	for rows.Next() {
		scan, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		scan.Findings = []Finding{}
		scans = append(scans, scan)
	}

	return scans, rows.Err()
}

func (p *Postgres) LatestDone(ctx context.Context, targetURL, owner string) (Scan, error) {
	scan, err := p.scanBy(ctx, `WHERE url = $1 AND status = 'done' AND owner_id = $2 ORDER BY created_at DESC LIMIT 1`, targetURL, owner)
	if err != nil {
		return Scan{}, err
	}

	findings, err := p.findingsBy(ctx, scan.ID)
	if err != nil {
		return Scan{}, err
	}
	scan.Findings = findings

	return scan, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (p *Postgres) scanBy(ctx context.Context, where string, args ...any) (Scan, error) {
	row := p.pool.QueryRow(ctx, `
		SELECT id::text, owner_id, url, depth, status, lang, version, COALESCE(previous_id::text, ''),
		       score, pages_count, checked_links, unverified_links, error,
		       stats, changes, created_at
		FROM scans `+where, args...)

	scan, err := scanRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Scan{}, ErrNotFound
	}
	return scan, err
}

func scanRow(row rowScanner) (Scan, error) {
	var (
		scan        Scan
		status      string
		previousID  string
		statsJSON   []byte
		changesJSON []byte
	)

	err := row.Scan(
		&scan.ID, &scan.Owner, &scan.URL, &scan.Depth, &status, &scan.Lang, &scan.Version, &previousID,
		&scan.Score, &scan.Pages, &scan.CheckedLinks, &scan.UnverifiedLinks,
		&scan.Error, &statsJSON, &changesJSON, &scan.CreatedAt,
	)
	if err != nil {
		return Scan{}, err
	}

	scan.Status = Status(status)
	scan.PreviousID = previousID
	if len(statsJSON) > 0 {
		if err := json.Unmarshal(statsJSON, &scan.Stats); err != nil {
			return Scan{}, err
		}
	}
	if len(changesJSON) > 0 {
		if err := json.Unmarshal(changesJSON, &scan.Changes); err != nil {
			return Scan{}, err
		}
	}

	return scan, nil
}

func (p *Postgres) findingsBy(ctx context.Context, scanID string) ([]Finding, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id::text, code, kind, severity, title, element, recommendation, wcag,
		       title_alt, recommendation_alt
		FROM findings
		WHERE scan_id = $1
		ORDER BY created_at, id`, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	findings := make([]Finding, 0)
	for rows.Next() {
		var finding Finding
		var severity string
		if err := rows.Scan(
			&finding.ID, &finding.Code, &finding.Kind, &severity,
			&finding.Title, &finding.Element, &finding.Recommendation, &finding.WCAG,
			&finding.TitleAlt, &finding.RecommendationAlt,
		); err != nil {
			return nil, err
		}
		finding.Severity = Severity(severity)
		findings = append(findings, finding)
	}

	return findings, rows.Err()
}

func insertFindings(ctx context.Context, tx pgx.Tx, scanID string, findings []Finding, createdAt time.Time) error {
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	for _, finding := range findings {
		_, err := tx.Exec(ctx, `
			INSERT INTO findings (
				id, scan_id, kind, severity, code, title, element,
				recommendation, wcag, created_at, title_alt, recommendation_alt
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			uuid.NewString(), scanID, finding.Kind, string(finding.Severity),
			finding.Code, finding.Title, finding.Element, finding.Recommendation,
			finding.WCAG, createdAt, finding.TitleAlt, finding.RecommendationAlt,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func nullableUUID(id string) any {
	if id == "" {
		return nil
	}
	return id
}

func finishedAt(scan Scan) any {
	if scan.Status == StatusDone || scan.Status == StatusFailed {
		return time.Now().UTC()
	}
	return nil
}
