package server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type DataSource interface {
	Schema(ctx context.Context) *tfprotov6.Schema
	Validate(ctx context.Context, config map[string]tftypes.Value) (diags []*tfprotov6.Diagnostic, err error)
	Read(ctx context.Context, config map[string]tftypes.Value) (state map[string]tftypes.Value, diags []*tfprotov6.Diagnostic, err error)
}
