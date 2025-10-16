package query

import (
	"strconv"

	"github.com/shopspring/decimal"
)

type ValueType string

const (
	ValueTypeString  ValueType = "string"
	ValueTypeNumber  ValueType = "number"
	ValueTypeBoolean ValueType = "boolean"
)

func StringToType(s string, valueType ValueType) (any, error) {
	switch valueType {
	case ValueTypeString:
		return s, nil
	case ValueTypeNumber:
		return parseNumber(s)
	case ValueTypeBoolean:
		return parseBoolean(s)
	default:
		return s, nil
	}
}

func StringsToType(ss []string, valueType ValueType) (any, error) {
	switch valueType {
	case ValueTypeString:
		return ss, nil
	case ValueTypeNumber:
		result := make([]decimal.Decimal, 0, len(ss))
		for _, s := range ss {
			v, err := parseNumber(s)
			if err != nil {
				return nil, err
			}

			result = append(result, v)
		}

		return result, nil
	case ValueTypeBoolean:
		result := make([]bool, 0, len(ss))
		for _, s := range ss {
			v, err := parseBoolean(s)
			if err != nil {
				return nil, err
			}

			result = append(result, v)
		}

		return result, nil
	default:
		return ss, nil
	}
}

func parseNumber(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(s)
}

func parseBoolean(s string) (bool, error) {
	return strconv.ParseBool(s)
}
