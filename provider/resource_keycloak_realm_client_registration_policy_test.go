package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakRealmClientRegistrationPolicy_basic(t *testing.T) {
	t.Parallel()

	policyName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmClientRegistrationPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientRegistrationPolicy_basic(policyName),
				Check:  testAccCheckRealmClientRegistrationPolicyExists("keycloak_realm_client_registration_policy.test"),
			},
			{
				ResourceName:      "keycloak_realm_client_registration_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getRealmClientRegistrationPolicyImportId("keycloak_realm_client_registration_policy.test"),
			},
		},
	})
}

func TestAccKeycloakRealmClientRegistrationPolicy_trustedHosts(t *testing.T) {
	t.Parallel()

	policyName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmClientRegistrationPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientRegistrationPolicy_trustedHosts(policyName, "claude.ai"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmClientRegistrationPolicyExists("keycloak_realm_client_registration_policy.test"),
					resource.TestCheckResourceAttr("keycloak_realm_client_registration_policy.test", "config.trusted-hosts", "claude.ai"),
					resource.TestCheckResourceAttr("keycloak_realm_client_registration_policy.test", "config.host-sending-registration-request-must-match", "true"),
					resource.TestCheckResourceAttr("keycloak_realm_client_registration_policy.test", "config.client-uris-must-match", "true"),
				),
			},
			{
				Config: testKeycloakRealmClientRegistrationPolicy_trustedHosts(policyName, "claude.ai,opencode.ai"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmClientRegistrationPolicyExists("keycloak_realm_client_registration_policy.test"),
					resource.TestCheckResourceAttr("keycloak_realm_client_registration_policy.test", "config.trusted-hosts", "claude.ai,opencode.ai"),
				),
			},
		},
	})
}

func TestAccKeycloakRealmClientRegistrationPolicy_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var policy = &keycloak.RealmClientRegistrationPolicy{}

	policyName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmClientRegistrationPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientRegistrationPolicy_basic(policyName),
				Check:  testAccCheckRealmClientRegistrationPolicyFetch("keycloak_realm_client_registration_policy.test", policy),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteRealmClientRegistrationPolicy(testCtx, policy.RealmId, policy.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealmClientRegistrationPolicy_basic(policyName),
				Check:  testAccCheckRealmClientRegistrationPolicyFetch("keycloak_realm_client_registration_policy.test", policy),
			},
		},
	})
}

func TestAccKeycloakRealmClientRegistrationPolicy_providerIdValidation(t *testing.T) {
	t.Parallel()

	policyName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmClientRegistrationPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakRealmClientRegistrationPolicy_withProviderId(policyName, "invalid-provider"),
				ExpectError: regexp.MustCompile("expected provider_id to be one of .+, got invalid-provider"),
			},
			{
				Config: testKeycloakRealmClientRegistrationPolicy_withProviderId(policyName, "trusted-hosts"),
				Check:  testAccCheckRealmClientRegistrationPolicyExists("keycloak_realm_client_registration_policy.test"),
			},
		},
	})
}

func TestAccKeycloakRealmClientRegistrationPolicy_subTypeValidation(t *testing.T) {
	t.Parallel()

	policyName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmClientRegistrationPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakRealmClientRegistrationPolicy_withSubType(policyName, "invalid-subtype"),
				ExpectError: regexp.MustCompile("expected sub_type to be one of .+, got invalid-subtype"),
			},
			{
				Config: testKeycloakRealmClientRegistrationPolicy_withSubType(policyName, "anonymous"),
				Check:  testAccCheckRealmClientRegistrationPolicyExists("keycloak_realm_client_registration_policy.test"),
			},
			{
				Config: testKeycloakRealmClientRegistrationPolicy_withSubType(policyName, "authenticated"),
				Check:  testAccCheckRealmClientRegistrationPolicyExists("keycloak_realm_client_registration_policy.test"),
			},
		},
	})
}

func testAccCheckRealmClientRegistrationPolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakRealmClientRegistrationPolicyFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckRealmClientRegistrationPolicyFetch(resourceName string, policy *keycloak.RealmClientRegistrationPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedPolicy, err := getKeycloakRealmClientRegistrationPolicyFromState(s, resourceName)
		if err != nil {
			return err
		}

		policy.Id = fetchedPolicy.Id
		policy.RealmId = fetchedPolicy.RealmId

		return nil
	}
}

func testAccCheckRealmClientRegistrationPolicyDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_client_registration_policy" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			policy, _ := keycloakClient.GetRealmClientRegistrationPolicy(testCtx, realm, id)
			if policy != nil {
				return fmt.Errorf("client registration policy with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakRealmClientRegistrationPolicyFromState(s *terraform.State, resourceName string) (*keycloak.RealmClientRegistrationPolicy, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	policy, err := keycloakClient.GetRealmClientRegistrationPolicy(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting client registration policy with id %s: %s", id, err)
	}

	return policy, nil
}

func getRealmClientRegistrationPolicyImportId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realmId := rs.Primary.Attributes["realm_id"]

		return fmt.Sprintf("%s/%s", realmId, id), nil
	}
}

func testKeycloakRealmClientRegistrationPolicy_basic(policyName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_client_registration_policy" "test" {
	realm_id    = data.keycloak_realm.realm.id
	name        = "%s"
	provider_id = "trusted-hosts"
	sub_type    = "anonymous"

	config = {
		"trusted-hosts"                                = "example.com"
		"host-sending-registration-request-must-match" = "true"
		"client-uris-must-match"                       = "true"
	}
}
	`, testAccRealmUserFederation.Realm, policyName)
}

func testKeycloakRealmClientRegistrationPolicy_trustedHosts(policyName, trustedHosts string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_client_registration_policy" "test" {
	realm_id    = data.keycloak_realm.realm.id
	name        = "%s"
	provider_id = "trusted-hosts"
	sub_type    = "anonymous"

	config = {
		"trusted-hosts"                                = "%s"
		"host-sending-registration-request-must-match" = "true"
		"client-uris-must-match"                       = "true"
	}
}
	`, testAccRealmUserFederation.Realm, policyName, trustedHosts)
}

func testKeycloakRealmClientRegistrationPolicy_withProviderId(policyName, providerId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_client_registration_policy" "test" {
	realm_id    = data.keycloak_realm.realm.id
	name        = "%s"
	provider_id = "%s"
	sub_type    = "anonymous"
}
	`, testAccRealmUserFederation.Realm, policyName, providerId)
}

func testKeycloakRealmClientRegistrationPolicy_withSubType(policyName, subType string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_client_registration_policy" "test" {
	realm_id    = data.keycloak_realm.realm.id
	name        = "%s"
	provider_id = "trusted-hosts"
	sub_type    = "%s"
}
	`, testAccRealmUserFederation.Realm, policyName, subType)
}
