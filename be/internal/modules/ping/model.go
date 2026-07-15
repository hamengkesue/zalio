package ping

import "time"

// Ping adalah entitas contoh untuk membuktikan pola vertical-slice.
// Modul asli (product, inventory, dst.) akan meniru bentuk file ini.
type Ping struct {
	ID        int       `json:"id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
