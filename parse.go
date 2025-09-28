package query

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	keyFields = "_fields"
	keySort   = "_sort"
	keyLimit  = "_limit"
	keyOffset = "_offset"
)

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
	o := &optionQuery{
		SkipUnderscore: true,
	}
	for _, opt := range opts {
		opt(o)
	}

	result := &Query{}

	var err error
	query, err = url.QueryUnescape(query)
	if err != nil {
		return nil, err
	}

	// Split the query by & to get key-value pairs
	for _, pair := range split(query, '&') {
		// Handle each parameter
		if pair == "" {
			continue
		}

		if isParenthesesAny(pair) {
			// Handle standalone parentheses expression
			exprs, err := parseFilter(pair)
			if err != nil {
				return nil, err
			}

			var expr Expression
			switch length := len(exprs); {
			case length > 1:
				expr = &ExpressionLogic{
					Operator: OperatorAnd,
					List:     exprs,
				}
			case length == 1:
				expr = exprs[0]
			default:
				continue
			}

			processed := resultAddExpr(result, expr, o)
			if processed != nil {
				result.Where = append(result.Where, processed)
			}

			continue
		}

		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			kv = append(kv, "")
		}

		key := kv[0]
		value := kv[1]

		switch key {
		case keyFields:
			// Handle field selection
			if value == "" {
				continue
			}

			for field := range strings.SplitSeq(value, ",") {
				if field != "" {
					result.Select = append(result.Select, field)
				}
			}
		case keySort:
			// Handle sorting
			if value == "" {
				continue
			}
			result.Sort = parseSort(value)
		case keyLimit:
			// Handle limit
			if value == "" {
				continue
			}
			limit, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid limit value: %s", value)
			}
			result.Limit = &limit
		case keyOffset:
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
			expr, err := parseFilterExpr(key, value)
			if err != nil {
				return nil, err
			}

			processed := resultAddExpr(result, expr, o)
			if processed != nil {
				result.Where = append(result.Where, processed)
			}
		}
	}

	if result.Offset == nil && o.DefaultOffset != nil {
		result.Offset = o.DefaultOffset
	}
	if result.Limit == nil && o.DefaultLimit != nil {
		result.Limit = o.DefaultLimit
	}

	for key, value := range o.Value {
		if result.Values == nil {
			result.Values = make(map[string][]*ExpressionCmp)
		}

		result.Values[key] = append(result.Values[key], value)
		result.Where = append(result.Where, value)
	}

	return result, nil
}

func resultAddExpr(result *Query, expr Expression, o *optionQuery) Expression {
	if expr == nil {
		return nil
	}

	if cmp, ok := expr.(*ExpressionCmp); ok {
		if result.Values == nil {
			result.Values = make(map[string][]*ExpressionCmp)
		}

		result.Values[cmp.Field] = append(result.Values[cmp.Field], cmp)

		if o.SkipUnderscore && strings.HasPrefix(cmp.Field, "_") {
			return nil
		}

		if _, ok := o.Skip[cmp.Field]; ok {
			return nil
		}

		return cmp
	}

	if exprLogic, ok := expr.(*ExpressionLogic); ok {
		newList := make([]Expression, 0, len(exprLogic.List))

		for _, e := range exprLogic.List {
			processed := resultAddExpr(result, e, o)
			if processed != nil {
				newList = append(newList, processed)
			}
		}

		if len(newList) == 0 {
			return nil
		}

		return &ExpressionLogic{
			Operator: exprLogic.Operator,
			List:     newList,
		}
	}

	// Other expression types
	return expr
}

