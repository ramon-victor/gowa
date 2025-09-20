
package webhook

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/webhook"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/google/uuid"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) webhook.IWebhookRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) InitializeSchema() error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	
	// Defer rollback in case of error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Create the table with all columns for PostgreSQL
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
	
	// Create indexes for PostgreSQL
	indexQueries := []string{
		"CREATE INDEX IF NOT EXISTS idx_webhooks_enabled ON webhooks(enabled)",
		"CREATE INDEX IF NOT EXISTS idx_webhooks_created_at ON webhooks(created_at)",
	}
	
	for _, indexQuery := range indexQueries {
		_, err = tx.Exec(indexQuery)
		if err != nil {
			// If IF NOT EXISTS fails, try without it
			if strings.Contains(err.Error(), "syntax error") {
				simpleQuery := strings.Replace(indexQuery, "IF NOT EXISTS ", "", 1)
				_, err = tx.Exec(simpleQuery)
				if err != nil {
					// If index already exists, continue
					if strings.Contains(err.Error(), "already exists") {
						continue
					}
					return err
				}
			} else {
				return err
			}
		}
	}
	
	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return err
	}
	
	return nil
}

func (r *PostgresRepository) Create(wh *webhook.Webhook) error {
	if wh.ID == "" {
		wh.ID = uuid.New().String()
	}
	if wh.CreatedAt.IsZero() {
		wh.CreatedAt = time.Now()
	}
	wh.UpdatedAt = time.Now()

	eventsJSON, err := json.Marshal(wh.Events)
	if err != nil {
		return pkgError.InternalServerError("failed to marshal events: " + err.Error())
	}

	query := `
		INSERT INTO webhooks (id, url, secret, events, enabled, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = r.db.Exec(query, wh.ID, wh.URL, wh.Secret, eventsJSON, wh.Enabled, wh.Description, wh.CreatedAt, wh.UpdatedAt)
	if err != nil {
		return pkgError.InternalServerError("failed to create webhook: " + err.Error())
	}

	return nil
}

func (r *PostgresRepository) Update(wh *webhook.Webhook) error {
	wh.UpdatedAt = time.Now()

	eventsJSON, err := json.Marshal(wh.Events)
	if err != nil {
		return pkgError.InternalServerError("failed to marshal events: " + err.Error())
	}

	query := `
		UPDATE webhooks 
		SET url = $1, secret = $2, events = $3, enabled = $4, description = $5, updated_at = $6
		WHERE id = $7
	`
	_, err = r.db.Exec(query, wh.URL, wh.Secret, eventsJSON, wh.Enabled, wh.Description, wh.UpdatedAt, wh.ID)
	if err != nil {
		return pkgError.InternalServerError("failed to update webhook: " + err.Error())
	}

	return nil
}

func (r *PostgresRepository) Delete(id string) error {
	query := `DELETE FROM webhooks WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return pkgError.InternalServerError("failed to delete webhook: " + err.Error())
	}

	return nil
}

func (r *PostgresRepository) FindByID(id string) (*webhook.Webhook, error) {
	query := `SELECT id, url, secret, events, enabled, description, created_at, updated_at FROM webhooks WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var wh webhook.Webhook
	var eventsJSON string
	err := row.Scan(&wh.ID, &wh.URL, &wh.Secret, &eventsJSON, &wh.Enabled, &wh.Description, &wh.CreatedAt, &wh.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, pkgError.InternalServerError("failed to get webhook: " + err.Error())
	}

	err = json.Unmarshal([]byte(eventsJSON), &wh.Events)
	if err != nil {
		return nil, pkgError.InternalServerError("failed to unmarshal events: " + err.Error())
	}

	return &wh, nil
}

func (r *PostgresRepository) FindAll() ([]*webhook.Webhook, error) {
	query := `SELECT id, url, secret, events, enabled, description, created_at, updated_at FROM webhooks ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, pkgError.InternalServerError("failed to get webhooks: " + err.Error())
	}
	defer rows.Close()

	var webhooks []*webhook.Webhook
	for rows.Next() {
		var wh webhook.Webhook
		var eventsJSON string
		err := rows.Scan(&wh.ID, &wh.URL, &wh.Secret, &eventsJSON, &wh.Enabled, &wh.Description, &wh.CreatedAt, &wh.UpdatedAt)
		if err != nil {
			return nil, pkgError.InternalServerError("failed to scan webhook: " + err.Error())
		}

		err = json.Unmarshal([]byte(eventsJSON), &wh.Events)
		if err != nil {
			return nil, pkgError.InternalServerError("failed to unmarshal events: " + err.Error())
		}

		webhooks = append(webhooks, &wh)
	}

	if err = rows.Err(); err != nil {
		return nil, pkgError.InternalServerError("error iterating webhooks: " + err.Error())
	}

	return webhooks, nil
}

func (r *PostgresRepository) FindByEvent(event string) ([]*webhook.Webhook, error) {
	query := `SELECT id, url, secret, events, enabled, description, created_at, updated_at FROM webhooks WHERE enabled = true AND events LIKE $1`
	rows, err := r.db.Query(query, "%"+event+"%")
	if err != nil {
		return nil, pkgError.InternalServerError("failed to get webhooks by event: " + err.Error())
	}
	defer rows.Close()

	var webhooks []*webhook.Webhook
	for rows.Next() {
		var wh webhook.Webhook
		var eventsJSON string
		err := rows.Scan(&wh.ID, &wh.URL, &wh.Secret, &eventsJSON, &wh.Enabled, &wh.Description, &wh.CreatedAt, &wh.UpdatedAt)
		if err != nil {
			return nil, pkgError.InternalServerError("failed to scan webhook: " + err.Error())
		}

		err = json.Unmarshal([]byte(eventsJSON), &wh.Events)
		if err != nil {
			return nil, pkgError.InternalServerError("failed to unmarshal events: " + err.Error())
		}

		webhooks = append(webhooks, &wh)
	}

	if err = rows.Err(); err != nil {
		return nil, pkgError.InternalServerError("error iterating webhooks: " + err.Error())
	}

	return webhooks, nil
}

func (r *PostgresRepository) FindEnabled() ([]*webhook.Webhook, error) {
	query := `SELECT id, url, secret, events, enabled, description, created_at, updated_at FROM webhooks WHERE enabled = true ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, pkgError.InternalServerError("failed to get enabled webhooks: " + err.Error())
	}
	defer rows.Close()

	var webhooks []*webhook.Webhook
	for rows.Next() {
		var wh webhook.Webhook
		var eventsJSON string
		err := rows.Scan(&wh.ID, &wh.URL, &wh.Secret, &eventsJSON, &wh.Enabled, &wh.Description, &wh.CreatedAt, &wh.UpdatedAt)
		if err != nil {
			return nil, pkgError.InternalServerError("failed to scan webhook: " + err.Error())
		}

		err = json.Unmarshal([]byte(eventsJSON), &wh.Events)
		if err != nil {
			return nil, pkgError.InternalServerError("failed to unmarshal events: " + err.Error())
		}

		webhooks = append(webhooks, &wh)
	}

	if err = rows.Err(); err != nil {
		return nil, pkgError.InternalServerError("error iterating webhooks: " + err.Error())
	}

	return webhooks, nil
}
