# Terraform PayPal Provider
PayPal terraform provider for managing merchant data such as products, subscription plans and webhooks

## Requirements

*	[Terraform](https://www.terraform.io/downloads.html) 0.11.x to 0.13.x
*	[Go](https://golang.org/doc/install) 1.12 to 1.17 (to build the provider plugin)

## Build

Clone repository anywhere:

```sh
$ git clone https://github.com/ollieparsley/terraform-provider-paypal.git
```

Enter the provider directory and build the provider

```sh
$ cd terraform-provider-paypal
$ make compile
```

Or alternatively, to install it as a plugin, run

```sh
$ cd terraform-provider-paypal
$ make install
```

## Using the provider

If you're building the provider, follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin) After placing it into your plugins directory,  run `terraform init` to initialize it.

### Basic Usage

Set the following environment variables, `TF_VAR_paypal_client_id` to store your PayPal OAuth client credentials. This is the recommended way to not commit an access token into your version control system.

    export TF_VAR_paypal_client_id=<your paypal oauth client ID>
    export TF_VAR_paypal_client_secret=<your paypal oauth client secret>

These variables are then accessible in your Terraform configuration as
`var.paypal_client_id` and `var.paypal_client_secret`, and can be used to configure the provider. See the examples below for a full usage.

The example below demonstrates the following operations:

  * create a catalog product
  * create a subscription plan for that product
  * create a notification webhook endpoint listening to events

```hcl
terraform {
  required_providers {
    stripe = {
      source = "ollieparsley/paypal"
      version = "<SEE GITHUB RELEASES FOR THE LATEST VERSION>"
    }
  }
}

variable "paypal_client_id" {
    type = string
}
variable "paypal_client_secret" {
    type = string
}
variable "paypal_base_url" {
    type = string
    default = "https://api.paypal.com"
}

provider "paypal" {
  # NOTE: This is populated from the `TF_VAR_paypal_client_id` environment variable.
  client_id = var.paypal_client_id
  # NOTE: This is populated from the `TF_VAR_paypal_client_secret` environment variable.
  client_secret = var.paypal_client_secret
  # NOTE: This is populated from the `TF_VAR_paypal_base_url` environment variable.
  base_url = var.paypal_base_url
}

# Create a notification webhook to receive events
resource "paypal_notification_webhook" "my_awesome_webhook" {
  url = "https://myexamplewebhook.com/my/web/hook"
  event_types = [
    "PAYMENT.SALE.COMPLETED",
    "PAYMENT.SALE.REFUNDED",
    "PAYMENT.SALE.REVERSED",
    "BILLING.SUBSCRIPTION.PAYMENT.FAILED",
    "BILLING.SUBSCRIPTION.CREATED",
    "BILLING.SUBSCRIPTION.ACTIVATED",
    "BILLING.SUBSCRIPTION.UPDATED",
    "BILLING.SUBSCRIPTION.EXPIRED",
    "BILLING.SUBSCRIPTION.CANCELLED",
    "BILLING.SUBSCRIPTION.SUSPENDED"
  ]
}

# Create a product
resource "paypal_catalog_product" "my_awesome_product" {
  name = "My awesome product"
  description = "The awesome product with features: x, y and z"
  type = "service"
  category = "GENERAL_SOFTWARE"
  image_url = "https://google.com"
  home_url = "https://google.com"
}

# Create a subscription plan
resource "paypal_subscription_plan" "my_awesome_product_monthly_usd" {
  name = "My awesome product (USD/month)"
  description = "The full description of my awesomr product"
  product_id = paypal_catalog_product.my_awesome_product.id
  quantity_supported = false

  taxes {
    percentage = "20"
    inclusive = true
  }

  payment_preferences {
    auto_bill_outstanding = true
    payment_failure_threshold = 2
    setup_fee_failure_action = "continue"
    setup_fee {
      value = "2.10"
      currency_code = "USD"
    }
  }

  # Trial period
  billing_cycle {
    sequence = 1
    total_cycles = 1
    tenure_type = "trial"
    frequency {
      interval_unit = "month"
      interval_count = 1
    }
    pricing_scheme {
      fixed_price {
        value = "0.00"
        currency_code = "USD"
      }
    }

  }

  # Normal pricing after trial period
  billing_cycle {
    sequence = 2
    total_cycles = 0 # infinite
    tenure_type = "regular"
    frequency {
      interval_unit = "month"
      interval_count = 1
    }
    pricing_scheme {
      fixed_price {
        value = "4.99"
        currency_code = "USD"
      }
    }
  }
}


```
