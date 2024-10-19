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

func (projectModel *ProjectModel) CheckExistingProject(id int) (models.Project, error) {
	var project models.Project
	result := projectModel.DBConnectionPool.First(&project, id)

	if result.Error == gorm.ErrRecordNotFound {
		return project, models.ErrNoRecord
	} else if result.Error != nil {
		return project, result.Error
	}

	return project, nil
}
