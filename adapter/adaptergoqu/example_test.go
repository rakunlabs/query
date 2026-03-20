package adaptergoqu_test

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/rakunlabs/query"
	"github.com/rakunlabs/query/adapter/adaptergoqu"
)

func ExampleParseQuery() {
	urlStr := "http://example.com?name=foo,bar|nick=bar&age[lt]=1&_sort=-age&_limit=10&_offset=5&_fields=id,name&_events=true&meta[kv]=eyJhIjoxLCJiIjoyfQ"
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}

	query, err := query.Parse(parsedURL.RawQuery)
	if err != nil {
		panic(err)
	}

	sql, params, err := adaptergoqu.Select(query, goqu.From("test"), adaptergoqu.WithParameterized(false)).ToSQL()
	if err != nil {
		panic(err)
	}

	// Print the SQL query and parameters
	fmt.Println("SQL:", sql)
	fmt.Println("Params:", params)
	// Output:
	// SQL: SELECT "id", "name" FROM "test" WHERE ((("name" IN ('foo', 'bar')) OR ("nick" = 'bar')) AND ("age" < '1') AND "meta" @> '{"a":1,"b":2}') ORDER BY "age" DESC LIMIT 10 OFFSET 5
	// Params: []
}

func ExampleParseQuery_parameterized() {
	urlStr := "http://example.com?name=foo,bar|nick=bar&age[lt]=1&_sort=-age&_limit=10&_offset=5&_fields=id,name"
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}

	query, err := query.Parse(parsedURL.RawQuery)
	if err != nil {
		panic(err)
	}

	sql, params, err := adaptergoqu.Select(query, goqu.From("test"), adaptergoqu.WithParameterized(true)).ToSQL()
	if err != nil {
		panic(err)
	}

	// Print the parameterized SQL query and parameters
	fmt.Println("SQL:", sql)
	fmt.Println("Params:", params)
	// Output:
	// SQL: SELECT "id", "name" FROM "test" WHERE ((("name" IN (?, ?)) OR ("nick" = ?)) AND ("age" < ?)) ORDER BY "age" DESC LIMIT ? OFFSET ?
	// Params: [foo bar bar 1 10 5]
}

func BenchmarkParseQuery(b *testing.B) {
	urlStr := "http://example.com?name=foo,bar|nick=bar&age[lt]=1&_sort=-age&_limit=10&_offset=5&_fields=id,name"

	b.ReportAllocs()

	for b.Loop() {
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			b.Fatalf("failed to parse URL: %v", err)
		}

		// Parse the query string
		q, err := query.Parse(parsedURL.RawQuery)
		if err != nil {
			b.Fatalf("failed to parse query: %v", err)
		}

		// Generate SQL using adaptergoqu
		_, _, err = adaptergoqu.Select(q, goqu.From("test")).ToSQL()
		if err != nil {
			b.Fatalf("failed to generate SQL: %v", err)
		}
	}
}

func TestSQLInjection(t *testing.T) {
	// Test with potentially malicious input that could cause SQL injection
	maliciousURL := "http://example.com?name=foo'; DROP TABLE users;--"
	parsedURL, err := url.Parse(maliciousURL)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	query, err := query.Parse(parsedURL.RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	// Test with literal SQL (default)
	sql, params, err := adaptergoqu.Select(query, goqu.From("test"), adaptergoqu.WithParameterized(false)).ToSQL()
	if err != nil {
		t.Fatalf("failed to generate SQL: %v", err)
	}

	t.Logf("Literal SQL: %s", sql)
	t.Logf("Params: %v", params)

	// Check that the malicious input is properly escaped
	if !strings.Contains(sql, `foo''; DROP TABLE users;--`) {
		t.Error("Malicious input was not properly escaped in literal SQL")
	}

	// Test with parameterized SQL
	sql, params, err = adaptergoqu.Select(query, goqu.From("test"), adaptergoqu.WithParameterized(true)).ToSQL()
	if err != nil {
		t.Fatalf("failed to generate parameterized SQL: %v", err)
	}

	t.Logf("Parameterized SQL: %s", sql)
	t.Logf("Params: %v", params)

	// Check that parameterized SQL uses placeholders
	if !strings.Contains(sql, "?") || len(params) == 0 {
		t.Error("Parameterized SQL was not generated correctly")
	}
}

func TestJInOperatorSQL(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		opts    []query.OptionQuery
		wantSQL string
	}{
		{
			name:    "jin with multiple values",
			query:   "tags[jin]=admin,editor",
			wantSQL: `SELECT * FROM "test" WHERE "tags" ?| array['admin','editor']`,
		},
		{
			name:    "jin with single value",
			query:   "tags[jin]=admin",
			wantSQL: `SELECT * FROM "test" WHERE "tags" ?| array['admin']`,
		},
		{
			name:    "njin with multiple values",
			query:   "tags[njin]=admin,editor",
			wantSQL: `SELECT * FROM "test" WHERE NOT ("tags" ?| array['admin','editor'])`,
		},
		{
			name:    "jin via WithKeyOperator",
			query:   "tags=admin,editor",
			opts:    []query.OptionQuery{query.WithKeyOperator("tags", query.OperatorJIn)},
			wantSQL: `SELECT * FROM "test" WHERE "tags" ?| array['admin','editor']`,
		},
		{
			name:    "jin with rename",
			query:   "tags[jin]=admin,editor",
			wantSQL: `SELECT * FROM "test" WHERE "data"."tags" ?| array['admin','editor']`,
		},
		{
			name:    "jin sql injection prevention",
			query:   "tags[jin]=admin,'; DROP TABLE users;--",
			wantSQL: `SELECT * FROM "test" WHERE "tags" ?| array['admin','''; DROP TABLE users;--']`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := query.Parse(tt.query, tt.opts...)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			var adapterOpts []adaptergoqu.Option
			adapterOpts = append(adapterOpts, adaptergoqu.WithParameterized(false))
			if tt.name == "jin with rename" {
				adapterOpts = append(adapterOpts, adaptergoqu.WithRename(map[string]string{
					"tags": "data.tags",
				}))
			}

			sql, _, err := adaptergoqu.Select(q, goqu.From("test"), adapterOpts...).ToSQL()
			if err != nil {
				t.Fatalf("ToSQL() error = %v", err)
			}

			if sql != tt.wantSQL {
				t.Errorf("SQL = %s, want %s", sql, tt.wantSQL)
			}
		})
	}
}
