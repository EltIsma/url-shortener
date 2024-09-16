package domain

import "time"

type User struct {
	ID           string
	Nickname     string
	PasswordHash string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type Session struct {
	RefreshToken string
	ExpiresAt    time.Time
}
