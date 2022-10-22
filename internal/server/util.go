package server

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// potential terraform-plugin-go convenience funcs
func unmarshalDynamicValueObject(dv *tfprotov6.DynamicValue, ty tftypes.Object) (tftypes.Value, map[string]tftypes.Value, error) {
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

func diagsHaveError(diags []*tfprotov6.Diagnostic) bool {
	for _, diag := range diags {
		if diag != nil && diag.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}

	return false
}

func schemaAsObject(schema *tfprotov6.Schema) tftypes.Object {
	return blockAsObject(schema.Block)
}

func blockAsObject(block *tfprotov6.SchemaBlock) tftypes.Object {
	o := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{},
	}

	for _, b := range block.BlockTypes {
		o.AttributeTypes[b.TypeName] = nestedBlockAsObject(b)
	}

	for _, s := range block.Attributes {
		if s.Type != nil {
			o.AttributeTypes[s.Name] = s.Type
		}
		if s.NestedType != nil {
			o.AttributeTypes[s.Name] = nestedTypeAsObject(s.Name, s.NestedType)
		}
	}

	return o
}

func nestedBlockAsObject(nestedBlock *tfprotov6.SchemaNestedBlock) tftypes.Type {
	switch nestedBlock.Nesting {
	case tfprotov6.SchemaNestedBlockNestingModeSingle:
		return blockAsObject(nestedBlock.Block)
	case tfprotov6.SchemaNestedBlockNestingModeList:
		return tftypes.List{
			ElementType: blockAsObject(nestedBlock.Block),
		}
	}

	panic(fmt.Sprintf("nested block type of %s for %s not supported", nestedBlock.Nesting, nestedBlock.TypeName))
}

func nestedTypeAsObject(name string, nestedType *tfprotov6.SchemaObject) tftypes.Type {
	o := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{},
	}
	for _, s := range nestedType.Attributes {
		if s.Type != nil {
			o.AttributeTypes[s.Name] = s.Type
		}
		if s.NestedType != nil {
			o.AttributeTypes[s.Name] = nestedTypeAsObject(s.Name, s.NestedType)
		}
	}
	switch nestedType.Nesting {
	case tfprotov6.SchemaObjectNestingModeSingle:
		return o
	case tfprotov6.SchemaObjectNestingModeList:
		return tftypes.List{
			ElementType: o,
		}
	}

	panic(fmt.Sprintf("nested type of %s for %s not supported", nestedType.Nesting, name))
}
