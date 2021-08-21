//

/*

gosq is a parsing engine for a simplicity-focused, template-based SQL
query builder for Go.

It provides a very simple syntax to inject arbitrary conditional query piece.

 q, err := gosq.Compile(`
   SELECT
     products.*
     {{ [if] .IncludeReviews [then] ,json_agg(reviews) AS reviews }}
   FROM products
   {{ [if] .IncludeReviews [then] LEFT JOIN reviews ON reviews.product_id = products.id }}
   WHERE category = $1
   OFFSET 100
   LIMIT 10
 `, map[string]interface{}{
   IncludeReviews: true,
 })


Limitations:
 - The predicate (expression between [if] and [then]) must be a one word: either true, false or
   a parameter that evaluates to a boolean. There's no support to evaluate an arbitrary predicate
   expression yet.
 - gosq does not validate nor executes the query itself. The only thing it does is
   build the query in string out of a template.

*/
package gosq // import "github.com/sanggonlee/gosq"
