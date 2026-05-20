package amqprpc

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"gl.eda1.ru/go/go-service-template/internal/dto"
	"gl.eda1.ru/go/go-service-template/internal/dto/cmd"
	"gl.eda1.ru/go/go-service-template/pkg/rabbitmq/rmq_rpc/server"
)

type exampleRoutes struct{}

func newExampleRoutes(routes map[string]server.CallHandler) {
	r := &exampleRoutes{}
	routes["kitchen"] = r.parseCommand()
}

func (r *exampleRoutes) parseCommand() server.CallHandler {
	return func(d *amqp.Delivery) (res interface{}, err error) {
		var c dto.AmpqQueueCommand
		err = json.Unmarshal(d.Body, &c)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - exampleRoutes - parseCommand - Unmarshal - msgBody = %s: %w", string(d.Body), err)
		}

		switch c.Command {
		case "ExampleCommand":
			var cmdData cmd.ExampleCmdData
			if err = decodeData(c, &cmdData); err != nil {
				return nil, fmt.Errorf("amqp_rpc - exampleRoutes - parseCommand - AddProductsChanges - decodeData: %w", err)
			}
			err = r.ExampleCommand(cmdData)
		default:
			return nil, fmt.Errorf("amqp_rpc - exampleRoutes - parseCommand: %s %q", dto.ErrUnknownCommand, c.Command)
		}

		return nil, err
	}
}

func (r *exampleRoutes) ExampleCommand(c cmd.ExampleCmdData) error {
	// res, err = r.s.Kitchen.AddProductsChanges(c.ReceiptID, c.Changes.ToEntity())
	// if err != nil {
	//	return nil, fmt.Errorf("amqp_rpc - exampleRoutes - CommandAddProductsChanges - r.uc.Kitchen.AddProductsChanges: %w", err)
	//}

	println(c.Id)
	return nil
}
