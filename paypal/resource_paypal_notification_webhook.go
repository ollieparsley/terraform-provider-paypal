package paypal

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	paypalSdk "github.com/plutov/paypal/v4"

	"log"
)

// Event types https://developer.paypal.com/docs/api-basics/notifications/webhooks/event-names/

type WebhookResource struct{}

func (r WebhookResource) Resource() *schema.Resource {
	return &schema.Resource{
		Schema: r.Schema(),
		Create: r.Create,
		Read:   r.Read,
		Update: r.Update,
		Delete: r.Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func (r WebhookResource) Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The URL that Paypal will send notifications to",
		},
		"event_types": {
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Required:    true,
			Description: "A list of event types",
		},
	}
}

// Create - Creating notification webhook in Paypal
func (r WebhookResource) Create(d *schema.ResourceData, m interface{}) error {
	client := m.(*paypalSdk.Client)

	eventTypeNamesInterfaces := d.Get("event_types").([]interface{})
	eventTypeNames := []string{}
	for _, eventTypeNameInterface := range eventTypeNamesInterfaces {
		eventTypeNames = append(eventTypeNames, eventTypeNameInterface.(string))
	}

	//panic(fmt.Sprintf("eventTypeNames %+v", eventTypeNames))

	eventTypes := r.eventTypeNamesToEventTypes(eventTypeNames)

	webhook, err := client.CreateWebhook(context.Background(), &paypalSdk.CreateWebhookRequest{
		URL:        d.Get("url").(string),
		EventTypes: eventTypes,
	})
	if err != nil {
		log.Printf("Error creating notifications webhook: %s", err.Error())
		return err
	}

	d.SetId(webhook.ID)
	d.Set("url", webhook.URL)
	d.Set("enabled_events", r.eventTypesToEventTypeNames(webhook.EventTypes))

	log.Printf("Created notifications webhook with ID: %s", webhook.ID)

	return nil
}

// Read - Get notification webhook in Paypal
func (r WebhookResource) Read(d *schema.ResourceData, m interface{}) error {
	client := m.(*paypalSdk.Client)

	webhook, err := client.GetWebhook(context.Background(), d.Id())
	if err != nil {
		log.Printf("Error getting notifications webhook %s: %s", d.Id(), err.Error())
		return err
	}

	d.Set("url", webhook.URL)
	d.Set("enabled_events", r.eventTypesToEventTypeNames(webhook.EventTypes))

	return nil
}

// Update - Update notification webhook in Paypal
func (r WebhookResource) Update(d *schema.ResourceData, m interface{}) error {
	client := m.(*paypalSdk.Client)

	webhook, err := client.UpdateWebhook(context.Background(), d.Id(), []paypalSdk.WebhookField{
		{
			Operation: "replace",
			Path:      "/url",
			Value:     d.Get("url").(string),
		},
		{
			Operation: "replace",
			Path:      "/event_types",
			Value:     r.eventTypeNamesToEventTypes(d.Get("event_types").([]string)),
		},
	})
	if err != nil {
		log.Printf("Error updating notifications webhook %s: %s", d.Id(), err.Error())
		return err
	}

	d.Set("url", webhook.URL)
	d.Set("enabled_events", r.eventTypesToEventTypeNames(webhook.EventTypes))

	return r.Read(d, m)
}

// Delete - Delete the notification webhook in Paypal
func (r WebhookResource) Delete(d *schema.ResourceData, m interface{}) error {
	client := m.(*paypalSdk.Client)

	err := client.DeleteWebhook(context.Background(), d.Id())
	if err != nil {
		log.Printf("Error deleting notifications webhook %s: %s", d.Id(), err.Error())
		return err
	}

	d.SetId("")

	return nil
}

// eventTypeNamesToEventTypes Convert the event_types object into an array of event type names
func (r WebhookResource) eventTypeNamesToEventTypes(eventTypeNames []string) []paypalSdk.WebhookEventType {
	eventTypes := []paypalSdk.WebhookEventType{}
	for _, eventTypeName := range eventTypeNames {
		eventTypes = append(eventTypes, paypalSdk.WebhookEventType{
			Name: strings.ToUpper(eventTypeName),
		})
	}
	return eventTypes
}

// eventTypesToEventTypeNames Convert the array of string event type names into the event types object used in the API
func (r WebhookResource) eventTypesToEventTypeNames(eventTypes []paypalSdk.WebhookEventType) []string {
	eventTypeNames := []string{}
	for _, eventType := range eventTypes {
		eventTypeNames = append(eventTypeNames, eventType.Name)
	}
	return eventTypeNames
}
