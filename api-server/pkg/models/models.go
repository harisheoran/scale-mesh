/*
Contains the model/schema of tables
*/
package models

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrNoRecord           = errors.New("MODELS: no matching record found")
	ErrInvalidCredentials = errors.New("MODELS: invalid credentials")
	ErrDuplicateEmails    = errors.New("MODELS: email already exists")
)

type User struct {
	gorm.Model
	ID             uint `gorm:"primaryKey"`
	Name           string
	Email          string `gorm:"unique"`
	HashedPassword string
}

type Project struct {
	gorm.Model
	ID        uint `gorm:"primaryKey"`
	Name      string
	GitUrl    string
	Domain    string
	CreatedBy string
}

type LoginUser struct {
	Email    string
	Password string
}
