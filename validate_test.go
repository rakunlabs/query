package query

import (
	"net/url"
	"testing"
)

func TestQuery_Validate(t *testing.T) {
	type fields struct {
		URL string
	}
	type args struct {
		opts []OptionValidateSet
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "invalid age",
			fields: fields{
				URL: "http://example.com?age=10000&fields=name,age&sort=age,-name&offset=10&limit=20",
			},
			args: args{
				opts: []OptionValidateSet{
					WithValue("age", WithMin("0"), WithMax("200")),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid age",
			fields: fields{
				URL: "http://example.com?age=10000,500&fields=name,age&sort=age,-name&offset=10&limit=20",
			},
			args: args{
				opts: []OptionValidateSet{
					WithValue("age", WithMin("0"), WithMax("200")),
				},
			},
			wantErr: true,
		},
		{
			name: "not a member",
			fields: fields{
				URL: "http://example.com?member=Z",
			},
			args: args{
				opts: []OptionValidateSet{
					WithValue("member", WithIn("X", "Y")),
				},
			},
			wantErr: true,
		},
		{
			name: "not a member list",
			fields: fields{
				URL: "http://example.com?member=X,Z",
			},
			args: args{
				opts: []OptionValidateSet{
					WithValue("member", WithIn("X", "Y")),
				},
			},
			wantErr: true,
		},
		{
			name: "not in",
			fields: fields{
				URL: "http://example.com?member=X,Z",
			},
			args: args{
				opts: []OptionValidateSet{
					WithValue("member", WithNotIn("O", "P", "S")),
				},
			},
			wantErr: false,
		},
		{
			name: "required",
			fields: fields{
				URL: "http://example.com?member=X",
			},
			args: args{
				opts: []OptionValidateSet{
					WithValue("member", WithRequired()),
				},
			},
			wantErr: false,
		},
		{
			name: "required",
			fields: fields{
				URL: "http://example.com?fields=name,age&sort=age,-name&offset=10&limit=20",
			},
			args: args{
				opts: []OptionValidateSet{
					WithField("fields", WithIn("X")),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlParsed, err := url.Parse(tt.fields.URL)
			if err != nil {
				t.Fatalf("failed to parse URL: %v", err)
			}

			q, err := Parse(urlParsed.RawQuery)
			if err != nil {
				t.Fatalf("failed to parse query: %v", err)
			}

			validate, err := NewValidator(tt.args.opts...)
			if err != nil {
				t.Fatalf("failed to create validator: %v", err)
			}

			if err := q.Validate(validate); (err != nil) != tt.wantErr {
				t.Errorf("Query.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
