package models

import (
	"time"

	"gorm.io/gorm"
)

type PasswordReset struct {
	UUID           string         `gorm:"type:varchar(255);primary_key" json:"uuid"`
	Email          string         `json:"email" validate:"required,email"`
	Token          string         `gorm:"type:varchar(255);uniqueIndex" json:"-"`
	ExpirationTime time.Time      `json:"-"`
	Used           bool           `gorm:"default:false" json:"-"`
	CreatedAt      time.Time      `json:"-"`
	UpdatedAt      time.Time      `json:"-"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type Reset struct {
	Password        string `json:"password" validate:"required,min=8"`
	PasswordConfirm string `json:"password_confirm" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}