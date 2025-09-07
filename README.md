# Query

[![License](https://img.shields.io/github/license/rakunlabs/query?color=red&style=flat-square)](https://raw.githubusercontent.com/rakunlabs/query/main/LICENSE)
[![Coverage](https://img.shields.io/sonar/coverage/rakunlabs_query?logo=sonarcloud&server=https%3A%2F%2Fsonarcloud.io&style=flat-square)](https://sonarcloud.io/summary/overall?id=rakunlabs_query)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/rakunlabs/query/test.yml?branch=main&logo=github&style=flat-square&label=ci)](https://github.com/rakunlabs/query/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/rakunlabs/query?style=flat-square)](https://goreportcard.com/report/github.com/rakunlabs/query)
[![Go PKG](https://raw.githubusercontent.com/rakunlabs/guide/main/badge/custom/reference.svg)](https://pkg.go.dev/github.com/rakunlabs/query)

Query is an adaptor of http query to expressions. Check adapters to convert it to sql or other expressions.

```sh
go get github.com/rakunlabs/query
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

`limit` and `offset` are used to limit the number of rows returned. _0_ limit means no limit.  
`fields` is used to select the fields to be returned, comma separated.  
`sort` is used to sort the result set, can be prefixed with `-` to indicate descending order and comma separated to indicate multiple fields.  
`[]` empty operator means `in` operator.

### Validation

`query.WithField` is used to validate the field names.
- `WithNotIn` is used to validate the field names that are not allowed.
- `WithIn` is used to validate the field names that are allowed.
- `WithNotAllowed` is used to validate the field names totally not allowed.

`query.WithValues` is used to validate the values of the fields.
- `WithNotIn` is used to validate the values that are not allowed.
- `WithIn` is used to validate the values that are allowed.
- `WithNotAllowed` is used to validate the values totally not allowed.

`query.WithValue` is used to validate the values of the fields.
- `WithRequired` is used to validate the values that are required.
- `WithNotEmpty` is used to validate the values that are not empty.
- `WithNotIn` is used to validate the values that are not allowed.
- `WithIn` is used to validate the values that are allowed.
- `WithNotAllowed` is used to validate the value that are not allowed.
- `WithOperator` is used to validate the operator that is allowed.
- `WithNotOperator` is used to validate the operator that is not allowed.
- `WithMax` is used to validate the maximum of value, value must be a number.
- `WithMin` is used to validate the minimum of value, value must be a number.

`query.WithOffset` is used to validate the offset value.
- `WithMax` is used to validate the maximum of offset, value must be a number.
- `WithMin` is used to validate the minimum of offset, value must be a number.
- `WithNotAllowed` is used to validate the offset value that are not allowed.

`query.WithLimit` is used to validate the limit value.
- `WithMax` is used to validate the maximum of limit, value must be a number.
- `WithMin` is used to validate the minimum of limit, value must be a number.
- `WithNotAllowed` is used to validate the limit value that are not allowed.

`query.WithSort` is used to validate the sort value.
- `WithNotIn` is used to validate the sort value that are not allowed.
- `WithIn` is used to validate the sort value that are allowed.
- `WithNotAllowed` is used to validate the sort value that are not allowed.

Example of validation:

```go
validator := query.NewValidator(
    WithValue("member", query.WithRequired(), query.WithNotIn("O", "P", "S")),
    WithValues(query.WithIn("age", "test", "member")),
    WithValue("age", query.WithOperator(OperatorEq), query.WithNotOperator(OperatorIn)),
    WithField(query.WithNotAllowed()),
)

// after that use it to validate
err := validator.Validate(query)
if err != nil {
    // handle error
}

// or pass with when parsing
query, err := query.Parse(rawQuery, query.WithValidator(validator))
// ...
```
