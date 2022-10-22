package server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type Provider interface {
	Schema(ctx context.Context) *tfprotov6.Schema
	Validate(ctx context.Context, config map[string]tftypes.Value) (diags []*tfprotov6.Diagnostic, err error)
	Configure(ctx context.Context, config map[string]tftypes.Value) (diags []*tfprotov6.Diagnostic, err error)
}
