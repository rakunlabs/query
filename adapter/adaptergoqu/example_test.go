package adaptergoqu_test

import (
	"fmt"
	"net/url"

	"github.com/doug-martin/goqu/v9"
	"github.com/worldline-go/query"
	"github.com/worldline-go/query/adapter/adaptergoqu"
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
