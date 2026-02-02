package query

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

func (q *Query) MarshalText() ([]byte, error) {
	// convert to url.Values
	values := bytes.Buffer{}

	if len(q.Select) > 0 {
		values.WriteString(keyFields)
		values.WriteString("=")
		values.WriteString(strings.Join(q.Select, ","))
	}

	if len(q.Sort) > 0 {
		if values.Len() > 0 {
			values.WriteString("&")
		}

		sortParts := make([]string, 0, len(q.Sort))
		for _, s := range q.Sort {
			decs := ""
			if s.Desc {
				decs = ":desc"
			}

			sortParts = append(sortParts, fmt.Sprintf("%s%s", s.Field, decs))
		}

		values.WriteString(keySort)
		values.WriteString("=")
		values.WriteString(strings.Join(sortParts, ","))
	}

	if q.Limit != nil {
		if values.Len() > 0 {
			values.WriteString("&")
		}

		values.WriteString(keyLimit)
		values.WriteString("=")
		values.WriteString(strconv.FormatUint(*q.Limit, 10))
	}

	if q.Offset != nil {
		if values.Len() > 0 {
			values.WriteString("&")
		}

		values.WriteString(keyOffset)
		values.WriteString("=")
		values.WriteString(strconv.FormatUint(*q.Offset, 10))
	}

	for _, expr := range q.Where {
		if values.Len() > 0 {
			values.WriteString("&")
		}

		values.WriteString(expr.String())
	}

	return values.Bytes(), nil
}

// Base64URLEncode encodes v using base64 URL encoding without padding.
//   - This useful for non-URL-safe strings that need to be included in URLs.
//   - OperatorKV understands this encoding.
func Base64URLEncode(v []byte) string {
	return base64.RawURLEncoding.EncodeToString(v)
}

// Base64URLDecode decodes a base64 URL encoded string without padding.
func Base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
