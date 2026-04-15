package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	UUID      string `gorm:"type:varchar(255);primary_key" json:"uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Fullname        string     `gorm:"not null" json:"fullname"`
	Email           string     `gorm:"unique; not null" json:"email"`
	Telephone       string     `gorm:"unique" json:"telephone"`

	Password        string     `json:"password"`
	PasswordConfirm string     `json:"password_confirm" gorm:"-"`
	Role            string     `json:"role"` // 'Directeur', 'Secretaire', 'Chef du bureau', 'Agent', 'SuperAdmin'
	Permission      string     `json:"permission"`
	Status          bool       `gorm:"default:false" json:"status"`
}

type UserResponse struct {
	UUID            string     `json:"uuid"`
	Fullname        string     `json:"fullname"`
	Email           string     `json:"email"`
	Telephone       string     `json:"telephone"`
 
	Password        string     `json:"password"`
	PasswordConfirm string     `json:"password_confirm"`
	Role            string     `json:"role"` // 'Directeur', 'Secretaire', 'Chef du bureau', 'Agent', 'SuperAdmin'
	Permission      string     `json:"permission"`
	Status          bool       `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Login struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

func (a *User) SetPassword(p string) {
	hp, _ := bcrypt.GenerateFromPassword([]byte(p), 14)
	a.Password = string(hp)
}

func (a *User) ComparePassword(p string) error {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(p))
	return err
}