/*
Contains the model/schema of tables
*/
package models

import (
	"errors"

	"gorm.io/gorm"
)

var ErrNoRecord = errors.New("MODELS: no matching recod found")

type Project struct {
	gorm.Model
	ID        uint `gorm:"primaryKey"`
	Name      string
	GitUrl    string
	Domain    string
	CreatedBy string
}
