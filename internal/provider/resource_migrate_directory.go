package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"

	"github.com/paultyng/terraform-provider-sql/internal/migration"
	"github.com/paultyng/terraform-provider-sql/internal/server"
)

type resourceMigrateDirectory struct {
	resourceMigrateCommon
}

func newResourceMigrateDirectory(p *provider) (*resourceMigrateDirectory, error) {
	return &resourceMigrateDirectory{
		resourceMigrateCommon: resourceMigrateCommon{
			p: p,
		},
	}, nil
}

var (
	_ server.Resource        = (*resourceMigrateDirectory)(nil)
	_ server.ResourceUpdater = (*resourceMigrateDirectory)(nil)
)

func (r *resourceMigrateDirectory) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Block: &tfprotov5.SchemaBlock{
			Attributes: []*tfprotov5.SchemaAttribute{
				{
					Name:     "path",
					Required: true,
					Description: "The path of the SQL migration files. For a path relative to the current module, " +
						"use `path.module`.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Type:            tftypes.String,
				},
				{
					Name:     "single_file_split",
					Optional: true,
					Description: fmt.Sprintf("Set this to a value if your migration up and down are in a single "+
						"file, split on some constant string (ie. in the case of [shmig](https://github.com/mbucc/shmig) "+
						"you would use `%s`).", migration.SHMigSplit),
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Type:            tftypes.String,
				},
				completeMigrationsAttribute(),
				deprecatedIDAttribute(),
			},
		},
	}
}

func (r *resourceMigrateDirectory) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (r *resourceMigrateDirectory) PlanCreate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	return r.plan(ctx, proposed)
}

func (r *resourceMigrateDirectory) PlanUpdate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	return r.plan(ctx, proposed)
}

func (r *resourceMigrateDirectory) plan(ctx context.Context, proposed map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	if !proposed["path"].IsFullyKnown() || !proposed["single_file_split"].IsFullyKnown() {
		return map[string]tftypes.Value{
			"id":                  tftypes.NewValue(tftypes.String, "static-id"),
			"path":                proposed["path"],
			"single_file_split":   proposed["single_file_split"],
			"complete_migrations": tftypes.NewValue(migration.ListTFType, tftypes.UnknownValue),
		}, nil, nil
	}

	var (
		err error

		path            string
		singleFileSplit string
	)

	err = proposed["path"].As(&path)
	if err != nil {
		return nil, nil, err
	}

	err = proposed["single_file_split"].As(&singleFileSplit)
	if err != nil {
		return nil, nil, err
	}

	migrations, err := migration.ReadDir(path, &migration.Options{
		StripLineComments: true,
		SingleFileSplit:   singleFileSplit,
	})
	// TODO: diagnostics here for common file issues, etc?
	if err != nil {
		return nil, nil, err
	}

	return map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "static-id"),
		"path":                proposed["path"],
		"single_file_split":   proposed["single_file_split"],
		"complete_migrations": migration.List(migrations),
	}, nil, nil
}
