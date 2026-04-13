package query

import (
	"testing"
)

// Sink variables to prevent dead-code elimination.
var (
	benchQuery *Query
	benchErr   error
)

func BenchmarkParse(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
		opts  []OptionQuery
	}{
		{
			name:  "simple",
			input: "name=foo",
		},
		{
			name:  "multi_filter",
			input: "name=foo&age=1&status=active",
		},
		{
			name:  "in_operator",
			input: "name=foo,bar,baz",
		},
		{
			name:  "bracket_ops",
			input: "age[gt]=18&age[lt]=65",
		},
		{
			name:  "or_condition",
			input: "name=foo|nick=bar",
		},
		{
			name:  "parentheses",
			input: "(name=foo|name=bar)&age=1",
		},
		{
			name:  "nested_parens",
			input: "(a=1|b=2)&(c=3|(d=4&e=5))",
		},
		{
			name:  "full_query",
			input: "name=foo,bar|nick=bar&age[lt]=1&_sort=-age&_limit=10&_offset=5&_fields=id,name",
		},
		{
			name:  "many_fields",
			input: "_fields=a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p,q,r,s,t",
		},
		{
			name:  "kv_operator",
			input: "meta[kv]=eyJhIjoxLCJiIjoyfQ",
		},
		{
			name:  "url_encoded",
			input: "name=foo%20bar&city=New%20York",
		},
		{
			name:  "no_encoding",
			input: "name=foo&age=1",
		},
		{
			name:  "with_key_operator",
			input: "name=foo&tags=admin,editor",
			opts:  []OptionQuery{WithKeyOperator("name", OperatorLike), WithKeyOperator("tags", OperatorJIn)},
		},
		{
			name:  "with_value_transform",
			input: "name=foo&title=bar",
			opts: []OptionQuery{
				WithKeyOperator("name", OperatorILike),
				WithKeyValueTransform("name", func(v string) string { return "%" + v + "%" }),
				WithKeyOperator("title", OperatorILike),
				WithKeyValueTransform("title", func(v string) string { return "%" + v + "%" }),
			},
		},
		{
			name:  "comma_split",
			input: "name[ilike]=foo,bar,baz&age[gt]=10",
			opts:  []OptionQuery{WithCommaSplit("name")},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchQuery, benchErr = Parse(bm.input, bm.opts...)
			}
		})
	}
}

func BenchmarkSplit(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
		delim byte
	}{
		{
			name:  "simple_ampersand",
			input: "name=foo&age=1&status=active",
			delim: '&',
		},
		{
			name:  "with_parens",
			input: "name=foo|(test=2&age=1)&nick=bar",
			delim: '&',
		},
		{
			name:  "nested_parens",
			input: "name=foo|(test=2&(age=1|age=2))&nick=bar",
			delim: '&',
		},
		{
			name:  "pipe_split",
			input: "name=foo|nick=bar|age=1",
			delim: '|',
		},
	}

	var result []string
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				result = split(bm.input, bm.delim)
			}
		})
	}
	_ = result
}
