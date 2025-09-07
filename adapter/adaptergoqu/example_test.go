package adaptergoqu_test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/rakunlabs/query"
	"github.com/rakunlabs/query/adapter/adaptergoqu"
)

func ExampleParseQuery() {
	urlStr := "http://example.com?name=foo,bar|nick=bar&age[lt]=1&sort=-age&limit=10&offset=5&fields=id,name"
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}

	query, err := query.Parse(parsedURL.RawQuery)
	if err != nil {
		panic(err)
	}

	sql, params, err := adaptergoqu.Select(query, goqu.From("test")).ToSQL()
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

func BenchmarkParseQuery(b *testing.B) {
	urlStr := "http://example.com?name=foo,bar|nick=bar&age[lt]=1&sort=-age&limit=10&offset=5&fields=id,name"

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
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
