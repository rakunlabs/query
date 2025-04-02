package query

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type Query struct {
	Select []any
	Where  []exp.Expression
	Order  []exp.OrderedExpression
	Offset *uint64
	Limit  *uint64
}

func (q *Query) GoquSelect(qq *goqu.SelectDataset) *goqu.SelectDataset {
	if q == nil {
		return qq
	}

	if len(q.Select) > 0 {
		qq = qq.Select(q.Select...)
	}

	if len(q.Where) > 0 {
		qq = qq.Where(q.Where...)
	}

	if len(q.Order) > 0 {
		qq = qq.Order(q.Order...)
	}

	if q.Offset != nil {
		qq = qq.Offset(uint(*q.Offset))
	}

	if q.Limit != nil {
		qq = qq.Limit(uint(*q.Limit))
	}

	return qq
}

// Parse parses a query string into a Query struct.
func Parse(query string) (*Query, error) {
	result := &Query{}

	if query == "" {
		return &Query{}, nil
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
			result.Where = append(result.Where, expr)
		}
	}

	return result, nil
}

// parseSort parses the sort parameter and returns the ordered expressions.
func parseSort(value string) []exp.OrderedExpression {
	if value == "" {
		return nil
	}

	fields := strings.Split(value, ",")
	orderedExpressions := make([]exp.OrderedExpression, 0, len(fields))

	for _, field := range fields {
		if strings.HasPrefix(field, "-") {
			// Descending order
			orderedExpressions = append(orderedExpressions, goqu.I(field[1:]).Desc())
		} else {
			// Ascending order
			orderedExpressions = append(orderedExpressions, goqu.I(field).Asc())
		}
	}

	return orderedExpressions
}

// parseFilter parses filter expressions from key-value pairs.
func parseFilter(key, value string) (exp.Expression, error) {
	switch {
	case strings.Contains(value, "|"):
		// Handle OR conditions with different fields
		parts := strings.Split(value, "|")
		exs := make([]exp.Expression, 0, len(parts))

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

		return goqu.Or(exs...), nil
	default:
		return parseExpression(key, value)
	}
}
