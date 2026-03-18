package query

type optionQuery struct {
	DefaultOffset *uint64
	DefaultLimit  *uint64

	Value          map[string]*ExpressionCmp
	Skip           map[string]struct{}
	SkipUnderscore bool

	UnderscorePrefix *bool

	KeyType map[string]ValueType
}

type OptionQuery func(*optionQuery)

// WithDefaultOffset sets the default offset value.
func WithDefaultOffset(offset uint64) OptionQuery {
	return func(o *optionQuery) {
		o.DefaultOffset = &offset
	}
}

// WithDefaultLimit sets the default limit value.
func WithDefaultLimit(limit uint64) OptionQuery {
	return func(o *optionQuery) {
		o.DefaultLimit = &limit
	}
}

// WithExpressionCmp sets the expression comparison for a given key.
func WithExpressionCmp(key string, value *ExpressionCmp) OptionQuery {
	return func(o *optionQuery) {
		if o.Value == nil {
			o.Value = make(map[string]*ExpressionCmp)
		}

		o.Value[key] = value
	}
}

// WithSkipExpressionCmp sets the keys to be skipped in the query.
func WithSkipExpressionCmp(key ...string) OptionQuery {
	return func(o *optionQuery) {
		if o.Skip == nil {
			o.Skip = make(map[string]struct{})
		}

		for _, k := range key {
			o.Skip[k] = struct{}{}
		}
	}
}

// WithSkipUnderscore sets whether to skip keys starting with underscore.
//   - Default is true.
func WithSkipUnderscore(v bool) OptionQuery {
	return func(o *optionQuery) {
		o.SkipUnderscore = v
	}
}

// WithUnderscorePrefix sets whether the special query keys use an underscore prefix.
//   - Default is true: _limit, _offset, _sort, _fields.
//   - When set to false: limit, offset, sort, fields.
func WithUnderscorePrefix(v bool) OptionQuery {
	return func(o *optionQuery) {
		o.UnderscorePrefix = &v
	}
}

func WithKeyType(key string, valueType ValueType) OptionQuery {
	return func(o *optionQuery) {
		if o.KeyType == nil {
			o.KeyType = make(map[string]ValueType)
		}

		o.KeyType[key] = valueType
	}
}
