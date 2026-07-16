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

// ascOrDesc: true → "DESC", false → "ASC".
func ascOrDesc(desc bool) string {
	if desc {
		return "DESC"
	}
	return "ASC"
}

// Kolom aman (tanpa password_hash). id di-cast ke text (UUID),
// full_name -> Name, whatsapp_number -> Whatsapp.
const userCols = `id::text, full_name, username, email, COALESCE(whatsapp_number, '') AS whatsapp, COALESCE(profile_image, '') AS profile_image, COALESCE(group_access, '') AS group_access, role, is_active, created_at`

func scanUser(row interface {
	Scan(dest ...any) error
}) (*User, error) {
	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.Username, &u.Email, &u.Whatsapp, &u.ProfileImage, &u.GroupAccess, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

// ListParams: parameter query daftar user (pagination + search + sort + filter).
type ListParams struct {
	Limit  int
	Offset int
	Search string
	Sort   string // "name"|"username"|"role"|"" (default = created_at terbaru)
	Desc   bool
	Role   string // ""=semua, "admin", "staff"
	Status string // ""=semua, "active", "inactive"
}

// ListPaged mengambil user dengan pagination server-side + search + sort + filter.
// Mengembalikan baris untuk (limit, offset) DAN total baris yang cocok (untuk pager).
func (r *Repo) ListPaged(ctx context.Context, p ListParams) ([]User, int, error) {
	// Filter (semua parameter pakai trik "$n = '' OR ..." supaya posisi param tetap):
	//   $1 = search (full_name/username/email), $2 = role, $3 = status.
	const filter = ` WHERE ($1 = '' OR full_name ILIKE '%'||$1||'%' OR username ILIKE '%'||$1||'%' OR email ILIKE '%'||$1||'%')` +
		` AND ($2 = '' OR role = $2)` +
		` AND ($3 = '' OR ($3 = 'active' AND is_active) OR ($3 = 'inactive' AND NOT is_active))`

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM m_internal_user`+filter, p.Search, p.Role, p.Status).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Kolom sort di-whitelist (JANGAN dari input mentah → cegah SQL injection).
	// Default (tanpa sort) = created_at terbaru dulu (DESC).
	orderCol := "created_at"
	dir := "DESC"
	switch p.Sort {
	case "name":
		orderCol = "full_name"
		dir = ascOrDesc(p.Desc)
	case "username":
		orderCol = "username"
		dir = ascOrDesc(p.Desc)
	case "role":
		orderCol = "role"
		dir = ascOrDesc(p.Desc)
	}

	// Sort sekunder pakai id supaya urutan stabil antar-halaman (hindari baris
	// dobel/terlewat saat nilai sort seri).
	rows, err := r.pool.Query(ctx,
		`SELECT `+userCols+` FROM m_internal_user`+filter+
			` ORDER BY `+orderCol+` `+dir+`, id LIMIT $4 OFFSET $5`,
		p.Search, p.Role, p.Status, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *u)
	}
	return list, total, rows.Err()
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

func (r *Repo) Create(ctx context.Context, name, username, email, whatsapp, profileImage, groupAccess, passwordHash, role string) (*User, error) {
	return scanUser(r.pool.QueryRow(ctx,
		`INSERT INTO m_internal_user (full_name, username, email, whatsapp_number, profile_image, group_access, password_hash, role)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING `+userCols,
		name, username, email, whatsapp, profileImage, groupAccess, passwordHash, role))
}

// Update mengubah data user (username tidak diubah). modified_at diurus trigger.
func (r *Repo) Update(ctx context.Context, id, name, email, whatsapp, profileImage, groupAccess, role string) (*User, error) {
	return scanUser(r.pool.QueryRow(ctx,
		`UPDATE m_internal_user
		 SET full_name = $1, email = $2, whatsapp_number = $3, profile_image = $4, group_access = $5, role = $6
		 WHERE id = $7::uuid
		 RETURNING `+userCols,
		name, email, whatsapp, profileImage, groupAccess, role, id))
}

func (r *Repo) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE m_internal_user SET password_hash = $1 WHERE id = $2::uuid`, passwordHash, id)
	return err
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
