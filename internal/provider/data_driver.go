package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/paultyng/terraform-provider-sql/internal/server"
)

type dataDriver struct {
	driver string
}

var _ server.DataSource = (*dataDriver)(nil)

func newDataDriver(driver driverName) (*dataDriver, error) {
	return &dataDriver{
		driver: string(driver),
	}, nil
}

func (d *dataDriver) Schema(context.Context) *tfprotov6.Schema {
	return &tfprotov6.Schema{
		Block: &tfprotov6.SchemaBlock{
			Description: "The `sql_driver` datasource allows you to determine which driver is in use by the provider. This " +
				"is mostly useful for module development when you may communicate with multiple types of databases.",
			DescriptionKind: tfprotov6.StringKindMarkdown,
			Attributes: []*tfprotov6.SchemaAttribute{
				{
					Name:            "name",
					Computed:        true,
					Description:     "The name of the driver, currently this will be one of `pgx`, `mysql`, or `sqlserver`.",
					DescriptionKind: tfprotov6.StringKindMarkdown,
					Type:            tftypes.String,
				},

				deprecatedIDAttribute(),
			},
		},
	}
}

func (d *dataDriver) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov6.Diagnostic, error) {
	return nil, nil
}

func (d *dataDriver) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov6.Diagnostic, error) {
	return map[string]tftypes.Value{
		"name": tftypes.NewValue(
			tftypes.String,
			d.driver,
		),

		// just a placeholder, see deprecatedIDAttribute
		"id": tftypes.NewValue(
			tftypes.String,
			"",
		),
	}, nil, nil
}
