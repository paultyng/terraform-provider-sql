package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type dataQuery struct {
	p *provider
}

func (d *dataQuery) Schema(context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Block: &tfprotov5.SchemaBlock{
			Description:     "The `sql_query` datasource allows you to execute a SQL query against the database of your choice.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes: []*tfprotov5.SchemaAttribute{
				{
					Name:            "query",
					Required:        true,
					Description:     "The query to execute. The types in this query will be reflected in the typing of the `result` attribute.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Type:            tftypes.String,
				},
				// {
				// 	Name:            "parameters",
				// 	Optional:        true,
				// 	DescriptionKind: tfprotov5.StringKindMarkdown,
				// 	Type:            tftypes.DynamicPseudoType,
				// },

				{
					Name:     "result",
					Computed: true,
					Description: "The result of the query. This will be a list of objects. Each object will have attributes " +
						"with names that match column names and types that match column types. The exact translation of types " +
						"is dependent upon the database driver.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Type: tftypes.List{
						ElementType: tftypes.DynamicPseudoType,
					},
				},

				// TODO: remove this once its not needed by testing
				{
					Name:       "id",
					Computed:   true,
					Deprecated: true,
					Description: "This attribute is only present for some compatibility issues and should not be used. It " +
						"will be removed in a future version.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Type:            tftypes.String,
				},
			},
		},
	}
}

func (d *dataQuery) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	// TODO: if connected to server, validate query against it?
	return nil, nil
}

func (d *dataQuery) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	var (
		query string
	)
	err := config["query"].As(&query)
	if err != nil {
		return nil, nil, err
	}

	rows, err := d.p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var rowType tftypes.Type
	rowSet := []tftypes.Value{}
	for rows.Next() {
		row, ty, err := d.p.db.valuesForRow(rows)
		if err != nil {
			return nil, []*tfprotov5.Diagnostic{
				{
					Severity: tfprotov5.DiagnosticSeverityError,
					Attribute: &tftypes.AttributePath{
						Steps: []tftypes.AttributePathStep{
							tftypes.AttributeName("result"),
						},
					},
					Summary: fmt.Sprintf("unable to convert value from database: %s", err),
				},
			}, nil
		}

		if rowType == nil {
			rowType = tftypes.Object{
				AttributeTypes: ty,
			}
		}

		rowSet = append(rowSet, tftypes.NewValue(
			rowType,
			row,
		))
	}

	return map[string]tftypes.Value{
		"id":    config["query"],
		"query": config["query"],
		// "parameters": config["parameters"],
		"result": tftypes.NewValue(
			tftypes.List{
				ElementType: rowType,
			},
			rowSet,
		),
	}, nil, nil
}