// parseSort parses the sort parameter and returns the ordered expressions.
func parseSort(value string) []ExpressionSort {
	if value == "" {
		return nil
	}

	fields := strings.Split(value, ",")
	orderedExpressions := make([]ExpressionSort, 0, len(fields))

	for _, field := range fields {
		switch {
		case field == "":
			// Skip empty fields
		case strings.HasPrefix(field, "+"):
			// Ascending order
			orderedExpressions = append(orderedExpressions, ExpressionSort{
				Field: field[1:],
				Desc:  false,
			})
		case strings.HasPrefix(field, "-"):
			// Descending order
			orderedExpressions = append(orderedExpressions, ExpressionSort{
				Field: field[1:],
				Desc:  true,
			})
		case strings.HasSuffix(field, ":asc"):
			// Ascending order
			orderedExpressions = append(orderedExpressions, ExpressionSort{
				Field: field[:len(field)-4],
				Desc:  false,
			})
		case strings.HasSuffix(field, ":desc"):
			// Descending order
			orderedExpressions = append(orderedExpressions, ExpressionSort{
				Field: field[:len(field)-5],
				Desc:  true,
			})
		default:
			// Ascending order
			orderedExpressions = append(orderedExpressions, ExpressionSort{
				Field: field,
				Desc:  false,
			})
		}
	}

	return orderedExpressions
}

func parseFilter(value string) ([]Expression, error) {
	if isParentheses(value) {
		// Strip surrounding parentheses
		value = value[1 : len(value)-1]
	}

	// Handle & conditions
	parts := split(value, '&')
	exs := make([]Expression, 0, len(parts))

	for _, part := range parts {
		if part == "" {
			continue
		}

		if isParentheses(part) {
			// Nested parentheses
			nestedExpr, err := parseFilter(part)
			if err != nil {
				return nil, err
			}
			exs = append(exs, &ExpressionLogic{
				Operator: OperatorAnd,
				List:     nestedExpr,
			})

			continue
		}

		if parts := split(part, '|'); len(parts) > 1 {
			exsInternal := make([]Expression, 0, len(parts))
			for _, p := range parts {
				nestedExpr, err := parseFilter(p)
				if err != nil {
					return nil, err
				}

				switch length := len(nestedExpr); {
				case length > 1:
					exsInternal = append(exsInternal, &ExpressionLogic{
						Operator: OperatorAnd,
						List:     nestedExpr,
					})
				case length == 1:
					exsInternal = append(exsInternal, nestedExpr[0])
				default:
					continue
				}
			}

			exs = append(exs, &ExpressionLogic{
				Operator: OperatorOr,
				List:     exsInternal,
			})

			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			kv = append(kv, "")
		}

		exp, err := parseFilterExpr(kv[0], kv[1])
		if err != nil {
			return nil, err
		}

		exs = append(exs, exp)
	}

	return exs, nil
}

// parseFilterExpr parses filter expressions from key-value pairs.
func parseFilterExpr(key, value string) (Expression, error) {
	switch {
	case strings.Contains(value, "|"):
		// Handle OR conditions
		parts := strings.Split(value, "|")
		exs := make([]Expression, 0, len(parts))

		exp, err := ParseExpression(key, parts[0])
		if err != nil {
			return nil, err
		}

		exs = append(exs, exp)

		for _, part := range parts[1:] {
			if strings.Contains(part, "=") {
				// Different field
				kv := strings.SplitN(part, "=", 2)
				if len(kv) != 2 {
					kv = append(kv, "")
				}

				exp, err := ParseExpression(kv[0], kv[1])
				if err != nil {
					return nil, err
				}

				exs = append(exs, exp)
			} else {
				// Same field
				exp, err := ParseExpression(key, part)
				if err != nil {
					return nil, err
				}

				exs = append(exs, exp)
			}
		}

		return &ExpressionLogic{
			Operator: OperatorOr,
			List:     exs,
		}, nil
	default:
		return ParseExpression(key, value)
	}
}

func isParentheses(value string) bool {
	return strings.HasPrefix(value, "(") && strings.HasSuffix(value, ")")
}

func isParenthesesAny(value string) bool {
	return strings.Contains(value, "(") && strings.Contains(value, ")")
}

// split with & or |
func split(value string, delim byte) []string {
	var result []string
	var current strings.Builder
	depth := 0

	for i := 0; i < len(value); i++ {
		v := value[i]
		switch v {
		case '(':
			depth++
			current.WriteByte(v)
		case ')':
			depth--
			current.WriteByte(v)
		case delim:
			if depth == 0 {
				if current.Len() > 0 {
					result = append(result, current.String())
					current.Reset()
				}
			} else {
				current.WriteByte(v)
			}
		default:
			current.WriteByte(v)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}
