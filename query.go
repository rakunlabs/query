package query

import (
	"fmt"
	"strconv"
	"strings"
)

type Query struct {
	Values map[string][]ExpressionCmp

	Select []string
	Where  []Expression
	Order  []ExpressionOrder
	Offset *uint64
	Limit  *uint64
}

func (q *Query) Has(v string) bool {
	if _, ok := q.Values[v]; ok {
		return true
	}

	return false
}

func (q *Query) HasAny(vList ...string) bool {
	for _, v := range vList {
		if _, ok := q.Values[v]; ok {
			return true
		}
	}

	return false
}

func (q *Query) GetOffset() uint64 {
	if q.Offset != nil {
		return *q.Offset
	}

	return 0
}

func (q *Query) GetLimit() uint64 {
	if q.Limit != nil {
		return *q.Limit
	}

	return 0
}

func ParseWithValidator(query string, validator *Validator, opts ...OptionQuery) (*Query, error) {
	q, err := Parse(query, opts...)
	if err != nil {
		return nil, err
	}

	if validator == nil {
		return q, nil
	}

	if err := q.Validate(validator); err != nil {
		return nil, err
	}

	return q, nil
}

// Parse parses a query string into a Query struct.
func Parse(query string, opts ...OptionQuery) (*Query, error) {
	o := &optionQuery{}
	for _, opt := range opts {
		opt(o)
	}

	result := &Query{
		Values: make(map[string][]ExpressionCmp),
	}

	// Split the query by & to get key-value pairs
	for pair := range strings.SplitSeq(query, "&") {
		// Handle each parameter
		if pair == "" {
			continue
		}

		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			kv = append(kv, "")
		}

		key := kv[0]
		value := kv[1]

		switch {
		case key == "fields":
			// Handle field selection
			if value == "" {
				continue
			}

			for _, field := range strings.Split(value, ",") {
				if field != "" {
					result.Select = append(result.Select, field)
				}
			}
		case key == "sort":
			// Handle sorting
			if value == "" {
				continue
			}
			result.Order = parseSort(value)
		case key == "limit":
			// Handle limit
			if value == "" {
				continue
			}
			limit, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid limit value: %s", value)
			}
			result.Limit = &limit
		case key == "offset":
			// Handle offset
			if value == "" {
				continue
			}
			offset, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid offset value: %s", value)
			}
			result.Offset = &offset
		default:
			// Handle filtering
			expr, err := parseFilter(key, value)
			if err != nil {
				return nil, err
			}

			if exprLogic, ok := expr.(ExpressionLogic); ok {
				for _, e := range exprLogic.List {
					if e == nil {
						continue
					}
					if cmp, ok := e.(ExpressionCmp); ok {
						result.Values[key] = append(result.Values[key], cmp)
					}
				}
			} else {
				if cmp, ok := expr.(ExpressionCmp); ok {
					result.Values[key] = append(result.Values[key], cmp)
				}
			}

			result.Where = append(result.Where, expr)
		}
	}

	if result.Offset == nil && o.DefaultOffset != nil {
		result.Offset = o.DefaultOffset
	}
	if result.Limit == nil && o.DefaultLimit != nil {
		result.Limit = o.DefaultLimit
	}

	for key, value := range o.Value {
		result.Values[key] = append(result.Values[key], value)
		result.Where = append(result.Where, value)
	}

	return result, nil
}

// parseSort parses the sort parameter and returns the ordered expressions.
func parseSort(value string) []ExpressionOrder {
	if value == "" {
		return nil
	}

	fields := strings.Split(value, ",")
	orderedExpressions := make([]ExpressionOrder, 0, len(fields))

	for _, field := range fields {
		switch {
		case field == "":
			// Skip empty fields
		case strings.HasPrefix(field, "+"):
			// Ascending order
			orderedExpressions = append(orderedExpressions, ExpressionOrder{
				Field: field[1:],
				Desc:  false,
			})
		case strings.HasPrefix(field, "-"):
			// Descending order
			orderedExpressions = append(orderedExpressions, ExpressionOrder{
				Field: field[1:],
				Desc:  true,
			})
		case strings.HasSuffix(field, ":asc"):
			// Ascending order
			orderedExpressions = append(orderedExpressions, ExpressionOrder{
				Field: field[:len(field)-4],
				Desc:  false,
			})
		case strings.HasSuffix(field, ":desc"):
			// Descending order
			orderedExpressions = append(orderedExpressions, ExpressionOrder{
				Field: field[:len(field)-5],
				Desc:  true,
			})
		default:
			// Ascending order
			orderedExpressions = append(orderedExpressions, ExpressionOrder{
				Field: field,
				Desc:  false,
			})
		}
	}

	return orderedExpressions
}

// parseFilter parses filter expressions from key-value pairs.
func parseFilter(key, value string) (Expression, error) {
	switch {
	case strings.Contains(value, "|"):
		// Handle OR conditions with different fields
		parts := strings.Split(value, "|")
		exs := make([]Expression, 0, len(parts))

		exp, err := parseExpression(key, parts[0])
		if err != nil {
			return nil, err
		}

		parts = parts[1:]

		exs = append(exs, exp)

		for _, part := range parts {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) != 2 {
				kv = append(kv, "")
			}

			exp, err := parseExpression(kv[0], kv[1])
			if err != nil {
				return nil, err
			}

			exs = append(exs, exp)
		}

		return ExpressionLogic{
			Operator: OperatorOr,
			List:     exs,
		}, nil
	default:
		return parseExpression(key, value)
	}
}
