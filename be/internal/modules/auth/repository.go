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

// Kolom aman (tanpa password_hash).
const userCols = `id, name, email, role, is_active, created_at`

func scanUser(row interface {
	Scan(dest ...any) error
}) (*User, error) {
	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) List(ctx context.Context) ([]User, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+userCols+` FROM tb_user ORDER BY created_at`)
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

func (r *Repo) GetByID(ctx context.Context, id int) (*User, error) {
	return scanUser(r.pool.QueryRow(ctx, `SELECT `+userCols+` FROM tb_user WHERE id = $1`, id))
}

// GetCredentialsByEmail mengambil hash + status untuk proses login.
func (r *Repo) GetCredentialsByEmail(ctx context.Context, email string) (*Credentials, error) {
	var cr Credentials
	err := r.pool.QueryRow(ctx,
		`SELECT id, password_hash, role, is_active FROM tb_user WHERE email = $1`, email).
		Scan(&cr.ID, &cr.PasswordHash, &cr.Role, &cr.IsActive)
	if err != nil {
		return nil, err
	}
	return &cr, nil
}

func (r *Repo) Create(ctx context.Context, name, email, passwordHash, role string) (*User, error) {
	return scanUser(r.pool.QueryRow(ctx,
		`INSERT INTO tb_user (name, email, password_hash, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+userCols,
		name, email, passwordHash, role))
}

func (r *Repo) ToggleActive(ctx context.Context, id int, isActive bool) (*User, error) {
	return scanUser(r.pool.QueryRow(ctx,
		`UPDATE tb_user SET is_active = $1, modified_at = now()
		 WHERE id = $2
		 RETURNING `+userCols,
		isActive, id))
}

func (r *Repo) CountAll(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM tb_user`).Scan(&n)
	return n, err
}
