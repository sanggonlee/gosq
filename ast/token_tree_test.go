package ast

import (
	"testing"

	"github.com/go-test/deep"
)

func TestTokenTree_BuildTokenTree(t *testing.T) {
	cases := []struct {
		desc     string
		input    string
		expected *TokenTree
	}{
		{
			desc:  "Flat free",
			input: `ABC DEF GHI JKL MNO`,
			expected: &TokenTree{
				chunks: []chunk{
					&literal{"ABC"},
					&literal{"DEF"},
					&literal{"GHI"},
					&literal{"JKL"},
					&literal{"MNO"},
				},
			},
		},
		{
			desc: "Nested expressions",
			input: `ABC {{ DEF GHI }} JKL
				{{ MNO PQR {{ [if] STU [then] VWX {{ YZ }} }} }}`,
			expected: &TokenTree{
				chunks: []chunk{
					&literal{"ABC"},
					&TokenTree{
						chunks: []chunk{
							&literal{"DEF"},
							&literal{"GHI"},
						},
					},
					&literal{"JKL"},
					&TokenTree{
						chunks: []chunk{
							&literal{"MNO"},
							&literal{"PQR"},
							&TokenTree{
								chunks: []chunk{
									&literal{"[if]"},
									&literal{"STU"},
									&literal{"[then]"},
									&literal{"VWX"},
									&TokenTree{
										chunks: []chunk{
											&literal{"YZ"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			output := BuildTokenTree(c.input)
			if diff := deep.Equal(c.expected, output); diff != nil {
				t.Errorf("Wrong result: %v", diff)
			}
		})
	}
}

func TestBuildTokenTree_Parse(t *testing.T) {
	cases := []struct {
		desc           string
		inputTokenTree *TokenTree
		isError        bool
		expected       LanguageNode
	}{
		{
			desc:           "Empty tree",
			inputTokenTree: &TokenTree{},
			isError:        true,
		},
		{
			desc: "Token tree with if blocks",
			inputTokenTree: &TokenTree{
				chunks: []chunk{
					&literal{"ABC"},
					&TokenTree{
						chunks: []chunk{
							&literal{"DEF"},
							&literal{"GHI"},
						},
					},
					&literal{"JKL"},
					&TokenTree{
						chunks: []chunk{
							&literal{"MNO"},
							&literal{"PQR"},
							&TokenTree{
								chunks: []chunk{
									&literal{"[if]"},
									&literal{"STU"},
									&literal{"[then]"},
									&literal{"VWX"},
									&TokenTree{
										chunks: []chunk{
											&literal{"YZ"},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &SyntaxTree{
				children: []LanguageNode{
					&literal{"ABC"},
					&SyntaxTree{
						children: []LanguageNode{
							&literal{"DEF"},
							&literal{"GHI"},
						},
					},
					&literal{"JKL"},
					&SyntaxTree{
						children: []LanguageNode{
							&literal{"MNO"},
							&literal{"PQR"},
							&SyntaxTree{
								children: []LanguageNode{
									&ifBlock{
										predicateExpr: []string{"STU"},
										expr: &SyntaxTree{
											children: []LanguageNode{
												&literal{"VWX"},
												&SyntaxTree{
													children: []LanguageNode{
														&literal{"YZ"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			output, err := c.inputTokenTree.Parse()
			if err != nil {
				if !c.isError {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			} else if c.isError {
				t.Errorf("Expected error, got nil")
			}

			if diff := deep.Equal(c.expected, output); diff != nil {
				t.Errorf("Wrong result: %v", diff)
			}
		})
	}
}
