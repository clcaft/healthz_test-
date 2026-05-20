package domain

import (
	"gl.eda1.ru/go/go-service-template/internal/domain/example"
	"gl.eda1.ru/go/go-service-template/pkg/logger"
	"gl.eda1.ru/go/go-service-template/pkg/mysql"
)

type Repositories struct {
	Example example.ExampleRepository
}

func NewRepositories(l logger.Interface, db *mysql.Connector) *Repositories {
	return &Repositories{
		Example: example.NewMysqlExampleRepository(l, db),
	}
}
