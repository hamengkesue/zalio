package auth

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// EnsureSeed membuat admin default kalau tabel user masih kosong,
// supaya kamu bisa login pertama kali. GANTI password-nya segera.
func EnsureSeed(ctx context.Context, pool *pgxpool.Pool) error {
	repo := NewRepo(pool)
	n, err := repo.CountAll(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := repo.Create(ctx, "Administrator", "admin@zalio.local", string(hash), "admin"); err != nil {
		return err
	}

	log.Println("┌──────────────────────────────────────────────────────┐")
	log.Println("│ Admin default dibuat:                                │")
	log.Println("│   email    : admin@zalio.local                       │")
	log.Println("│   password : admin123                                │")
	log.Println("│ GANTI password ini sebelum dipakai sungguhan!        │")
	log.Println("└──────────────────────────────────────────────────────┘")
	return nil
}
