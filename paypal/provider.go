package paypal

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	paypalSdk "github.com/plutov/paypal/v4"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {

	// Internal mapping of resources to ensure matching interface
	internalResourceMapping := map[string]TerraformResource{
		"paypal_notification_webhook": WebhookResource{},
		"paypal_catalog_product":      CatalogProductResource{},
		"paypal_subscription_plan":    SubscriptionPlanResource{},
	}

	// Map to the terraform resource from our internal representation
	providerResourceMap := map[string]*schema.Resource{}
	for resourceName, resource := range internalResourceMapping {
		providerResourceMap[resourceName] = resource.Resource()
	}

	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Your PayPal OAuth Client ID. You can get this from your developer dashboard https://developer.paypal.com/developer/applications",
				DefaultFunc: schema.EnvDefaultFunc("PAYPAL_CLIENT_ID", nil),
			},
			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Your PayPal OAuth Client Secret. You can get this from your developer dashboard https://developer.paypal.com/developer/applications",
				DefaultFunc: schema.EnvDefaultFunc("PAYPAL_CLIENT_SECRET", nil),
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     paypalSdk.APIBaseLive,
				Description: fmt.Sprintf("The base API url. Default is production %s, but you can set it to the sandbox URL", paypalSdk.APIBaseLive),
				DefaultFunc: schema.EnvDefaultFunc("PAYPAL_BASE_URL", paypalSdk.APIBaseLive),
			},
		},
		ResourcesMap:  providerResourceMap,
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		ClientID:     d.Get("client_id").(string),
		ClientSecret: d.Get("client_secret").(string),
		BaseURL:      d.Get("base_url").(string),
	}

	if config.ClientID == "" {
		return nil, errors.New("a PayPal client_id is required")
	}
	if config.ClientSecret == "" {
		return nil, errors.New("a PayPal client_secret is required")
	}

	log.Println("[INFO] Initializing Paypal client with client credentials")

	client, clientErr := config.Client()
	if clientErr != nil {
		return client, clientErr
	}

	_, accessTokenErr := client.GetAccessToken(context.Background())
	if accessTokenErr != nil {
		return client, accessTokenErr
	}

	return client, clientErr
}
