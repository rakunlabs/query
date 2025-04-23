package query

import (
	"fmt"
	"math/big"
	"slices"
)

type funcType int

const (
	fieldsType funcType = iota
	valuesType
	valueType
	offsetType
	limitType
	sortType
)

type Validator struct {
	fields []func(q *Query) error
	values []func(q *Query) error
	value  map[string][]func(q *Query) error

	offset []func(q *Query) error
	limit  []func(q *Query) error
	sort   []func(q *Query) error
}

type (
	optionValidateFunc func(string, *Validator, funcType) error
	OptionValidate     func(string, ...optionValidateFunc)
	OptionValidateSet  func(v *Validator) error
)

func NewValidator(opts ...OptionValidateSet) (*Validator, error) {
	v := &Validator{
		value: make(map[string][]func(q *Query) error),
	}

	for _, opt := range opts {
		if err := opt(v); err != nil {
			return nil, err
		}
	}

	return v, nil
}

func WithField(opts ...optionValidateFunc) OptionValidateSet {
	return func(v *Validator) error {
		for _, opt := range opts {
			if err := opt("", v, fieldsType); err != nil {
				return err
			}
		}

		return nil
	}
}

func WithValue(key string, opts ...optionValidateFunc) OptionValidateSet {
	return func(v *Validator) error {
		for _, opt := range opts {
			if err := opt(key, v, valueType); err != nil {
				return err
			}
		}

		return nil
	}
}

func WithValues(opts ...optionValidateFunc) OptionValidateSet {
	return func(v *Validator) error {
		for _, opt := range opts {
			if err := opt("", v, valuesType); err != nil {
				return err
			}
		}

		return nil
	}
}

func WithOffset(opts ...optionValidateFunc) OptionValidateSet {
	return func(v *Validator) error {
		for _, opt := range opts {
			if err := opt("", v, offsetType); err != nil {
				return err
			}
		}

		return nil
	}
}

func WithLimit(opts ...optionValidateFunc) OptionValidateSet {
	return func(v *Validator) error {
		for _, opt := range opts {
			if err := opt("", v, limitType); err != nil {
				return err
			}
		}

		return nil
	}
}

func WithSort(opts ...optionValidateFunc) OptionValidateSet {
	return func(v *Validator) error {
		for _, opt := range opts {
			if err := opt("", v, sortType); err != nil {
				return err
			}
		}

		return nil
	}
}

// ////////////////////////////////////////////////////////////////////////////
// ////////////////////////////////////////////////////////////////////////////

// WithMin to validate the minimum of a value.
//   - Usable for 'WithValue', WithSort', 'WithLimit'
func WithMin(min string) optionValidateFunc {
	return func(key string, v *Validator, t funcType) error {
		vMinBig, ok := new(big.Float).SetString(min)
		if !ok {
			return fmt.Errorf("min value [%s] is not a valid number", min)
		}

		switch t {
		case offsetType:
			v.offset = append(v.offset, func(q *Query) error {
				if q.Offset != nil {
					if new(big.Float).SetUint64(*q.Offset).Cmp(vMinBig) < 0 {
						return fmt.Errorf("offset [%d] is less than min [%s]", q.Offset, min)
					}
				}

				return nil
			})
		case limitType:
			v.limit = append(v.limit, func(q *Query) error {
				if q.Limit != nil {
					if new(big.Float).SetUint64(*q.Limit).Cmp(vMinBig) < 0 {
						return fmt.Errorf("limit [%d] is less than min [%s]", q.Limit, min)
					}
				}
				return nil
			})
		case valueType:
			v.value[key] = append(v.value[key], func(q *Query) error {
				for _, cmp := range q.Values[key] {
					if cmp.Operator == OperatorEq && cmp.Value != nil {
						cmpStr, ok := cmp.Value.(string)
						if !ok {
							return fmt.Errorf("value [%v] is not a string for number", cmp.Value)
						}

						cmpBig, ok := new(big.Float).SetString(cmpStr)
						if !ok {
							return fmt.Errorf("value [%s] is not a number", cmpStr)
						}

						if cmpBig.Cmp(vMinBig) < 0 {
							return fmt.Errorf("value [%s] is less than min [%s]", cmpStr, min)
						}
					}
					if cmp.Operator == OperatorIn && cmp.Value != nil {
						cmpIn, ok := cmp.Value.([]string)
						if !ok {
							return fmt.Errorf("value [%v] is not a slice for number", cmp.Value)
						}
						for _, val := range cmpIn {
							cmpBig, ok := new(big.Float).SetString(val)
							if !ok {
								return fmt.Errorf("value [%s] is not a number", val)
							}

							if cmpBig.Cmp(vMinBig) < 0 {
								return fmt.Errorf("value [%s] is less than min [%s]", val, min)
							}
						}
					}
				}

				return nil
			})
		}

		return nil
	}
}

