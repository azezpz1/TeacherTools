package models

// User represents an application user.
type User struct {
	Email        string `firestore:"email" json:"email"`
	PasswordHash string `firestore:"passwordHash" json:"-"`
}
