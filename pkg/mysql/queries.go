package mysql

import (
	"context"
	"database/sql"
	"errors"
)

func ExecuteQuery(c *Connector, ctx context.Context, query string, args ...interface{}) error {
	_, span := c.traces.Span(ctx, query)
	defer span.End()

	db, err := c.GetMaster(ctx)
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)
	return err
}

func QueryRowSlave[T any](c *Connector, ctx context.Context, query string, args ...interface{}) (*T, error) {
	_, span := c.traces.Span(ctx, query)
	defer span.End()

	db, err := c.GetSlave(ctx)
	if err != nil {
		return nil, err
	}
	var result T
	err = db.Get(&result, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func QueryRowsSlave[T any](c *Connector, ctx context.Context, query string, args ...interface{}) ([]T, error) {
	_, span := c.traces.Span(ctx, query)
	defer span.End()

	db, err := c.GetSlave(ctx)
	if err != nil {
		return nil, err
	}
	var result []T
	err = db.Select(&result, query, args...)
	if err != nil {
		return nil, err
	}
	return result, nil
}
