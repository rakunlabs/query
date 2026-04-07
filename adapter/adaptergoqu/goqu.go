package adaptergoqu

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/rakunlabs/query"
)

func Expression(q *query.Query, opts ...Option) []exp.Expression {
	opt := &option{}
	for _, o := range opts {
		o(opt)
	}

	if opt.Edit != nil {
		q = opt.Edit(q)
	}

	if q == nil {
		return nil
	}

	if len(q.Where) > 0 {
		where := []exp.Expression{}
		stack := [][]goqu.Expression{{}}
		q.Walk(func(t query.Token) error {
			currentStack := &stack[len(stack)-1]
			switch t.Type {
			case query.WalkCurrent:
				if exprCmp, ok := t.Expression.(*query.ExpressionCmp); ok {
					e, err := exprCmpToGoqu(exprCmp, opt.Rename)
					if err != nil {
						return err
					}

					*currentStack = append(*currentStack, e)
				}
			case query.WalkStart:
				// add new stack
				stack = append(stack, []goqu.Expression{})
			case query.WalkEnd:
				if exprLogic, ok := t.Expression.(*query.ExpressionLogic); ok {
					e, err := exprLogicToGoqu(exprLogic, *currentStack)
					if err != nil {
						return err
					}

					if len(stack) > 1 {
						// pop stack
						stack = stack[:len(stack)-1]
						// add to parent stack
						stack[len(stack)-1] = append(stack[len(stack)-1], e)
					} else {
						// add to where
						where = append(where, e)
					}
				} else {
					return fmt.Errorf("unexpected expression type: %T", t.Expression)
				}
			default:
				return fmt.Errorf("unsupported walk type: %d", t.Type)
			}

			return nil
		})

		return where
	}

	return nil
}

func Select(q *query.Query, qq *goqu.SelectDataset, opts ...Option) *goqu.SelectDataset {
	opt := &option{
		Parameterized: true,
	}
	for _, o := range opts {
		o(opt)
	}

	if opt.Edit != nil {
		q = opt.Edit(q)
	}

	if q == nil {
		return qq
	}

	var selects []string
	if len(q.Select) > 0 {
		selects = q.Select
	} else if len(opt.DefaultSelect) > 0 {
		selects = opt.DefaultSelect
	}

	if len(selects) > 0 {
		selectsAny := make([]any, 0, len(selects))
		for _, s := range selects {
			if rename, ok := opt.Rename[s]; ok {
				s = rename
			}

			selectsAny = append(selectsAny, goqu.I(s))
		}

		qq = qq.Select(selectsAny...)
	}

	if len(q.Where) > 0 {
		qq = qq.Where(Expression(q, opts...)...)
	}

	if len(q.Sort) > 0 {
		order := make([]exp.OrderedExpression, 0, len(q.Sort))
		for _, o := range q.Sort {
			field := o.Field
			if rename, ok := opt.Rename[field]; ok {
				field = rename
			}

			if o.Desc {
				order = append(order, goqu.I(field).Desc())
			} else {
				order = append(order, goqu.I(field).Asc())
			}
		}

		qq = qq.Order(order...)
	}

	if q.Offset != nil {
		qq = qq.Offset(uint(*q.Offset))
	}

	if q.Limit != nil {
		if *q.Limit != 0 {
			qq = qq.Limit(uint(*q.Limit))
		}
	}

	if opt.Parameterized {
		qq = qq.Prepared(true)
	}

	return qq
}

func exprLogicToGoqu(e *query.ExpressionLogic, stack []goqu.Expression) (goqu.Expression, error) {
	switch e.Operator {
	case query.OperatorAnd:
		return goqu.And(stack...), nil
	case query.OperatorOr:
		return goqu.Or(stack...), nil
	}

	return nil, fmt.Errorf("unsupported operator: [%s]", e.Operator)
}

