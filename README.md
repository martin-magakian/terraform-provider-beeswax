# Terraform Provider Beeswax

This Terraform module was done using [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) starting from a fork of the Terraform example [terraform-provider-scaffolding](https://github.com/hashicorp/terraform-provider-scaffolding).

## Usage

```
terraform {
  required_providers {
    beeswax = {
      source = "martin-magakian/beeswax"
      version = "1.0.4"     # check latest version (https://registry.terraform.io/providers/martin-magakian/beeswax/latest)
    }
  }
}

provider "beeswax" {
  host     = "https://myorg.api.beeswax.com"
  email    = "myemail@myorg.com"
  password = "myPasswd"
}

resource "beeswax_user" "example" {
  super_user         = false
  email              = "myemail@myorg.com"
  first_name         = "Martin"
  last_name          = "Magakian"
  role_id            = 1
  account_id         = 1
  active             = true
  all_account_access = true
  account_group_ids  = []
}
```


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

## Deploying a new version to Terraform Registry

When code is merge in `main` simply tag the version and push. Github action will automatically release the provider to Terraform Registry.

Example:
````
git tag v1.0.4
git push origin v1.0.4
```

## Limitation

* Only user and role are supported. See [Beeswax documentation](https://api-docs.freewheel.tv/beeswax/v2.0/reference) for all resources available.
* data resource can only use ID.
