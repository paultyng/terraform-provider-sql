package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"

	"github.com/paultyng/terraform-provider-sql/internal/migration"
)

type resourceMigrate struct {
	resourceMigrateCommon
}

func (r *resourceMigrate) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Block: &tfprotov5.SchemaBlock{
			BlockTypes: []*tfprotov5.SchemaNestedBlock{
				{
					TypeName: "migration",
					Nesting:  tfprotov5.SchemaNestedBlockNestingModeList,
					Block: &tfprotov5.SchemaBlock{
						Attributes: []*tfprotov5.SchemaAttribute{
							{
								Name:            "id",
								Required:        true,
								Description:     "Identifier can be any string to help identifying the migration in the source.",
								DescriptionKind: tfprotov5.StringKindMarkdown,
								Type:            tftypes.String,
							},
							{
								Name:            "up",
								Required:        true,
								Description:     "The query to run when applying this migration.",
								DescriptionKind: tfprotov5.StringKindMarkdown,
								Type:            tftypes.String,
							},
							{
								Name:            "down",
								Required:        true,
								Description:     "The query to run when undoing this migration.",
								DescriptionKind: tfprotov5.StringKindMarkdown,
								Type:            tftypes.String,
							},
						},
					},
				},
			},
			Attributes: []*tfprotov5.SchemaAttribute{
				completeMigrationsAttribute(),
				deprecatedIDAttribute(),
			},
		},
	}
}

func (r *resourceMigrate) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	migrationValue := config["migration"]

	if !migrationValue.IsFullyKnown() {
		return nil, nil
	}

	migrations, err := migration.FromListValue(migrationValue)
	if err != nil {
		return nil, err
	}

	if len(migrations) == 0 {
		return []*tfprotov5.Diagnostic{
			{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "At least one migration is required.",
			},
		}, nil
	}

	ids := map[string]bool{}
	for i, m := range migrations {
		if strings.TrimSpace(m.ID) == "" {
			return []*tfprotov5.Diagnostic{
				{
					Severity: tfprotov5.DiagnosticSeverityError,
					Summary:  "ID cannot be empty.",
					Attribute: &tftypes.AttributePath{
						Steps: []tftypes.AttributePathStep{
							tftypes.AttributeName("migration"),
							tftypes.ElementKeyInt(i),
							tftypes.AttributeName("id"),
						},
					},
				},
			}, nil
		}
		if ids[m.ID] {
			return []*tfprotov5.Diagnostic{
				{
					Severity: tfprotov5.DiagnosticSeverityError,
					Summary:  fmt.Sprintf("Duplicate ID value of %q.", m.ID),
					Attribute: &tftypes.AttributePath{
						Steps: []tftypes.AttributePathStep{
							tftypes.AttributeName("migration"),
							tftypes.ElementKeyInt(i),
							tftypes.AttributeName("id"),
						},
					},
				},
			}, nil
		}
	}

	return nil, nil
}

func (r *resourceMigrate) Plan(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	return map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "static-id"),
		"migration":           proposed["migration"],
		"complete_migrations": proposed["migration"],
	}, nil, nil
}
