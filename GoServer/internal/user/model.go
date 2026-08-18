package user

import "time"

// User represents a user in the domain
type User struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    CreatedAt time.Time `json:"created_at"`
}