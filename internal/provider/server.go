package provider

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"os"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type provider struct {
	db *db
}

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

type dataSource interface {
	Schema(context.Context) *tfprotov5.Schema
	Validate(context.Context, map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error)
	Read(context.Context, map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error)
}

func (p *provider) NewDataSource(typeName string) (dataSource, error) {
	switch typeName {
	case "sql_query":
		return &dataQuery{
			p: p,
		}, nil
	}
	return nil, fmt.Errorf("unexpected data source type %q", typeName)
}

type resource interface {
	Schema(context.Context) *tfprotov5.Schema
	Validate(context.Context, map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error)
	Read(ctx context.Context, current map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error)
	// TODO: split this to PlanCreate, PlanUpdate, etc.
	Plan(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error)
	Destroy(context.Context, map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error)
	Create(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error)
	// Read(context.Context, map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error)
}

type resourceUpdater interface {
	Update(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error)
}

func (p *provider) NewResource(typeName string) (resource, error) {
	switch typeName {
	case "sql_migrate":
		return &resourceMigrate{
			resourceMigrateCommon: resourceMigrateCommon{
				p: p,
			},
		}, nil
	case "sql_migrate_directory":
		return &resourceMigrateDirectory{
			resourceMigrateCommon: resourceMigrateCommon{
				p: p,
			},
		}, nil
	}
	return nil, fmt.Errorf("unexpected resource type %q", typeName)
}

func New(version string) func() tfprotov5.ProviderServer {
	return func() tfprotov5.ProviderServer {
		return &provider{}
	}
}

// ProviderServer methods

func (p *provider) GetProviderSchema(ctx context.Context, req *tfprotov5.GetProviderSchemaRequest) (*tfprotov5.GetProviderSchemaResponse, error) {
	resp := &tfprotov5.GetProviderSchemaResponse{
		Provider:          p.Schema(ctx),
		DataSourceSchemas: map[string]*tfprotov5.Schema{},
		ResourceSchemas:   map[string]*tfprotov5.Schema{},
	}

	for _, typeName := range []string{"sql_query"} {
		ds, err := p.NewDataSource(typeName)
		if err != nil {
			return nil, err
		}
		resp.DataSourceSchemas[typeName] = ds.Schema(ctx)
	}

	for _, typeName := range []string{"sql_migrate", "sql_migrate_directory"} {
		ds, err := p.NewResource(typeName)
		if err != nil {
			return nil, err
		}
		resp.ResourceSchemas[typeName] = ds.Schema(ctx)
	}

	return resp, nil
}

func (p *provider) PrepareProviderConfig(ctx context.Context, req *tfprotov5.PrepareProviderConfigRequest) (*tfprotov5.PrepareProviderConfigResponse, error) {
	// TODO: default handling?
	// TODO: validate URL value
	return &tfprotov5.PrepareProviderConfigResponse{
		PreparedConfig: req.Config,
	}, nil
}

