package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
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

func (d *dataDriver) Schema(context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Block: &tfprotov5.SchemaBlock{
			Description: "The `sql_driver` datasource allows you to determine which driver is in use by the provider. This " +
				"is mostly useful for module development when you may communicate with multiple types of databases.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes: []*tfprotov5.SchemaAttribute{
				{
					Name:            "name",
					Computed:        true,
					Description:     "The name of the driver, currently this will be one of `pgx`, `mysql`, or `sqlserver`.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Type:            tftypes.String,
				},

				deprecatedIDAttribute(),
			},
		},
	}
}

func (d *dataDriver) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (d *dataDriver) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
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
