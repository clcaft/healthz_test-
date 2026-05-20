package example

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"gl.eda1.ru/go/go-service-template/pkg/logger"
	"gl.eda1.ru/go/go-service-template/pkg/mysqlGorm"
)

type gormExampleRepository struct {
	l  logger.Interface
	db *mysqlGorm.Connector
}

func NewGormExampleRepository(l logger.Interface, db *mysqlGorm.Connector) ExampleRepository {
	return &gormExampleRepository{l, db}
}

func (r *gormExampleRepository) FindCounter(ctx context.Context, id int64) (*ExampleCounter, error) {
	db, err := r.db.GetSlave(ctx)
	if err != nil {
		return nil, err
	}

	var model ExampleCounter
	if err = db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// ничего не найдено
			return nil, nil
		}
		return nil, err
	}
	return &model, nil
}

func (r *gormExampleRepository) InsertCounter(ctx context.Context, counter *ExampleCounter) error {
	db, err := r.db.GetMaster(ctx)
	if err != nil {
		return err
	}
	return db.Create(counter).Error
}

func (r *gormExampleRepository) UpdateCounter(ctx context.Context, counter *ExampleCounter) error {
	db, err := r.db.GetMaster(ctx)
	if err != nil {
		return err
	}
	return db.Save(counter).Error
}
