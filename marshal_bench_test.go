package query

import (
	"testing"
)

var benchMarshalBytes []byte

func BenchmarkMarshalText(b *testing.B) {
	benchmarks := []struct {
		name  string
		query *Query
	}{
		{
			name: "simple",
			query: &Query{
				Values: map[string][]*ExpressionCmp{
					"name": {NewExpressionCmp(OperatorEq, "name", "foo")},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorEq, "name", "foo"),
				},
			},
		},
		{
			name: "complex",
			query: &Query{
				Select: []string{"id", "name", "age", "email"},
				Where: []Expression{
					&ExpressionLogic{
						Operator: OperatorOr,
						List: []Expression{
							NewExpressionCmp(OperatorEq, "name", "foo"),
							NewExpressionCmp(OperatorEq, "nick", "bar"),
							&ExpressionLogic{
								Operator: OperatorAnd,
								List: []Expression{
									NewExpressionCmp(OperatorEq, "test", "1"),
									NewExpressionCmp(OperatorEq, "test2", "2"),
								},
							},
						},
					},
					NewExpressionCmp(OperatorGt, "age", "18"),
					NewExpressionCmp(OperatorKV, "meta", `{"a":1,"b":2}`),
				},
				Sort: []ExpressionSort{
					{Field: "age", Desc: true},
					{Field: "name", Desc: false},
				},
				Offset: ptr(10),
				Limit:  ptr(50),
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchMarshalBytes, benchErr = bm.query.MarshalText()
			}
		})
	}
}

func BenchmarkRoundTrip(b *testing.B) {
	input := "_fields=id,name&_sort=age:desc&_limit=10&_offset=5&(name=foo|nick=bar|(test=1&test2=2))&age=1&meta[kv]=eyJhIjoxLCJiIjoyfQ"

	b.ReportAllocs()
	for b.Loop() {
		q, err := Parse(input)
		if err != nil {
			b.Fatal(err)
		}
		benchMarshalBytes, benchErr = q.MarshalText()
	}
}
