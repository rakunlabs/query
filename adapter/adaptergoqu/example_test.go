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
	urlStr := "http://example.com?name=foo,bar|nick=bar&age[lt]=1&_sort=-age&_limit=10&_offset=5&_fields=id,name&_events=true"
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
	// SQL: SELECT "id", "name" FROM "test" WHERE ((("name" IN ('foo', 'bar')) OR ("nick" = 'bar')) AND ("age" < '1')) ORDER BY "age" DESC LIMIT 10 OFFSET 5
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
