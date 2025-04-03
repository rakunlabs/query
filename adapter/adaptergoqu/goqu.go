package adaptergoqu

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/worldline-go/query"
)

func Select(q *query.Query, qq *goqu.SelectDataset) *goqu.SelectDataset {
	if q == nil {
		return qq
	}

	if len(q.Select) > 0 {
		qq = qq.Select(q.Select...)
	}

	if len(q.Where) > 0 {
		where := []exp.Expression{}
		stack := [][]goqu.Expression{{}}
		q.Walk(func(t query.Token) error {
			currentStack := &stack[len(stack)-1]
			switch t.Type {
			case query.WalkCurrent:
				if exprCmp, ok := t.Expression.(query.ExpressionCmp); ok {
					e, err := exprCmpToGoqu(exprCmp)
					if err != nil {
						return err
					}

					*currentStack = append(*currentStack, e)
				}
			case query.WalkStart:
				// add new stack
				stack = append(stack, []goqu.Expression{})
			case query.WalkEnd:
				if exprLogic, ok := t.Expression.(query.ExpressionLogic); ok {
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

		qq = qq.Where(where...)
	}

	if len(q.Order) > 0 {
		order := make([]exp.OrderedExpression, 0, len(q.Order))
		for _, o := range q.Order {
			if o.Desc {
				order = append(order, goqu.I(o.Field).Desc())
			} else {
				order = append(order, goqu.I(o.Field).Asc())
			}
		}

		qq = qq.Order(order...)
	}

	if q.Offset != nil {
		qq = qq.Offset(uint(*q.Offset))
	}

	if q.Limit != nil {
		qq = qq.Limit(uint(*q.Limit))
	}

	return qq
}

func exprLogicToGoqu(e query.ExpressionLogic, stack []goqu.Expression) (goqu.Expression, error) {
	switch e.Operator {
	case query.OperatorAnd:
		return goqu.And(stack...), nil
	case query.OperatorOr:
		return goqu.Or(stack...), nil
	}

	return nil, fmt.Errorf("unsupported operator: [%s]", e.Operator)
}

func exprCmpToGoqu(e query.ExpressionCmp) (goqu.Expression, error) {
	switch e.Operator {
	case query.OperatorEq:
		return goqu.C(e.Field).Eq(e.Value), nil
	case query.OperatorNe:
		return goqu.C(e.Field).Neq(e.Value), nil
	case query.OperatorGt:
		return goqu.C(e.Field).Gt(e.Value), nil
	case query.OperatorLt:
		return goqu.C(e.Field).Lt(e.Value), nil
	case query.OperatorGte:
		return goqu.C(e.Field).Gte(e.Value), nil
	case query.OperatorLte:
		return goqu.C(e.Field).Lte(e.Value), nil
	case query.OperatorLike:
		return goqu.C(e.Field).Like(e.Value), nil
	case query.OperatorILike:
		return goqu.C(e.Field).ILike(e.Value), nil
	case query.OperatorNLike:
		return goqu.C(e.Field).NotLike(e.Value), nil
	case query.OperatorNILike:
		return goqu.C(e.Field).NotILike(e.Value), nil
	case query.OperatorIn:
		return goqu.C(e.Field).In(e.Value), nil
	case query.OperatorNIn:
		return goqu.C(e.Field).NotIn(e.Value), nil
	case query.OperatorIs:
		return goqu.C(e.Field).IsNull(), nil
	case query.OperatorIsNot:
		return goqu.C(e.Field).IsNotNull(), nil
	}

	return nil, fmt.Errorf("unsupported operator: [%s]", e.Operator)
}