// WithMax to validate the maximum of a value.
//   - Usable for 'WithValue', 'WithLimit', 'WithOffset'
func WithMax(max string) optionValidateFunc {
	return func(key string, v *Validator, t funcType) error {
		vMaxBig, ok := new(big.Float).SetString(max)
		if !ok {
			return fmt.Errorf("max value [%s] is not a valid number", max)
		}

		switch t {
		case offsetType:
			v.offset = append(v.offset, func(q *Query) error {
				if q.Offset != nil {
					if new(big.Float).SetUint64(*q.Offset).Cmp(vMaxBig) > 0 {
						return fmt.Errorf("offset [%d] is greater than max [%s]", q.Offset, max)
					}
				}

				return nil
			})
		case limitType:
			v.limit = append(v.limit, func(q *Query) error {
				if q.Limit != nil {
					if new(big.Float).SetUint64(*q.Limit).Cmp(vMaxBig) > 0 {
						return fmt.Errorf("limit [%d] is greater than max [%s]", q.Limit, max)
					}
				}
				return nil
			})
		case valueType:
			v.value[key] = append(v.value[key], func(q *Query) error {
				for _, cmp := range q.Values[key] {
					if cmp.Operator == OperatorEq && cmp.Value != nil {
						cmpStr, ok := cmp.Value.(string)
						if !ok {
							return fmt.Errorf("value [%v] is not a string for number", cmp.Value)
						}

						cmpBig, ok := new(big.Float).SetString(cmpStr)
						if !ok {
							return fmt.Errorf("value [%s] is not a number", cmpStr)
						}

						if cmpBig.Cmp(vMaxBig) > 0 {
							return fmt.Errorf("value [%s] is greater than max [%s]", cmpStr, max)
						}
					}
					if cmp.Operator == OperatorIn && cmp.Value != nil {
						cmpIn, ok := cmp.Value.([]string)
						if !ok {
							return fmt.Errorf("value [%v] is not a slice for number", cmp.Value)
						}
						for _, val := range cmpIn {
							cmpBig, ok := new(big.Float).SetString(val)
							if !ok {
								return fmt.Errorf("value [%s] is not a number", val)
							}

							if cmpBig.Cmp(vMaxBig) > 0 {
								return fmt.Errorf("value [%s] is greater than max [%s]", val, max)
							}
						}
					}
				}

				return nil
			})
		}

		return nil
	}
}

// WithIn checks if the value is in the list of values.
//   - Usable for 'WithValue', 'WithSort', 'WithValues', 'WithFields'
func WithIn(values ...string) optionValidateFunc {
	valuesMap := make(map[string]struct{}, len(values))
	for _, val := range values {
		valuesMap[val] = struct{}{}
	}

	return func(key string, v *Validator, t funcType) error {
		switch t {
		case sortType:
			v.sort = append(v.sort, func(q *Query) error {
				for _, cmp := range q.Sort {
					if _, ok := valuesMap[cmp.Field]; !ok {
						return fmt.Errorf("value [%s] is not in %v", cmp.Field, values)
					}
				}

				return nil
			})
		case valuesType:
			v.values = append(v.values, func(q *Query) error {
				for vKey := range q.Values {
					if _, ok := valuesMap[vKey]; !ok {
						return fmt.Errorf("value [%s] is not in %v", vKey, values)
					}
				}

				return nil
			})
		case fieldsType:
			v.fields = append(v.fields, func(q *Query) error {
				for _, cmp := range q.Select {
					if _, ok := valuesMap[cmp]; !ok {
						return fmt.Errorf("value [%s] is not in %v", cmp, values)
					}
				}

				return nil
			})
		case valueType:
			v.value[key] = append(v.value[key], func(q *Query) error {
				for _, cmp := range q.Values[key] {
					if cmp.Operator == OperatorEq && cmp.Value != nil {
						cmpStr, ok := cmp.Value.(string)
						if !ok {
							return fmt.Errorf("value [%v] is not a string for in", cmp.Value)
						}

						if _, ok := valuesMap[cmpStr]; !ok {
							return fmt.Errorf("value [%s] is not in %v", cmpStr, values)
						}

						return nil
					}

					if cmp.Operator == OperatorIn && cmp.Value != nil {
						cmpIn, ok := cmp.Value.([]string)
						if !ok {
							return fmt.Errorf("value [%v] is not a string for in", cmp.Value)
						}

						for _, val := range cmpIn {
							if _, ok := valuesMap[val]; !ok {
								return fmt.Errorf("value [%s] is not in %v", val, values)
							}
						}

						return nil
					}
				}

				return nil
			})
		}

		return nil
	}
}

