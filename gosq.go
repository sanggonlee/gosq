package gosq

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/sanggonlee/gosq/ast"
)

// Apply receives a query template and a map of parameters, and replaces
// the expressions in the query template based on the values of the parameters.
//
// "args" can either be a map of parameters (map[string]interface{}), or a
// custom struct.
//
// The parameters given in "args" must be accessed by a preceeding dot (.)
// in the template.
//
// The values of the parameters can be anything, but it will be evaluated as a
// string, using `fmt.Sprintf("%v", v)`.
//
// The following are the supported syntax in the expressions:
//  - {{ [if] predicate [then] clause }}
//  - {{ [if] predicate [then] clause [else] clause }}
//
// Recursive expressions are supported, as long as they're parts of a [then] or
// [else] clause. For example:
//  {{ [if] predicate [then]
//    {{ [if] predicate [then] clause }}
//  }}
//
// If you need grammar for a more complex expression and you think it's a common
// use case, please file an issue in GitHub.
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
	var _m map[string]interface{}
	var ok bool

	if reflect.ValueOf(args).Kind() == reflect.Struct {
		_m = convertStructToMap(args)
	} else if _m, ok = args.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("unsupported args type: %T", args)
	}

	m := make(map[string]interface{})
	for k, v := range _m {
		m["."+k] = v
		delete(m, k)
	}

	return m, nil
}

func convertStructToMap(args interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	t := reflect.TypeOf(args)
	v := reflect.ValueOf(args)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		m[f.Name] = v.Field(i).Interface()
	}
	return m
}
