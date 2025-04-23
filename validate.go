package query

import (
	"encoding/json"
	"fmt"
	"slices"
)

type funcType int

const (
	fieldsType funcType = iota
	valuesType
	valueType
)

type Validator struct {
	fields []func(q *Query) error
	values []func(q *Query) error
	value  map[string][]func(q *Query) error
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

// WithMin to validate the minimum of a value.
//   - only usable for 'WithValue'
func WithMin(min json.Number) optionValidateFunc {
	return func(key string, v *Validator, t funcType) error {
		vMinFloat, err := json.Number(min).Float64()
		if err != nil {
			return fmt.Errorf("min value [%s] is not a number", min)
		}

		switch t {
		case valueType:
			v.value[key] = append(v.value[key], func(q *Query) error {
				for _, cmp := range q.Values[key] {
					if cmp.Operator == OperatorEq && cmp.Value != nil {
						cmpStr, ok := cmp.Value.(string)
						if !ok {
							return fmt.Errorf("value [%v] is not a string for number", cmp.Value)
						}

						cmpFloat, err := json.Number(cmpStr).Float64()
						if err != nil {
							return fmt.Errorf("value [%s] is not a number", cmpStr)
						}

						if cmpFloat < vMinFloat {
							return fmt.Errorf("value [%s] is less than min [%s]", cmpStr, min)
						}

						return nil
					}
					if cmp.Operator == OperatorIn && cmp.Value != nil {
						cmpIn, ok := cmp.Value.([]string)
						if !ok {
							return fmt.Errorf("value [%v] is not a slice for number", cmp.Value)
						}
						for _, val := range cmpIn {
							cmpFloat, err := json.Number(val).Float64()
							if err != nil {
								return fmt.Errorf("value [%s] is not a number", val)
							}

							if cmpFloat < vMinFloat {
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
//   - only usable for 'WithValue'
func WithMax(max json.Number) optionValidateFunc {
	return func(key string, v *Validator, t funcType) error {
		vMaxFloat, err := json.Number(max).Float64()
		if err != nil {
			return fmt.Errorf("max value [%s] is not a number", max)
		}

		switch t {
		case valueType:
			v.value[key] = append(v.value[key], func(q *Query) error {
				for _, cmp := range q.Values[key] {
					if cmp.Operator == OperatorEq && cmp.Value != nil {
						cmpStr, ok := cmp.Value.(string)
						if !ok {
							return fmt.Errorf("value [%v] is not a string for number", cmp.Value)
						}

						cmpFloat, err := json.Number(cmpStr).Float64()
						if err != nil {
							return fmt.Errorf("value [%s] is not a number", cmpStr)
						}

						if cmpFloat > vMaxFloat {
							return fmt.Errorf("value [%s] is greater than max [%s]", cmpStr, max)
						}

						return nil
					}
					if cmp.Operator == OperatorIn && cmp.Value != nil {
						cmpIn, ok := cmp.Value.([]string)
						if !ok {
							return fmt.Errorf("value [%v] is not a slice for number", cmp.Value)
						}
						for _, val := range cmpIn {
							cmpFloat, err := json.Number(val).Float64()
							if err != nil {
								return fmt.Errorf("value [%s] is not a number", val)
							}

							if cmpFloat > vMaxFloat {
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
func WithIn(values ...string) optionValidateFunc {
	valuesMap := make(map[string]struct{}, len(values))
	for _, val := range values {
		valuesMap[val] = struct{}{}
	}

	return func(key string, v *Validator, t funcType) error {
		switch t {
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

func WithNotIn(values ...string) optionValidateFunc {
	valuesMap := make(map[string]struct{}, len(values))
	for _, val := range values {
		valuesMap[val] = struct{}{}
	}

	return func(key string, v *Validator, t funcType) error {
		switch t {
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
//   - only usable for 'WithValue'
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
//   - only usable for 'WithValue'
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

func WithNotAllowed() optionValidateFunc {
	return func(key string, v *Validator, t funcType) error {
		switch t {
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
//   - only usable for 'WithValue'
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
//   - only usable for 'WithValue'
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

	return nil
}
