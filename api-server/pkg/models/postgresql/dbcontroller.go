package postgresql

import (
	"gitlab.com/harisheoran/scale-mesh/api-server/pkg/models"
	"gorm.io/gorm"
)

type ProjectModel struct {
	DBConnectionPool *gorm.DB
}

func (projectModel *ProjectModel) Insert(project models.Project) (int, error) {
	result := projectModel.DBConnectionPool.Create(&project)

	if result.Error != nil {
		return 0, nil
	}

	return int(project.ID), nil
}

func (projectModel *ProjectModel) Get() (models.Project, error) {

	return models.Project{}, nil
}