func exprCmpToGoqu(e *query.ExpressionCmp, rename map[string]string) (goqu.Expression, error) {
	field := e.Field
	if rename, ok := rename[field]; ok {
		field = rename
	}

	fieldI := goqu.I(field)

	// Handle comma-split []string values for operators that support it.
	if values, ok := e.Value.([]string); ok && len(values) > 1 {
		if fn, logicOp, supported := commaSplitGoquFn(e.Operator, fieldI); supported {
			exprs := make([]goqu.Expression, len(values))
			for i, v := range values {
				exprs[i] = fn(v)
			}

			if logicOp == "and" {
				return goqu.And(exprs...), nil
			}

			return goqu.Or(exprs...), nil
		}
	}

	switch e.Operator {
	case query.OperatorEq:
		return fieldI.Eq(e.Value), nil
	case query.OperatorNe:
		return fieldI.Neq(e.Value), nil
	case query.OperatorGt:
		return fieldI.Gt(e.Value), nil
	case query.OperatorLt:
		return fieldI.Lt(e.Value), nil
	case query.OperatorGte:
		return fieldI.Gte(e.Value), nil
	case query.OperatorLte:
		return fieldI.Lte(e.Value), nil
	case query.OperatorLike:
		return fieldI.Like(e.Value), nil
	case query.OperatorILike:
		return fieldI.ILike(e.Value), nil
	case query.OperatorNLike:
		return fieldI.NotLike(e.Value), nil
	case query.OperatorNILike:
		return fieldI.NotILike(e.Value), nil
	case query.OperatorIn:
		return fieldI.In(e.Value), nil
	case query.OperatorNIn:
		return fieldI.NotIn(e.Value), nil
	case query.OperatorIs:
		return fieldI.IsNull(), nil
	case query.OperatorIsNot:
		return fieldI.IsNotNull(), nil
	case query.OperatorKV:
		// For JSONB containment (@>) operator
		return goqu.L("? @> ?", fieldI, e.Value), nil
	case query.OperatorJIn:
		// For JSONB array "has any" (?|) operator
		return goqu.L("? ?| "+buildArrayLiteral(e.Value), fieldI), nil
	case query.OperatorNJIn:
		// For negated JSONB array "has any" (NOT ?|) operator
		return goqu.L("NOT (? ?| "+buildArrayLiteral(e.Value)+")", fieldI), nil
	}

	return nil, fmt.Errorf("unsupported operator: [%s]", e.Operator)
}

// commaSplitGoquFn returns a function that creates a goqu expression for a single value,
// the logic operator to combine multiple expressions ("or" or "and"),
// and whether the operator supports comma splitting.
// Negated operators (ne, nlike, nilike) use AND; positive operators use OR.
func commaSplitGoquFn(op query.OperatorCmpType, fieldI exp.IdentifierExpression) (func(string) goqu.Expression, string, bool) {
	switch op {
	case query.OperatorEq:
		return func(v string) goqu.Expression { return fieldI.Eq(v) }, "or", true
	case query.OperatorNe:
		return func(v string) goqu.Expression { return fieldI.Neq(v) }, "and", true
	case query.OperatorGt:
		return func(v string) goqu.Expression { return fieldI.Gt(v) }, "or", true
	case query.OperatorLt:
		return func(v string) goqu.Expression { return fieldI.Lt(v) }, "or", true
	case query.OperatorGte:
		return func(v string) goqu.Expression { return fieldI.Gte(v) }, "or", true
	case query.OperatorLte:
		return func(v string) goqu.Expression { return fieldI.Lte(v) }, "or", true
	case query.OperatorLike:
		return func(v string) goqu.Expression { return fieldI.Like(v) }, "or", true
	case query.OperatorILike:
		return func(v string) goqu.Expression { return fieldI.ILike(v) }, "or", true
	case query.OperatorNLike:
		return func(v string) goqu.Expression { return fieldI.NotLike(v) }, "and", true
	case query.OperatorNILike:
		return func(v string) goqu.Expression { return fieldI.NotILike(v) }, "and", true
	default:
		return nil, "", false
	}
}

// buildArrayLiteral constructs a SQL array literal from a value.
// The value is expected to be []string from the jin/njin operators.
func buildArrayLiteral(v any) string {
	values, ok := v.([]string)
	if !ok {
		return "array[]"
	}

	quoted := make([]string, len(values))
	for i, s := range values {
		quoted[i] = "'" + strings.ReplaceAll(s, "'", "''") + "'"
	}

	return "array[" + strings.Join(quoted, ",") + "]"
}
