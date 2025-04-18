package query

import (
	"fmt"
	"net/url"
	"strings"
)

type OperatorCmpType string

const (
	// OperatorEq is the equality operator.
	OperatorEq OperatorCmpType = "eq"
	// OperatorNe is the not equal operator.
	OperatorNe OperatorCmpType = "ne"
	// OperatorGt is the greater than operator.
	OperatorGt OperatorCmpType = "gt"
	// OperatorLt is the less than operator.
	OperatorLt OperatorCmpType = "lt"
	// OperatorGte is the greater than or equal operator.
	OperatorGte OperatorCmpType = "gte"
	// OperatorLte is the less than or equal operator.
	OperatorLte OperatorCmpType = "lte"
	// OperatorLike is the like operator.
	OperatorLike OperatorCmpType = "like"
	// OperatorILike is the case insensitive like operator.
	OperatorILike OperatorCmpType = "ilike"
	// OperatorNLike is the not like operator.
	OperatorNLike OperatorCmpType = "nlike"
	// OperatorNILike is the case insensitive not like operator.
	OperatorNILike OperatorCmpType = "nilike"
	// OperatorIn is the in operator.
	OperatorIn OperatorCmpType = "in"
	// OperatorNIn is the not in operator.
	OperatorNIn OperatorCmpType = "nin"
	// OperatorIs is the is null operator.
	OperatorIs OperatorCmpType = "is"
	// OperatorIsNot is the is not null operator.
	OperatorIsNot OperatorCmpType = "not"
)

type OperatorLogicType string

const (
	// OperatorAnd is the AND operator.
	OperatorAnd OperatorLogicType = "and"
	// OperatorOr is the OR operator.
	OperatorOr OperatorLogicType = "or"
)

type Expression interface {
	Expression() Expression
}

type ExpressionCmp struct {
	Operator OperatorCmpType
	Field    string
	Value    any
}

func (e ExpressionCmp) Expression() Expression {
	return e
}

type ExpressionLogic struct {
	Operator OperatorLogicType
	List     []Expression
}

func (e ExpressionLogic) Expression() Expression {
	return e
}

type ExpressionOrder struct {
	// Field is the field name to order by.
	Field string
	// Desc indicates whether the order is descending.
	Desc bool
}

func newExpressionCmp(operator OperatorCmpType, field string, value any) ExpressionCmp {
	return ExpressionCmp{
		Operator: operator,
		Field:    field,
		Value:    value,
	}
}

// parseFieldWithOperator a field name possibly containing an operator.
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
func parseExpression(key, valueRaw string) (ExpressionCmp, error) {
	value, err := url.QueryUnescape(valueRaw)
	if err != nil {
		return ExpressionCmp{}, err
	}

	field, operator, hasOperator := parseFieldWithOperator(key)

	if !hasOperator {
		if strings.Contains(value, ",") {
			return newExpressionCmp(OperatorIn, field, strings.Split(value, ",")), nil
		}

		return newExpressionCmp(OperatorEq, field, value), nil
	}

	switch OperatorCmpType(operator) {
	case OperatorEq:
		return newExpressionCmp(OperatorEq, field, value), nil
	case OperatorNe:
		return newExpressionCmp(OperatorNe, field, value), nil
	case OperatorGt:
		return newExpressionCmp(OperatorGt, field, value), nil
	case OperatorLt:
		return newExpressionCmp(OperatorLt, field, value), nil
	case OperatorGte:
		return newExpressionCmp(OperatorGte, field, value), nil
	case OperatorLte:
		return newExpressionCmp(OperatorLte, field, value), nil
	case OperatorLike:
		return newExpressionCmp(OperatorLike, field, value), nil
	case OperatorILike:
		return newExpressionCmp(OperatorILike, field, value), nil
	case OperatorNLike:
		return newExpressionCmp(OperatorNLike, field, value), nil
	case OperatorNILike:
		return newExpressionCmp(OperatorNILike, field, value), nil
	case OperatorIn:
		if strings.Contains(value, ",") {
			return newExpressionCmp(OperatorIn, field, strings.Split(value, ",")), nil
		}

		return newExpressionCmp(OperatorIn, field, value), nil
	case OperatorNIn:
		if strings.Contains(value, ",") {
			return newExpressionCmp(OperatorNIn, field, strings.Split(value, ",")), nil
		}

		return newExpressionCmp(OperatorNIn, field, value), nil
	case OperatorIs:
		return newExpressionCmp(OperatorIs, field, nil), nil
	case OperatorIsNot:
		return newExpressionCmp(OperatorIsNot, field, nil), nil
	}

	return ExpressionCmp{}, fmt.Errorf("unsupported operator: [%s]", operator)
}
