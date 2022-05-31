terraform {
  required_providers {
    vaulted-null = {
      source = "syndbg/vaulted-null"
    }
  }
}

provider "vaulted-null" {
  private_key_path = "./private.pem"
}

data "vaulted-null_content" "example" {
  content = "$VED;1.0::IhQnapipQSRi9gR9OD2Bph3souOeg2DzyyhciRNOasXItG05/BaZ/w5M33i34BrkW6kZJGbmi9jZ73qRGXjHHS004b8L9pPUvSzoeSRoC9XX4DgeKk3ywMaMOi02SJWdohhxsHpDJZE2Y2/t8IQvv75J8ObhdzFO6+0rQm3WGf20tA1tERlGVhWoq0Jgx2ZMyfqgTDbFJPz0DjRkzidpCzaNjCbz+X0/d/az7U9Y926lbXUlXwJE0tdHz5qhI+3NkRMwiZOw4sqFEaiZEMrLFI3FCVP2C71ruOgb2UTygSgv6uiVtqbcaWmytUjEdSs6gvsb2OMq5lNFDIgoG67IQNJLTns/WRj4E7UwrGh7PkAYfiXKqbElEBUvvBrp3RhYeP09JnD8/WlSZrptH9GatwbFCoq0pLzCnEbQJNEmIS+Ez8Wz/4alk8GjTQKjOEkvI0J0pPp9ifYgNin2lfj80UvR8imceBkVdUdojVfB5GLdTmFGnsqmta+Y69onn+4cetB+fLw8SiDROSIsswMMJdLMQ2INYr6SSkhUGvvgFZkRGBpVhl3H+qI1c0klTURoNyUFoUxMOZWlDKM/ZIZPN87dhirLwYNNCGleyYDel7IbxOIhOvKIRQCS3jepI+gC5Mj7M0lYNZxGQ3Z+/k1ppYmhKbU9x4pjLIs1TEc5ESc=::BT/tWV1YcgySqoB/AsVkyChpet1V8HYiJzI1S0OsPtkKqNyA"
}

resource "vaulted-null_encrypt_content" "example" {
  plaintext = "EXAMPLE"
}

output "example" {
  value = data.vaulted-null_content.example.decrypted
}

output "new-encrypted-example" {
  value = vaulted-null_encrypt_content.example.encrypted
}
