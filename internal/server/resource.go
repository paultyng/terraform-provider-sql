package server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type Resource interface {
	Schema(ctx context.Context) *tfprotov5.Schema
	Validate(ctx context.Context, config map[string]tftypes.Value) (diags []*tfprotov5.Diagnostic, err error)
	Read(ctx context.Context, config map[string]tftypes.Value) (state map[string]tftypes.Value, diags []*tfprotov5.Diagnostic, err error)
	Destroy(ctx context.Context, prior map[string]tftypes.Value) (diags []*tfprotov5.Diagnostic, err error)
	PlanCreate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value) (planned map[string]tftypes.Value, diags []*tfprotov5.Diagnostic, err error)
	Create(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (state map[string]tftypes.Value, diags []*tfprotov5.Diagnostic, err error)
}

type ResourceUpdater interface {
	PlanUpdate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (planned map[string]tftypes.Value, diags []*tfprotov5.Diagnostic, err error)
	Update(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (state map[string]tftypes.Value, diags []*tfprotov5.Diagnostic, err error)
}
