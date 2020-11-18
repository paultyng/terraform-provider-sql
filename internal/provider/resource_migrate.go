package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type resourceMigrate struct {
	p *provider
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
								Optional:        true,
								Computed:        true,
								Description:     "Identifier can be any string to help identifying the migration in the source.",
								DescriptionKind: tfprotov5.StringKindMarkdown,
								Type:            tftypes.String,
							},
							{
								Name:            "up",
								Required:        true,
								DescriptionKind: tfprotov5.StringKindMarkdown,
								Type:            tftypes.String,
							},
							{
								Name:            "down",
								Required:        true,
								DescriptionKind: tfprotov5.StringKindMarkdown,
								Type:            tftypes.String,
							},
						},
					},
				},
				// {
				// 	TypeName: "files",
				// 	Nesting:  tfprotov5.SchemaNestedBlockNestingModeSingle,
				// 	MaxItems: 1,
				// 	Block: &tfprotov5.SchemaBlock{
				// 		Attributes: []*tfprotov5.SchemaAttribute{
				// 			{
				// 				Name:            "directory",
				// 				Optional:        true,
				// 				DescriptionKind: tfprotov5.StringKindMarkdown,
				// 				Type:            tftypes.String,
				// 			},
				// 			{
				// 				// standard, shmig
				// 				Name:            "mode",
				// 				Optional:        true,
				// 				Computed:        true,
				// 				DescriptionKind: tfprotov5.StringKindMarkdown,
				// 				Type:            tftypes.String,
				// 			},
				// 		},
				// 	},
				// },
			},
			Attributes: []*tfprotov5.SchemaAttribute{
				// {
				// 	Name:            "complete_migration_ids",
				// 	Computed:        true,
				// 	DescriptionKind: tfprotov5.StringKindMarkdown,
				// 	Type: tftypes.List{
				// 		ElementType: tftypes.String,
				// 	},
				// },

				// TODO: remove this once its not needed by testing
				{
					Name:            "id",
					Computed:        true,
					Deprecated:      true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Type:            tftypes.String,
				},
			},
		},
	}
}

func (r *resourceMigrate) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	if migrationValue, ok := config["migration"]; ok {
		if !migrationValue.IsFullyKnown() {
			return nil, nil
		}

		var migrations []tftypes.Value
		err := migrationValue.As(&migrations)
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
	} else {
		return []*tfprotov5.Diagnostic{
			{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "At least one migration is required.",
			},
		}, nil
	}

	return nil, nil
}

func (r *resourceMigrate) Read(ctx context.Context, current map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	// just roundtrip state here
	return current, nil, nil
}

func (r *resourceMigrate) Plan(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	var migrations []tftypes.Value
	err := proposed["migration"].As(&migrations)
	if err != nil {
		return nil, nil, err
	}

	migrationType := schemaAsObject(r.Schema(ctx)).AttributeTypes["migration"]

	for i, mObj := range migrations {
		var m map[string]tftypes.Value
		err := mObj.As(&m)
		if err != nil {
			return nil, nil, err
		}

		if !m["id"].IsFullyKnown() {
			m["id"] = tftypes.NewValue(tftypes.String, fmt.Sprintf("migration%d", i))
		}
		migrations[i] = tftypes.NewValue(migrationType, m)
	}

	return map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, "static-id"),
		"migration": tftypes.NewValue(tftypes.List{
			ElementType: migrationType,
		}, migrations),
	}, nil, nil
}

func (r *resourceMigrate) Create(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	migrations, err := valueAsMigrationSlice(planned["migration"])
	if err != nil {
		return nil, nil, err
	}

	err = r.runMigrations(ctx, migrations, true, nil)
	if err != nil {
		return nil, nil, err
	}

	return map[string]tftypes.Value{
		"id":        planned["id"],
		"migration": planned["migration"],
	}, nil, nil
}

func (r *resourceMigrate) Update(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	priorMigrations, err := valueAsMigrationSlice(prior["migration"])
	if err != nil {
		return nil, nil, err
	}

	priorMigrationIDs := map[string]bool{}
	for _, m := range priorMigrations {
		priorMigrationIDs[m.ID] = true
	}

	migrations, err := valueAsMigrationSlice(planned["migration"])
	if err != nil {
		return nil, nil, err
	}

	err = r.runMigrations(ctx, migrations, true, priorMigrationIDs)
	if err != nil {
		return nil, nil, err
	}

	// TODO: handle running down any removed migrations in proper ordering?

	return map[string]tftypes.Value{
		"id":        planned["id"],
		"migration": planned["migration"],
	}, nil, nil
}

func (r *resourceMigrate) Destroy(ctx context.Context, prior map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	priorMigrations, err := valueAsMigrationSlice(prior["migration"])
	if err != nil {
		return nil, err
	}

	priorMigrationIDs := map[string]bool{}
	for _, m := range priorMigrations {
		priorMigrationIDs[m.ID] = true
	}

	err = r.runMigrations(ctx, priorMigrations, false, priorMigrationIDs)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

type migration struct {
	ID   string
	Up   string
	Down string
}

func valueAsMigration(v tftypes.Value) (migration, error) {
	m := migration{}

	var valueMap map[string]tftypes.Value
	err := v.As(&valueMap)
	if err != nil {
		return m, err
	}

	err = valueMap["id"].As(&m.ID)
	if err != nil {
		return m, err
	}

	err = valueMap["up"].As(&m.Up)
	if err != nil {
		return m, err
	}

	err = valueMap["down"].As(&m.Down)
	if err != nil {
		return m, err
	}

	return m, nil
}

func valueAsMigrationSlice(v tftypes.Value) ([]migration, error) {
	var migrationValues []tftypes.Value
	err := v.As(&migrationValues)
	if err != nil {
		return nil, err
	}

	migrations := []migration{}
	for _, mValue := range migrationValues {
		m, err := valueAsMigration(mValue)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	return migrations, nil
}

func (r *resourceMigrate) runMigrations(ctx context.Context, migrations []migration, up bool, alreadyRun map[string]bool) error {
	// TODO: add diagnostics to track down specific migration at error

	if up {
		for _, m := range migrations {
			// TODO: should probably filter this somewhere else...
			if alreadyRun[m.ID] {
				continue
			}

			_, err := r.p.db.ExecContext(ctx, m.Up)
			if err != nil {
				return err
			}
		}
	} else {
		for i := len(migrations) - 1; i >= 0; i-- {
			m := migrations[i]

			if !alreadyRun[m.ID] {
				continue
			}

			_, err := r.p.db.ExecContext(ctx, m.Down)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
