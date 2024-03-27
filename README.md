# Terraform Provider Beeswax

This Terraform module was done using [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) starting from a fork of the Terraform example [terraform-provider-scaffolding](https://github.com/hashicorp/terraform-provider-scaffolding).

## Developement

During development it's faster to using a locally build provider.

Compile the plugin the the `/bin` directory:
```
make build
```

Edit your `~/.terraformrc`:
```
provider_installation {
  dev_overrides {
    "registry.terraform.io/martin-magakian/beeswax" = "/<PAHT>/terraform-provider-beeswax/bin/"
  }
  direct {}
}
```

Use the provider in your Terraform project:
```
terraform {
  required_providers {
    beeswax = {
      source = "registry.terraform.io/martin-magakian/beeswax"
    }
  }
}
```

Your `terraform apply` will now use the locally build plugin.


## Generate documentation

```
make doc
```

## Limitation

* Only user and role are supported. See [Beeswax documentation](https://api-docs.freewheel.tv/beeswax/v2.0/reference) for all resources available
* data resource can only use ID
