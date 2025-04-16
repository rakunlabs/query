# Query

[![License](https://img.shields.io/github/license/worldline-go/query?color=red&style=flat-square)](https://raw.githubusercontent.com/worldline-go/query/main/LICENSE)
[![Coverage](https://img.shields.io/sonar/coverage/worldline-go_query?logo=sonarcloud&server=https%3A%2F%2Fsonarcloud.io&style=flat-square)](https://sonarcloud.io/summary/overall?id=worldline-go_query)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/worldline-go/query/test.yml?branch=main&logo=github&style=flat-square&label=ci)](https://github.com/worldline-go/query/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/worldline-go/query?style=flat-square)](https://goreportcard.com/report/github.com/worldline-go/query)
[![Go PKG](https://raw.githubusercontent.com/worldline-go/guide/main/badge/custom/reference.svg)](https://pkg.go.dev/github.com/worldline-go/query)

Query is an adaptor of http query to expressions. Check adapters to convert it to sql or other expressions.

```sh
go get github.com/worldline-go/query
```

## Usage

Parse url and extract query parameters with RAW, give to `query.Parse` to convert it to expression.  
Use an adapter to convert it to sql or other expressions.

```go
urlStr := "http://example.com?name=foo,bar|nick=bar&age[lt]=1&sort=-age&limit=10&offset=5&fields=id,name"
parsedURL, err := url.Parse(urlStr)
// ...
query, err := query.Parse(parsedURL.RawQuery)
// ...
sql, params, err := adaptergoqu.Select(query, goqu.From("test")).ToSQL()
// ...

// SQL: SELECT "id", "name" FROM "test" WHERE ((("name" IN ('foo', 'bar')) OR ("nick" = 'bar')) AND ("age" < '1')) ORDER BY "age" DESC LIMIT 10 OFFSET 5
// Params: []
```

If some value separated by `,` it will be converted to `IN` operator.  
There are a list of `[ ]` operators that can be used in the query string:  
`eq, ne, gt, lt, gte, lte, like, ilike, nlike, nilike, in, nin, is, not`

### Validation

```go
validator := query.NewValidator(
    WithValue("member", WithRequired(), WithNotIn("O", "P", "S")),
    WithValues(WithIn("age", "test", "member")),
    WithValue("age", WithOperator(OperatorEq), WithNotOperator(OperatorIn)),
    WithField(WithNotAllowed()),
)

// after that use it to validate
err := validator.Validate(query)
if err != nil {
    // handle error
}
```
