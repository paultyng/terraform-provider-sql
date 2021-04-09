package server

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/go-argmapper"
)

// TypeName represents the resource / data source type name as passed by Terraform
type TypeName string

func callResourceFactory(f *argmapper.Func, provider Provider, typeName TypeName, converter *argmapper.Func) argmapper.Result {
	opts := []argmapper.Arg{
		argmapper.Typed(typeName),
		argmapper.Typed(provider),

		argmapper.ConverterFunc(converter),
	}

	opts = append(opts, namedArgsFromStruct(provider)...)

	return f.Call(opts...)
}

func namedArgsFromStruct(v interface{}) []argmapper.Arg {
	// TODO: maybe eventually there is an easier way to pass a bunch of args from a struct, see
	// https://github.com/hashicorp/go-argmapper/issues/3

	rv := reflect.ValueOf(v)
	sv := structValueOf(rv)
	if sv.Kind() == reflect.Invalid {
		panic(fmt.Sprintf("only struct or pointer to struct types are supported in FromStruct, got %T", v))
	}
	st := sv.Type()

	var args []argmapper.Arg
	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)
		fv := sv.Field(i)

		if f.PkgPath != "" {
			// skip unexported
			continue
		}
		args = append(args, argmapper.Named(f.Name, fv.Interface()))

		// TODO: if not a primitive type only?
		args = append(args, argmapper.Typed(fv.Interface()))
	}

	return args
}

func structValueOf(rv reflect.Value) reflect.Value {
	if k := rv.Kind(); k != reflect.Struct && k != reflect.Ptr {
		return reflect.Value{}
	}

	sv := rv
	if sv.Kind() == reflect.Ptr {
		// unwrap ptr
		sv = sv.Elem()
		if sv.Kind() != reflect.Struct {
			return reflect.Value{}
		}
	}

	return sv
}
