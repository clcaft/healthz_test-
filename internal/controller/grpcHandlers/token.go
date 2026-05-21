package grpcHandlers

import (
	"context"

	dto2 "gl.eda1.ru/go/go-service-template/internal/controller/http/v1/dto"
	proto "gl.eda1.ru/go/go-service-template/internal/proto/go"
	"gl.eda1.ru/go/go-service-template/internal/usecase"
	"gl.eda1.ru/go/go-service-template/pkg/logger"
)

type ExampleHandler struct {
	proto.UnimplementedExampleServer
	l  logger.Interface
	uc *usecase.UseCases
}

func NewExampleHandler(l logger.Interface, uc *usecase.UseCases) *ExampleHandler {
	return &ExampleHandler{
		l:  l,
		uc: uc,
	}
}

func (t ExampleHandler) Plus(ctx context.Context, request *proto.ExampleRequest) (*proto.ExampleResponse, error) {
	result, err := t.uc.Example.PlusValue(ctx, request.CounterId, &dto2.ExampleRequest{
		Value: request.Value,
	})
	if err != nil {
		return nil, err
	}

	return &proto.ExampleResponse{
		NewValue: result.NewValue,
	}, nil
}