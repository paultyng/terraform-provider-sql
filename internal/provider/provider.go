package provider

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"os"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"

	"github.com/paultyng/terraform-provider-sql/internal/server"
)

func New(version string) func() tfprotov5.ProviderServer {
	return func() tfprotov5.ProviderServer {
		s := server.MustNew(func() server.Provider {
			return &provider{}
		})

		// data sources
		s.MustRegisterDataSource("sql_driver", newDataDriver)
		s.MustRegisterDataSource("sql_query", newDataQuery)

		// resources
		s.MustRegisterResource("sql_migrate", newResourceMigrate)
		s.MustRegisterResource("sql_migrate_directory", newResourceMigrateDirectory)

		return s
	}
}

// TODO: use consts for driver names?
type driverName string

type provider struct {
	DB *sql.DB `argmapper:",typeOnly"`

	Driver driverName
}

var _ server.Provider = (*provider)(nil)

func (p *provider) Schema(context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Block: &tfprotov5.SchemaBlock{
			Attributes: []*tfprotov5.SchemaAttribute{
				{
					Name:     "url",
					Optional: true,
					Computed: true,
					Description: "Database connection strings are specified via URLs. The URL format is driver dependent " +
						"but generally has the form: `dbdriver://username:password@host:port/dbname?param1=true&param2=false`. " +
						"You can optionally set the `SQL_URL` environment variable instead.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Type:            tftypes.String,
				},
				{
					Name:     "max_open_conns",
					Optional: true,
					Description: "Sets the maximum number of open connections to the database. Default is `0` (unlimited). " +
						"See Go's documentation on [DB.SetMaxOpenConns](https://golang.org/pkg/database/sql/#DB.SetMaxOpenConns).",
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Type:            tftypes.Number,
				},
				{
					Name:     "max_idle_conns",
					Optional: true,
					Description: "Sets the maximum number of connections in the idle connection pool. Default is `2`. " +
						"See Go's documentation on [DB.SetMaxIdleConns](https://golang.org/pkg/database/sql/#DB.SetMaxIdleConns).",
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Type:            tftypes.Number,
				},
			},
		},
	}
}

func (p *provider) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (p *provider) Configure(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	if p.DB != nil {
		// if reconfiguring, close existing connection
		_ = p.DB.Close()
	}

	var err error

	var (
		url          string
		maxOpenConns *big.Float
		maxIdleConns *big.Float
	)
	if v := config["url"]; v.IsNull() {
		url = os.Getenv("SQL_URL")
	} else {
		err = config["url"].As(&url)
		if err != nil {
			// TODO: diag with path
			return nil, fmt.Errorf("ConfigureProvider - unable to read url: %w", err)
		}
	}

	if url == "" {
		return []*tfprotov5.Diagnostic{
			{
				Severity: tfprotov5.DiagnosticSeverityError,
				Attribute: &tftypes.AttributePath{Steps: []tftypes.AttributePathStep{
					tftypes.AttributeName("url"),
				}},
				Summary: "A `url` is required to connect to your database.",
			},
		}, nil
	}

	if v := config["max_open_conns"]; v.IsNull() {
		maxOpenConns = big.NewFloat(float64(0))
	} else {
		maxOpenConns = &big.Float{}
		err = config["max_open_conns"].As(&maxOpenConns)
		if err != nil {
			// TODO: diag with path
			return nil, fmt.Errorf("ConfigureProvider - unable to read max_open_conns: %w", err)
		}
	}

	if v := config["max_idle_conns"]; v.IsNull() {
		maxIdleConns = big.NewFloat(float64(2))
	} else {
		maxIdleConns = &big.Float{}
		err = v.As(&maxIdleConns)
		if err != nil {
			// TODO: diag with path
			return nil, fmt.Errorf("ConfigureProvider - unable to read max_idle_conns: %w", err)
		}
	}

	err = p.connect(url)
	if err != nil {
		return nil, fmt.Errorf("ConfigureProvider - unable to open database: %w", err)
	}

	maxOpen, acc := maxOpenConns.Int64()
	if acc != big.Exact {
		return nil, fmt.Errorf("ConfigureProvider - results for max_open_conns is not exact")
	}

	maxIdle, acc := maxIdleConns.Int64()
	if acc != big.Exact {
		return nil, fmt.Errorf("ConfigureProvider - results for max_open_conns is not exact")
	}

	p.DB.SetMaxOpenConns(int(maxOpen))
	p.DB.SetMaxIdleConns(int(maxIdle))

	err = p.DB.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("ConfigureProvider - unable to ping database: %w", err)
	}

	return nil, nil
}
