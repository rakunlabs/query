package query

type optionQuery struct {
	DefaultOffset *uint64
	DefaultLimit  *uint64

	Value map[string]ExpressionCmp
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

func WithExpressionCmp(key string, value ExpressionCmp) OptionQuery {
	return func(o *optionQuery) {
		if o.Value == nil {
			o.Value = make(map[string]ExpressionCmp)
		}

		o.Value[key] = value
	}
}
