package query

import (
	"testing"
)

var benchValidateErr error

func BenchmarkValidate(b *testing.B) {
	benchmarks := []struct {
		name      string
		input     string
		validator func() *Validator
	}{
		{
			name:  "pass_simple",
			input: "age=25&name=foo&_limit=10&_offset=0&_sort=-age&_fields=id,name",
			validator: func() *Validator {
				v, _ := NewValidator(
					WithValues(WithIn("age", "name")),
					WithValue("age", WithMin("0"), WithMax("200")),
					WithLimit(WithMin("1"), WithMax("100")),
					WithSort(WithIn("age", "name")),
					WithField(WithIn("id", "name", "age")),
				)
				return v
			},
		},
		{
			name:  "fail_early_value",
			input: "age=500&name=foo",
			validator: func() *Validator {
				v, _ := NewValidator(
					WithValue("age", WithMax("200")),
				)
				return v
			},
		},
		{
			name:  "pass_many_rules",
			input: "member=X&age=50&_limit=20&_offset=5&_sort=age&_fields=id,name",
			validator: func() *Validator {
				v, _ := NewValidator(
					WithValue("member", WithRequired(), WithNotEmpty(), WithIn("X", "Y", "Z")),
					WithValue("age", WithMin("0"), WithMax("200"), WithOperator(OperatorEq)),
					WithValues(WithIn("age", "member")),
					WithLimit(WithMin("1"), WithMax("100")),
					WithOffset(WithMin("0"), WithMax("1000")),
					WithSort(WithIn("age", "name")),
					WithField(WithIn("id", "name", "age")),
				)
				return v
			},
		},
		{
			name:  "no_validator",
			input: "name=foo&age=1",
			validator: func() *Validator {
				return nil
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			q, err := Parse(bm.input)
			if err != nil {
				b.Fatal(err)
			}
			v := bm.validator()

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				benchValidateErr = q.Validate(v)
			}
		})
	}
}
