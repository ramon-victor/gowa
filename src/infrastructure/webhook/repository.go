package webhook

import (
	"database/sql"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	domainwebhook "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/webhook"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/google/uuid"
)

type Repository struct {
	db         *sql.DB
	isPostgres bool
}

func NewRepository(db *sql.DB, isPostgres bool) domainwebhook.IWebhookRepository {
	return &Repository{db: db, isPostgres: isPostgres}
}

func (r *Repository) InitializeSchema() error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `
		CREATE TABLE IF NOT EXISTS webhooks (
			id TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			secret TEXT,
			events TEXT NOT NULL DEFAULT '[]',
			enabled BOOLEAN DEFAULT TRUE,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err = tx.Exec(query)
	if err != nil {
		return err
	}

	indexQueries := []string{
		"CREATE INDEX IF NOT EXISTS idx_webhooks_enabled ON webhooks(enabled)",
		"CREATE INDEX IF NOT EXISTS idx_webhooks_created_at ON webhooks(created_at)",
	}

	for _, indexQuery := range indexQueries {
		_, err = tx.Exec(indexQuery)
		if err != nil {
			// Handle PostgreSQL vs SQLite syntax differences
			if strings.Contains(err.Error(), "syntax error") && r.isPostgres {
				simpleQuery := strings.Replace(indexQuery, "IF NOT EXISTS ", "", 1)
				_, err = tx.Exec(simpleQuery)
				if err != nil && !strings.Contains(err.Error(), "already exists") {
					return err
				}
			} else if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *Repository) Create(wh *domainwebhook.Webhook) error {
	if wh.ID == "" {
		wh.ID = uuid.New().String()
	}

	if err := validateWebhookURL(wh.URL); err != nil {
		return err
	}

	eventsJSON, err := json.Marshal(wh.Events)
	if err != nil {
		return pkgError.InternalServerError("failed to marshal events: " + err.Error())
	}

	now := time.Now()
	wh.CreatedAt = now
	wh.UpdatedAt = now

	query := `INSERT INTO webhooks (id, url, secret, events, enabled, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	if r.isPostgres {
		query = `INSERT INTO webhooks (id, url, secret, events, enabled, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	}

	_, err = r.db.Exec(query, wh.ID, wh.URL, wh.Secret, string(eventsJSON), wh.Enabled, wh.Description, wh.CreatedAt, wh.UpdatedAt)
	if err != nil {
		return pkgError.InternalServerError("failed to create webhook: " + err.Error())
	}

	return nil
}

func (r *Repository) Update(wh *domainwebhook.Webhook) error {
	if err := validateWebhookURL(wh.URL); err != nil {
		return err
	}

	eventsJSON, err := json.Marshal(wh.Events)
	if err != nil {
		return pkgError.InternalServerError("failed to marshal events: " + err.Error())
	}

	wh.UpdatedAt = time.Now()

	query := `UPDATE webhooks SET url = ?, secret = ?, events = ?, enabled = ?, description = ?, updated_at = ? WHERE id = ?`
	if r.isPostgres {
		query = `UPDATE webhooks SET url = $1, secret = $2, events = $3, enabled = $4, description = $5, updated_at = $6 WHERE id = $7`
	}

	_, err = r.db.Exec(query, wh.URL, wh.Secret, string(eventsJSON), wh.Enabled, wh.Description, wh.UpdatedAt, wh.ID)
	if err != nil {
		return pkgError.InternalServerError("failed to update webhook: " + err.Error())
	}

	return nil
}

func (r *Repository) Delete(id string) error {
	query := `DELETE FROM webhooks WHERE id = ?`
	if r.isPostgres {
		query = `DELETE FROM webhooks WHERE id = $1`
	}

	_, err := r.db.Exec(query, id)
	if err != nil {
		return pkgError.InternalServerError("failed to delete webhook: " + err.Error())
	}

	return nil
}

func (r *Repository) FindByID(id string) (*domainwebhook.Webhook, error) {
	query := `SELECT id, url, secret, events, enabled, description, created_at, updated_at FROM webhooks WHERE id = ?`
	if r.isPostgres {
		query = `SELECT id, url, secret, events, enabled, description, created_at, updated_at FROM webhooks WHERE id = $1`
	}

	row := r.db.QueryRow(query, id)
	return r.scanWebhook(row)
}

func (r *Repository) FindAll() ([]*domainwebhook.Webhook, error) {
	query := `SELECT id, url, secret, events, enabled, description, created_at, updated_at FROM webhooks ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, pkgError.InternalServerError("failed to get webhooks: " + err.Error())
	}
	defer rows.Close()

	var webhooks []*domainwebhook.Webhook
	for rows.Next() {
		wh, err := r.scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		webhooks = append(webhooks, wh)
	}

	if err = rows.Err(); err != nil {
		return nil, pkgError.InternalServerError("error iterating webhooks: " + err.Error())
	}

	return webhooks, nil
}

func (r *Repository) FindByEvent(event string) ([]*domainwebhook.Webhook, error) {
	query := `SELECT id, url, secret, events, enabled, description, created_at, updated_at FROM webhooks WHERE enabled = true`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, pkgError.InternalServerError("failed to get webhooks by event: " + err.Error())
	}
	defer rows.Close()

	var webhooks []*domainwebhook.Webhook
	for rows.Next() {
		wh, err := r.scanWebhook(rows)
		if err != nil {
			return nil, err
		}

		for _, e := range wh.Events {
			if e == event {
				webhooks = append(webhooks, wh)
				break
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, pkgError.InternalServerError("error iterating webhooks: " + err.Error())
	}

	return webhooks, nil
}

func (r *Repository) FindEnabled() ([]*domainwebhook.Webhook, error) {
	query := `SELECT id, url, secret, events, enabled, description, created_at, updated_at FROM webhooks WHERE enabled = true ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, pkgError.InternalServerError("failed to get enabled webhooks: " + err.Error())
	}
	defer rows.Close()

	var webhooks []*domainwebhook.Webhook
	for rows.Next() {
		wh, err := r.scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		webhooks = append(webhooks, wh)
	}

	if err = rows.Err(); err != nil {
		return nil, pkgError.InternalServerError("error iterating webhooks: " + err.Error())
	}

	return webhooks, nil
}

func (r *Repository) scanWebhook(scanner interface{ Scan(...any) error }) (*domainwebhook.Webhook, error) {
	var wh domainwebhook.Webhook
	var eventsJSON string
	var secret, description sql.NullString

	err := scanner.Scan(
		&wh.ID, &wh.URL, &secret, &eventsJSON, &wh.Enabled, &description, &wh.CreatedAt, &wh.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if secret.Valid {
		wh.Secret = secret.String
	}
	if description.Valid {
		wh.Description = description.String
	}

	if err := json.Unmarshal([]byte(eventsJSON), &wh.Events); err != nil {
		return nil, pkgError.InternalServerError("failed to unmarshal events: " + err.Error())
	}

	return &wh, nil
}

func validateWebhookURL(urlStr string) error {
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		return pkgError.ValidationError("URL cannot be empty")
	}

	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return pkgError.ValidationError("Invalid URL format: " + err.Error())
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return pkgError.ValidationError("URL scheme must be http or https")
	}

	if parsedURL.Host == "" {
		return pkgError.ValidationError("URL must contain a host")
	}

	return nil
}
