package models

import "time"

type (
	// User
	User struct {
		ID        int       `json:"id"`
		RoleID    int       `json:"role_id"`
		Username  string    `json:"username"`
		Password  string    `json:"password"`
		Active    int       `json:"active"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		DeletedAt time.Time `json:"deleted_at"`
		Role      Role
	}

	Role struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
)
