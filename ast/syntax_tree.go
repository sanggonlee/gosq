package ast

import (
	"strings"
)

// LanguageNode represents a node in the AST.
type LanguageNode interface {
	//String() string
	SubstituteVars(map[string]interface{}) error
	Evaluate() string
}

// func BuildSyntaxTree(tt *TokenTree) (*SyntaxTree, error) {
// 	isIf, err := isIfBlock(tt)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "checking an expression for if block")
// 	}
// 	if isIf {
// 		ifBlock, err := parseIfBlock(tt)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "parsing an expression for if block")
// 		}
// 		return &SyntaxTree{children: []LanguageNode{ifBlock}}, nil
// 	}

// 	t := &SyntaxTree{
// 		children: make([]LanguageNode, 0, len(tt.chunks)),
// 	}

// 	for _, chunk := range tt.chunks {
// 		if e, ok := chunk.(*TokenTree); ok {
// 			node, err := BuildSyntaxTree(e)
// 			if err != nil {
// 				return nil, errors.Wrap(err, "building a node from an expression")
// 			}
// 			t.children = append(t.children, node)
// 		} else if l, ok := chunk.(*literal); ok {
// 			t.children = append(t.children, l)
// 		}
// 	}

// 	return t, nil
// }

// SyntaxTree is a concrete implementation of the AST.
type SyntaxTree struct {
	children []LanguageNode
}

// // String returns the stringified SyntaxTree.
// func (t *SyntaxTree) String() string {
// 	// TODO: implement
// 	var s string
// 	for _, node := range t.children {
// 		s += node.String() + "\n"
// 	}
// 	return s
// }

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
