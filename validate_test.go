package query

import (
	"net/url"
	"testing"
)

func TestQuery_Validate(t *testing.T) {
	type subCase struct {
		URL     string
		wantErr bool
	}
	type args struct {
		opts []OptionValidateSet
	}
	tests := []struct {
		name  string
		cases []subCase
		args  args
	}{
		{
			name: "invalid age",
			cases: []subCase{
				{
					URL:     "http://example.com?age=10000&_fields=name,age&_sort=age,-name&_offset=10&_limit=20",
					wantErr: true,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithValue("age", WithMin("0"), WithMax("200")),
				},
			},
		},
		{
			name: "invalid age",
			cases: []subCase{
				{
					URL:     "http://example.com?age=10000,500&_fields=name,age&_sort=age,-name&_offset=10&_limit=20",
					wantErr: true,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithValue("age", WithMin("0"), WithMax("200")),
				},
			},
		},
		{
			name: "not a member",
			cases: []subCase{
				{
					URL:     "http://example.com?member=Z",
					wantErr: true,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithValue("member", WithIn("X", "Y")),
				},
			},
		},
		{
			name: "not a member list",
			cases: []subCase{
				{
					URL:     "http://example.com?member=X,Z",
					wantErr: true,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithValue("member", WithIn("X", "Y")),
				},
			},
		},
		{
			name: "not in",
			cases: []subCase{
				{
					URL:     "http://example.com?member=X,Z",
					wantErr: false,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithValue("member", WithNotIn("O", "P", "S")),
				},
			},
		},
		{
			name: "required",
			cases: []subCase{
				{
					URL:     "http://example.com?member[in]=x",
					wantErr: false,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithValue("member", WithNotEmpty()),
				},
			},
		},
		{
			name: "required",
			cases: []subCase{
				{
					URL:     "http://example.com?_fields=name,age&_sort=age,-name&_offset=10&_limit=20",
					wantErr: true,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithField(WithIn("X")),
				},
			},
		},
		{
			name: "not allowed",
			cases: []subCase{
				{
					URL:     "http://example.com?age=5&fields=name,age&_sort=age,-name&_offset=10&_limit=20",
					wantErr: true,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithField(WithNotAllowed()),
					WithValues(WithNotAllowed()),
				},
			},
		},
		{
			name: "values in",
			cases: []subCase{
				{
					URL:     "http://example.com?age=5",
					wantErr: false,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithValues(WithIn("age", "test")),
				},
			},
		},
		{
			name: "values operator",
			cases: []subCase{
				{
					URL:     "http://example.com?age=5",
					wantErr: false,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithValues(WithIn("age")),
					WithValue("age", WithOperator(OperatorEq), WithNotOperator(OperatorIn)),
					WithValue("test", WithNotAllowed()),
				},
			},
		},
		{
			name: "limit min max",
			cases: []subCase{
				{
					URL:     "http://example.com?_limit=200",
					wantErr: true,
				},
				{
					URL:     "http://example.com?_limit=100",
					wantErr: false,
				},
				{
					URL:     "http://example.com?_limit=0",
					wantErr: true,
				},
				{
					URL:     "http://example.com?_limit=1",
					wantErr: false,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithLimit(WithMin("1"), WithMax("100")),
				},
			},
		},
		{
			name: "offset min max",
			cases: []subCase{
				{
					URL:     "http://example.com?_offset=200",
					wantErr: true,
				},
				{
					URL:     "http://example.com?_offset=100",
					wantErr: false,
				},
				{
					URL:     "http://example.com?_offset=0",
					wantErr: true,
				},
				{
					URL:     "http://example.com?_offset=1",
					wantErr: false,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithOffset(WithMin("1"), WithMax("100")),
				},
			},
		},
		{
			name: "limit not allowed",
			cases: []subCase{
				{
					URL:     "http://example.com?_limit=200",
					wantErr: true,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithLimit(WithNotAllowed()),
				},
			},
		},
		{
			name: "offset not allowed",
			cases: []subCase{
				{
					URL:     "http://example.com?_offset=200",
					wantErr: true,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithOffset(WithNotAllowed()),
				},
			},
		},
		{
			name: "sort not allowed",
			cases: []subCase{
				{
					URL:     "http://example.com?_sort=age,-name",
					wantErr: true,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithSort(WithNotAllowed()),
				},
			},
		},
		{
			name: "sort in",
			cases: []subCase{
				{
					URL:     "http://example.com?_sort=age,-name",
					wantErr: false,
				},
				{
					URL:     "http://example.com?_sort=age,-name,test",
					wantErr: true,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithSort(WithIn("age", "name")),
				},
			},
		},
		{
			name: "sort not in",
			cases: []subCase{
				{
					URL:     "http://example.com?_sort=age,-name",
					wantErr: true,
				},
				{
					URL:     "http://example.com?_sort=-name,test",
					wantErr: false,
				},
			},
			args: args{
				opts: []OptionValidateSet{
					WithSort(WithNotIn("age")),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate, err := NewValidator(tt.args.opts...)
			if err != nil {
				t.Fatalf("failed to create validator: %v", err)
			}

			for _, tc := range tt.cases {
				t.Logf("[%s] case %s", tt.name, tc.URL)

				urlParsed, err := url.Parse(tc.URL)
				if err != nil {
					t.Fatalf("failed to parse URL: %v", err)
				}

				q, err := Parse(urlParsed.RawQuery)
				if err != nil {
					t.Fatalf("failed to parse query: %v", err)
				}

				if err := q.Validate(validate); (err != nil) != tc.wantErr {
					t.Errorf("Query.Validate() error = %v, wantErr %v", err, tc.wantErr)
				}
			}
		})
	}
}
