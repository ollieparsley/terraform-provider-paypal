package paypal

import (
	"reflect"
	"testing"

	"github.com/go-test/deep"
	"github.com/hashicorp/terraform/helper/schema"
	paypalSdk "github.com/plutov/paypal/v4"
)

func TestWebhookResourceSchema(t *testing.T) {
	resource := WebhookResource{}
	actualSchema := resource.Schema()
	expectedSchema := map[string]*schema.Schema{
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

	differences := deep.Equal(expectedSchema, actualSchema)
	if len(differences) > 0 {
		t.Errorf("Expected correct simplified schema. Got differences: %+v", differences)
	}

}

func TestWebhookResourceEventTypeNamesToEventTypes(t *testing.T) {
	resource := WebhookResource{}
	eventTypeNames := []string{
		"TEST_NaMe_1",
		"TEST_NAME_2",
		"test_name_3",
	}

	expectedEventTypes := []paypalSdk.WebhookEventType{
		{Name: "TEST_NAME_1"},
		{Name: "TEST_NAME_2"},
		{Name: "TEST_NAME_3"},
	}

	actualEventTypes := resource.eventTypeNamesToEventTypes(eventTypeNames)

	if reflect.DeepEqual(expectedEventTypes, actualEventTypes) == false {
		t.Errorf("Expected Event types didn't match. Got: %+v", actualEventTypes)
	}
}

func TestWebhookResourceEventTypesToEventTypesNames(t *testing.T) {
	resource := WebhookResource{}
	eventTypes := []paypalSdk.WebhookEventType{
		{Name: "TEST_NAME_1"},
		{Name: "TEST_NAME_2"},
		{Name: "test_name_3"},
	}

	expectedEventTypeNames := []string{
		"TEST_NAME_1",
		"TEST_NAME_2",
		"test_name_3",
	}

	actualEventTypeNames := resource.eventTypesToEventTypeNames(eventTypes)

	if reflect.DeepEqual(expectedEventTypeNames, actualEventTypeNames) == false {
		t.Errorf("Expected Event types names didn't match. Got: %+v", actualEventTypeNames)
	}
}
