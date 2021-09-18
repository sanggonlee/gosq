# gosq

**Go** **S**imple **Q**uery builder.

[![Go Reference](https://pkg.go.dev/badge/github.com/sanggonlee/gosq.svg)](https://pkg.go.dev/github.com/sanggonlee/gosq)
[![Go Report Card](https://goreportcard.com/badge/github.com/sanggonlee/gosq)](https://goreportcard.com/report/github.com/sanggonlee/gosq)

gosq is a parsing engine for a simplicity-focused, template-based SQL
query builder for Go.

It provides syntax to inject arbitrary conditional query piece.

## Usage

```go
q, err := gosq.Compile(`
  SELECT
    products.*
    {{ [if] .IncludeReviews [then] ,json_agg(reviews) AS reviews }}
  FROM products
  {{ [if] .IncludeReviews [then] LEFT JOIN reviews ON reviews.product_id = products.id }}
  WHERE category = $1
  OFFSET 100
  LIMIT 10
`, struct{
  IncludeReviews bool
}{
  IncludeReviews: true,
})
```

or

```go
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
  "IncludeReviews": true,
})
```

Or if you prefer the syntax from [text/template](https://pkg.go.dev/text/template) package:

```go
q, err := gosq.Execute(`
  SELECT
    products.*
    {{if .IncludeReviews}} ,json_agg(reviews) AS reviews {{end}}
  FROM products
  {{if .IncludeReviews}} LEFT JOIN reviews ON reviews.product_id = products.id {{end}}
  WHERE category = $1
  OFFSET 100
  LIMIT 10
`, map[string]interface{}{
  "IncludeReviews": true,
})
```

## Installation

```
go get github.com/sanggonlee/gosq
```

## Documentation

[godoc](https://pkg.go.dev/github.com/sanggonlee/gosq)

## Why?

For me, if I want to run a query, the best way to build such query is to just write the entire thing in the native SQL and pass it to the runner. For example:
```go
func getProducts(db *sql.DB) {
  q := `
    SELECT *
    FROM products
    WHERE category = $1
    OFFSET 100
    LIMIT 10
  `
  category := "electronics"
  rows, err := db.Query(q, category)
  ...
```
It's declarative, easy to understand, and everything is in a single place. What You Give Is What You Get.

But we're living in a dynamic world of requirements, and writing static queries like this will quickly get out of hand as new requirements come in. For example, what if we want to optionally join with a table called "reviews"?

* * *

I could define a clause and optionally concatenate to the query, like this:

```go
func getProducts(db *sql.DB, includeReviews bool) {
  var (
    reviewsColumn string
    reviewsJoinClause string
  )
  if includeReviews {
    reviewsColumn = `, json_agg(reviews) AS reviews`
    reviewsJoinClause = `LEFT JOIN reviews ON reviews.product_id = products.id`
  }

  q := `
    SELECT products.*
    `+reviewsColumn+`
    FROM products
    WHERE category = $1
    `+reviewsJoinClause+`
    OFFSET 100
    LIMIT 10
  `
  category := "electronics"
  rows, err := db.Query(q, category)
```

I don't know about you, but I'm already starting to get uncomfortable here. I can think of several reasons here:
- Dynamically concatenating strings is prone to errors. For example the comma at the start of `, json_agg(reviews) AS reviews` is very easy to miss.
- The query parts are starting to scatter around, and you have to jump between the conditional cases to understand what's going on.
- It's harder to see the overall, cohesive structure of the query. It might not show on this simple example, but as the query gets complex it's often hard to see even the most primary goal of the query.

* * *

There are some SQL builder libraries out there, like [squirrel](https://github.com/Masterminds/squirrel) or [dbr](https://github.com/gocraft/dbr). Maybe they will help?

```go
import sq "github.com/Masterminds/squirrel"

func getProducts(db *sql.DB, includeReviews bool) {
  category := "electronics"
  qb := sq.Select("products.*").
    From("products").
    Where(sq.Eq{"category": category}).
    Offset(100).
    Limit(10)

  if includeReviews {
    qb = qb.Column("json_agg(reviews) AS reviews").
      LeftJoin("reviews ON reviews.product_id = products.id")
  }

  q, args, _ := qb.ToSql()
  rows, err := db.Query(q, args...)
```

That looks a lot better! It's easier to understand, and we've addressed some of the issues we saw earlier, especially around missing commas.

But I'm still not 100% happy. That's too much Go code sneaked into what is really just a SQL query. Still a little hard to understand it as a whole. Also it didn't solve the problem of having to jump around the conditional cases to understand logic. This will get only worse as we have more and more conditional statements.

At the end, what I'd like is a SQL query that can dynamically respond to arbitrary requirements.

* * *

How about some really simple conditionals embedded in a SQL query, rather than SQL query chunks embedded in application code? Something like this, maybe?

```go
func getProducts(includeReviews bool) {
  type queryArgs struct {
    IncludeReviews bool
  }
  q, err := gosq.Compile(`
    SELECT
      products.*
      {{ [if] .IncludeReviews [then] ,json_agg(reviews) AS reviews }}
    FROM products
    {{ [if] .IncludeReviews [then] LEFT JOIN reviews ON reviews.product_id = products.id }}
    WHERE category = $1
    OFFSET 100
    LIMIT 10
  `, queryArgs{
    IncludeReviews: true,
  })
  rows, err := db.Query(q, category)
```

And here we are, `gosq` is born.

Note, this still doesn't address the problem with the preceeding comma. I can't think of a good way to address it in this solution - any suggestion for improvement is welcome.

## Benchmarks

```
BenchmarkExecute-8         57698             19530 ns/op
BenchmarkCompile-8        260319              4570 ns/op
```