/*
Contains the model/schema of tables
*/
package models

import (
	"errors"

	"gorm.io/gorm"
)

// ENUM for status of deployment
type Status int64

const (
	QUEUE Status = iota
	PROGRESS
	READY
	FAIL
)

// For custom errors
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
	Projects       []Project `gorm:"foreignKey:UserID"`
}

type Project struct {
	gorm.Model
	ID          uint `gorm:"primaryKey"`
	Name        string
	GitUrl      string
	Domain      string
	UserID      uint         // foreign key to User
	User        User         `gorm:"constraint:OnDelete:CASCADE;"`
	Deployments []Deployment `gorm:"foreignKey:ProjectID"`
}

type Deployment struct {
	gorm.Model
	ID        uint    `gorm:"primaryKey"`
	ProjectID uint    // foreign key to Project
	Project   Project `gorm:"constraint:OnDelete:CASCADE;"`
	// Status    Status
}

type LoginUser struct {
	Email    string
	Password string
}
