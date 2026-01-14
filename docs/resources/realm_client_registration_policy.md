---
page_title: "keycloak_realm_client_registration_policy Resource"
---

# keycloak\_realm\_client\_registration\_policy Resource

Allows for creating and managing Client Registration Policies within Keycloak.

Client Registration Policies control how clients can be dynamically registered through the Keycloak Client Registration Service.
These policies can be applied to anonymous or authenticated client registration requests.

## Example Usage

### Trusted Hosts Policy

```hcl
resource "keycloak_realm" "realm" {
	realm = "my-realm"
}

resource "keycloak_realm_client_registration_policy" "trusted_hosts" {
	realm_id    = keycloak_realm.realm.id
	name        = "Trusted Hosts"
	provider_id = "trusted-hosts"
	sub_type    = "anonymous"

	config = {
		"trusted-hosts"                                = "claude.ai,opencode.ai"
		"host-sending-registration-request-must-match" = "true"
		"client-uris-must-match"                       = "true"
	}
}
```

### Max Clients Policy

```hcl
resource "keycloak_realm_client_registration_policy" "max_clients" {
	realm_id    = keycloak_realm.realm.id
	name        = "Max Clients"
	provider_id = "max-clients"
	sub_type    = "anonymous"

	config = {
		"max-clients" = "100"
	}
}
```

### Consent Required Policy

```hcl
resource "keycloak_realm_client_registration_policy" "consent_required" {
	realm_id    = keycloak_realm.realm.id
	name        = "Consent Required"
	provider_id = "consent-required"
	sub_type    = "authenticated"
}
```

## Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm this client registration policy belongs to.
- `name` - (Required) The name of the client registration policy.
- `provider_id` - (Required) The type of client registration policy. Valid values are:
  - `trusted-hosts` - Controls which hosts can register clients
  - `allowed-protocol-mappers` - Controls allowed protocol mappers
  - `allowed-client-templates` - Controls allowed client scopes
  - `consent-required` - Requires consent for registered clients
  - `scope` - Full scope disabled policy
  - `max-clients` - Maximum clients limit
- `sub_type` - (Required) The sub-type of the policy. Valid values are:
  - `anonymous` - Applies to anonymous client registration requests
  - `authenticated` - Applies to authenticated client registration requests
- `config` - (Optional) Configuration options for the policy. The available options depend on the `provider_id`.

### Configuration Options by Provider

#### trusted-hosts
- `trusted-hosts` - Comma-separated list of trusted host domains
- `host-sending-registration-request-must-match` - Boolean string ("true"/"false"). If enabled, the host sending the registration request must match one of the trusted hosts
- `client-uris-must-match` - Boolean string ("true"/"false"). If enabled, client URIs must match one of the trusted hosts

#### max-clients
- `max-clients` - Maximum number of clients that can be registered

#### allowed-protocol-mappers
- `allowed-protocol-mappers` - List of allowed protocol mapper provider IDs

#### allowed-client-templates
- `allowed-client-templates` - List of allowed client scope names

#### consent-required
- No configuration options required

#### scope
- No configuration options required

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - The ID of the client registration policy.

## Import

Client registration policies can be imported using the format: `{{realmId}}/{{policyId}}`.

Example:

```bash
$ terraform import keycloak_realm_client_registration_policy.trusted_hosts my-realm/3080544e-a34a-4f75-afd3-a464b7373f2a
```
