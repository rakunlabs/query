# Query

Convert rest-API query to goqu's expressions.

```sh
go get github.com/worldline-go/query
```

## Usage

Parse url and extract query parameters with RAW, after that give to `query.Parse` for parsing goqu.

```go
urlStr := "http://example.com?name=foo,bar|nick=bar&age[lt]=1&sort=-age&limit=10&offset=5&fields=id,name"
parsedURL, err := url.Parse(urlStr)
// ...
query, err := query.Parse(parsedURL.RawQuery)
// ...
sql, params, err := query.GoquSelect(goqu.From("test")).ToSQL()
// ...

// SQL: SELECT "id", "name" FROM "test" WHERE ((("name" IN ('foo', 'bar')) OR ("nick" = 'bar')) AND ("age" < '1')) ORDER BY "age" DESC LIMIT 10 OFFSET 5
// Params: []
```

If some value separated by `,` it will be converted to `IN` operator.  
There are a list of `[ ]` operators that can be used in the query string:  
`eq, ne, gt, lt, gte, lte, like, ilike, nlike, nilike, in, nin, is, not`
