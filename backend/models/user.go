package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint
	Username  string
	FirstName string
	LastName  string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BeforeSave : hook before a user is saved
func (user *User) BeforeSave(tx *gorm.DB) (err error) {
	if user.Password != "" {
		bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil
		}

		user.Password = string(bytes)
	}

	return
}
