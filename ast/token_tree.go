package ast

import (
	"strings"

	"github.com/pkg/errors"
)

// BuildTokenTree builds a tree of tokens from a template string.
func BuildTokenTree(q string) *TokenTree {
	tt := &TokenTree{}
	for _, token := range strings.Fields(q) {
		if token == keywordLanguageStart {
			child := &TokenTree{parent: tt}
			tt.chunks = append(tt.chunks, child)
			tt = child
			continue
		}

		if token == keywordLanguageEnd {
			tt = tt.parent
			continue
		}

		tt.chunks = append(tt.chunks, &literal{token})
	}

	return tt
}

// TokenTree represents a tree of token chunks.
type TokenTree struct {
	chunks []chunk
	parent *TokenTree
}

// Parse parses the TokenTree and returns the AST built from it.
func (tt *TokenTree) Parse() (LanguageNode, error) {
	isIf, err := isIfBlock(tt)
	if err != nil {
		return nil, errors.Wrap(err, "checking an expression for if block")
	}
	if isIf {
		ifBlock, err := parseIfBlock(tt)
		if err != nil {
			return nil, errors.Wrap(err, "parsing an expression for if block")
		}
		return &SyntaxTree{children: []LanguageNode{ifBlock}}, nil
	}

	st := &SyntaxTree{
		children: make([]LanguageNode, 0, len(tt.chunks)),
	}
	for _, chunk := range tt.chunks {
		node, err := chunk.Parse()
		if err != nil {
			return nil, errors.Wrap(err, "building a node from token chunk")
		}
		st.children = append(st.children, node)
	}

	return st, nil
}

type chunk interface {
	Parse() (LanguageNode, error)
}
