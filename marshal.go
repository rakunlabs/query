package query

import (
	"bytes"
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
