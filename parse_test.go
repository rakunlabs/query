package query

import (
	"reflect"
	"testing"
)

func TestMarshalText(t *testing.T) {
	q := &Query{
		Select: []string{"id", "name"},
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
			NewExpressionCmp(OperatorEq, "age", "1"),
		},
		Sort: []ExpressionSort{
			{
				Field: "age",
				Desc:  true,
			},
		},
		Offset: ptr(5),
		Limit:  ptr(10),
	}

	b, err := q.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}

	expected := "fields=id,name&sort=age:desc&limit=10&offset=5&(name=foo|nick=bar|(test=1&test2=2))&age=1"
	if string(b) != expected {
		t.Fatalf("MarshalText = %s, want %s", string(b), expected)
	}

	unmarshaled, err := Parse(string(b))
	if err != nil {
		t.Fatalf("UnmarshalText error: %v", err)
	}

	if unmarshaled.Limit != nil && q.Limit != nil {
		if *unmarshaled.Limit != *q.Limit {
			t.Fatalf("UnmarshalText() Limit = %d, want %d", *unmarshaled.Limit, *q.Limit)
		}
	}
	if unmarshaled.Offset != nil && q.Offset != nil {
		if *unmarshaled.Offset != *q.Offset {
			t.Fatalf("UnmarshalText() Offset = %d, want %d", *unmarshaled.Offset, *q.Offset)
		}
	}

	if !reflect.DeepEqual(unmarshaled.Where, q.Where) {
		t.Fatalf("UnmarshalText() Where = %v, want %v", unmarshaled.Where, q.Where)
	}

	if !reflect.DeepEqual(unmarshaled.Sort, q.Sort) {
		t.Fatalf("UnmarshalText() Sort = %v, want %v", unmarshaled.Sort, q.Sort)
	}

	if !reflect.DeepEqual(unmarshaled.Select, q.Select) {
		t.Fatalf("UnmarshalText() Select = %v, want %v", unmarshaled.Select, q.Select)
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    *Query
		wantErr bool
	}{
		{
			name:  "test 1",
			value: "(amount=50.12|method=CARD)|name=test",
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"amount": {NewExpressionCmp(OperatorEq, "amount", "50.12")},
					"method": {NewExpressionCmp(OperatorEq, "method", "CARD")},
					"name":   {NewExpressionCmp(OperatorEq, "name", "test")},
				},
				Where: []Expression{
					&ExpressionLogic{
						Operator: OperatorOr,
						List: []Expression{
							&ExpressionLogic{
								Operator: OperatorOr,
								List: []Expression{
									NewExpressionCmp(OperatorEq, "amount", "50.12"),
									NewExpressionCmp(OperatorEq, "method", "CARD"),
								},
							},
							NewExpressionCmp(OperatorEq, "name", "test"),
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_split(t *testing.T) {
	tests := []struct {
		name  string
		value string
		split byte
		want  []string
	}{
		{
			name:  "test 1",
			value: "name=foo|(test=2&age=1)&nick=bar",
			split: '&',
			want:  []string{"name=foo|(test=2&age=1)", "nick=bar"},
		},
		{
			name:  "test 2",
			value: "(name=foo|nick=bar)&age=1",
			split: '&',
			want:  []string{"(name=foo|nick=bar)", "age=1"},
		},
		{
			name:  "test 3",
			value: "name=foo|(test=2&(age=1|age=2))&nick=bar",
			split: '&',
			want:  []string{"name=foo|(test=2&(age=1|age=2))", "nick=bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := split(tt.value, tt.split)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("split() = %v, want %v", got, tt.want)
			}
		})
	}
}
