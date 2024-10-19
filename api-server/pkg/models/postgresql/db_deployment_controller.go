package postgresql

import (
	"gitlab.com/harisheoran/scale-mesh/api-server/pkg/models"
	"gorm.io/gorm"
)

type DeploymentController struct {
	DatabaseConnectionPool *gorm.DB
}

func (dc *DeploymentController) Insert(deployement models.Deployment) (int, error) {
	result := dc.DatabaseConnectionPool.Create(&deployement)

	if result.Error != nil {
		return 0, nil
	}

	return int(deployement.ID), nil
}
