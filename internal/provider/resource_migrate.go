package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/paultyng/terraform-provider-sql/internal/migration"
	"github.com/paultyng/terraform-provider-sql/internal/server"
)

type resourceMigrate struct {
	resourceMigrateCommon
}

var _ server.Resource = (*resourceMigrate)(nil)
var _ server.ResourceUpdater = (*resourceMigrate)(nil)

func newResourceMigrate(db dbExecer) (*resourceMigrate, error) {
	return &resourceMigrate{
		resourceMigrateCommon: resourceMigrateCommon{
			db: db,
		},
	}, nil
}

var (
	_ server.Resource        = (*resourceMigrate)(nil)
	_ server.ResourceUpdater = (*resourceMigrate)(nil)
)

func (r *resourceMigrate) Schema(ctx context.Context) *tfprotov6.Schema {
	return &tfprotov6.Schema{
		Block: &tfprotov6.SchemaBlock{
			BlockTypes: []*tfprotov6.SchemaNestedBlock{
				{
					TypeName: "migration",
					Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:            "id",
								Required:        true,
								Description:     "Identifier can be any string to help identifying the migration in the source.",
								DescriptionKind: tfprotov6.StringKindMarkdown,
								Type:            tftypes.String,
							},
							{
								Name:            "up",
								Required:        true,
								Description:     "The query to run when applying this migration.",
								DescriptionKind: tfprotov6.StringKindMarkdown,
								Type:            tftypes.String,
							},
							{
								Name:            "down",
								Required:        true,
								Description:     "The query to run when undoing this migration.",
								DescriptionKind: tfprotov6.StringKindMarkdown,
								Type:            tftypes.String,
							},
						},
					},
				},
			},
			Attributes: []*tfprotov6.SchemaAttribute{
				completeMigrationsAttribute(),
				deprecatedIDAttribute(),
			},
		},
	}
}

func (r *resourceMigrate) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov6.Diagnostic, error) {
	migrationValue := config["migration"]

	if !migrationValue.IsFullyKnown() {
		return nil, nil
	}

	migrations, err := migration.FromListValue(migrationValue)
	if err != nil {
		return nil, err
	}

	if len(migrations) == 0 {
		return []*tfprotov6.Diagnostic{
			{
				Severity: tfprotov6.DiagnosticSeverityError,
				Summary:  "At least one migration is required.",
			},
		}, nil
	}

	ids := map[string]bool{}
	for i, m := range migrations {
		if strings.TrimSpace(m.ID) == "" {
			return []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityError,
					Summary:  "ID cannot be empty.",
					Attribute: tftypes.NewAttributePathWithSteps([]tftypes.AttributePathStep{
						tftypes.AttributeName("migration"),
						tftypes.ElementKeyInt(i),
						tftypes.AttributeName("id"),
					}),
				},
			}, nil
		}
		if ids[m.ID] {
			return []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityError,
					Summary:  fmt.Sprintf("Duplicate ID value of %q.", m.ID),
					Attribute: tftypes.NewAttributePathWithSteps([]tftypes.AttributePathStep{
						tftypes.AttributeName("migration"),
						tftypes.ElementKeyInt(i),
						tftypes.AttributeName("id"),
					}),
				},
			}, nil
		}
	}

	return nil, nil
}

func (r *resourceMigrate) PlanCreate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov6.Diagnostic, error) {
	return r.plan(ctx, proposed)
}

func (r *resourceMigrate) PlanUpdate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov6.Diagnostic, error) {
	return r.plan(ctx, proposed)
}

func (r *resourceMigrate) plan(ctx context.Context, proposed map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov6.Diagnostic, error) {
	return map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "static-id"),
		"migration":           proposed["migration"],
		"complete_migrations": proposed["migration"],
	}, nil, nil
}
