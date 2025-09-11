package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Challenge    []byte    `json:"-"`
	ServerSecret []byte    `json:"-"`
	Token        string    `json:"token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *Session) error {
	query := `
		INSERT INTO sessions (user_id, challenge, server_secret, token, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		session.UserID,
		session.Challenge,
		session.ServerSecret,
		session.Token,
		session.ExpiresAt,
	).Scan(&session.ID, &session.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (*Session, error) {
	query := `
		SELECT id, user_id, challenge, server_secret, token, expires_at, created_at
		FROM sessions
		WHERE id = $1
	`

	var session Session
	err := r.db.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.Challenge,
		&session.ServerSecret,
		&session.Token,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &session, nil
}

func (r *SessionRepository) GetByToken(ctx context.Context, token string) (*Session, error) {
	query := `
		SELECT id, user_id, challenge, server_secret, token, expires_at, created_at
		FROM sessions
		WHERE token = $1 AND expires_at > NOW()
	`

	var session Session
	err := r.db.QueryRow(ctx, query, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Challenge,
		&session.ServerSecret,
		&session.Token,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &session, nil
}

func (r *SessionRepository) GetActiveByUserID(ctx context.Context, userID string) ([]*Session, error) {
	query := `
		SELECT id, user_id, challenge, server_secret, token, expires_at, created_at
		FROM sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.Challenge,
			&session.ServerSecret,
			&session.Token,
			&session.ExpiresAt,
			&session.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

func (r *SessionRepository) Update(ctx context.Context, session *Session) error {
	query := `
		UPDATE sessions
		SET challenge = $2, server_secret = $3, token = $4, expires_at = $5
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		session.ID,
		session.Challenge,
		session.ServerSecret,
		session.Token,
		session.ExpiresAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.Exec(ctx, query, userID)
	return err
}
