package query_test

import (
	"fmt"
	"testing"

	"github.com/rakunlabs/query"
)

func TestStringToType(t *testing.T) {
	tests := []struct {
		name      string // description of this test case
		s         string
		valueType query.ValueType
		want      any
		wantErr   bool
	}{
		{
			name:      "string to string",
			s:         "12.99",
			valueType: query.ValueTypeNumber,
			want:      "12.99",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := query.StringToType(tt.s, tt.valueType)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("StringToType() failed: %v", gotErr)
				}
				return
			}

			if fmt.Sprintf("%v", got) != tt.want {
				t.Errorf("StringToType() = %v, want %v", got, tt.want)
			}
		})
	}
}
