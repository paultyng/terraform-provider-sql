package server

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/go-argmapper"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

var (
	ifaceType = reflect.TypeOf((*interface{})(nil)).Elem()

	resourceType   = reflect.TypeOf((*Resource)(nil)).Elem()
	dataSourceType = reflect.TypeOf((*DataSource)(nil)).Elem()
)

func New(providerFactoryFunc interface{}) (*Server, error) {
	s := &Server{
		dsf: map[string]*argmapper.Func{},
		rf:  map[string]*argmapper.Func{},
	}

	f, err := argmapper.NewFunc(func(p Provider) {
		s.p = p
	})
	if err != nil {
		return nil, err
	}

	res := f.Call(
		argmapper.Converter(providerFactoryFunc),
	)
	if err := res.Err(); err != nil {
		return nil, err
	}

	return s, nil
}

type Server struct {
	p Provider

	dsf map[string]*argmapper.Func
	rf  map[string]*argmapper.Func
}

func assertValidFactory(fn *argmapper.Func, target reflect.Type) error {
	outputs := fn.Output().Values()
	if len(outputs) != 1 {
		return fmt.Errorf("factory functions should have exactly one non-error output, the implementation")
	}

	typ := outputs[0].Type
	if typ != ifaceType && !typ.Implements(target) {
		return fmt.Errorf("factory output should implement interface: %s", target)
	}

	return nil
}

func (s *Server) MustRegisterDataSource(typeName string, factory interface{}) {
	err := s.RegisterDataSource(typeName, factory)
	if err != nil {
		panic(err)
	}
}

func (s *Server) RegisterDataSource(typeName string, factory interface{}) error {
	f, err := argmapper.NewFunc(
		factory,
	)
	if err != nil {
		return err
	}

	err = assertValidFactory(f, dataSourceType)
	if err != nil {
		return err
	}

	s.dsf[typeName] = f

	return nil
}

func (s *Server) dataSource(typeName string) (DataSource, error) {
	conv, ok := s.dsf[typeName]
	if !ok {
		return nil, fmt.Errorf("unable to find %q", typeName)
	}

	var ds DataSource
	f, err := argmapper.NewFunc(func(p DataSource) {
		ds = p
	})
	if err != nil {
		return nil, err
	}

	res := f.Call(
		argmapper.Named("typeName", typeName),

		argmapper.Typed(s.p),

		argmapper.ConverterFunc(conv),
	)
	if err := res.Err(); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *Server) MustRegisterResource(typeName string, fn interface{}) {
	err := s.RegisterResource(typeName, fn)
	if err != nil {
		panic(err)
	}
}

func (s *Server) RegisterResource(typeName string, fn interface{}) error {
	f, err := argmapper.NewFunc(fn)
	if err != nil {
		return err
	}

	err = assertValidFactory(f, resourceType)
	if err != nil {
		return err
	}

	s.rf[typeName] = f

	return nil
}

func (s *Server) resource(typeName string) (Resource, error) {
	conv, ok := s.rf[typeName]
	if !ok {
		return nil, fmt.Errorf("unable to find %q", typeName)
	}
	var r Resource
	f, err := argmapper.NewFunc(func(p Resource) {
		r = p
	})
	if err != nil {
		return nil, err
	}

	res := f.Call(
		argmapper.Named("typeName", typeName),

		argmapper.Typed(s.p),

		argmapper.ConverterFunc(conv),
	)
	if err := res.Err(); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Server) GetProviderSchema(ctx context.Context, req *tfprotov5.GetProviderSchemaRequest) (*tfprotov5.GetProviderSchemaResponse, error) {
	resp := &tfprotov5.GetProviderSchemaResponse{
		Provider:          s.p.Schema(ctx),
		DataSourceSchemas: map[string]*tfprotov5.Schema{},
		ResourceSchemas:   map[string]*tfprotov5.Schema{},
	}

	for typeName := range s.dsf {
		ds, err := s.dataSource(typeName)
		if err != nil {
			return nil, err
		}
		resp.DataSourceSchemas[typeName] = ds.Schema(ctx)
	}

	for typeName := range s.rf {
		r, err := s.resource(typeName)
		if err != nil {
			return nil, err
		}
		resp.ResourceSchemas[typeName] = r.Schema(ctx)
	}

	return resp, nil
}

func (s *Server) PrepareProviderConfig(ctx context.Context, req *tfprotov5.PrepareProviderConfigRequest) (*tfprotov5.PrepareProviderConfigResponse, error) {
	schemaObjectType := schemaAsObject(s.p.Schema(ctx))

	_, config, err := unmarshalDynamicValueObject(req.Config, schemaObjectType)
	if err != nil {
		return nil, fmt.Errorf("ConfigureProvider - unmarshalDynamicValueObject(req.Config): %w", err)
	}

	diags, err := s.p.Validate(ctx, config)
	if err != nil {
		return nil, err
	}
	if diagsHaveError(diags) {
		return &tfprotov5.PrepareProviderConfigResponse{
			Diagnostics: diags,
		}, nil
	}

	// TODO: defaulting?

	return &tfprotov5.PrepareProviderConfigResponse{
		Diagnostics:    diags,
		PreparedConfig: req.Config,
	}, nil
}