func (p *provider) ConfigureProvider(ctx context.Context, req *tfprotov5.ConfigureProviderRequest) (*tfprotov5.ConfigureProviderResponse, error) {
	if p.db != nil {
		// if reconfiguring, close existing connection
		p.db.Close()
	}

	schemaObjectType := schemaAsObject(p.Schema(ctx))

	_, config, err := unmarshalDynamicValueObject(req.Config, schemaObjectType)
	if err != nil {
		return nil, fmt.Errorf("ConfigureProvider - unmarshalDynamicValueObject(req.Config): %w", err)
	}

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
		diags := []*tfprotov5.Diagnostic{
			{
				Severity: tfprotov5.DiagnosticSeverityError,
				Attribute: &tftypes.AttributePath{Steps: []tftypes.AttributePathStep{
					tftypes.AttributeName("url"),
				}},
				Summary: "A `url` is required to connect to your database.",
			},
		}
		return &tfprotov5.ConfigureProviderResponse{
			Diagnostics: diags,
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

	p.db, err = newDB(url, func(db *sql.DB) error {
		maxOpen, acc := maxOpenConns.Int64()
		if acc != big.Exact {
			return fmt.Errorf("ConfigureProvider - results for max_open_conns is not exact")
		}

		maxIdle, acc := maxIdleConns.Int64()
		if acc != big.Exact {
			return fmt.Errorf("ConfigureProvider - results for max_open_conns is not exact")
		}

		db.SetMaxOpenConns(int(maxOpen))
		db.SetMaxIdleConns(int(maxIdle))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("ConfigureProvider - unable to open database: %w", err)
	}

	err = p.db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("ConfigureProvider - unable to ping database: %w", err)
	}

	return &tfprotov5.ConfigureProviderResponse{}, nil
}

func (p *provider) StopProvider(ctx context.Context, req *tfprotov5.StopProviderRequest) (*tfprotov5.StopProviderResponse, error) {
	// TODO: close/reopen? db connection
	panic("not implemented")
}

// ResourceServer methods

func (p *provider) ValidateResourceTypeConfig(ctx context.Context, req *tfprotov5.ValidateResourceTypeConfigRequest) (*tfprotov5.ValidateResourceTypeConfigResponse, error) {
	r, err := p.NewResource(req.TypeName)
	if err != nil {
		return nil, err
	}

	schemaObjectType := schemaAsObject(r.Schema(ctx))

	_, config, err := unmarshalDynamicValueObject(req.Config, schemaObjectType)
	if err != nil {
		return nil, fmt.Errorf("ValidateResourceTypeConfig - unmarshalDynamicValueObject(req.Config): %w", err)
	}

	diags, err := r.Validate(ctx, config)
	if err != nil {
		return nil, err
	}
	return &tfprotov5.ValidateResourceTypeConfigResponse{
		Diagnostics: diags,
	}, nil
}

func (p *provider) UpgradeResourceState(ctx context.Context, req *tfprotov5.UpgradeResourceStateRequest) (*tfprotov5.UpgradeResourceStateResponse, error) {
	r, err := p.NewResource(req.TypeName)
	if err != nil {
		return nil, err
	}

	schemaObjectType := schemaAsObject(r.Schema(ctx))

	rawStateObject, err := req.RawState.Unmarshal(schemaObjectType)
	if err != nil {
		return nil, err
	}

	rawStateValue, err := tfprotov5.NewDynamicValue(schemaObjectType, rawStateObject)
	if err != nil {
		return nil, err
	}

	return &tfprotov5.UpgradeResourceStateResponse{
		UpgradedState: &rawStateValue,
	}, nil
}

func (p *provider) ReadResource(ctx context.Context, req *tfprotov5.ReadResourceRequest) (*tfprotov5.ReadResourceResponse, error) {
	r, err := p.NewResource(req.TypeName)
	if err != nil {
		return nil, err
	}

	schemaObjectType := schemaAsObject(r.Schema(ctx))

	_, currentState, err := unmarshalDynamicValueObject(req.CurrentState, schemaObjectType)
	if err != nil {
		return nil, fmt.Errorf("ReadResource - unmarshalDynamicValueObject(req.CurrentState): %w", err)
	}

	newState, diags, err := r.Read(ctx, currentState)
	if err != nil {
		return nil, err
	}

	if diagsHaveError(diags) {
		return &tfprotov5.ReadResourceResponse{
			Diagnostics: diags,
		}, nil
	}

	newStateValue, err := tfprotov5.NewDynamicValue(schemaObjectType, tftypes.NewValue(schemaObjectType, newState))
	if err != nil {
		return nil, fmt.Errorf("ApplyResourceChange - error NewDynamicValue: %w", err)
	}

	return &tfprotov5.ReadResourceResponse{
		NewState:    &newStateValue,
		Diagnostics: diags,
	}, nil
}

func (p *provider) PlanResourceChange(ctx context.Context, req *tfprotov5.PlanResourceChangeRequest) (*tfprotov5.PlanResourceChangeResponse, error) {
	r, err := p.NewResource(req.TypeName)
	if err != nil {
		return nil, err
	}

	schemaObjectType := schemaAsObject(r.Schema(ctx))

	proposedObject, proposed, err := unmarshalDynamicValueObject(req.ProposedNewState, schemaObjectType)
	if err != nil {
		return nil, fmt.Errorf("PlanResourceChange - unmarshalDynamicValueObject(req.ProposedNewState): %w", err)
	}

	if proposedObject.IsNull() {
		// short circuit, this is a destroy
		return &tfprotov5.PlanResourceChangeResponse{
			PlannedState: req.ProposedNewState,
		}, nil
	}

	_, config, err := unmarshalDynamicValueObject(req.Config, schemaObjectType)
	if err != nil {
		return nil, fmt.Errorf("PlanResourceChange - unmarshalDynamicValueObject(req.Config): %w", err)
	}

	diags, err := r.Validate(ctx, config)
	if err != nil {
		return nil, err
	}

	if diagsHaveError(diags) {
		return &tfprotov5.PlanResourceChangeResponse{
			Diagnostics: diags,
		}, nil
	}

	_, prior, err := unmarshalDynamicValueObject(req.PriorState, schemaObjectType)
	if err != nil {
		return nil, fmt.Errorf("PlanResourceChange - unmarshalDynamicValueObject(req.PriorState): %w", err)
	}

	state, planDiags, err := r.Plan(ctx, proposed, config, prior)
	if err != nil {
		return nil, err
	}
	diags = append(diags, planDiags...)

	if diagsHaveError(diags) {
		return &tfprotov5.PlanResourceChangeResponse{
			Diagnostics: diags,
		}, nil
	}

	stateValue, err := tfprotov5.NewDynamicValue(schemaObjectType, tftypes.NewValue(schemaObjectType, state))
	if err != nil {
		return nil, fmt.Errorf("ApplyResourceChange - error NewDynamicValue: %w", err)
	}

	return &tfprotov5.PlanResourceChangeResponse{
		PlannedState: &stateValue,
		Diagnostics:  diags,
	}, nil
}

func (p *provider) ApplyResourceChange(ctx context.Context, req *tfprotov5.ApplyResourceChangeRequest) (*tfprotov5.ApplyResourceChangeResponse, error) {
	r, err := p.NewResource(req.TypeName)
	if err != nil {
		return nil, err
	}

	schemaObjectType := schemaAsObject(r.Schema(ctx))

	plannedObject, planned, err := unmarshalDynamicValueObject(req.PlannedState, schemaObjectType)
	if err != nil {
		return nil, fmt.Errorf("ApplyResourceChange - unmarshalDynamicValueObject(req.PlannedState): %w", err)
	}

	priorObject, prior, err := unmarshalDynamicValueObject(req.PriorState, schemaObjectType)
	if err != nil {
		return nil, fmt.Errorf("ApplyResourceChange - unmarshalDynamicValueObject(req.PriorState): %w", err)
	}

	if plannedObject.IsNull() {

		// short circuit, this is a destroy
		diags, err := r.Destroy(ctx, prior)
		if err != nil {
			return nil, err
		}

		if diagsHaveError(diags) {
			return &tfprotov5.ApplyResourceChangeResponse{
				Diagnostics: diags,
			}, nil
		}

		return &tfprotov5.ApplyResourceChangeResponse{
			Diagnostics: diags,
			NewState:    req.PlannedState,
		}, nil
	}

	_, config, err := unmarshalDynamicValueObject(req.Config, schemaObjectType)
	if err != nil {
		return nil, fmt.Errorf("ApplyResourceChange - unmarshalDynamicValueObject(req.Config): %w", err)
	}

	diags, err := r.Validate(ctx, config)
	if err != nil {
		return nil, err
	}

	if diagsHaveError(diags) {
		return &tfprotov5.ApplyResourceChangeResponse{
			Diagnostics: diags,
		}, nil
	}

	var state map[string]tftypes.Value
	if priorObject.IsNull() {
		var createDiags []*tfprotov5.Diagnostic
		state, createDiags, err = r.Create(ctx, planned, config, prior)
		if err != nil {
			return nil, err
		}
		diags = append(diags, createDiags...)
	} else {
		updater, ok := r.(resourceUpdater)
		if !ok {
			return nil, fmt.Errorf("attempting to update resource with no Update implementation")
		}
		var updateDiags []*tfprotov5.Diagnostic
		state, updateDiags, err = updater.Update(ctx, planned, config, prior)
		if err != nil {
			return nil, err
		}
		diags = append(diags, updateDiags...)
	}

	if diagsHaveError(diags) {
		return &tfprotov5.ApplyResourceChangeResponse{
			Diagnostics: diags,
		}, nil
	}

	stateValue, err := tfprotov5.NewDynamicValue(schemaObjectType, tftypes.NewValue(schemaObjectType, state))
	if err != nil {
		return nil, fmt.Errorf("ApplyResourceChange - error NewDynamicValue: %w", err)
	}

	return &tfprotov5.ApplyResourceChangeResponse{
		NewState:    &stateValue,
		Diagnostics: diags,
	}, nil
}

func (p *provider) ImportResourceState(ctx context.Context, req *tfprotov5.ImportResourceStateRequest) (*tfprotov5.ImportResourceStateResponse, error) {
	panic("not implemented")
}

// DataSourceServer methods

func (p *provider) ValidateDataSourceConfig(ctx context.Context, req *tfprotov5.ValidateDataSourceConfigRequest) (*tfprotov5.ValidateDataSourceConfigResponse, error) {
	ds, err := p.NewDataSource(req.TypeName)
	if err != nil {
		return nil, err
	}

	schemaObjectType := schemaAsObject(ds.Schema(ctx))

	_, config, err := unmarshalDynamicValueObject(req.Config, schemaObjectType)
	if err != nil {
		return nil, fmt.Errorf("ValidateDataSourceConfig - unmarshalDynamicValueObject(req.Config): %w", err)
	}

	diags, err := ds.Validate(ctx, config)
	if err != nil {
		return nil, err
	}

	return &tfprotov5.ValidateDataSourceConfigResponse{
		Diagnostics: diags,
	}, nil
}

func (p *provider) ReadDataSource(ctx context.Context, req *tfprotov5.ReadDataSourceRequest) (*tfprotov5.ReadDataSourceResponse, error) {
	ds, err := p.NewDataSource(req.TypeName)
	if err != nil {
		return nil, err
	}

	schemaObjectType := schemaAsObject(ds.Schema(ctx))

	_, config, err := unmarshalDynamicValueObject(req.Config, schemaObjectType)
	if err != nil {
		return nil, fmt.Errorf("ReadDataSource - unmarshalDynamicValueObject(req.Config): %w", err)
	}

	diags, err := ds.Validate(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("ReadDataSource - error ds.Validate: %w", err)
	}
	if diagsHaveError(diags) {
		return &tfprotov5.ReadDataSourceResponse{
			Diagnostics: diags,
		}, nil
	}
	state, diags, err := ds.Read(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("ReadDataSource - error ds.Read: %w", err)
	}
	if diagsHaveError(diags) {
		return &tfprotov5.ReadDataSourceResponse{
			Diagnostics: diags,
		}, nil
	}

	// TODO: should NewDynamicValue return a pointer?
	stateValue, err := tfprotov5.NewDynamicValue(schemaObjectType, tftypes.NewValue(schemaObjectType, state))
	if err != nil {
		return nil, fmt.Errorf("ReadDataSource - error NewDynamicValue: %w", err)
	}

	return &tfprotov5.ReadDataSourceResponse{
		State:       &stateValue,
		Diagnostics: diags,
	}, nil
}
