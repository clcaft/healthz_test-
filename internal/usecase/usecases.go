package usecase

import (
	"context"

	"gl.eda1.ru/go/go-service-template/internal/domain"
	"gl.eda1.ru/go/go-service-template/pkg/logger"
)

type UseCases struct {
	Example Example
}

func New(ctx context.Context, l logger.Interface, r *domain.Repositories) *UseCases {
	return &UseCases{
		Example: NewExample(ctx, l, r),
	}
}
