package example

import (
	"context"

	"gl.eda1.ru/go/go-service-template/pkg/logger"
	"gl.eda1.ru/go/go-service-template/pkg/mysql"
)

type mysqlExampleRepository struct {
	l  logger.Interface
	db *mysql.Connector
}

func NewMysqlExampleRepository(l logger.Interface, db *mysql.Connector) ExampleRepository {
	return &mysqlExampleRepository{l, db}
}

func (r *mysqlExampleRepository) FindCounter(ctx context.Context, id int64) (*ExampleCounter, error) {
	sql := "SELECT id, value FROM example_counters WHERE id=?"
	return mysql.QueryRowSlave[ExampleCounter](r.db, ctx, sql, id)
}

func (r *mysqlExampleRepository) InsertCounter(ctx context.Context, counter *ExampleCounter) error {
	sql := "INSERT INTO example_counters(id, value) VALUES (?, ?)"
	return mysql.ExecuteQuery(r.db, ctx, sql, counter.ID, counter.Value)
}

func (r *mysqlExampleRepository) UpdateCounter(ctx context.Context, counter *ExampleCounter) error {
	sql := "UPDATE example_counters SET value=? WHERE id=?"
	return mysql.ExecuteQuery(r.db, ctx, sql, counter.Value, counter.ID)
}
