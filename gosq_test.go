package gosq_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/sanggonlee/gosq"
)

func TestApply(t *testing.T) {
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
			expected: `SELECT *	FROM products`,
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
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			result, err := gosq.Apply(c.inputTemplate, c.inputArgs)
			if whitespaceNormalized(result) != whitespaceNormalized(c.expected) {
				t.Errorf("Expected %s, got %s", c.expected, result)
			}
			if err != c.expectedError {
				t.Errorf("Expected error %s, got %s", c.expectedError, err)
			}
		})
	}
}

func whitespaceNormalized(s string) string {
	whitespaceRegex := regexp.MustCompile(`\s+`)
	return whitespaceRegex.ReplaceAllString(strings.TrimSpace(s), " ")
}
