package ping

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo bertugas membaca/menulis data ping ke database.
type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) List(ctx context.Context) ([]Ping, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, message, created_at
		 FROM tb_ping
		 ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Ping{}
	for rows.Next() {
		var p Ping
		if err := rows.Scan(&p.ID, &p.Message, &p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (r *Repo) Create(ctx context.Context, message string) (*Ping, error) {
	var p Ping
	err := r.pool.QueryRow(ctx,
		`INSERT INTO tb_ping (message)
		 VALUES ($1)
		 RETURNING id, message, created_at`,
		message).Scan(&p.ID, &p.Message, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
