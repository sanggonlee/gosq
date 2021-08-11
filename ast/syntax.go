package ast

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const (
	keywordIf            = "[if]"
	keywordThen          = "[then]"
	keywordElse          = "[else]"
	keywordLanguageStart = "{{"
	keywordLanguageEnd   = "}}"
)

func isKeyword(s string) bool {
	return s == keywordIf ||
		s == keywordThen ||
		s == keywordElse ||
		s == keywordLanguageStart ||
		s == keywordLanguageEnd
}

// literal represents a token of a literal string in the template.
type literal struct {
	s string
}

// String is a string representation of literal.
func (l *literal) String() string {
	if l == nil {
		return ""
	}
	return l.s
}

// Parse converts the literal to the LanguageNode interface value.
func (l *literal) Parse() (LanguageNode, error) {
	return l, nil
}

// SubstituteVars performs var substitution on this literal instance.
func (l *literal) SubstituteVars(vars map[string]interface{}) error {
	if l == nil {
		return nil
	}
	v, ok := vars[string(l.s)]
	if ok {
		l.s = fmt.Sprintf("%v", v)
	}
	return nil
}

// Evaluate returns the evaluated value of the literal.
func (l *literal) Evaluate() string {
	if l == nil {
		return ""
	}
	return l.s
}

// ifBlock represents a parsed syntax state of an [if] block.
type ifBlock struct {
	predicateExpr []string
	expr          *SyntaxTree
}

// SubstituteVars performs var substitution on the predicate and expression of
// this ifBlock instance.
func (ib *ifBlock) SubstituteVars(vars map[string]interface{}) error {
	if ib == nil {
		return nil
	}
	if len(ib.predicateExpr) == 0 {
		return errors.New("predicate expression not found")
	} else if len(ib.predicateExpr) > 1 {
		return errors.New("multi-token expression predicate is not supported yet")
	}
	v, ok := vars[ib.predicateExpr[0]]
	if ok {
		ib.predicateExpr = []string{fmt.Sprintf("%v", v)}
	}
	if boolExpr := strings.ToLower(ib.predicateExpr[0]); boolExpr != "true" && boolExpr != "false" {
		return errors.New("predicate must be a boolean expression")
	}

	if err := ib.expr.SubstituteVars(vars); err != nil {
		return err
	}

	return nil
}

// Evaluate returns the evaluated value of this ifBlock's expression if
// predicate evaluates to true, otherwise returns an empty string
func (ib *ifBlock) Evaluate() string {
	if ib == nil {
		return ""
	}
	if strings.ToLower(ib.predicateExpr[0]) == "true" {
		return ib.expr.Evaluate()
	}
	return ""
}

// isIfBlock checks if the TokenTree is analyzed to an if block.
func isIfBlock(tt *TokenTree) (bool, error) {
	if len(tt.chunks) == 0 {
		return false, errors.New("expression with empty chunks")
	}

	if maybeIf, ok := tt.chunks[0].(*literal); !ok || maybeIf.String() != keywordIf {
		return false, nil
	}

	var (
		thenIndex int
	)
	for i, chunk := range tt.chunks {
		literalChunk, isLiteral := chunk.(*literal)
		if i == 1 {
			if !isLiteral || isKeyword(string(literalChunk.String())) {
				return false, errors.New("[if] must be followed by a predicate")
			}
		}
		if literalChunk.String() == keywordThen {
			thenIndex = i
		}
	}
	if thenIndex == 0 {
		return false, errors.New("[if] must be followed by a [then] clause")
	}

	return true, nil
}

// parseIfBlock parses the TokenTree and returns the parsed ifBlock.
// It assumes the TokenTree is a valid if block (make sure to call isIfBlock first).
// If it's not, returns an error.
func parseIfBlock(tt *TokenTree) (*ifBlock, error) {
	ib := &ifBlock{}
	var (
		isIf   bool = true
		isThen bool
	)
	for i, chunk := range tt.chunks {
		if i == 0 {
			continue
		}

		exprChunk, isExpr := chunk.(*TokenTree)
		if isExpr {
			// Assumes it's a then/else clause
			node, err := exprChunk.Parse()
			if err != nil {
				return nil, errors.Wrap(err, "parsing an expression for if block")
			}

			ib.expr.children = append(ib.expr.children, node)
		}

		literalChunk, isLiteral := chunk.(*literal)
		if isLiteral {
			if literalChunk.String() == keywordThen {
				isIf = false
				isThen = true
				ib.expr = &SyntaxTree{}
			} else if isIf {
				ib.predicateExpr = append(ib.predicateExpr, literalChunk.String())
			} else if isThen {
				ib.expr.children = append(ib.expr.children, literalChunk)
			}
		}
	}

	return ib, nil
}
