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
			NewExpressionCmp(OperatorKV, "meta", `{"a":1,"b":2}`),
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

	expected := "_fields=id,name&_sort=age:desc&_limit=10&_offset=5&(name=foo|nick=bar|(test=1&test2=2))&age=1&meta[kv]=eyJhIjoxLCJiIjoyfQ"
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

func TestParseWithKeyOperator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		opts    []OptionQuery
		want    *Query
		wantErr bool
	}{
		{
			name:  "key operator like",
			value: "name=foo",
			opts:  []OptionQuery{WithKeyOperator("name", OperatorLike)},
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"name": {NewExpressionCmp(OperatorLike, "name", "foo")},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorLike, "name", "foo"),
				},
			},
		},
		{
			name:  "key operator kv with json",
			value: `meta={"a":1}`,
			opts:  []OptionQuery{WithKeyOperator("meta", OperatorKV)},
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"meta": {NewExpressionCmp(OperatorKV, "meta", `{"a":1}`)},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorKV, "meta", `{"a":1}`),
				},
			},
		},
		{
			name:  "key operator does not override explicit bracket operator",
			value: "name[eq]=foo",
			opts:  []OptionQuery{WithKeyOperator("name", OperatorLike)},
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"name": {NewExpressionCmp(OperatorEq, "name", "foo")},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorEq, "name", "foo"),
				},
			},
		},
		{
			name:  "key operator only applies to configured key",
			value: "name=foo&age=30",
			opts:  []OptionQuery{WithKeyOperator("name", OperatorILike)},
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"name": {NewExpressionCmp(OperatorILike, "name", "foo")},
					"age":  {NewExpressionCmp(OperatorEq, "age", "30")},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorILike, "name", "foo"),
					NewExpressionCmp(OperatorEq, "age", "30"),
				},
			},
		},
		{
			name:  "key operator lt with numeric type",
			value: "age=30",
			opts: []OptionQuery{
				WithKeyOperator("age", OperatorLt),
				WithKeyType("age", ValueTypeNumber),
			},
			want: func() *Query {
				expr := NewExpressionCmp(OperatorLt, "age", "30")
				return &Query{
					Values: map[string][]*ExpressionCmp{
						"age": {expr},
					},
					Where: []Expression{expr},
				}
			}(),
		},
		{
			name:  "key operator in parenthesized expression",
			value: "(name=foo|name=bar)",
			opts:  []OptionQuery{WithKeyOperator("name", OperatorLike)},
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"name": {
						NewExpressionCmp(OperatorLike, "name", "foo"),
						NewExpressionCmp(OperatorLike, "name", "bar"),
					},
				},
				Where: []Expression{
					&ExpressionLogic{
						Operator: OperatorOr,
						List: []Expression{
							NewExpressionCmp(OperatorLike, "name", "foo"),
							NewExpressionCmp(OperatorLike, "name", "bar"),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.value, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseWithKeyValueTransform(t *testing.T) {
	wrapPercent := func(v string) string { return "%" + v + "%" }

	tests := []struct {
		name    string
		value   string
		opts    []OptionQuery
		want    *Query
		wantErr bool
	}{
		{
			name:  "value transform alone",
			value: "name=test",
			opts:  []OptionQuery{WithKeyValueTransform("name", wrapPercent)},
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"name": {NewExpressionCmp(OperatorEq, "name", "%test%")},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorEq, "name", "%test%"),
				},
			},
		},
		{
			name:  "value transform with key operator",
			value: "name=test",
			opts: []OptionQuery{
				WithKeyOperator("name", OperatorLike),
				WithKeyValueTransform("name", wrapPercent),
			},
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"name": {NewExpressionCmp(OperatorLike, "name", "%test%")},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorLike, "name", "%test%"),
				},
			},
		},
		{
			name:  "value transform with explicit bracket operator",
			value: "name[like]=test",
			opts:  []OptionQuery{WithKeyValueTransform("name", wrapPercent)},
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"name": {NewExpressionCmp(OperatorLike, "name", "%test%")},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorLike, "name", "%test%"),
				},
			},
		},
		{
			name:  "value transform only affects configured key",
			value: "name=test&age=30",
			opts:  []OptionQuery{WithKeyValueTransform("name", wrapPercent)},
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"name": {NewExpressionCmp(OperatorEq, "name", "%test%")},
					"age":  {NewExpressionCmp(OperatorEq, "age", "30")},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorEq, "name", "%test%"),
					NewExpressionCmp(OperatorEq, "age", "30"),
				},
			},
		},
		{
			name:  "value transform in parenthesized expression",
			value: "(name=foo|name=bar)",
			opts: []OptionQuery{
				WithKeyOperator("name", OperatorLike),
				WithKeyValueTransform("name", wrapPercent),
			},
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"name": {
						NewExpressionCmp(OperatorLike, "name", "%foo%"),
						NewExpressionCmp(OperatorLike, "name", "%bar%"),
					},
				},
				Where: []Expression{
					&ExpressionLogic{
						Operator: OperatorOr,
						List: []Expression{
							NewExpressionCmp(OperatorLike, "name", "%foo%"),
							NewExpressionCmp(OperatorLike, "name", "%bar%"),
						},
					},
				},
			},
		},
		{
			name:  "value transform with suffix function",
			value: "key=test",
			opts: []OptionQuery{
				WithKeyValueTransform("key", func(v string) string { return v + "x" }),
			},
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"key": {NewExpressionCmp(OperatorEq, "key", "testx")},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorEq, "key", "testx"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.value, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseWithJInOperator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		opts    []OptionQuery
		want    *Query
		wantErr bool
	}{
		{
			name:  "explicit jin with multiple values",
			value: "tags[jin]=admin,editor",
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"tags": {NewExpressionCmp(OperatorJIn, "tags", []string{"admin", "editor"})},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorJIn, "tags", []string{"admin", "editor"}),
				},
			},
		},
		{
			name:  "explicit jin with single value",
			value: "tags[jin]=admin",
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"tags": {NewExpressionCmp(OperatorJIn, "tags", []string{"admin"})},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorJIn, "tags", []string{"admin"}),
				},
			},
		},
		{
			name:  "explicit njin with multiple values",
			value: "tags[njin]=admin,editor",
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"tags": {NewExpressionCmp(OperatorNJIn, "tags", []string{"admin", "editor"})},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorNJIn, "tags", []string{"admin", "editor"}),
				},
			},
		},
		{
			name:  "jin via WithKeyOperator",
			value: "tags=admin,editor",
			opts:  []OptionQuery{WithKeyOperator("tags", OperatorJIn)},
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"tags": {NewExpressionCmp(OperatorJIn, "tags", []string{"admin", "editor"})},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorJIn, "tags", []string{"admin", "editor"}),
				},
			},
		},
		{
			name:  "njin via WithKeyOperator",
			value: "tags=admin",
			opts:  []OptionQuery{WithKeyOperator("tags", OperatorNJIn)},
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"tags": {NewExpressionCmp(OperatorNJIn, "tags", []string{"admin"})},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorNJIn, "tags", []string{"admin"}),
				},
			},
		},
		{
			name:  "jin with other filters",
			value: "tags[jin]=admin,editor&name=foo",
			want: &Query{
				Values: map[string][]*ExpressionCmp{
					"tags": {NewExpressionCmp(OperatorJIn, "tags", []string{"admin", "editor"})},
					"name": {NewExpressionCmp(OperatorEq, "name", "foo")},
				},
				Where: []Expression{
					NewExpressionCmp(OperatorJIn, "tags", []string{"admin", "editor"}),
					NewExpressionCmp(OperatorEq, "name", "foo"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.value, tt.opts...)
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
