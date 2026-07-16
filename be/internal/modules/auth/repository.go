package auth

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// Kolom aman (tanpa password_hash). id di-cast ke text (UUID),
// full_name -> Name, whatsapp_number -> Whatsapp.
const userCols = `id::text, full_name, username, email, COALESCE(whatsapp_number, '') AS whatsapp, role, is_active, created_at`

func scanUser(row interface {
	Scan(dest ...any) error
}) (*User, error) {
	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.Username, &u.Email, &u.Whatsapp, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) List(ctx context.Context) ([]User, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+userCols+` FROM m_internal_user ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *u)
	}
	return list, rows.Err()
}

func (r *Repo) GetByID(ctx context.Context, id string) (*User, error) {
	return scanUser(r.pool.QueryRow(ctx, `SELECT `+userCols+` FROM m_internal_user WHERE id = $1::uuid`, id))
}

// GetCredentialsByUsername mengambil hash + status untuk proses login.
func (r *Repo) GetCredentialsByUsername(ctx context.Context, username string) (*Credentials, error) {
	var cr Credentials
	err := r.pool.QueryRow(ctx,
		`SELECT id::text, password_hash, role, is_active FROM m_internal_user WHERE username = $1`, username).
		Scan(&cr.ID, &cr.PasswordHash, &cr.Role, &cr.IsActive)
	if err != nil {
		return nil, err
	}
	return &cr, nil
}

func (r *Repo) Create(ctx context.Context, name, username, email, whatsapp, passwordHash, role string) (*User, error) {
	return scanUser(r.pool.QueryRow(ctx,
		`INSERT INTO m_internal_user (full_name, username, email, whatsapp_number, password_hash, role)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING `+userCols,
		name, username, email, whatsapp, passwordHash, role))
}

func (r *Repo) ToggleActive(ctx context.Context, id string, isActive bool) (*User, error) {
	// modified_at diurus oleh trigger update_modified_at_column.
	return scanUser(r.pool.QueryRow(ctx,
		`UPDATE m_internal_user SET is_active = $1
		 WHERE id = $2::uuid
		 RETURNING `+userCols,
		isActive, id))
}

func (r *Repo) CountAll(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM m_internal_user`).Scan(&n)
	return n, err
}
