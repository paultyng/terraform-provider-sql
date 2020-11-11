package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

func unquoteIfQuoted(value interface{}) (string, error) {
	var bytes []byte

	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return "", fmt.Errorf("could not convert value '%+v' to byte array of type '%T'",
			value, value)
	}

	// If the amount is quoted, strip the quotes
	if len(bytes) > 2 && bytes[0] == '"' && bytes[len(bytes)-1] == '"' {
		bytes = bytes[1 : len(bytes)-1]
	}
	return string(bytes), nil
}

// potential terraform-plugin-go convenience funcs

func diagsHaveError(diags []*tfprotov5.Diagnostic) bool {
	for _, diag := range diags {
		if diag != nil && diag.Severity == tfprotov5.DiagnosticSeverityError {
			return true
		}
	}

	return false
}

func schemaAsObject(schema *tfprotov5.Schema) tftypes.Object {
	o := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{},
	}

	// TODO: block types

	for _, s := range schema.Block.Attributes {
		o.AttributeTypes[s.Name] = s.Type
	}

	return o
}
