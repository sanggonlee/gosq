//

// gosq is a parsing engine for a simplicity-focused, template-based SQL
// query builder for Go.
//
// It provides a very simple syntax to inject arbitrary conditional query piece.
//
//  q, err := gosq.Apply(`
//    SELECT
//      products.*
//      {{ [if] .IncludeReviews [then] ,json_agg(reviews) AS reviews }}
//    FROM products
//    {{ [if] .IncludeReviews [then] LEFT JOIN reviews ON reviews.product_id = products.id }}
//    WHERE category = $1
//    OFFSET 100
//    LIMIT 10
//  `, map[string]interface{}{
//    IncludeReviews: true,
//  })
//
package gosq // import "github.com/sanggonlee/gosq"
