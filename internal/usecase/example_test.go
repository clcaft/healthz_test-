package usecase

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/mock/gomock"

	"gl.eda1.ru/go/go-service-template/internal/controller/http/v1/dto"
	"gl.eda1.ru/go/go-service-template/internal/domain"
	"gl.eda1.ru/go/go-service-template/internal/domain/example"
	"gl.eda1.ru/go/go-service-template/internal/domain/mocks"
	"gl.eda1.ru/go/go-service-template/pkg/logger"
)

func TestCreateNewCounter(t *testing.T) {
	c := gomock.NewController(t)
	repo := mocks.NewMockExampleRepository(c)
	uc := ExampleUseCase{
		l: logger.NewNopLogger(),
		r: &domain.Repositories{
			Example: repo,
		},
	}

	repo.EXPECT().FindCounter(gomock.Any(), int64(1)).Return(nil, nil)
	repo.EXPECT().InsertCounter(gomock.Any(), &example.ExampleCounter{ID: 1, Value: 3}).Return(nil)

	model, err := uc.PlusValue(context.TODO(), 1, &dto.ExampleRequest{Value: 3})
	if err != nil {
		t.Error(err)
	}
	if model.NewValue != 3 {
		t.Error(fmt.Sprintf("model.NewValue expected 3 but got %d", model.NewValue))
	}
}

func TestUpdateCounter(t *testing.T) {
	c := gomock.NewController(t)
	repo := mocks.NewMockExampleRepository(c)
	uc := ExampleUseCase{
		l: logger.NewNopLogger(),
		r: &domain.Repositories{
			Example: repo,
		},
	}

	repo.EXPECT().FindCounter(gomock.Any(), int64(1)).Return(&example.ExampleCounter{ID: 1, Value: 2}, nil)
	repo.EXPECT().UpdateCounter(gomock.Any(), &example.ExampleCounter{ID: 1, Value: 5}).Return(nil)

	model, err := uc.PlusValue(context.TODO(), 1, &dto.ExampleRequest{Value: 3})
	if err != nil {
		t.Error(err)
	}
	if model.NewValue != 5 {
		t.Error(fmt.Sprintf("model.NewValue expected 3 but got %d", model.NewValue))
	}
}
