package adaptergoqu

import "github.com/rakunlabs/query"

type option struct {
	Edit          func(q *query.Query) *query.Query
	Rename        map[string]string
	DefaultSelect []string
	Parameterized bool
}

type Option func(*option)

func WithEdit(edit func(q *query.Query) *query.Query) Option {
	return func(o *option) {
		o.Edit = edit
	}
}

func WithRename(rename map[string]string) Option {
	return func(o *option) {
		o.Rename = rename
	}
}

func WithDefaultSelect(selects ...string) Option {
	return func(o *option) {
		o.DefaultSelect = selects
	}
}

func WithParameterized(parameterized bool) Option {
	return func(o *option) {
		o.Parameterized = parameterized
	}
}
