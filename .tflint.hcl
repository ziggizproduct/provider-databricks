# Workspace standard, copied from ziggizproduct/arch. Do not edit in product
# repos. Run `tflint --init` once to fetch plugins, then `tflint --recursive`.
config {
  call_module_type = "local"
}

plugin "terraform" {
  enabled = true
  preset  = "recommended"
}

plugin "aws" {
  enabled = true
  version = "0.49.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}
