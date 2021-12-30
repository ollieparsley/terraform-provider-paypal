package paypal

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	paypalSdk "github.com/plutov/paypal/v4"

	"log"
)

type SubscriptionPlanResource struct{}

func (r SubscriptionPlanResource) Resource() *schema.Resource {
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

func (r SubscriptionPlanResource) Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"product_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The ID of the product this plan is for",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the subscription plan",
		},
		"description": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The drescription of the subscription plan",
		},
		"status": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The status of the subscription plan",
		},
		"quantity_supported": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Indicates whether you can subscribe to this plan by providing a quantity for the goods or service",
		},
		"billing_cycle": {
			Type:     schema.TypeList,
			MaxItems: 3,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"sequence": {
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(1, 99),
						Description:  "The order in which this cycle is to run among other billing cycles. For example, a trial billing cycle has a sequence of 1 while a regular billing cycle has a sequence of 2, so that trial cycle runs before the regular cycle.",
					},
					"total_cycles": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "he number of times this billing cycle gets executed. Trial billing cycles can only be executed a finite number of times (value between 1 and 999 for total_cycles). Regular billing cycles can be executed infinite times (value of 0 for total_cycles) or a finite number of times (value between 1 and 999 for total_cycles).",
					},
					"tenure_type": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice(r.tenureTypes(), true),
						Description:  fmt.Sprintf("One of: %s", strings.Join(r.tenureTypes(), ",")),
					},
					"frequency": {
						Type:     schema.TypeList,
						MaxItems: 1,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"interval_unit": {
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringInSlice(r.frequencyIntervalTypes(), true),
									Description:  fmt.Sprintf("One of: %s", strings.Join(r.frequencyIntervalTypes(), ",")),
								},
								"interval_count": {
									Type:        schema.TypeInt,
									Required:    true,
									Description: "The number of intervals after which a subscriber is billed. For example, if the interval_unit is DAY with an interval_count of 2, the subscription is billed once every two days.",
								},
							},
						},
					},

					"pricing_scheme": {
						Type:     schema.TypeList,
						MaxItems: 1,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"version": {
									Type:     schema.TypeInt,
									Optional: true,
									Computed: true,
								},
								"fixed_price": {
									Type:     schema.TypeList,
									MaxItems: 1,
									Required: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"value": {
												Type:             schema.TypeString,
												Required:         true,
												DiffSuppressFunc: r.diffSuppressFuncForFloats,
												Description:      "More info: https://developer.paypal.com/docs/api/payments.billing-plans/v1/#definition-currency",
											},
											"currency_code": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "The three-character ISO-4217 currency code",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},

		"payment_preferences": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"auto_bill_outstanding": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Indicates whether to automatically bill the outstanding amount in the next billing cycle.",
					},
					"setup_fee": {
						Type:        schema.TypeList,
						MaxItems:    1,
						Required:    true,
						Description: "The initial set-up fee for the service.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"value": {
									Type:             schema.TypeString,
									Required:         true,
									DiffSuppressFunc: r.diffSuppressFuncForFloats,
									Description:      "More info: https://developer.paypal.com/docs/api/payments.billing-plans/v1/#definition-currency",
								},
								"currency_code": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "The three-character ISO-4217 currency code",
								},
							},
						},
					},
					"payment_failure_threshold": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "The maximum number of payment failures before a subscription is suspended. For example, if payment_failure_threshold is 2, the subscription automatically updates to the SUSPEND state if two consecutive payments fail.",
					},
					"setup_fee_failure_action": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice(r.setupFeeFailureActions(), true),
						Description:  fmt.Sprintf("One of: %s", strings.Join(r.setupFeeFailureActions(), ",")),
					},
				},
			},
		},

		"taxes": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"percentage": {
						Type:             schema.TypeString,
						Required:         true,
						DiffSuppressFunc: r.diffSuppressFuncForFloats,
					},
					"inclusive": {
						Type:     schema.TypeBool,
						Required: true,
					},
				},
			},
		},
	}
}

// Create - Creating a subscription plan in Paypal
func (r SubscriptionPlanResource) Create(d *schema.ResourceData, m interface{}) error {
	client := m.(*paypalSdk.Client)

	subscriptionPlan := r.sdkObjectFromResourceData(d)

	// Create the plan
	billingResponse, err := client.CreateSubscriptionPlan(context.Background(), subscriptionPlan)

	if err != nil {
		log.Printf("Error creating billing plan: %s", err.Error())
		return err
	}

	d.SetId(billingResponse.ID)
	r.Read(d, m)

	log.Printf("Created billing plan with ID: %s", billingResponse.ID)

	return nil
}

