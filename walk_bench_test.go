package query

import (
	"testing"
)

var benchWalkCount int

func BenchmarkWalk(b *testing.B) {
	benchmarks := []struct {
		name  string
		query *Query
	}{
		{
			name: "flat_5",
			query: &Query{
				Values: map[string][]*ExpressionCmp{},
				Where: []Expression{
					NewExpressionCmp(OperatorEq, "a", "1"),
					NewExpressionCmp(OperatorEq, "b", "2"),
					NewExpressionCmp(OperatorGt, "c", "3"),
					NewExpressionCmp(OperatorLt, "d", "4"),
					NewExpressionCmp(OperatorEq, "e", "5"),
				},
			},
		},
		{
			name: "nested_or_and",
			query: &Query{
				Values: map[string][]*ExpressionCmp{},
				Where: []Expression{
					&ExpressionLogic{
						Operator: OperatorOr,
						List: []Expression{
							NewExpressionCmp(OperatorEq, "name", "foo"),
							NewExpressionCmp(OperatorEq, "name", "bar"),
						},
					},
					&ExpressionLogic{
						Operator: OperatorAnd,
						List: []Expression{
							NewExpressionCmp(OperatorGt, "age", "18"),
							NewExpressionCmp(OperatorLt, "age", "65"),
						},
					},
					NewExpressionCmp(OperatorEq, "status", "active"),
				},
			},
		},
		{
			name: "deep_3_levels",
			query: &Query{
				Values: map[string][]*ExpressionCmp{},
				Where: []Expression{
					&ExpressionLogic{
						Operator: OperatorOr,
						List: []Expression{
							&ExpressionLogic{
								Operator: OperatorAnd,
								List: []Expression{
									NewExpressionCmp(OperatorEq, "a", "1"),
									&ExpressionLogic{
										Operator: OperatorOr,
										List: []Expression{
											NewExpressionCmp(OperatorEq, "b", "2"),
											NewExpressionCmp(OperatorEq, "c", "3"),
										},
									},
								},
							},
							NewExpressionCmp(OperatorEq, "d", "4"),
						},
					},
					NewExpressionCmp(OperatorEq, "e", "5"),
				},
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				count := 0
				bm.query.Walk(func(t Token) error {
					count++
					return nil
				})
				benchWalkCount = count
			}
		})
	}
}
