package gosq

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sanggonlee/gosq/ast"
)

func Apply(template string, args interface{}) (string, error) {
	if args == nil {
		return template, nil
	}

	argsLookup, err := initArgsLookupTable(args)
	if err != nil {
		return "", err
	}

	// First pass - Build a tree of tokens
	tt := ast.BuildTokenTree(template)

	// First walk - Convert the tree into an AST by parsing the expression languages
	st, err := tt.Parse()
	if err != nil {
		return "", errors.Wrap(err, "building AST")
	}

	// Second walk - Subsitute the arguments in the AST
	if err = st.SubstituteVars(argsLookup); err != nil {
		return "", errors.Wrap(err, "substituting args")
	}

	// Third walk - Evaluate the expressions
	return st.Evaluate(), nil
}

func initArgsLookupTable(args interface{}) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	switch t := args.(type) {
	case map[string]interface{}:
		for k, v := range args.(map[string]interface{}) {
			m["."+k] = v
			delete(m, k)
		}
	default:
		return m, fmt.Errorf("unsupported args type: %s", t)
	}

	return m, nil
}