// Read - Get a subscription plan in Paypal - https://developer.paypal.com/docs/api/subscriptions/v1/#plans_get
func (r SubscriptionPlanResource) Read(d *schema.ResourceData, m interface{}) error {
	client := m.(*paypalSdk.Client)

	subscriptionPlan, err := client.GetSubscriptionPlan(context.Background(), d.Id())
	if err != nil {
		log.Printf("Error getting subscription plan %s: %s", d.Id(), err.Error())
		return err
	}

	// Taxes to resource data map
	taxes := []map[string]interface{}{{
		"percentage": subscriptionPlan.Taxes.Percentage,
		"inclusive":  subscriptionPlan.Taxes.Inclusive,
	}}

	// Payment preferences to map
	paymentPreferences := []map[string]interface{}{{
		"auto_bill_outstanding": subscriptionPlan.PaymentPreferences.AutoBillOutstanding,
		"setup_fee": map[string]interface{}{
			"value":         subscriptionPlan.PaymentPreferences.SetupFee.Value,
			"currency_code": subscriptionPlan.PaymentPreferences.SetupFee.Currency,
		},
		"payment_failure_threshold": subscriptionPlan.PaymentPreferences.PaymentFailureThreshold,
		"setup_fee_failure_action":  subscriptionPlan.PaymentPreferences.SetupFeeFailureAction,
	}}

	// Billing cycles to array of maps
	billingCycles := []map[string]interface{}{}
	for _, billingCycleObj := range subscriptionPlan.BillingCycles {
		billingCycle := map[string]interface{}{
			"sequence":     billingCycleObj.Sequence,
			"total_cycles": billingCycleObj.TotalCycles,
			"tenure_type":  strings.ToLower(string(billingCycleObj.TenureType)),
			"frequency": []map[string]interface{}{{
				"interval_unit":  strings.ToLower(string(billingCycleObj.Frequency.IntervalUnit)),
				"interval_count": billingCycleObj.Frequency.IntervalCount,
			}},
			"pricing_scheme": []map[string]interface{}{{
				"version": billingCycleObj.PricingScheme.Version,
				"fixed_price": []map[string]interface{}{{
					"value":         billingCycleObj.PricingScheme.FixedPrice.Value,
					"currency_code": billingCycleObj.PricingScheme.FixedPrice.Currency,
				}},
			}},
		}
		billingCycles = append(billingCycles, billingCycle)
	}

	d.SetId(subscriptionPlan.ID)
	d.Set("status", subscriptionPlan.Status)
	d.Set("product_id", subscriptionPlan.ProductId)
	d.Set("name", subscriptionPlan.Name)
	d.Set("description", subscriptionPlan.Description)
	d.Set("quantity_supported", subscriptionPlan.QuantitySupported)
	d.Set("taxes", taxes)
	d.Set("payment_preferences", paymentPreferences)
	d.Set("billing_cycle", billingCycles)

	return nil
}

// Update - Update a subscription plan in Paypal
func (r SubscriptionPlanResource) Update(d *schema.ResourceData, m interface{}) error {
	client := m.(*paypalSdk.Client)

	subscriptionPlan := r.sdkObjectFromResourceData(d)

	// Update as much as we can from subscription plan
	err := client.UpdateSubscriptionPlan(context.Background(), subscriptionPlan)
	if err != nil {
		log.Printf("Error updating subscription plan %s: %s", d.Id(), err.Error())
		return err
	}

	// Update pricing separately
	pricingSchemeUpdates := []paypalSdk.PricingSchemeUpdate{}
	for _, billingCycleObj := range subscriptionPlan.BillingCycles {
		pricingSchemeUpdates = append(pricingSchemeUpdates, paypalSdk.PricingSchemeUpdate{
			BillingCycleSequence: billingCycleObj.Sequence,
			PricingScheme:        billingCycleObj.PricingScheme,
		})
	}
	pricingErr := client.UpdateSubscriptionPlanPricing(context.Background(), subscriptionPlan.ID, pricingSchemeUpdates)
	if pricingErr != nil {
		log.Printf("Error updating subcription plan pricing %s: %s", d.Id(), pricingErr.Error())
		return err
	}

	return r.Read(d, m)
}

// Delete - Delete a subscription plan in Paypal - Subscription plans cannot be deleted
// so we will update the name with a (removed) suffix and remove our reference to it
func (r SubscriptionPlanResource) Delete(d *schema.ResourceData, m interface{}) error {

	// Deactivate and delete
	// https://developer.paypal.com/docs/api/subscriptions/v1/#plans_deactivate
	// we cannot delete, but we can deactivate
	client := m.(*paypalSdk.Client)
	err := client.DeactivateSubscriptionPlans(context.Background(), d.Id())
	if err != nil {
		log.Printf("Error deactivating subscription plan %s: %s", d.Id(), err.Error())
		return err
	}

	// Even though we can't delete it, we can remove our id reference
	d.SetId("")

	return nil
}