// WithNotIn checks if the value is not in the list of values.
//   - Usable for 'WithValue', 'WithSort', 'WithValues', 'WithFields'
func WithNotIn(values ...string) optionValidateFunc {
	valuesMap := make(map[string]struct{}, len(values))
	for _, val := range values {
		valuesMap[val] = struct{}{}
	}

	return func(key string, v *Validator, t funcType) error {
		switch t {
		case sortType:
			v.sort = append(v.sort, func(q *Query) error {
				for _, cmp := range q.Sort {
					if _, ok := valuesMap[cmp.Field]; ok {
						return fmt.Errorf("value [%s] is in %v", cmp.Field, values)
					}
				}

				return nil
			})
		case valuesType:
			v.values = append(v.values, func(q *Query) error {
				for vKey := range q.Values {
					if _, ok := valuesMap[vKey]; ok {
						return fmt.Errorf("value [%s] is in %v", vKey, values)
					}
				}

				return nil
			})
		case fieldsType:
			v.fields = append(v.fields, func(q *Query) error {
				for _, cmp := range q.Select {
					if _, ok := valuesMap[cmp]; ok {
						return fmt.Errorf("value [%s] is in %v", cmp, values)
					}

					return nil
				}

				return nil
			})
		case valueType:
			v.value[key] = append(v.value[key], func(q *Query) error {
				for _, cmp := range q.Values[key] {
					if cmp.Operator == OperatorEq && cmp.Value != nil {
						cmpStr, ok := cmp.Value.(string)
						if !ok {
							return fmt.Errorf("value [%v] is not a string for in", cmp.Value)
						}

						if _, ok := valuesMap[cmpStr]; ok {
							return fmt.Errorf("value [%s] is in %v", cmpStr, values)
						}

						return nil
					}

					if cmp.Operator == OperatorIn && cmp.Value != nil {
						cmpIn, ok := cmp.Value.([]string)
						if !ok {
							return fmt.Errorf("value [%v] is not a string for in", cmp.Value)
						}

						for _, val := range cmpIn {
							if _, ok := valuesMap[val]; ok {
								return fmt.Errorf("value [%s] is in %v", val, values)
							}
						}

						return nil
					}
				}

				return nil
			})
		}

		return nil
	}
}

// WithNotEmpty to validate the value is not empty.
//   - Usable for 'WithValue'
func WithNotEmpty() optionValidateFunc {
	return func(key string, v *Validator, t funcType) error {
		switch t {
		case valueType:
			v.value[key] = append(v.value[key], func(q *Query) error {
				values := q.GetValues(key)
				if len(values) == 0 {
					return fmt.Errorf("value [%s] is empty", key)
				}
				if slices.Contains(values, "") {
					return fmt.Errorf("value [%s] is empty", key)
				}

				return nil
			})
		}

		return nil
	}
}

// WithRequired to validate the value is required.
//   - Usable for 'WithValue'
func WithRequired() optionValidateFunc {
	return func(key string, v *Validator, t funcType) error {
		switch t {
		case valueType:
			v.value[key] = append(v.value[key], func(q *Query) error {
				values := q.GetValues(key)
				if len(values) > 0 {
					return nil
				}

				return fmt.Errorf("value [%s] is required", key)
			})
		}

		return nil
	}
}

