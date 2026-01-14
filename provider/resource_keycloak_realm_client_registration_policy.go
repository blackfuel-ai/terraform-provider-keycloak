package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var (
	keycloakRealmClientRegistrationPolicyProviderIds = []string{
		"trusted-hosts",
		"allowed-protocol-mappers",
		"allowed-client-templates",
		"consent-required",
		"scope",
		"max-clients",
	}
	keycloakRealmClientRegistrationPolicySubTypes = []string{
		"anonymous",
		"authenticated",
	}
)

func resourceKeycloakRealmClientRegistrationPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmClientRegistrationPolicyCreate,
		ReadContext:   resourceKeycloakRealmClientRegistrationPolicyRead,
		UpdateContext: resourceKeycloakRealmClientRegistrationPolicyUpdate,
		DeleteContext: resourceKeycloakRealmClientRegistrationPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakRealmClientRegistrationPolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The realm this client registration policy belongs to.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the client registration policy.",
			},
			"provider_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(keycloakRealmClientRegistrationPolicyProviderIds, false),
				Description:  "The type of client registration policy. Valid values are: trusted-hosts, allowed-protocol-mappers, allowed-client-templates, consent-required, scope, max-clients.",
			},
			"sub_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(keycloakRealmClientRegistrationPolicySubTypes, false),
				Description:  "The sub-type of the policy. Valid values are: anonymous (for anonymous client registration), authenticated (for authenticated client registration).",
			},
			"config": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Configuration options for the policy. The available options depend on the provider_id.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func getRealmClientRegistrationPolicyFromData(data *schema.ResourceData) *keycloak.RealmClientRegistrationPolicy {
	config := make(map[string][]string)

	if v, ok := data.GetOk("config"); ok {
		configMap := v.(map[string]interface{})
		for key, value := range configMap {
			strValue := value.(string)
			// Handle comma-separated values for fields like trusted-hosts
			if key == "trusted-hosts" && strings.Contains(strValue, ",") {
				// Split by comma and trim spaces
				values := strings.Split(strValue, ",")
				for i, val := range values {
					values[i] = strings.TrimSpace(val)
				}
				config[key] = values
			} else {
				config[key] = []string{strValue}
			}
		}
	}

	policy := &keycloak.RealmClientRegistrationPolicy{
		Id:         data.Id(),
		Name:       data.Get("name").(string),
		RealmId:    data.Get("realm_id").(string),
		ProviderId: data.Get("provider_id").(string),
		SubType:    data.Get("sub_type").(string),
		Config:     config,
	}

	return policy
}

func setRealmClientRegistrationPolicyData(data *schema.ResourceData, policy *keycloak.RealmClientRegistrationPolicy) error {
	data.SetId(policy.Id)

	if err := data.Set("realm_id", policy.RealmId); err != nil {
		return err
	}
	if err := data.Set("name", policy.Name); err != nil {
		return err
	}
	if err := data.Set("provider_id", policy.ProviderId); err != nil {
		return err
	}
	if err := data.Set("sub_type", policy.SubType); err != nil {
		return err
	}

	// Convert config from []string to map[string]string for Terraform
	configMap := make(map[string]string)
	for key, values := range policy.Config {
		if len(values) > 0 {
			// Join array values with comma for fields like trusted-hosts
			if key == "trusted-hosts" && len(values) > 1 {
				configMap[key] = strings.Join(values, ",")
			} else {
				configMap[key] = values[0]
			}
		}
	}

	if err := data.Set("config", configMap); err != nil {
		return err
	}

	return nil
}

func resourceKeycloakRealmClientRegistrationPolicyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	policy := getRealmClientRegistrationPolicyFromData(data)

	err := keycloakClient.NewRealmClientRegistrationPolicy(ctx, policy)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := setRealmClientRegistrationPolicyData(data, policy); err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmClientRegistrationPolicyRead(ctx, data, meta)
}

func resourceKeycloakRealmClientRegistrationPolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	policy, err := keycloakClient.GetRealmClientRegistrationPolicy(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	if err := setRealmClientRegistrationPolicyData(data, policy); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmClientRegistrationPolicyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	policy := getRealmClientRegistrationPolicyFromData(data)

	err := keycloakClient.UpdateRealmClientRegistrationPolicy(ctx, policy)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := setRealmClientRegistrationPolicyData(data, policy); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmClientRegistrationPolicyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteRealmClientRegistrationPolicy(ctx, realmId, id))
}

func resourceKeycloakRealmClientRegistrationPolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")

	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{policyId}}")
	}

	d.Set("realm_id", parts[0])
	d.SetId(parts[1])

	return []*schema.ResourceData{d}, nil
}
