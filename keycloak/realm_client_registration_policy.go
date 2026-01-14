package keycloak

import (
	"context"
	"fmt"
	"strings"
)

type RealmClientRegistrationPolicy struct {
	Id         string
	Name       string
	RealmId    string
	ProviderId string
	SubType    string
	Config     map[string][]string
}

func convertFromRealmClientRegistrationPolicyToComponent(policy *RealmClientRegistrationPolicy) *component {
	return &component{
		Id:           policy.Id,
		Name:         policy.Name,
		ParentId:     policy.RealmId,
		ProviderId:   policy.ProviderId,
		ProviderType: "org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy",
		SubType:      policy.SubType,
		Config:       policy.Config,
	}
}

func convertFromComponentToRealmClientRegistrationPolicy(component *component, realmId string) *RealmClientRegistrationPolicy {
	policy := &RealmClientRegistrationPolicy{
		Id:         component.Id,
		Name:       component.Name,
		RealmId:    realmId,
		ProviderId: component.ProviderId,
		SubType:    component.SubType,
		Config:     component.Config,
	}

	return policy
}

func (keycloakClient *KeycloakClient) NewRealmClientRegistrationPolicy(ctx context.Context, policy *RealmClientRegistrationPolicy) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", policy.RealmId), convertFromRealmClientRegistrationPolicyToComponent(policy))
	if err != nil {
		return err
	}

	policy.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetRealmClientRegistrationPolicy(ctx context.Context, realmId, id string) (*RealmClientRegistrationPolicy, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	// Verify this is a client registration policy
	if component.ProviderType != "org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy" {
		return nil, fmt.Errorf("component with id %s is not a client registration policy", id)
	}

	return convertFromComponentToRealmClientRegistrationPolicy(component, realmId), nil
}

func (keycloakClient *KeycloakClient) UpdateRealmClientRegistrationPolicy(ctx context.Context, policy *RealmClientRegistrationPolicy) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", policy.RealmId, policy.Id), convertFromRealmClientRegistrationPolicyToComponent(policy))
}

func (keycloakClient *KeycloakClient) DeleteRealmClientRegistrationPolicy(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}

// GetRealmClientRegistrationPolicies returns all client registration policies for a realm
func (keycloakClient *KeycloakClient) GetRealmClientRegistrationPolicies(ctx context.Context, realmId string) ([]*RealmClientRegistrationPolicy, error) {
	var components []*component

	params := map[string]string{
		"type": "org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy",
	}

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components", realmId), &components, params)
	if err != nil {
		return nil, err
	}

	var policies []*RealmClientRegistrationPolicy
	for _, component := range components {
		policies = append(policies, convertFromComponentToRealmClientRegistrationPolicy(component, realmId))
	}

	return policies, nil
}

// Helper function to convert config map with single values to []string format
func convertConfigMapToArrayFormat(config map[string]string) map[string][]string {
	result := make(map[string][]string)
	for key, value := range config {
		// For comma-separated values like trusted-hosts, split into array
		if key == "trusted-hosts" && strings.Contains(value, ",") {
			result[key] = strings.Split(value, ",")
		} else {
			result[key] = []string{value}
		}
	}
	return result
}

// Helper function to convert config array format to single values
func convertConfigArrayToMapFormat(config map[string][]string) map[string]string {
	result := make(map[string]string)
	for key, values := range config {
		if len(values) > 0 {
			// For arrays like trusted-hosts, join with comma
			if key == "trusted-hosts" && len(values) > 1 {
				result[key] = strings.Join(values, ",")
			} else {
				result[key] = values[0]
			}
		}
	}
	return result
}
