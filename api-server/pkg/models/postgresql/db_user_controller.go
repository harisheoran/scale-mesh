package postgresql

import (
	"errors"

	"gitlab.com/harisheoran/scale-mesh/api-server/pkg/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserDBController struct {
	DatabaseConnectionPool *gorm.DB
}

func (uc *UserDBController) Insert(user models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.HashedPassword), 12)
	user.HashedPassword = string(hashedPassword)

	if err != nil {
		return err
	}

	result := uc.DatabaseConnectionPool.Create(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// check user exists and if yes then return the user ID.
func (uc *UserDBController) Authenticate(email, password string) (int, error) {
	user := models.User{}
	result := uc.DatabaseConnectionPool.Where("Email= ?", email).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, models.ErrNoRecord
		}
		return 0, result.Error
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	return int(user.ID), nil
}

// fetch user details
func (uc *UserDBController) GetUser(id int) (*models.User, error) {

	return nil, nil
}
