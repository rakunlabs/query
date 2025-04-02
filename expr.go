package query

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// Parse a field name possibly containing an operator
func parseFieldWithOperator(input string) (field string, op string, hasOp bool) {
	openBracket := strings.Index(input, "[")
	closeBracket := strings.LastIndex(input, "]")

	if openBracket != -1 && closeBracket != -1 && closeBracket > openBracket {
		field = input[:openBracket]
		op = input[openBracket+1 : closeBracket]
		return field, op, true
	}

	return input, "", false
}

// parseExpression parses a single expression from key-value pairs.
//   - key -> key[eq]
//   - eq, ne, gt, lt, gte, lte, like, ilike, nlike, nilike, in, nin, is, not
func parseExpression(key, value string) (exp.Expression, error) {
	field, operator, hasOperator := parseFieldWithOperator(key)

	if !hasOperator {
		if strings.Contains(value, ",") {
			return goqu.C(field).In(strings.Split(value, ",")), nil
		}

		return exp.Ex{field: value}, nil
	}

	switch operator {
	case "eq":
		return goqu.C(field).Eq(value), nil
	case "ne":
		return goqu.C(field).Neq(value), nil
	case "gt":
		return goqu.C(field).Gt(value), nil
	case "lt":
		return goqu.C(field).Lt(value), nil
	case "gte":
		return goqu.C(field).Gte(value), nil
	case "lte":
		return goqu.C(field).Lte(value), nil
	case "like":
		return goqu.C(field).Like(value), nil
	case "ilike":
		return goqu.C(field).ILike(value), nil
	case "nlike":
		return goqu.C(field).NotLike(value), nil
	case "nilike":
		return goqu.C(field).NotILike(value), nil
	case "in":
		if strings.Contains(value, ",") {
			return goqu.C(field).In(strings.Split(value, ",")), nil
		}

		return goqu.C(field).In(value), nil
	case "nin":
		if strings.Contains(value, ",") {
			return goqu.C(field).NotIn(strings.Split(value, ",")), nil
		}

		return goqu.C(field).NotIn(value), nil
	case "is":
		return goqu.C(field).IsNull(), nil
	case "not":
		return goqu.C(field).IsNotNull(), nil
	}

	return nil, fmt.Errorf("unsupported operator: [%s]", operator)
}
