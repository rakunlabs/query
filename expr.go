package query

import (
	"fmt"
	"net/url"
	"strings"
)

type operatorCmpType string

const (
	// OperatorEmpty is the empty operator.
	OperatorEmpty operatorCmpType = ""
	// OperatorEq is the equality operator.
	OperatorEq operatorCmpType = "eq"
	// OperatorNe is the not equal operator.
	OperatorNe operatorCmpType = "ne"
	// OperatorGt is the greater than operator.
	OperatorGt operatorCmpType = "gt"
	// OperatorLt is the less than operator.
	OperatorLt operatorCmpType = "lt"
	// OperatorGte is the greater than or equal operator.
	OperatorGte operatorCmpType = "gte"
	// OperatorLte is the less than or equal operator.
	OperatorLte operatorCmpType = "lte"
	// OperatorLike is the like operator.
	OperatorLike operatorCmpType = "like"
	// OperatorILike is the case insensitive like operator.
	OperatorILike operatorCmpType = "ilike"
	// OperatorNLike is the not like operator.
	OperatorNLike operatorCmpType = "nlike"
	// OperatorNILike is the case insensitive not like operator.
	OperatorNILike operatorCmpType = "nilike"
	// OperatorIn is the in operator.
	OperatorIn operatorCmpType = "in"
	// OperatorNIn is the not in operator.
	OperatorNIn operatorCmpType = "nin"
	// OperatorIs is the is null operator.
	OperatorIs operatorCmpType = "is"
	// OperatorIsNot is the is not null operator.
	OperatorIsNot operatorCmpType = "not"
	// OperatorKV is the contains operator JSONB types.
	OperatorKV operatorCmpType = "kv"
)

type operatorLogicType string

const (
	// OperatorAnd is the AND operator.
	OperatorAnd operatorLogicType = "and"
	// OperatorOr is the OR operator.
	OperatorOr operatorLogicType = "or"
)

type Expression interface {
	Expression() Expression
}

type ExpressionCmp struct {
	Operator operatorCmpType
	Field    string
	Value    any
}

func (e ExpressionCmp) Expression() Expression {
	return e
}

type ExpressionLogic struct {
	Operator operatorLogicType
	List     []Expression
}

func (e ExpressionLogic) Expression() Expression {
	return e
}

type ExpressionSort struct {
	// Field is the field name to order by.
	Field string
	// Desc indicates whether the order is descending.
	Desc bool
}

func NewExpressionCmp(operator operatorCmpType, field string, value any) ExpressionCmp {
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

// ParseExpression parses a single expression from key-value pairs.
//   - key -> key[eq]
//   - eq, ne, gt, lt, gte, lte, like, ilike, nlike, nilike, in, nin, is, not, kv
func ParseExpression(key, valueRaw string) (ExpressionCmp, error) {
	value, err := url.QueryUnescape(valueRaw)
	if err != nil {
		return ExpressionCmp{}, err
	}

	field, operator, hasOperator := parseFieldWithOperator(key)

	if !hasOperator {
		if strings.Contains(value, ",") {
			return NewExpressionCmp(OperatorIn, field, strings.Split(value, ",")), nil
		}

		return NewExpressionCmp(OperatorEq, field, value), nil
	}

	return ParseExpressionWithOperator(operator, field, value)
}

func ParseExpressionWithOperator(operator string, key string, value string) (ExpressionCmp, error) {
	switch operatorCmpType(operator) {
	case OperatorEq:
		return NewExpressionCmp(OperatorEq, key, value), nil
	case OperatorNe:
		return NewExpressionCmp(OperatorNe, key, value), nil
	case OperatorGt:
		return NewExpressionCmp(OperatorGt, key, value), nil
	case OperatorLt:
		return NewExpressionCmp(OperatorLt, key, value), nil
	case OperatorGte:
		return NewExpressionCmp(OperatorGte, key, value), nil
	case OperatorLte:
		return NewExpressionCmp(OperatorLte, key, value), nil
	case OperatorLike:
		return NewExpressionCmp(OperatorLike, key, value), nil
	case OperatorILike:
		return NewExpressionCmp(OperatorILike, key, value), nil
	case OperatorNLike:
		return NewExpressionCmp(OperatorNLike, key, value), nil
	case OperatorNILike:
		return NewExpressionCmp(OperatorNILike, key, value), nil
	case OperatorIn, OperatorEmpty:
		if strings.Contains(value, ",") {
			return NewExpressionCmp(OperatorIn, key, strings.Split(value, ",")), nil
		}

		return NewExpressionCmp(OperatorIn, key, value), nil
	case OperatorNIn:
		if strings.Contains(value, ",") {
			return NewExpressionCmp(OperatorNIn, key, strings.Split(value, ",")), nil
		}

		return NewExpressionCmp(OperatorNIn, key, value), nil
	case OperatorIs:
		return NewExpressionCmp(OperatorIs, key, nil), nil
	case OperatorIsNot:
		return NewExpressionCmp(OperatorIsNot, key, nil), nil
	case OperatorKV:
		valueParts := strings.Split(value, ",")

		build := strings.Builder{}
		build.WriteString(`{`)
		for i := range valueParts {
			kv := strings.SplitN(valueParts[i], ":", 2)
			if len(kv) != 2 {
				return ExpressionCmp{}, fmt.Errorf("invalid kv format: [%s]", valueParts[i])
			}

			// Trim spaces
			kv[0] = strings.TrimSpace(kv[0])
			kv[1] = strings.TrimSpace(kv[1])

			build.WriteString(`"`)
			build.WriteString(kv[0])
			build.WriteString(`":"`)
			build.WriteString(kv[1])
			build.WriteString(`"`)

			if i < len(valueParts)-1 {
				build.WriteString(`,`)
			}
		}

		build.WriteString(`}`)

		return NewExpressionCmp(OperatorKV, key, build.String()), nil
	}

	return ExpressionCmp{}, fmt.Errorf("unsupported operator: [%s]", operator)
}
