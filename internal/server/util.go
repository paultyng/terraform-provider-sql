package server

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

// potential terraform-plugin-go convenience funcs
func unmarshalDynamicValueObject(dv *tfprotov5.DynamicValue, ty tftypes.Object) (tftypes.Value, map[string]tftypes.Value, error) {
	obj, err := dv.Unmarshal(ty)
	if err != nil {
		return tftypes.Value{}, nil, fmt.Errorf("error dv.Unmarshal: %w", err)
	}

	objMap := map[string]tftypes.Value{}
	err = obj.As(&objMap)
	if err != nil {
		return tftypes.Value{}, nil, fmt.Errorf("error obj.As: %w", err)
	}

	return obj, objMap, nil
}

func diagsHaveError(diags []*tfprotov5.Diagnostic) bool {
	for _, diag := range diags {
		if diag != nil && diag.Severity == tfprotov5.DiagnosticSeverityError {
			return true
		}
	}

	return false
}

func schemaAsObject(schema *tfprotov5.Schema) tftypes.Object {
	return blockAsObject(schema.Block)
}

func blockAsObject(block *tfprotov5.SchemaBlock) tftypes.Object {
	o := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{},
	}

	for _, b := range block.BlockTypes {
		o.AttributeTypes[b.TypeName] = nestedBlockAsObject(b)
	}

	for _, s := range block.Attributes {
		o.AttributeTypes[s.Name] = s.Type
	}

	return o
}

func nestedBlockAsObject(nestedBlock *tfprotov5.SchemaNestedBlock) tftypes.Type {
	switch nestedBlock.Nesting {
	case tfprotov5.SchemaNestedBlockNestingModeSingle:
		return blockAsObject(nestedBlock.Block)
	case tfprotov5.SchemaNestedBlockNestingModeList:
		return tftypes.List{
			ElementType: blockAsObject(nestedBlock.Block),
		}
	}

	panic(fmt.Sprintf("nested type of %s for %s not supported", nestedBlock.Nesting, nestedBlock.TypeName))
}
