package query

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

type operatorCmpType string

const (
	// OperatorEmpty is the empty operator which is equal to OperatorIn.
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
	// OperatorKV is the contains operator JSON types.
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
	String() string
}

type ExpressionCmp struct {
	Operator operatorCmpType
	Field    string
	Value    any
}

func (e *ExpressionCmp) Expression() Expression {
	return e
}

func (e ExpressionCmp) String() string {
	key := e.Field
	if e.Operator != OperatorEq {
		key += "[" + string(e.Operator) + "]"
	}

	val := formatValue(e.Operator, e.Value)

	return key + "=" + val
}

type ExpressionLogic struct {
	Operator operatorLogicType
	List     []Expression
}

func (e *ExpressionLogic) Expression() Expression {
	return e
}

func (e ExpressionLogic) String() string {
	if e.Operator == OperatorOr {
		// Check if all are ExpressionCmp with same field
		if len(e.List) > 0 {
			if cmp, ok := e.List[0].(*ExpressionCmp); ok {
				field := cmp.Field
				values := make([]string, 0, len(e.List))
				allSame := true
				for _, sub := range e.List {
					if c, ok := sub.(*ExpressionCmp); ok && c.Field == field && c.Operator == OperatorEq {
						values = append(values, formatValue(c.Operator, c.Value))
					} else {
						allSame = false
						break
					}
				}
				if allSame && len(values) > 1 {
					return field + "=(" + strings.Join(values, "|") + ")"
				}
			}
		}
	}
	// Default
	sep := "&"
	if e.Operator == OperatorOr {
		sep = "|"
	}
	parts := make([]string, len(e.List))
	for i, sub := range e.List {
		parts[i] = sub.String()
	}
	joined := strings.Join(parts, sep)
	if e.Operator == OperatorOr || e.Operator == OperatorAnd {
		joined = "(" + joined + ")"
	}

	return joined
}

func formatValue(operator operatorCmpType, v any) string {
	if operator == OperatorKV {
		vStr, _ := v.(string)

		return base64.RawURLEncoding.EncodeToString([]byte(vStr))
	}

	if s, ok := v.(string); ok {
		return url.QueryEscape(s)
	}

	if ss, ok := v.([]string); ok {
		escaped := make([]string, len(ss))
		for i, s := range ss {
			escaped[i] = url.QueryEscape(s)
		}

		return strings.Join(escaped, ",")
	}

	return url.QueryEscape(fmt.Sprintf("%v", v))
}

type ExpressionSort struct {
	// Field is the field name to order by.
	Field string
	// Desc indicates whether the order is descending.
	Desc bool
}

// NewExpressionCmp creates a new ExpressionCmp.
func NewExpressionCmp(operator operatorCmpType, field string, value any) *ExpressionCmp {
	return &ExpressionCmp{
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

func getKey(input string) string {
	if openBracket := strings.Index(input, "["); openBracket != -1 {
		return input[:openBracket]
	}

	return input
}

// ParseExpression parses a single expression from key-value pairs.
//   - key -> key[eq]
//   - eq, ne, gt, lt, gte, lte, like, ilike, nlike, nilike, in, nin, is, not, kv
func ParseExpression(key, value string, valueType ValueType) (*ExpressionCmp, error) {
	field, operator, hasOperator := parseFieldWithOperator(key)

	if !hasOperator {
		if strings.Contains(value, ",") {
			v, err := StringsToType(strings.Split(value, ","), valueType)
			if err != nil {
				return nil, err
			}

			return NewExpressionCmp(OperatorIn, field, v), nil
		}

		v, err := StringToType(value, valueType)
		if err != nil {
			return nil, err
		}

		return NewExpressionCmp(OperatorEq, field, v), nil
	}

	return ParseExpressionWithOperator(operator, field, value, valueType)
}

func ParseExpressionWithOperator(operator string, key string, value string, valueType ValueType) (*ExpressionCmp, error) {
	switch operatorCmpType(operator) {
	case OperatorEq:
		v, err := StringToType(value, valueType)
		if err != nil {
			return nil, err
		}

		return NewExpressionCmp(OperatorEq, key, v), nil
	case OperatorNe:
		v, err := StringToType(value, valueType)
		if err != nil {
			return nil, err
		}

		return NewExpressionCmp(OperatorNe, key, v), nil
	case OperatorGt:
		v, err := StringToType(value, valueType)
		if err != nil {
			return nil, err
		}

		return NewExpressionCmp(OperatorGt, key, v), nil
	case OperatorLt:
		v, err := StringToType(value, valueType)
		if err != nil {
			return nil, err
		}

		return NewExpressionCmp(OperatorLt, key, v), nil
	case OperatorGte:
		v, err := StringToType(value, valueType)
		if err != nil {
			return nil, err
		}

		return NewExpressionCmp(OperatorGte, key, v), nil
	case OperatorLte:
		v, err := StringToType(value, valueType)
		if err != nil {
			return nil, err
		}

		return NewExpressionCmp(OperatorLte, key, v), nil
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
			v, err := StringsToType(strings.Split(value, ","), valueType)
			if err != nil {
				return nil, err
			}

			return NewExpressionCmp(OperatorIn, key, v), nil
		}

		v, err := StringToType(value, valueType)
		if err != nil {
			return nil, err
		}

		return NewExpressionCmp(OperatorIn, key, v), nil
	case OperatorNIn:
		if strings.Contains(value, ",") {
			v, err := StringsToType(strings.Split(value, ","), valueType)
			if err != nil {
				return nil, err
			}

			return NewExpressionCmp(OperatorNIn, key, v), nil
		}

		v, err := StringToType(value, valueType)
		if err != nil {
			return nil, err
		}

		return NewExpressionCmp(OperatorNIn, key, v), nil
	case OperatorIs:
		return NewExpressionCmp(OperatorIs, key, nil), nil
	case OperatorIsNot:
		return NewExpressionCmp(OperatorIsNot, key, nil), nil
	case OperatorKV:
		if !strings.Contains(value, ":") {
			// base64 URL encoded JSON
			valueDecoded, err := base64.RawURLEncoding.DecodeString(value)
			if err != nil {
				return nil, fmt.Errorf("invalid base64 encoding for kv operator: %v", err)
			}

			return NewExpressionCmp(OperatorKV, key, string(valueDecoded)), nil
		}

		valueParts := strings.Split(value, ",")

		build := strings.Builder{}
		build.WriteString(`{`)
		for i := range valueParts {
			kv := strings.SplitN(valueParts[i], ":", 2)
			if len(kv) != 2 {
				return nil, fmt.Errorf("invalid kv format: [%s]", valueParts[i])
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

	return nil, fmt.Errorf("unsupported operator: [%s]", operator)
}
