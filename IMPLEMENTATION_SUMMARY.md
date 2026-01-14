# Client Registration Policy Resource Implementation

## Overview

This implementation adds support for managing Keycloak Client Registration Policies via Terraform. Client Registration Policies control how clients can be dynamically registered through the Keycloak Client Registration Service.

This addresses the upstream issue: https://github.com/keycloak/terraform-provider-keycloak/issues/715

## Files Created/Modified

### New Files

1. **keycloak/realm_client_registration_policy.go**
   - Keycloak API client methods for client registration policies
   - CRUD operations using the Keycloak Components API
   - Helper functions for config conversion

2. **provider/resource_keycloak_realm_client_registration_policy.go**
   - Terraform resource implementation
   - Schema definition with validation
   - CRUD handlers
   - Import support

3. **provider/resource_keycloak_realm_client_registration_policy_test.go**
   - Comprehensive test suite
   - Tests for basic functionality, import, validation, and updates
   - Tests for trusted-hosts policy with multiple hosts

4. **docs/resources/realm_client_registration_policy.md**
   - User documentation
   - Usage examples for different policy types
   - Configuration reference

### Modified Files

1. **keycloak/component.go**
   - Added `SubType` field to component struct to support client registration policy sub-types

2. **provider/provider.go**
   - Registered new resource `keycloak_realm_client_registration_policy`

## Features

### Supported Policy Types

The implementation supports all Keycloak client registration policy types:

1. **trusted-hosts** - Controls which hosts can register clients
2. **allowed-protocol-mappers** - Controls allowed protocol mappers
3. **allowed-client-templates** - Controls allowed client scopes
4. **consent-required** - Requires consent for registered clients
5. **scope** - Full scope disabled policy
6. **max-clients** - Maximum clients limit

### Supported Sub-Types

- **anonymous** - Applies to anonymous client registration requests
- **authenticated** - Applies to authenticated client registration requests

### Key Features

- Full CRUD operations (Create, Read, Update, Delete)
- Import support with format: `{{realmId}}/{{policyId}}`
- Validation for provider_id and sub_type fields
- Flexible config map for policy-specific settings
- Smart handling of comma-separated values (e.g., trusted-hosts)
- Comprehensive test coverage

## Usage Example

### Basic Trusted Hosts Policy

```hcl
resource "keycloak_realm" "staging" {
  realm   = "staging"
  enabled = true
}

resource "keycloak_realm_client_registration_policy" "trusted_hosts" {
  realm_id    = keycloak_realm.staging.id
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
  realm_id    = keycloak_realm.staging.id
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
  realm_id    = keycloak_realm.staging.id
  name        = "Consent Required"
  provider_id = "consent-required"
  sub_type    = "authenticated"
}
```

## Testing

The implementation includes comprehensive tests:

1. **Basic functionality** - Create and read policy
2. **Import test** - Verify import with correct format
3. **Update test** - Modify trusted-hosts list
4. **Validation tests** - Ensure invalid provider_id and sub_type are rejected
5. **Manual destroy test** - Verify recreation after manual deletion

### Running Tests

```bash
# Format code
make fmt

# Run vet
make vet

# Build provider
make build

# Run acceptance tests (requires local Keycloak)
make testacc TESTARGS='-run=TestAccKeycloakRealmClientRegistrationPolicy'
```

## API Details

### Keycloak Components API

Client Registration Policies are managed via the Keycloak Components API:

**Base endpoint:** `/admin/realms/{realm}/components`

**Provider type:** `org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy`

**Component structure:**
```json
{
  "id": "3080544e-a34a-4f75-afd3-a464b7373f2a",
  "name": "Trusted Hosts",
  "providerId": "trusted-hosts",
  "providerType": "org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy",
  "parentId": "staging",
  "subType": "anonymous",
  "config": {
    "host-sending-registration-request-must-match": ["true"],
    "client-uris-must-match": ["true"],
    "trusted-hosts": ["claude.ai", "opencode.ai"]
  }
}
```

## Implementation Notes

### Config Handling

The implementation handles config values intelligently:

- Single values are stored as `[]string` with one element
- Comma-separated values (like trusted-hosts) are split into arrays in the API
- On read, arrays are joined with commas for Terraform state
- This allows users to specify: `"trusted-hosts" = "claude.ai,opencode.ai"`

### ForceNew Fields

The following fields require resource recreation:

- `realm_id` - Policies cannot be moved between realms
- `provider_id` - Policy type is immutable
- `sub_type` - Sub-type is immutable

### Import Format

Policies can be imported using: `{{realmId}}/{{policyId}}`

Example:
```bash
terraform import keycloak_realm_client_registration_policy.trusted_hosts staging/3080544e-a34a-4f75-afd3-a464b7373f2a
```

## Build Status

The implementation successfully compiles and builds:

```bash
$ make build
# Build successful: terraform-provider-keycloak_v5.6.0+bf.2
```

## Next Steps

1. **Test with local Keycloak instance** - Run acceptance tests
2. **Create PR** - Submit to upstream repository
3. **Document in changelog** - Add entry for new resource
4. **Update examples** - Add to provider examples directory

## Related Resources

- Upstream Issue: https://github.com/keycloak/terraform-provider-keycloak/issues/715
- Keycloak Components API: https://www.keycloak.org/docs-api/latest/rest-api/index.html#_component_resource
- Client Registration: https://www.keycloak.org/docs/latest/securing_apps/index.html#_client_registration
