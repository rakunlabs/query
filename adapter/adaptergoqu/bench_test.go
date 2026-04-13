package adaptergoqu_test

import (
	"net/url"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/rakunlabs/query"
	"github.com/rakunlabs/query/adapter/adaptergoqu"
)

// Sink variables to prevent dead-code elimination.
var (
	benchSQL    string
	benchParams []any
	benchE2EErr error
)

func BenchmarkE2E(b *testing.B) {
	benchmarks := []struct {
		name  string
		url   string
		opts  []query.OptionQuery
		aOpts []adaptergoqu.Option
	}{
		{
			name: "simple",
			url:  "http://example.com?name=foo",
		},
		{
			name: "full",
			url:  "http://example.com?name=foo,bar|nick=bar&age[lt]=1&_sort=-age&_limit=10&_offset=5&_fields=id,name",
		},
		{
			name: "complex_nested",
			url:  "http://example.com?(name=foo|name=bar)&(age[gt]=18|age[lt]=5)&status=active&_sort=-age,name&_limit=20&_offset=0&_fields=id,name,age,status",
		},
		{
			name: "kv_and_jin",
			url:  "http://example.com?meta[kv]=eyJhIjoxfQ&tags[jin]=admin,editor&status=active",
		},
		{
			name: "many_filters",
			url:  "http://example.com?a=1&b=2&c=3&d=4&e=5&f=6&g=7&h=8&_sort=-a,b&_limit=50&_fields=a,b,c,d,e",
		},
		{
			name:  "parameterized",
			url:   "http://example.com?name=foo,bar|nick=bar&age[lt]=1&_sort=-age&_limit=10&_offset=5&_fields=id,name",
			aOpts: []adaptergoqu.Option{adaptergoqu.WithParameterized(true)},
		},
		{
			name:  "literal",
			url:   "http://example.com?name=foo,bar|nick=bar&age[lt]=1&_sort=-age&_limit=10&_offset=5&_fields=id,name",
			aOpts: []adaptergoqu.Option{adaptergoqu.WithParameterized(false)},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			parsedURL, err := url.Parse(bm.url)
			if err != nil {
				b.Fatal(err)
			}
			rawQuery := parsedURL.RawQuery

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				q, err := query.Parse(rawQuery, bm.opts...)
				if err != nil {
					b.Fatal(err)
				}
				benchSQL, benchParams, benchE2EErr = adaptergoqu.Select(q, goqu.From("test"), bm.aOpts...).ToSQL()
			}
		})
	}
}

// BenchmarkSelectOnly isolates the goqu adapter from the parse step.
// The query is pre-parsed; only the Select + ToSQL path is measured.
func BenchmarkSelectOnly(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
		opts  []query.OptionQuery
		aOpts []adaptergoqu.Option
	}{
		{
			name:  "simple",
			input: "name=foo",
		},
		{
			name:  "full",
			input: "name=foo,bar|nick=bar&age[lt]=1&_sort=-age&_limit=10&_offset=5&_fields=id,name",
		},
		{
			name:  "complex_nested",
			input: "(name=foo|name=bar)&(age[gt]=18|age[lt]=5)&status=active&_sort=-age,name&_limit=20&_offset=0&_fields=id,name,age,status",
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			q, err := query.Parse(bm.input, bm.opts...)
			if err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				benchSQL, benchParams, benchE2EErr = adaptergoqu.Select(q, goqu.From("test"), bm.aOpts...).ToSQL()
			}
		})
	}
}