// WithNotAllowed to validate the value is not allowed.
//   - Usable for 'WithValue', 'WithOffset', 'WithLimit', 'WithSort', 'WithValues', 'WithFields'
func WithNotAllowed() optionValidateFunc {
	return func(key string, v *Validator, t funcType) error {
		switch t {
		case offsetType:
			v.offset = append(v.offset, func(q *Query) error {
				if q.Offset != nil {
					return fmt.Errorf("offset is not allowed")
				}

				return nil
			})
		case limitType:
			v.limit = append(v.limit, func(q *Query) error {
				if q.Limit != nil {
					return fmt.Errorf("limit is not allowed")
				}

				return nil
			})
		case sortType:
			v.sort = append(v.sort, func(q *Query) error {
				if len(q.Sort) > 0 {
					return fmt.Errorf("sort is not allowed")
				}

				return nil
			})
		case valuesType:
			v.values = append(v.values, func(q *Query) error {
				if len(q.Values) > 0 {
					return fmt.Errorf("values is not allowed")
				}

				return nil
			})
		case fieldsType:
			v.fields = append(v.fields, func(q *Query) error {
				if len(q.Select) > 0 {
					return fmt.Errorf("fields is not allowed")
				}

				return nil
			})
		case valueType:
			v.value[key] = append(v.value[key], func(q *Query) error {
				if len(q.Values[key]) > 0 {
					return fmt.Errorf("value [%s] is not allowed", key)
				}

				return nil
			})
		}

		return nil
	}
}

// WithOperator to validate the operator is allowed.
//   - Usable for 'WithValue'
func WithOperator(operators ...OperatorCmpType) optionValidateFunc {
	operatorsMap := make(map[OperatorCmpType]struct{}, len(operators))
	for _, op := range operators {
		operatorsMap[op] = struct{}{}
	}

	return func(key string, v *Validator, t funcType) error {
		switch t {
		case valueType:
			v.value[key] = append(v.value[key], func(q *Query) error {
				for _, cmp := range q.Values[key] {
					if _, ok := operatorsMap[cmp.Operator]; !ok {
						return fmt.Errorf("operator [%s] is not allowed", cmp.Operator)
					}
				}

				return nil
			})
		}

		return nil
	}
}

// WithNotOperator to validate the operator is not allowed.
//   - Usable for 'WithValue'
func WithNotOperator(operators ...OperatorCmpType) optionValidateFunc {
	operatorsMap := make(map[OperatorCmpType]struct{}, len(operators))
	for _, op := range operators {
		operatorsMap[op] = struct{}{}
	}

	return func(key string, v *Validator, t funcType) error {
		switch t {
		case valueType:
			v.value[key] = append(v.value[key], func(q *Query) error {
				for _, cmp := range q.Values[key] {
					if _, ok := operatorsMap[cmp.Operator]; ok {
						return fmt.Errorf("operator [%s] is not allowed", cmp.Operator)
					}
				}

				return nil
			})
		}

		return nil
	}
}

// ///////////////////////////////////////////////////////////////////
// ///////////////////////////////////////////////////////////////////

func (q *Query) Validate(v *Validator) error {
	if v == nil {
		return nil
	}

	for key, f := range v.value {
		for _, fn := range f {
			if err := fn(q); err != nil {
				return fmt.Errorf("validate [%s]: %w", key, err)
			}
		}
	}

	for _, fn := range v.fields {
		if err := fn(q); err != nil {
			return fmt.Errorf("validate fields: %w", err)
		}
	}

	for _, fn := range v.values {
		if err := fn(q); err != nil {
			return fmt.Errorf("validate values: %w", err)
		}
	}

	for _, fn := range v.offset {
		if err := fn(q); err != nil {
			return fmt.Errorf("validate offset: %w", err)
		}
	}

	for _, fn := range v.limit {
		if err := fn(q); err != nil {
			return fmt.Errorf("validate limit: %w", err)
		}
	}

	for _, fn := range v.sort {
		if err := fn(q); err != nil {
			return fmt.Errorf("validate sort: %w", err)
		}
	}

	return nil
}
