package usecase

import (
	"context"
	"errors"
	"fmt"

	"gl.eda1.ru/go/go-service-template/internal/controller/http/v1/dto"
	"gl.eda1.ru/go/go-service-template/internal/domain"
	"gl.eda1.ru/go/go-service-template/internal/domain/example"
	"gl.eda1.ru/go/go-service-template/pkg/logger"
)

type ExampleUseCase struct {
	l logger.Interface
	r *domain.Repositories
}

func NewExample(ctx context.Context, l logger.Interface, r *domain.Repositories) *ExampleUseCase {
	return &ExampleUseCase{
		l, r,
	}
}

func (uc *ExampleUseCase) PlusValue(ctx context.Context, id int64, request *dto.ExampleRequest) (*dto.ExampleResponse, error) {
	if id < 0 {
		return nil, errors.New("ID should be positive integer")
	}

	counter, err := uc.r.Example.FindCounter(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usecaces - PluesValue: %w", err)
	}

	if counter == nil {
		counter = &example.ExampleCounter{ID: id, Value: request.Value}
		err = uc.r.Example.InsertCounter(ctx, counter)
		if err != nil {
			return nil, fmt.Errorf("usecaces - PluesValue - Insert: %w", err)
		}
	} else {
		counter.Increment(request.Value)
		err = uc.r.Example.UpdateCounter(ctx, counter)
		if err != nil {
			return nil, fmt.Errorf("usecaces - PluesValue - Insert: %w", err)
		}
	}

	result := dto.ExampleResponse{
		NewValue: counter.Value,
	}
	return &result, nil
}
