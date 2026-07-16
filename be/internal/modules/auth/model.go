package auth

import "time"

// User adalah data pengguna yang AMAN dikirim ke frontend
// (tanpa password_hash).
type User struct {
	ID        string    `json:"id"` // UUID
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Whatsapp  string    `json:"whatsapp"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Credentials dipakai HANYA di dalam proses login (berisi hash).
type Credentials struct {
	ID           string // UUID
	PasswordHash string
	Role         string
	IsActive     bool
}
