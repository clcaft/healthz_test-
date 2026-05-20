package amqprpc

import (
	"fmt"

	"gl.eda1.ru/go/go-service-template/internal/dto"
	"gl.eda1.ru/go/go-service-template/pkg/rabbitmq/rmq_rpc/server"
)

// NewRouter -.
func NewRouter() map[string]server.CallHandler {
	routes := make(map[string]server.CallHandler)
	{
		newExampleRoutes(routes)
	}

	return routes
}

func decodeData(cmd, cmdData interface{}) error {
	decoder, err := dto.CreateDefaultDecoder(&cmdData)
	if err != nil {
		return fmt.Errorf("error creating decoder: &%w", err)
	}

	ampqQueueCommand, ok := cmd.(dto.AmpqQueueCommand)
	if ok {
		err = decoder.Decode(ampqQueueCommand.Data)
	}

	if err != nil {
		return server.NewAckError(fmt.Errorf("can't convert data to %T: %w", cmdData, err))
	}

	return nil
}
