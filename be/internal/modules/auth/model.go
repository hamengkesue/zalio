package auth

import "time"

// User adalah data pengguna yang AMAN dikirim ke frontend
// (tanpa password_hash).
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Credentials dipakai HANYA di dalam proses login (berisi hash).
type Credentials struct {
	ID           int
	PasswordHash string
	Role         string
	IsActive     bool
}
