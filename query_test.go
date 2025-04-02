package query

import (
	"net/url"
	"slices"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

func ptr(i uint64) *uint64 {
	return &i
}

func TestParseQuery(t *testing.T) {
	type args struct {
		query string
	}
	tests := []struct {
		name    string
		args    args
		want    *Query
		wantErr bool
	}{
		{
			name: "test 1",
			args: args{
				query: "name=foo,bar&age=1&sort=-age&limit=10&offset=5&fields=id,name",
			},
			want: &Query{
				Select: []any{"id", "name"},
				Where: []exp.Expression{
					goqu.C("name").In("foo", "bar"),
					exp.Ex{
						"age": "1",
					},
				},
				Order: []exp.OrderedExpression{
					goqu.I("age").Desc(),
				},
				Offset: ptr(5),
				Limit:  ptr(10),
			},
			wantErr: false,
		},
		{
			name: "test 2",
			args: args{
				query: "name=foo|nick=bar&age=1&sort=age&limit=10",
			},
			want: &Query{
				Where: []exp.Expression{
					goqu.Or(
						exp.Ex{
							"name": "foo",
						},
						exp.Ex{
							"nick": "bar",
						},
					),
					exp.Ex{
						"age": "1",
					},
				},
				Order: []exp.OrderedExpression{
					goqu.I("age").Asc(),
				},
				Limit: ptr(10),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseQuery() error = %s, wantErr %#v", err, tt.wantErr)
			}

			gotSQL, gotParams, err := got.GoquSelect(goqu.From("test")).ToSQL()
			if err != nil {
				t.Fatalf("ParseQuery() error = %s", err)
			}

			wantSQL, wantParams, err := tt.want.GoquSelect(goqu.From("test")).ToSQL()
			if err != nil {
				t.Fatalf("ParseQuery() error = %s", err)
			}

			if gotSQL != wantSQL {
				t.Errorf("ParseQuery() gotSQL = \n%v\n, want \n%v", gotSQL, wantSQL)
			}
			if !slices.Equal(gotParams, wantParams) {
				t.Errorf("ParseQuery() gotParams = \n%v\n, want \n%v", gotParams, wantParams)
			}
		})
	}
}

func Test_URLQuery(t *testing.T) {
	testURL := "http://example.com?name=foo|nick=foo&age=1&sort=age&limit=10&offset=5&fields=id,name#test"
	parsedURL, err := url.Parse(testURL)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	if parsedURL.RawQuery != "name=foo|nick=foo&age=1&sort=age&limit=10&offset=5&fields=id,name" {
		t.Fatalf("parsed URL query does not match expected value: %s", parsedURL.RawQuery)
	}
}
