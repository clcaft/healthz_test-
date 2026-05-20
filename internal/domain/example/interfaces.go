package example

//go:generate mockgen -source=interfaces.go -destination=../mocks/example.go -package=mocks

import "context"

type ExampleRepository interface {
	FindCounter(ctx context.Context, id int64) (*ExampleCounter, error)
	InsertCounter(ctx context.Context, counter *ExampleCounter) error
	UpdateCounter(ctx context.Context, counter *ExampleCounter) error
}
