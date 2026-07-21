package database

import (
	"context"
	"time"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresNotificationRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresNotificationRepository(pool *pgxpool.Pool) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{pool: pool}
}

func (r *PostgresNotificationRepository) FindByID(ctx context.Context, id int64) (*entities.Notification, error) {
	n := &entities.Notification{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, user_id, title, message, channel, status, metadata, is_read, created_at
		 FROM notifications WHERE id = $1`, id,
	).Scan(
		&n.ID, &n.TenantID, &n.UserID, &n.Title, &n.Message,
		&n.Channel, &n.Status, &n.Metadata, &n.IsRead, &n.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (r *PostgresNotificationRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.Notification, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, user_id, title, message, channel, status, metadata, is_read, created_at
		 FROM notifications WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*entities.Notification
	for rows.Next() {
		n := &entities.Notification{}
		err := rows.Scan(
			&n.ID, &n.TenantID, &n.UserID, &n.Title, &n.Message,
			&n.Channel, &n.Status, &n.Metadata, &n.IsRead, &n.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (r *PostgresNotificationRepository) FindByUserID(ctx context.Context, userID int64) ([]*entities.Notification, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, user_id, title, message, channel, status, metadata, is_read, created_at
		 FROM notifications WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*entities.Notification
	for rows.Next() {
		n := &entities.Notification{}
		err := rows.Scan(
			&n.ID, &n.TenantID, &n.UserID, &n.Title, &n.Message,
			&n.Channel, &n.Status, &n.Metadata, &n.IsRead, &n.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (r *PostgresNotificationRepository) Create(ctx context.Context, notification *entities.Notification) error {
	notification.CreatedAt = time.Now()

	return r.pool.QueryRow(ctx,
		`INSERT INTO notifications (tenant_id, user_id, title, message, channel, status, metadata, is_read, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id`,
		notification.TenantID, notification.UserID, notification.Title, notification.Message,
		notification.Channel, notification.Status, notification.Metadata,
		notification.IsRead, notification.CreatedAt,
	).Scan(&notification.ID)
}

func (r *PostgresNotificationRepository) Update(ctx context.Context, notification *entities.Notification) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notifications SET title = $1, message = $2, channel = $3, status = $4,
		        metadata = $5, is_read = $6
		 WHERE id = $7`,
		notification.Title, notification.Message, notification.Channel,
		notification.Status, notification.Metadata, notification.IsRead, notification.ID,
	)
	return err
}

func (r *PostgresNotificationRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM notifications WHERE id = $1`, id)
	return err
}

func (r *PostgresNotificationRepository) MarkAsRead(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notifications SET is_read = TRUE WHERE id = $1`, id,
	)
	return err
}

func (r *PostgresNotificationRepository) MarkAllAsRead(ctx context.Context, userID int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`, userID,
	)
	return err
}

func (r *PostgresNotificationRepository) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`, userID,
	).Scan(&count)
	return count, err
}

type PostgresNotificationPreferenceRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresNotificationPreferenceRepository(pool *pgxpool.Pool) *PostgresNotificationPreferenceRepository {
	return &PostgresNotificationPreferenceRepository{pool: pool}
}

func (r *PostgresNotificationPreferenceRepository) FindByID(ctx context.Context, id int64) (*entities.NotificationPreference, error) {
	pref := &entities.NotificationPreference{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, user_id, email_enabled, push_enabled, webhook_enabled,
		        webhook_url, email_address, created_at, updated_at
		 FROM notification_preferences WHERE id = $1`, id,
	).Scan(
		&pref.ID, &pref.TenantID, &pref.UserID, &pref.EmailEnabled,
		&pref.PushEnabled, &pref.WebhookEnabled, &pref.WebhookURL,
		&pref.EmailAddress, &pref.CreatedAt, &pref.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return pref, nil
}

func (r *PostgresNotificationPreferenceRepository) FindByTenantAndUser(ctx context.Context, tenantID, userID int64) (*entities.NotificationPreference, error) {
	pref := &entities.NotificationPreference{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, user_id, email_enabled, push_enabled, webhook_enabled,
		        webhook_url, email_address, created_at, updated_at
		 FROM notification_preferences WHERE tenant_id = $1 AND user_id = $2`, tenantID, userID,
	).Scan(
		&pref.ID, &pref.TenantID, &pref.UserID, &pref.EmailEnabled,
		&pref.PushEnabled, &pref.WebhookEnabled, &pref.WebhookURL,
		&pref.EmailAddress, &pref.CreatedAt, &pref.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return pref, nil
}

func (r *PostgresNotificationPreferenceRepository) Create(ctx context.Context, pref *entities.NotificationPreference) error {
	now := time.Now()
	pref.CreatedAt = now
	pref.UpdatedAt = now

	return r.pool.QueryRow(ctx,
		`INSERT INTO notification_preferences (tenant_id, user_id, email_enabled, push_enabled,
		        webhook_enabled, webhook_url, email_address, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id`,
		pref.TenantID, pref.UserID, pref.EmailEnabled, pref.PushEnabled,
		pref.WebhookEnabled, pref.WebhookURL, pref.EmailAddress,
		pref.CreatedAt, pref.UpdatedAt,
	).Scan(&pref.ID)
}

func (r *PostgresNotificationPreferenceRepository) Update(ctx context.Context, pref *entities.NotificationPreference) error {
	pref.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx,
		`UPDATE notification_preferences SET email_enabled = $1, push_enabled = $2,
		        webhook_enabled = $3, webhook_url = $4, email_address = $5, updated_at = $6
		 WHERE id = $7`,
		pref.EmailEnabled, pref.PushEnabled, pref.WebhookEnabled,
		pref.WebhookURL, pref.EmailAddress, pref.UpdatedAt, pref.ID,
	)
	return err
}

func (r *PostgresNotificationPreferenceRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM notification_preferences WHERE id = $1`, id)
	return err
}

type PostgresNotificationTemplateRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresNotificationTemplateRepository(pool *pgxpool.Pool) *PostgresNotificationTemplateRepository {
	return &PostgresNotificationTemplateRepository{pool: pool}
}

func (r *PostgresNotificationTemplateRepository) FindByID(ctx context.Context, id int64) (*entities.NotificationTemplate, error) {
	tmpl := &entities.NotificationTemplate{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, channel, subject_template, body_template, is_system, created_at, updated_at
		 FROM notification_templates WHERE id = $1`, id,
	).Scan(
		&tmpl.ID, &tmpl.TenantID, &tmpl.Name, &tmpl.Channel,
		&tmpl.SubjectTemplate, &tmpl.BodyTemplate, &tmpl.IsSystem,
		&tmpl.CreatedAt, &tmpl.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (r *PostgresNotificationTemplateRepository) FindByName(ctx context.Context, name string) (*entities.NotificationTemplate, error) {
	tmpl := &entities.NotificationTemplate{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, channel, subject_template, body_template, is_system, created_at, updated_at
		 FROM notification_templates WHERE name = $1`, name,
	).Scan(
		&tmpl.ID, &tmpl.TenantID, &tmpl.Name, &tmpl.Channel,
		&tmpl.SubjectTemplate, &tmpl.BodyTemplate, &tmpl.IsSystem,
		&tmpl.CreatedAt, &tmpl.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (r *PostgresNotificationTemplateRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]*entities.NotificationTemplate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, channel, subject_template, body_template, is_system, created_at, updated_at
		 FROM notification_templates WHERE tenant_id = $1 OR tenant_id IS NULL ORDER BY name ASC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []*entities.NotificationTemplate
	for rows.Next() {
		tmpl := &entities.NotificationTemplate{}
		err := rows.Scan(
			&tmpl.ID, &tmpl.TenantID, &tmpl.Name, &tmpl.Channel,
			&tmpl.SubjectTemplate, &tmpl.BodyTemplate, &tmpl.IsSystem,
			&tmpl.CreatedAt, &tmpl.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		templates = append(templates, tmpl)
	}
	return templates, nil
}

func (r *PostgresNotificationTemplateRepository) Create(ctx context.Context, tmpl *entities.NotificationTemplate) error {
	now := time.Now()
	tmpl.CreatedAt = now
	tmpl.UpdatedAt = now

	return r.pool.QueryRow(ctx,
		`INSERT INTO notification_templates (tenant_id, name, channel, subject_template, body_template, is_system, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id`,
		tmpl.TenantID, tmpl.Name, tmpl.Channel, tmpl.SubjectTemplate,
		tmpl.BodyTemplate, tmpl.IsSystem, tmpl.CreatedAt, tmpl.UpdatedAt,
	).Scan(&tmpl.ID)
}

func (r *PostgresNotificationTemplateRepository) Update(ctx context.Context, tmpl *entities.NotificationTemplate) error {
	tmpl.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx,
		`UPDATE notification_templates SET name = $1, channel = $2, subject_template = $3,
		        body_template = $4, is_system = $5, updated_at = $6
		 WHERE id = $7`,
		tmpl.Name, tmpl.Channel, tmpl.SubjectTemplate,
		tmpl.BodyTemplate, tmpl.IsSystem, tmpl.UpdatedAt, tmpl.ID,
	)
	return err
}

func (r *PostgresNotificationTemplateRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM notification_templates WHERE id = $1`, id)
	return err
}