// sdkOjectFromResourceData Get a PayPal SDK object from resource data
func (r SubscriptionPlanResource) sdkObjectFromResourceData(d *schema.ResourceData) paypalSdk.SubscriptionPlan {
	subscriptionPlan := paypalSdk.SubscriptionPlan{
		ID:                d.Id(),
		ProductId:         d.Get("product_id").(string),
		Name:              d.Get("name").(string),
		Description:       d.Get("description").(string),
		QuantitySupported: d.Get("quantity_supported").(bool),
	}

	// Add taxes
	taxesData := d.Get("taxes").([]interface{})
	if len(taxesData) == 1 {
		taxData := taxesData[0].(map[string]interface{})
		subscriptionPlan.Taxes = &paypalSdk.Taxes{
			Percentage: taxData["percentage"].(string),
			Inclusive:  taxData["inclusive"].(bool),
		}
	}

	// Payment preferences
	paymentPreferencesData := d.Get("payment_preferences").([]interface{})
	if len(taxesData) == 1 {
		paymentPreferenceData := paymentPreferencesData[0].(map[string]interface{})
		setupFeesData := paymentPreferenceData["setup_fee"].([]interface{})
		setupFeeData := setupFeesData[0].(map[string]interface{})

		subscriptionPlan.PaymentPreferences = &paypalSdk.PaymentPreferences{
			AutoBillOutstanding:     paymentPreferenceData["auto_bill_outstanding"].(bool),
			PaymentFailureThreshold: paymentPreferenceData["payment_failure_threshold"].(int),
			SetupFee: &paypalSdk.Money{
				Currency: setupFeeData["currency_code"].(string),
				Value:    setupFeeData["value"].(string),
			},
			SetupFeeFailureAction: paypalSdk.SetupFeeFailureAction(paymentPreferenceData["setup_fee_failure_action"].(string)),
		}

	}

	// Billing cycle
	billingCyclesData := d.Get("billing_cycle").([]interface{})
	subscriptionPlan.BillingCycles = []paypalSdk.BillingCycle{}
	for _, billingCycleInterfaceData := range billingCyclesData {
		billingCycleData := billingCycleInterfaceData.(map[string]interface{})

		// Frequency
		frequenciesData := billingCycleData["frequency"].([]interface{})
		frequencyData := frequenciesData[0].(map[string]interface{})
		billingCycleFrequency := paypalSdk.Frequency{
			IntervalUnit:  paypalSdk.IntervalUnit(frequencyData["interval_unit"].(string)),
			IntervalCount: frequencyData["interval_count"].(int),
		}

		// Pricing Scheme
		pricingSchemesData := billingCycleData["pricing_scheme"].([]interface{})
		pricingSchemeData := pricingSchemesData[0].(map[string]interface{})
		billingCyclePricingScheme := paypalSdk.PricingScheme{
			Version: pricingSchemeData["version"].(int),
		}

		// Pricing scheme - Fixed price
		fixedPricesData := pricingSchemeData["fixed_price"].([]interface{})

		fixedPriceData := fixedPricesData[0].(map[string]interface{})
		billingCyclePricingSchemeDataFixedPrice := paypalSdk.Money{
			Currency: fixedPriceData["currency_code"].(string),
			Value:    fixedPriceData["value"].(string),
		}
		billingCyclePricingScheme.FixedPrice = billingCyclePricingSchemeDataFixedPrice

		// Put it all together
		billingCycle := paypalSdk.BillingCycle{
			Sequence:      billingCycleData["sequence"].(int),
			TotalCycles:   billingCycleData["total_cycles"].(int),
			TenureType:    paypalSdk.TenureType(billingCycleData["tenure_type"].(string)),
			Frequency:     billingCycleFrequency,
			PricingScheme: billingCyclePricingScheme,
		}
		subscriptionPlan.BillingCycles = append(subscriptionPlan.BillingCycles, billingCycle)

	}

	return subscriptionPlan
}

// tenureTypes List of acceptable tenure types
func (r SubscriptionPlanResource) tenureTypes() []string {
	return []string{
		"regular",
		"trial",
	}
}

// frequencyIntervalTypes List of acceptable frequency interval types
func (r SubscriptionPlanResource) frequencyIntervalTypes() []string {
	return []string{
		"day",
		"week",
		"month",
		"year",
	}
}

// paymentDefinitionTypes List of acceptable types
func (r SubscriptionPlanResource) paymentDefinitionTypes() []string {
	return []string{
		"trial",
		"regular",
	}
}

// paymentDefinitionTypes List of acceptable types
func (r SubscriptionPlanResource) paymentDefinitionFrequencies() []string {
	return []string{
		"day",
		"week",
		"month",
		"year",
	}
}

// setupFeeFailureActions List setup fee failure actions
func (r SubscriptionPlanResource) setupFeeFailureActions() []string {
	return []string{
		"continue",
		"cancel",
	}
}

// diffSuppressFuncForFloats Function to compare string versions of floats
func (r SubscriptionPlanResource) diffSuppressFuncForFloats(k, old, new string, d *schema.ResourceData) bool {
	oldFloat, oldErr := strconv.ParseFloat(old, 64)
	if oldErr != nil {
		return false
	}

	newFloat, newErr := strconv.ParseFloat(new, 64)
	if newErr != nil {
		return false
	}

	return oldFloat == newFloat
}
