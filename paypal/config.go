package paypal

import (
	"log"

	paypalSdk "github.com/plutov/paypal/v4"
)

// Config stores PayPal's API configuration
type Config struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
}

// Client returns a new Client for accessing Paypal by using the access token
func (c *Config) Client() (*paypalSdk.Client, error) {

	// Create a client instance
	client, err := paypalSdk.NewClient(c.ClientID, c.ClientSecret, c.BaseURL)
	if err != nil {
		return nil, err
	}
	//client.SetAccessToken(c.AccessToken)
	log.Printf("[INFO] Paypal Client configured.")

	return client, nil
}
