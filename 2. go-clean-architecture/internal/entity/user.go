package entity

import "time"

type User struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Status    string    `json:"status" db:"status"`
	JoinDate  time.Time `json:"join_date" db:"join_date"`
	DeletedAt time.Time `json:"deleted_at" db:"deleted_at"`
}
