package server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type Resource interface {
	Schema(ctx context.Context) *tfprotov6.Schema
	Validate(ctx context.Context, config map[string]tftypes.Value) (diags []*tfprotov6.Diagnostic, err error)
	Read(ctx context.Context, config map[string]tftypes.Value) (state map[string]tftypes.Value, diags []*tfprotov6.Diagnostic, err error)
	Destroy(ctx context.Context, prior map[string]tftypes.Value) (diags []*tfprotov6.Diagnostic, err error)
	PlanCreate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value) (planned map[string]tftypes.Value, diags []*tfprotov6.Diagnostic, err error)
	Create(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (state map[string]tftypes.Value, diags []*tfprotov6.Diagnostic, err error)
}

type ResourceUpdater interface {
	PlanUpdate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (planned map[string]tftypes.Value, diags []*tfprotov6.Diagnostic, err error)
	Update(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (state map[string]tftypes.Value, diags []*tfprotov6.Diagnostic, err error)
}
