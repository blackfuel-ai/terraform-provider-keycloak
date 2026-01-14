package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakRealmClientRegistrationPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakRealmClientRegistrationPolicyRead,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The realm this client registration policy belongs to.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the client registration policy to find.",
			},
			"provider_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by provider ID (e.g., 'trusted-hosts', 'max-clients').",
			},
			"sub_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by sub-type ('anonymous' or 'authenticated').",
			},
			// Computed attributes
			"config": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Configuration options for the policy.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceKeycloakRealmClientRegistrationPolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	name := data.Get("name").(string)
	providerId := data.Get("provider_id").(string)
	subType := data.Get("sub_type").(string)

	policies, err := keycloakClient.GetRealmClientRegistrationPolicies(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	var matchingPolicies []*keycloak.RealmClientRegistrationPolicy
	for _, policy := range policies {
		if policy.Name != name {
			continue
		}
		if providerId != "" && policy.ProviderId != providerId {
			continue
		}
		if subType != "" && policy.SubType != subType {
			continue
		}
		matchingPolicies = append(matchingPolicies, policy)
	}

	if len(matchingPolicies) == 0 {
		return diag.Errorf("no client registration policy found with name '%s' in realm '%s'", name, realmId)
	}

	if len(matchingPolicies) > 1 {
		var ids []string
		for _, p := range matchingPolicies {
			ids = append(ids, fmt.Sprintf("%s (providerId=%s, subType=%s)", p.Id, p.ProviderId, p.SubType))
		}
		return diag.Errorf("multiple client registration policies found with name '%s': %s. Use provider_id and/or sub_type to filter.", name, strings.Join(ids, ", "))
	}

	policy := matchingPolicies[0]

	data.SetId(policy.Id)
	data.Set("provider_id", policy.ProviderId)
	data.Set("sub_type", policy.SubType)

	// Convert config from []string to map[string]string for Terraform
	configMap := make(map[string]string)
	for key, values := range policy.Config {
		if len(values) > 0 {
			if key == "trusted-hosts" && len(values) > 1 {
				configMap[key] = strings.Join(values, ",")
			} else {
				configMap[key] = values[0]
			}
		}
	}
	data.Set("config", configMap)

	return nil
}
