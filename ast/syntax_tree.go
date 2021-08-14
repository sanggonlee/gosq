package ast

import (
	"strings"
)

// LanguageNode represents a node in the AST.
type LanguageNode interface {
	SubstituteVars(map[string]interface{}) error
	Evaluate() string
}

// SyntaxTree is a concrete implementation of the AST.
type SyntaxTree struct {
	children []LanguageNode
}

// SubstituteVars performs recursive vars substitution on the SyntaxTree.
func (t *SyntaxTree) SubstituteVars(vars map[string]interface{}) error {
	if t == nil {
		return nil
	}
	for _, node := range t.children {
		if err := node.SubstituteVars(vars); err != nil {
			return err
		}
	}
	return nil
}

// Evaluate returns the recursively evaluated SyntaxTree.
func (t *SyntaxTree) Evaluate() string {
	ns := make([]string, 0, len(t.children))
	for _, node := range t.children {
		n := node.Evaluate()
		if n != "" {
			ns = append(ns, n)
		}
	}
	return strings.Join(ns, " ")
}
