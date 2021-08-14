package gosq

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sanggonlee/gosq/ast"
)

// Apply receives a query template and a map of parameters, and replaces
// the expressions in the query template based on the values of the parameters.
// `args` can either be a map of parameters (map[string]interface{}), or a
// custom struct.
// The parameters given in `args` must be accessed by a preceeding dot (.)
// in the template.
// The values of the parameters can be anything, but it will be evaluated as a
// string, using `fmt.Sprintf("%v", v)`.
// The following are the supported syntax in the expressions:
//  - {{ [if] predicate [then] clause }}
//  - {{ [if] predicate [then] clause [else] clause }}
// If you need grammar for a more complex expression, you will need to pass
// the formatted expressions as part of the parameters. Or please file an issue
// if you think it's a common use case.
func Apply(template string, args interface{}) (string, error) {
	if args == nil {
		return template, nil
	}

	argsLookup, err := initArgsLookupTable(args)
	if err != nil {
		return "", err
	}

	tt := ast.BuildTokenTree(template)

	st, err := tt.Parse()
	if err != nil {
		return "", errors.Wrap(err, "building AST")
	}

	if err = st.SubstituteVars(argsLookup); err != nil {
		return "", errors.Wrap(err, "substituting args")
	}

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