func (s *Server) ConfigureProvider(ctx context.Context, req *tfprotov5.ConfigureProviderRequest) (*tfprotov5.ConfigureProviderResponse, error) {
	schemaObjectType := schemaAsObject(s.p.Schema(ctx))

	_, config, err := unmarshalDynamicValueObject(req.Config, schemaObjectType)
	if err != nil {
		return nil, fmt.Errorf("ConfigureProvider - unmarshalDynamicValueObject(req.Config): %w", err)
	}

	diags, err := s.p.Validate(ctx, config)
	if err != nil {
		return nil, err
	}
	if diagsHaveError(diags) {
		return &tfprotov5.ConfigureProviderResponse{
			Diagnostics: diags,
		}, nil
	}

	diags, err = s.p.Configure(ctx, config)
	if err != nil {
		return nil, err
	}
	if diagsHaveError(diags) {
		return &tfprotov5.ConfigureProviderResponse{
			Diagnostics: diags,
		}, nil
	}
	return &tfprotov5.ConfigureProviderResponse{
		Diagnostics: diags,
	}, nil
}

func (s *Server) StopProvider(ctx context.Context, req *tfprotov5.StopProviderRequest) (*tfprotov5.StopProviderResponse, error) {
	// TODO: close/reopen? db connection
	panic("not implemented")
}

// ResourceServer methods

func (s *Server) ValidateResourceTypeConfig(ctx context.Context, req *tfprotov5.ValidateResourceTypeConfigRequest) (*tfprotov5.ValidateResourceTypeConfigResponse, error) {
	r, err := s.resource(req.TypeName)
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

func (s *Server) UpgradeResourceState(ctx context.Context, req *tfprotov5.UpgradeResourceStateRequest) (*tfprotov5.UpgradeResourceStateResponse, error) {
	r, err := s.resource(req.TypeName)
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

func (s *Server) ReadResource(ctx context.Context, req *tfprotov5.ReadResourceRequest) (*tfprotov5.ReadResourceResponse, error) {
	r, err := s.resource(req.TypeName)
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

func (s *Server) PlanResourceChange(ctx context.Context, req *tfprotov5.PlanResourceChangeRequest) (*tfprotov5.PlanResourceChangeResponse, error) {
	r, err := s.resource(req.TypeName)
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

	priorObject, prior, err := unmarshalDynamicValueObject(req.PriorState, schemaObjectType)
	if err != nil {
		return nil, fmt.Errorf("PlanResourceChange - unmarshalDynamicValueObject(req.PriorState): %w", err)
	}

	var planned map[string]tftypes.Value
	if priorObject.IsNull() {
		var createDiags []*tfprotov5.Diagnostic
		planned, createDiags, err = r.PlanCreate(ctx, proposed, config)
		if err != nil {
			return nil, err
		}
		diags = append(diags, createDiags...)
	} else {
		updater, ok := r.(ResourceUpdater)
		if !ok {
			return nil, fmt.Errorf("attempting to update resource with no Update implementation")
		}
		var updateDiags []*tfprotov5.Diagnostic
		planned, updateDiags, err = updater.PlanUpdate(ctx, proposed, config, prior)
		if err != nil {
			return nil, err
		}
		diags = append(diags, updateDiags...)
	}

	if diagsHaveError(diags) {
		return &tfprotov5.PlanResourceChangeResponse{
			Diagnostics: diags,
		}, nil
	}

	plannedValue, err := tfprotov5.NewDynamicValue(schemaObjectType, tftypes.NewValue(schemaObjectType, planned))
	if err != nil {
		return nil, fmt.Errorf("ApplyResourceChange - error NewDynamicValue: %w", err)
	}

	return &tfprotov5.PlanResourceChangeResponse{
		PlannedState: &plannedValue,
		Diagnostics:  diags,
	}, nil
}

func (s *Server) ApplyResourceChange(ctx context.Context, req *tfprotov5.ApplyResourceChangeRequest) (*tfprotov5.ApplyResourceChangeResponse, error) {
	r, err := s.resource(req.TypeName)
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
		updater, ok := r.(ResourceUpdater)
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

func (s *Server) ImportResourceState(ctx context.Context, req *tfprotov5.ImportResourceStateRequest) (*tfprotov5.ImportResourceStateResponse, error) {
	panic("not implemented")
}

// DataSourceServer methods

func (s *Server) ValidateDataSourceConfig(ctx context.Context, req *tfprotov5.ValidateDataSourceConfigRequest) (*tfprotov5.ValidateDataSourceConfigResponse, error) {
	ds, err := s.dataSource(req.TypeName)
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

func (s *Server) ReadDataSource(ctx context.Context, req *tfprotov5.ReadDataSourceRequest) (*tfprotov5.ReadDataSourceResponse, error) {
	ds, err := s.dataSource(req.TypeName)
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
