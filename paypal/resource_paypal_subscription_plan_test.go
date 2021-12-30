package paypal

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/hashicorp/terraform/helper/schema"
)

func TestSubscriptionPlanResourceSchema(t *testing.T) {
	resource := SubscriptionPlanResource{}
	actualFullSchema := resource.Schema()

	expectedSchemaSimplified := map[string]SchemaSimplified{
		"product_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
		},
		"name": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
		},
		"description": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
		},
		"status": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"quantity_supported": {
			Type:     schema.TypeBool,
			Optional: true,
			Required: false,
		},
		"billing_cycle": {
			Type:     schema.TypeList,
			Required: true,
			Optional: false,
			Nested: map[string]SchemaSimplified{
				"sequence": {
					Type:     schema.TypeInt,
					Required: true,
					Optional: false,
				},
				"total_cycles": {
					Type:     schema.TypeInt,
					Required: true,
					Optional: false,
				},
				"tenure_type": {
					Type:     schema.TypeString,
					Required: true,
					Optional: false,
				},
				"frequency": {
					Type:     schema.TypeList,
					Required: true,
					Optional: false,
				},
				"pricing_scheme": {
					Type:     schema.TypeList,
					Required: true,
					Optional: false,
				},
			},
		},
		"payment_preferences": {
			Type:     schema.TypeList,
			Required: true,
			Optional: false,
			Nested: map[string]SchemaSimplified{
				"auto_bill_outstanding": {
					Type:     schema.TypeBool,
					Required: true,
					Optional: false,
				},
				"setup_fee": {
					Type:     schema.TypeList,
					Required: true,
					Optional: false,
				},
				"payment_failure_threshold": {
					Type:     schema.TypeInt,
					Required: true,
					Optional: false,
				},
				"setup_fee_failure_action": {
					Type:     schema.TypeString,
					Required: true,
					Optional: false,
				},
			},
		},
		"taxes": {
			Type:     schema.TypeList,
			Required: false,
			Optional: true,
			Nested: map[string]SchemaSimplified{
				"percentage": {
					Type:     schema.TypeString,
					Required: true,
					Optional: false,
				},
				"inclusive": {
					Type:     schema.TypeBool,
					Required: true,
					Optional: false,
				},
			},
		},
	}

	fullActualSchemaSimplified := map[string]SchemaSimplified{}
	for name, actualSchema := range actualFullSchema {
		actualSchemaSimplified := SchemaSimplified{
			Type:     actualSchema.Type,
			Optional: actualSchema.Optional,
			Required: actualSchema.Required,
		}
		if actualSchema.Type == schema.TypeList || actualSchema.Type == schema.TypeSet {
			nested := map[string]SchemaSimplified{}
			for nestedActualKey, nestedActualSchema := range actualSchema.Elem.(*schema.Resource).Schema {
				nested[nestedActualKey] = SchemaSimplified{
					Type:     nestedActualSchema.Type,
					Optional: nestedActualSchema.Optional,
					Required: nestedActualSchema.Required,
				}
			}
			actualSchemaSimplified.Nested = nested
		}
		fullActualSchemaSimplified[name] = actualSchemaSimplified
	}

	differences := deep.Equal(expectedSchemaSimplified, fullActualSchemaSimplified)
	if len(differences) > 0 {
		t.Errorf("Expected correct simplified schema. Got differences: %+v", differences)
	}
}

func TestSubscriptionPlanTenureTypes(t *testing.T) {
	resource := SubscriptionPlanResource{}
	expected := []string{
		"regular",
		"trial",
	}
	differences := deep.Equal(expected, resource.tenureTypes())
	if len(differences) > 0 {
		t.Errorf("Expected didn't match. Got differences: %+v", differences)
	}
}

func TestSubscriptionPlanFrequencyIntervalTypes(t *testing.T) {
	resource := SubscriptionPlanResource{}
	expected := []string{
		"day",
		"week",
		"month",
		"year",
	}
	differences := deep.Equal(expected, resource.frequencyIntervalTypes())
	if len(differences) > 0 {
		t.Errorf("Expected didn't match. Got differences: %+v", differences)
	}
}

func TestSubscriptionPlanPaymentDefinitionTypes(t *testing.T) {
	resource := SubscriptionPlanResource{}
	expected := []string{
		"trial",
		"regular",
	}
	differences := deep.Equal(expected, resource.paymentDefinitionTypes())
	if len(differences) > 0 {
		t.Errorf("Expected didn't match. Got differences: %+v", differences)
	}
}

func TestSubscriptionPlanPaymentDefinitionFrequencies(t *testing.T) {
	resource := SubscriptionPlanResource{}
	expected := []string{
		"day",
		"week",
		"month",
		"year",
	}
	differences := deep.Equal(expected, resource.paymentDefinitionFrequencies())
	if len(differences) > 0 {
		t.Errorf("Expected didn't match. Got differences: %+v", differences)
	}
}

func TestSubscriptionPlanSetupFeeFailureActions(t *testing.T) {
	resource := SubscriptionPlanResource{}
	expected := []string{
		"continue",
		"cancel",
	}
	differences := deep.Equal(expected, resource.setupFeeFailureActions())
	if len(differences) > 0 {
		t.Errorf("Expected didn't match. Got differences: %+v", differences)
	}
}
