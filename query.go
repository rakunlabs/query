package query

import (
	"fmt"
	"strconv"
	"strings"
)

type Query struct {
	Values map[string][]ExpressionCmp

	Select []any
	Where  []Expression
	Order  []ExpressionOrder
	Offset *uint64
	Limit  *uint64
}

// Parse parses a query string into a Query struct.
func Parse(query string) (*Query, error) {
	result := &Query{
		Values: make(map[string][]ExpressionCmp),
	}

	if query == "" {
		return result, nil
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
		if strings.HasPrefix(field, "-") {
			// Descending order
			orderedExpressions = append(orderedExpressions, ExpressionOrder{
				Field: field[1:],
				Desc:  true,
			})
			// orderedExpressions = append(orderedExpressions, goqu.I(field[1:]).Desc())
		} else {
			// Ascending order
			orderedExpressions = append(orderedExpressions, ExpressionOrder{
				Field: field,
				Desc:  false,
			})
			// orderedExpressions = append(orderedExpressions, goqu.I(field).Asc())
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
