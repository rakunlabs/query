package query

type Query struct {
	Values map[string][]*ExpressionCmp

	Select []string
	Where  []Expression
	Sort   []ExpressionSort
	Offset *uint64
	Limit  *uint64
}

func (q *Query) GetValues(v string) []string {
	if values, ok := q.Values[v]; ok {
		result := make([]string, 0, len(values))

		for _, v := range values {
			if vList, ok := v.Value.([]string); ok {
				result = append(result, vList...)
			} else {
				result = append(result, v.Value.(string))
			}
		}

		return result
	}

	return nil
}

func (q *Query) GetValue(v string) string {
	if values, ok := q.Values[v]; ok {
		for _, v := range values {
			if vList, ok := v.Value.([]string); ok {
				if len(vList) > 0 {
					return vList[0]
				}
			} else {
				return v.Value.(string)
			}
		}
	}

	return ""
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

func (q *Query) CloneLimit() *uint64 {
	if q.Limit != nil {
		v := *q.Limit

		return &v
	}

	return nil
}

func (q *Query) CloneOffset() *uint64 {
	if q.Offset != nil {
		v := *q.Offset

		return &v
	}

	return nil
}

func New() *Query {
	return &Query{
		Values: make(map[string][]*ExpressionCmp),
	}
}

func (q *Query) AddField(fields ...string) *Query {
	q.Select = append(q.Select, fields...)

	return q
}

func (q *Query) AddSort(sorts ...ExpressionSort) *Query {
	q.Sort = append(q.Sort, sorts...)

	return q
}

// SetLimit sets the limit for the query.
//   - if limit is <= 0, it means no limit.
func (q *Query) SetLimit(limit uint64) *Query {
	if limit <= 0 {
		q.Limit = nil
		return q
	}

	q.Limit = &limit
	return q
}

// SetOffset sets the offset for the query.
//   - if offset is <= 0, it means no offset.
func (q *Query) SetOffset(offset uint64) *Query {
	if offset <= 0 {
		q.Offset = nil
		return q
	}

	q.Offset = &offset
	return q
}

func (q *Query) AddWhere(exprs ...Expression) *Query {
	q.Where = append(q.Where, exprs...)

	for _, expr := range exprs {
		q.valuesExpression(expr)
	}

	return q
}

func (q *Query) valuesExpression(expr Expression) {
	if expr == nil {
		return
	}

	switch expr := expr.(type) {
	case *ExpressionCmp:
		q.Values[expr.Field] = append(q.Values[expr.Field], expr)
	case *ExpressionLogic:
		for _, e := range expr.List {
			if e == nil {
				continue
			}

			q.valuesExpression(e)
		}
	}
}
