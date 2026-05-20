package usecase

import (
	"context"

	"gl.eda1.ru/go/go-service-template/internal/controller/http/v1/dto"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks/mocks.go -package=mocks

type (
	Example interface {
		PlusValue(ctx context.Context, id int64, request *dto.ExampleRequest) (*dto.ExampleResponse, error)
	}
)
