[Terraform Registry](https://registry.terraform.io/providers/paultyng/sql/latest/docs)

# Terraform SQL Provider

This provider is an experiment using the new [terraform-plugin-go](https://github.com/hashicorp/terraform-plugin-go) SDK in order to utilize dynamic typing for its attributes.

Currently it only has a single data source (`sql_query`) which lets you execute a query against Microsoft SQL Server, PostreSQL, MySQL or other data base engines that are protocol compatible (Maria, CockroachDB, etc.).

## TODO

* Better decimal handling (maybe not string? use big.float?)
* Convert JSON (or other structured db types) to native HCL types, not strings
