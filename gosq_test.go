package gosq_test

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/sanggonlee/gosq"
	"github.com/stretchr/testify/assert"
)

func TestCompile(t *testing.T) {
	cases := []struct {
		desc          string
		inputTemplate string
		inputArgs     interface{}
		expected      string
		expectedError error
	}{
		{
			desc:          "No args",
			inputTemplate: `SELECT * FROM products`,
			inputArgs:     nil,
			expected:      `SELECT *	FROM products`,
		},
		{
			desc: "Simple case of falsey substitute from map",
			inputTemplate: `
				SELECT
					products.*
					{{ [if] .IncludeReviews [then] ,json_agg(reviews) AS reviews }}
				FROM products
				{{ [if] .IncludeReviews [then] LEFT JOIN reviews ON reviews.product_id = products.id }}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: map[string]interface{}{
				"IncludeReviews": false,
			},
			expected: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Simple case of truthy substitute from map",
			inputTemplate: `
				SELECT
					products.*
					{{ [if] .IncludeReviews [then] ,json_agg(reviews) AS reviews }}
				FROM products
				{{ [if] .IncludeReviews [then] LEFT JOIN reviews ON reviews.product_id = products.id }}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: map[string]interface{}{
				"IncludeReviews": true,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Recursive truthy expression",
			inputTemplate: `
				SELECT
					products.*
					{{
						[if] .IncludeReviews [then] ,json_agg(reviews) AS reviews
						{{
							[if] .IncludeCount [then] ,count(reviews) AS num_reviews
						}}
					}}
				FROM products
				{{ [if] .IncludeReviews [then] LEFT JOIN reviews ON reviews.product_id = products.id }}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: map[string]interface{}{
				"IncludeReviews": true,
				"IncludeCount":   true,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
					,count(reviews) AS num_reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Recursive falsey expression",
			inputTemplate: `
				SELECT
					products.*
					{{
						[if] .IncludeReviews [then] ,json_agg(reviews) AS reviews
						{{
							[if] .IncludeCount [then] ,count(reviews) AS num_reviews
						}}
					}}
				FROM products
				{{ [if] .IncludeReviews [then] LEFT JOIN reviews ON reviews.product_id = products.id }}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: map[string]interface{}{
				"IncludeReviews": true,
				"IncludeCount":   false,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Simple truthy if-then-else clause",
			inputTemplate: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT {{ [if] .GetMany [then] 100 [else] 10 }}
			`,
			inputArgs: map[string]interface{}{
				"GetMany": true,
			},
			expected: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT 100
			`,
		},
		{
			desc: "Simple falsey if-then-else clause",
			inputTemplate: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT {{ [if] .GetMany [then] 100 [else] 10 }}
			`,
			inputArgs: map[string]interface{}{
				"GetMany": false,
			},
			expected: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Simple case of truthy substitute from struct",
			inputTemplate: `
				SELECT
					products.*
					{{ [if] .IncludeReviews [then] ,json_agg(reviews) AS reviews }}
				FROM products
				{{ [if] .IncludeReviews [then] LEFT JOIN reviews ON reviews.product_id = products.id }}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: struct {
				IncludeReviews bool
			}{
				IncludeReviews: true,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Recursive falsey expression from struct",
			inputTemplate: `
				SELECT
					products.*
					{{
						[if] .IncludeReviews [then] ,json_agg(reviews) AS reviews
						{{
							[if] .IncludeCount [then] ,count(reviews) AS num_reviews
						}}
					}}
				FROM products
				{{ [if] .IncludeReviews [then] LEFT JOIN reviews ON reviews.product_id = products.id }}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: struct {
				IncludeReviews bool
				IncludeCount   bool
			}{
				IncludeReviews: true,
				IncludeCount:   false,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			result, err := gosq.Compile(c.inputTemplate, c.inputArgs)
			if whitespaceNormalized(result) != whitespaceNormalized(c.expected) {
				t.Errorf("Expected %s, got %s", c.expected, result)
			}
			if err != c.expectedError {
				t.Errorf("Expected error %s, got %s", c.expectedError, err)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	cases := []struct {
		desc          string
		inputTemplate string
		inputArgs     interface{}
		expected      string
		expectedError error
	}{
		{
			desc:          "No args",
			inputTemplate: `SELECT * FROM products`,
			inputArgs:     nil,
			expected:      `SELECT *	FROM products`,
		},
		{
			desc: "Simple case of falsey substitute from map",
			inputTemplate: `
				SELECT
					products.*
					{{if .IncludeReviews}} ,json_agg(reviews) AS reviews {{end}}
				FROM products
				{{if .IncludeReviews}} LEFT JOIN reviews ON reviews.product_id = products.id {{end}}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: map[string]interface{}{
				"IncludeReviews": false,
			},
			expected: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Simple case of truthy substitute from map",
			inputTemplate: `
				SELECT
					products.*
					{{if .IncludeReviews}} ,json_agg(reviews) AS reviews {{end}}
				FROM products
				{{if .IncludeReviews}} LEFT JOIN reviews ON reviews.product_id = products.id {{end}}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: map[string]interface{}{
				"IncludeReviews": true,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Recursive truthy expression",
			inputTemplate: `
				SELECT
					products.*
					{{if .IncludeReviews}}
						,json_agg(reviews) AS reviews
						{{if .IncludeCount}} ,count(reviews) AS num_reviews {{end}}
					{{end}}
				FROM products
				{{if .IncludeReviews}} LEFT JOIN reviews ON reviews.product_id = products.id {{end}}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: map[string]interface{}{
				"IncludeReviews": true,
				"IncludeCount":   true,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
					,count(reviews) AS num_reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Recursive falsey expression",
			inputTemplate: `
				SELECT
					products.*
					{{if .IncludeReviews}} ,json_agg(reviews) AS reviews
						{{if .IncludeCount}} ,count(reviews) AS num_reviews {{end}}
					{{end}}
				FROM products
				{{if .IncludeReviews}} LEFT JOIN reviews ON reviews.product_id = products.id {{end}}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: map[string]interface{}{
				"IncludeReviews": true,
				"IncludeCount":   false,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Simple truthy if-then-else clause",
			inputTemplate: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT {{if .GetMany}} 100 {{else}} 10 {{end}}
			`,
			inputArgs: map[string]interface{}{
				"GetMany": true,
			},
			expected: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT 100
			`,
		},
		{
			desc: "Simple falsey if-then-else clause",
			inputTemplate: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT {{if .GetMany}} 100 {{else}} 10 {{end}}
			`,
			inputArgs: map[string]interface{}{
				"GetMany": false,
			},
			expected: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Simple case of truthy substitute from struct",
			inputTemplate: `
				SELECT
					products.*
					{{if .IncludeReviews}} ,json_agg(reviews) AS reviews {{end}}
				FROM products
				{{if .IncludeReviews}} LEFT JOIN reviews ON reviews.product_id = products.id {{end}}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: struct {
				IncludeReviews bool
			}{
				IncludeReviews: true,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Recursive falsey expression from struct",
			inputTemplate: `
				SELECT
					products.*
					{{if .IncludeReviews}} ,json_agg(reviews) AS reviews
						{{if .IncludeCount}} ,count(reviews) AS num_reviews {{end}}
					{{end}}
				FROM products
				{{if .IncludeReviews}} LEFT JOIN reviews ON reviews.product_id = products.id {{end}}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: struct {
				IncludeReviews bool
				IncludeCount   bool
			}{
				IncludeReviews: true,
				IncludeCount:   false,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			result, err := gosq.Execute(c.inputTemplate, c.inputArgs)
			if whitespaceNormalized(result) != whitespaceNormalized(c.expected) {
				t.Errorf("Expected %s, got %s", c.expected, result)
			}
			if err != nil {
				assert.ErrorContains(t, err, c.expectedError.Error())
			}
		})
	}
}

func TestExecuteWithOption(t *testing.T) {
	cases := []struct {
		desc          string
		inputTemplate string
		inputArgs     interface{}
		option        string
		expected      string
		expectedError error
	}{
		{
			desc:          "No args",
			inputTemplate: `SELECT * FROM products`,
			inputArgs:     nil,
			expected:      `SELECT *	FROM products`,
		},
		{
			desc: "Simple case of falsey substitute from map",
			inputTemplate: `
				SELECT
					products.*
					{{if .IncludeReviews}} ,json_agg(reviews) AS reviews {{end}}
				FROM products
				{{if .IncludeReviews}} LEFT JOIN reviews ON reviews.product_id = products.id {{end}}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: map[string]interface{}{
				"IncludeReviews": false,
			},
			expected: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Simple case of truthy substitute from map",
			inputTemplate: `
				SELECT
					products.*
					{{if .IncludeReviews}} ,json_agg(reviews) AS reviews {{end}}
				FROM products
				{{if .IncludeReviews}} LEFT JOIN reviews ON reviews.product_id = products.id {{end}}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: map[string]interface{}{
				"IncludeReviews": true,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Recursive truthy expression",
			inputTemplate: `
				SELECT
					products.*
					{{if .IncludeReviews}}
						,json_agg(reviews) AS reviews
						{{if .IncludeCount}} ,count(reviews) AS num_reviews {{end}}
					{{end}}
				FROM products
				{{if .IncludeReviews}} LEFT JOIN reviews ON reviews.product_id = products.id {{end}}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: map[string]interface{}{
				"IncludeReviews": true,
				"IncludeCount":   true,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
					,count(reviews) AS num_reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Recursive falsey expression",
			inputTemplate: `
				SELECT
					products.*
					{{if .IncludeReviews}} ,json_agg(reviews) AS reviews
						{{if .IncludeCount}} ,count(reviews) AS num_reviews {{end}}
					{{end}}
				FROM products
				{{if .IncludeReviews}} LEFT JOIN reviews ON reviews.product_id = products.id {{end}}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: map[string]interface{}{
				"IncludeReviews": true,
				"IncludeCount":   false,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Simple truthy if-then-else clause",
			inputTemplate: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT {{if .GetMany}} 100 {{else}} 10 {{end}}
			`,
			inputArgs: map[string]interface{}{
				"GetMany": true,
			},
			expected: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT 100
			`,
		},
		{
			desc: "Simple falsey if-then-else clause",
			inputTemplate: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT {{if .GetMany}} 100 {{else}} 10 {{end}}
			`,
			inputArgs: map[string]interface{}{
				"GetMany": false,
			},
			expected: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Simple case of truthy substitute from struct",
			inputTemplate: `
				SELECT
					products.*
					{{if .IncludeReviews}} ,json_agg(reviews) AS reviews {{end}}
				FROM products
				{{if .IncludeReviews}} LEFT JOIN reviews ON reviews.product_id = products.id {{end}}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: struct {
				IncludeReviews bool
			}{
				IncludeReviews: true,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Recursive falsey expression from struct",
			inputTemplate: `
				SELECT
					products.*
					{{if .IncludeReviews}} ,json_agg(reviews) AS reviews
						{{if .IncludeCount}} ,count(reviews) AS num_reviews {{end}}
					{{end}}
				FROM products
				{{if .IncludeReviews}} LEFT JOIN reviews ON reviews.product_id = products.id {{end}}
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
			inputArgs: struct {
				IncludeReviews bool
				IncludeCount   bool
			}{
				IncludeReviews: true,
				IncludeCount:   false,
			},
			expected: `
				SELECT
					products.*
					,json_agg(reviews) AS reviews
				FROM products
				LEFT JOIN reviews ON reviews.product_id = products.id
				WHERE category = $1
				OFFSET 100
				LIMIT 10
			`,
		},
		{
			desc: "Missing variable, no value  is returned",
			inputTemplate: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT {{.Missing}}
			`,
			inputArgs: map[string]interface{}{
				"Any": 0,
			},
			option: "missingkey=default",
			expected: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT <no value>
			`,
		},
		{
			desc: "Missing variable, zero value is returned",
			inputTemplate: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT {{.Missing}}
			`,
			inputArgs: map[string]int{
				"Any": 0,
			},
			option: "missingkey=zero",
			expected: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT 0
			`,
		},
		{
			desc: "Missing variable, error is returned",
			inputTemplate: `
				SELECT
					products.*
				FROM products
				WHERE category = $1
				OFFSET 100
				LIMIT {{.Missing}}
			`,
			inputArgs: map[string]interface{}{
				"Any": 0,
			},
			option:        "missingkey=error",
			expectedError: errors.New("executing template"),
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			result, err := gosq.ExecuteWithOption(c.inputTemplate, c.inputArgs, c.option)
			if whitespaceNormalized(result) != whitespaceNormalized(c.expected) {
				t.Errorf("Expected %s, got %s", c.expected, result)
			}
			if err != nil {
				assert.ErrorContains(t, err, c.expectedError.Error())
			}
		})
	}
}

func whitespaceNormalized(s string) string {
	whitespaceRegex := regexp.MustCompile(`\s+`)
	return whitespaceRegex.ReplaceAllString(strings.TrimSpace(s), " ")
}

var benchmarkInputTmpl = `
SELECT
	products.*
	{{if .IncludeReviews}}
		,json_agg(reviews) AS reviews
		{{if .IncludeCount}} ,count(reviews) AS num_reviews {{end}}
	{{end}}
FROM products
{{if .IncludeReviews}} LEFT JOIN reviews ON reviews.product_id = products.id {{end}}
WHERE category = $1
OFFSET 100
LIMIT 10
`
var benchmarkInputArgs = map[string]interface{}{
	"IncludeReviews": true,
	"IncludeCount":   true,
}

func BenchmarkExecute(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = gosq.Execute(benchmarkInputTmpl, benchmarkInputArgs)
	}
}

func BenchmarkCompile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = gosq.Compile(benchmarkInputTmpl, benchmarkInputArgs)
	}
}
