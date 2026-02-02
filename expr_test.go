package query_test

import (
	"testing"

	"github.com/rakunlabs/query"
)

func TestExpr(t *testing.T) {
	q := query.New().AddWhere(
		&query.ExpressionLogic{
			Operator: query.OperatorOr,
			List: []query.Expression{
				query.NewExpressionCmp(
					query.OperatorEq,
					"name",
					"ELECTRONIC_REFUND",
				),
				query.NewExpressionCmp(
					query.OperatorEq,
					"name",
					"ELECTRONIC_CREDIT",
				),
			},
		},
		query.NewExpressionCmp(
			query.OperatorEq,
			"status",
			"completed",
		),
		query.NewExpressionCmp(
			query.OperatorEq,
			"amount",
			1234,
		),
		query.NewExpressionCmp(
			query.OperatorKV,
			"metadata",
			`{"group_id":1234}`,
		),
	).SetLimit(100)

	text, err := q.MarshalText()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "_limit=100&name=(ELECTRONIC_REFUND|ELECTRONIC_CREDIT)&status=completed&amount=1234&metadata[kv]=eyJncm91cF9pZCI6MTIzNH0"

	if string(text) != expected {
		t.Fatalf("unexpected text:\n- want: %s\n-  got: %s", expected, string(text))
	}
}
