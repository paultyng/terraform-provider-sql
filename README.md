View the docs on the [Terraform Registry](https://registry.terraform.io/providers/paultyng/sql/latest/docs).

# Terraform SQL Provider

This provider is an experiment using the new [terraform-plugin-go](https://github.com/hashicorp/terraform-plugin-go) SDK in order to utilize dynamic typing for its attributes.

## TODO

* Better decimal handling (maybe not string? use big.float?)
* Convert JSON (or other structured db types) to native HCL types, not strings
