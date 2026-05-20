package dto

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pavand239/decimal"
)

const (
	ErrUnknownCommand = "unknown command"
	ErrEmptyCommand   = "empty command"
)

type AmpqQueueCommand struct {
	Command string      `json:"command"`
	Data    interface{} `json:"data"`
}

// Вложенные json объекты из rabbit по умолчанию конвертируются как map[string]interface{}
// конвертируем в нужную структуру с помощью пакета mapstructure
// некоторые типы вроде decimal.Decimal, time.Time не поддерживаются
// для них нужно описать кастомное преобразование данных.
var decoderHook mapstructure.DecodeHookFuncType = func(inType, outType reflect.Type, in interface{}) (interface{}, error) {
	if outType == reflect.TypeOf(decimal.Decimal{}) {
		switch inType.Kind() {
		case reflect.String:
			return decimal.NewFromString(in.(string))
		case reflect.Int, reflect.Uint, reflect.Int64, reflect.Uint64, reflect.Int32, reflect.Uint32:
			return decimal.NewFromInt(in.(int64)), nil
		case reflect.Float32, reflect.Float64:
			return decimal.NewFromFloat(in.(float64)), nil
		default:
			return nil, fmt.Errorf("unexpected typr for decimal.Decimal: %v,  inVal: %v", inType, in)
		}
	}

	if outType == reflect.TypeOf(time.Time{}) {
		if inType.Kind() == reflect.String {
			if inString, ok := in.(string); ok {
				return time.Parse(time.RFC3339, inString)
			}

			return nil, fmt.Errorf("error convert in to string for time.Time conversion: %v, inVal: %v", inType, in)
		}

		return nil, fmt.Errorf("unexpected type for time.Time: %v, inVal: %v", inType, in)
	}

	return in, nil
}

// CreateDefaultDecoder
// *mapstructure.Decoder в который запихиваем нужный decode hook
// result must be pointer to a struct.
func CreateDefaultDecoder(result interface{}) (*mapstructure.Decoder, error) {
	return mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: decoderHook,
		Result:     result,
	})
}
