package ast

import (
	"testing"

	"github.com/go-test/deep"
)

func TestSyntaxTree_SubstituteVars(t *testing.T) {
	cases := []struct {
		desc               string
		inputSyntaxTree    *SyntaxTree
		inputVars          map[string]interface{}
		isError            bool
		expectedSyntaxTree *SyntaxTree
	}{
		{
			desc:               "Input syntax tree is nil",
			inputSyntaxTree:    nil,
			isError:            false,
			expectedSyntaxTree: nil,
		},
		{
			desc: "Well formatted nested syntax tree",
			inputSyntaxTree: &SyntaxTree{
				children: []LanguageNode{
					&literal{"ABC"},
					&ifBlock{
						predicateExpr: []string{"GHI"},
						expr: &SyntaxTree{
							children: []LanguageNode{
								&literal{"JKL"},
							},
						},
					},
					&literal{"DEF"},
					&SyntaxTree{
						children: []LanguageNode{
							&literal{"GHI"},
							&ifBlock{
								predicateExpr: []string{"false"},
								expr: &SyntaxTree{
									children: []LanguageNode{
										&literal{"JKL"},
									},
								},
							},
						},
					},
					&literal{"GHI"},
					&SyntaxTree{
						children: []LanguageNode{
							&literal{"ABC"},
						},
					},
				},
			},
			inputVars: map[string]interface{}{
				"GHI": true,
				"ABC": "Hello",
			},
			isError: false,
			expectedSyntaxTree: &SyntaxTree{
				children: []LanguageNode{
					&literal{"Hello"},
					&ifBlock{
						predicateExpr: []string{"true"},
						expr: &SyntaxTree{
							children: []LanguageNode{
								&literal{"JKL"},
							},
						},
					},
					&literal{"DEF"},
					&SyntaxTree{
						children: []LanguageNode{
							&literal{"true"},
							&ifBlock{
								predicateExpr: []string{"false"},
								expr: &SyntaxTree{
									children: []LanguageNode{
										&literal{"JKL"},
									},
								},
							},
						},
					},
					&literal{"true"},
					&SyntaxTree{
						children: []LanguageNode{
							&literal{"Hello"},
						},
					},
				},
			},
		},
		{
			desc: "Predicate expression has more than one tokens",
			inputSyntaxTree: &SyntaxTree{
				children: []LanguageNode{
					&literal{"ABC"},
					&ifBlock{
						predicateExpr: []string{"GHI", "JKL"},
						expr: &SyntaxTree{
							children: []LanguageNode{
								&literal{"JKL"},
							},
						},
					},
				},
			},
			inputVars: map[string]interface{}{
				"GHI": true,
				"ABC": "Hello",
			},
			isError: true,
		},
		{
			desc: "Predicate expression has no tokens",
			inputSyntaxTree: &SyntaxTree{
				children: []LanguageNode{
					&literal{"ABC"},
					&ifBlock{
						predicateExpr: []string{},
						expr: &SyntaxTree{
							children: []LanguageNode{
								&literal{"JKL"},
							},
						},
					},
				},
			},
			inputVars: map[string]interface{}{
				"GHI": true,
				"ABC": "Hello",
			},
			isError: true,
		},
		{
			desc: "Predicate expression evaluates to non-boolean",
			inputSyntaxTree: &SyntaxTree{
				children: []LanguageNode{
					&literal{"ABC"},
					&ifBlock{
						predicateExpr: []string{"GHI"},
						expr: &SyntaxTree{
							children: []LanguageNode{
								&literal{"JKL"},
							},
						},
					},
				},
			},
			inputVars: map[string]interface{}{
				"GHI": "haha",
				"ABC": "Hello",
			},
			isError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			err := c.inputSyntaxTree.SubstituteVars(c.inputVars)
			if err != nil {
				if !c.isError {
					t.Fatalf("Unexpected error: %v, got: %v", c.isError, err)
				}
				return
			} else if c.isError {
				t.Fatalf("Expected error but got nil error")
			}
			if diff := deep.Equal(c.inputSyntaxTree, c.expectedSyntaxTree); diff != nil {
				t.Fatalf("Wrong result: %v", diff)
			}
		})
	}
}

func TestSyntaxTree_Evaluate(t *testing.T) {
	cases := []struct {
		desc            string
		inputSyntaxTree *SyntaxTree
		expected        string
	}{
		{
			desc: "Evaluate a nested syntax tree",
			inputSyntaxTree: &SyntaxTree{
				children: []LanguageNode{
					&literal{"ABC"},
					&ifBlock{
						predicateExpr: []string{"true"},
						expr: &SyntaxTree{
							children: []LanguageNode{
								&literal{"JKL"},
							},
						},
					},
					&literal{"GHI"},
					&ifBlock{
						predicateExpr: []string{"false"},
						expr: &SyntaxTree{
							children: []LanguageNode{
								&literal{"DEF"},
							},
						},
					},
					&literal{"MNO"},
				},
			},
			expected: "ABC JKL GHI MNO",
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			output := c.inputSyntaxTree.Evaluate()
			if output != c.expected {
				t.Fatalf("Expected %v but got %v", c.expected, output)
			}
		})
	}
}
