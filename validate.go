package query

import (
	"encoding/json"
	"fmt"
)

type validator struct {
	Values map[string][]func(q *Query) error
}

type (
	optionValidateFunc func(string, *validator) error
	OptionValidate     func(string, ...optionValidateFunc)
	OptionValidateSet  func(v *validator) error
)

func NewValidator(opts ...OptionValidateSet) (*validator, error) {
	v := &validator{
		Values: make(map[string][]func(q *Query) error),
	}

	for _, opt := range opts {
		if err := opt(v); err != nil {
			return nil, err
		}
	}

	return v, nil
}

func WithValue(key string, opts ...optionValidateFunc) OptionValidateSet {
	return func(v *validator) error {
		for _, opt := range opts {
			if err := opt(key, v); err != nil {
				return err
			}
		}

		return nil
	}
}

func WithMin(min json.Number) optionValidateFunc {
	return func(key string, v *validator) error {
		vMinFloat, err := json.Number(min).Float64()
		if err != nil {
			return fmt.Errorf("min value %s is not a number", min)
		}

		v.Values[key] = append(v.Values[key], func(q *Query) error {
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

		return nil
	}
}

func WithMax(max json.Number) optionValidateFunc {
	return func(key string, v *validator) error {
		vMaxFloat, err := json.Number(max).Float64()
		if err != nil {
			return fmt.Errorf("max value %s is not a number", max)
		}

		v.Values[key] = append(v.Values[key], func(q *Query) error {
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

		return nil
	}
}

func WithIn(values ...string) optionValidateFunc {
	valuesMap := make(map[string]struct{}, len(values))
	for _, val := range values {
		valuesMap[val] = struct{}{}
	}

	return func(key string, v *validator) error {
		v.Values[key] = append(v.Values[key], func(q *Query) error {
			for _, cmp := range q.Values[key] {
				if cmp.Operator == OperatorEq && cmp.Value != nil {
					cmpStr, ok := cmp.Value.(string)
					if !ok {
						return fmt.Errorf("value [%v] is not a string for in", cmp.Value)
					}

					if _, ok := valuesMap[cmpStr]; !ok {
						return fmt.Errorf("value %s is not in %v", cmpStr, values)
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
							return fmt.Errorf("value %s is not in %v", val, values)
						}
					}

					return nil
				}
			}

			return nil
		})

		return nil
	}
}

func WithNotIn(values ...string) optionValidateFunc {
	valuesMap := make(map[string]struct{}, len(values))
	for _, val := range values {
		valuesMap[val] = struct{}{}
	}

	return func(key string, v *validator) error {
		v.Values[key] = append(v.Values[key], func(q *Query) error {
			for _, cmp := range q.Values[key] {
				if cmp.Operator == OperatorEq && cmp.Value != nil {
					cmpStr, ok := cmp.Value.(string)
					if !ok {
						return fmt.Errorf("value [%v] is not a string for in", cmp.Value)
					}

					if _, ok := valuesMap[cmpStr]; ok {
						return fmt.Errorf("value %s is in %v", cmpStr, values)
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
							return fmt.Errorf("value %s is in %v", val, values)
						}
					}

					return nil
				}
			}

			return nil
		})

		return nil
	}
}

func WithNotEmpty() optionValidateFunc {
	return func(key string, v *validator) error {
		v.Values[key] = append(v.Values[key], func(q *Query) error {
			for _, cmp := range q.Values[key] {
				if cmp.Operator == OperatorEq && cmp.Value != nil {
					cmpStr, ok := cmp.Value.(string)
					if !ok {
						return fmt.Errorf("value [%v] is not a string for not empty", cmp.Value)
					}

					if cmpStr == "" {
						return fmt.Errorf("value %s is empty", cmp)
					}

					return nil
				}
			}

			return nil
		})

		return nil
	}
}

func WithRequired() optionValidateFunc {
	return func(key string, v *validator) error {
		v.Values[key] = append(v.Values[key], func(q *Query) error {
			for _, cmp := range q.Values[key] {
				if cmp.Operator == OperatorEq && cmp.Value != nil {
					return nil
				}
			}

			return fmt.Errorf("value %s is required", key)
		})

		return nil
	}
}

func (q *Query) Validate(v *validator) error {
	if v == nil {
		return fmt.Errorf("validate is nil")
	}

	for key, f := range v.Values {
		for _, fn := range f {
			if err := fn(q); err != nil {
				return fmt.Errorf("validate %s: %w", key, err)
			}
		}
	}

	return nil
}
