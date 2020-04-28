provider "hashicups" {}

module "psl" {
  source = "./coffee"

  coffee_name = "Packer Spiced Latte"
}

output "psl" {
  value = module.psl.coffee
}

data "hashicups_ingredients" "psl" {
  coffee_id = values(module.psl.coffee)[0].id
}

output "psl_i" {
  value = data.hashicups_ingredients.psl
}
