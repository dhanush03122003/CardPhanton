package user

import "time"

// User represents a user in the domain
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"` // "USER" or "ADMIN"
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
