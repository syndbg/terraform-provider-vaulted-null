provider "hashicups" {
  username = "dos"
  password = "test123"
}

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

resource "hashicups_order" "first" {
  item {
    coffee_id = 2
    quantity  = 2
  }
}
