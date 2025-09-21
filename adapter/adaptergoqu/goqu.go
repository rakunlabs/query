package adaptergoqu

import (
	"fmt"

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
	opt := &option{}
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
	}

	return nil, fmt.Errorf("unsupported operator: [%s]", e.Operator)
}
