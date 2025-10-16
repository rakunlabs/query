package query

import (
	"strconv"
)

type ValueType string

const (
	ValueTypeString  ValueType = "string"
	ValueTypeNumber  ValueType = "number"
	ValueTypeBoolean ValueType = "boolean"
)

func StringToType(s string, valueType ValueType) (any, error) {
	switch valueType {
	case ValueTypeString, ValueTypeNumber:
		return s, nil
	case ValueTypeBoolean:
		return parseBoolean(s)
	default:
		return s, nil
	}
}

func StringsToType(ss []string, valueType ValueType) (any, error) {
	switch valueType {
	case ValueTypeString, ValueTypeNumber:
		return ss, nil
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

func parseBoolean(s string) (bool, error) {
	return strconv.ParseBool(s)
}
