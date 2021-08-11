package main

import (
	"fmt"

	"github.com/sanggonlee/gosq"
)

func main() {
	q := `
		SELECT
			products.*
			{{ [if] .IncludeReviews [then] ,json_agg(reviews) AS reviews }}
		FROM products
		{{ [if] .IncludeReviews [then] LEFT JOIN reviews ON reviews.product_id =
			{{ [if] .IncludeFoo [then] products.id }}
		}}
		WHERE category = $1
		OFFSET {{ .Offset }}
		{{ .LimitClause }}
	`
	// s := strings.ReplaceAll(q, "\n", " ")
	// s = strings.ReplaceAll(s, "\t", " ")
	// s = strings.ReplaceAll(s, "  ", " ")
	result, err := gosq.Apply(q, map[string]interface{}{
		"IncludeReviews": true,
		"IncludeFoo":     true,
		"Offset":         100,
		"LimitClause":    "LIMIT 10",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}
