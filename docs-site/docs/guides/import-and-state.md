# Import and state

Import establishes Terraform ownership of existing AIRS infrastructure. Write the matching resource configuration first and use the identifier accepted by that resource.

| Resource | Import identifier |
| --- | --- |
| Security profile | Profile name; selects highest numeric revision |
| Custom topic | Topic ID |
| API key | API-key ID |
| Customer app | Existing application name |
| Model Security group | Group UUID |
| Red Team target | UUID or matching `family/uuid` hint |
| Custom prompt set | Prompt-set UUID |

```bash
terraform import prisma-airs_security_profile.production production-profile
terraform import prisma-airs_customer_app.chatbot customer-support-chatbot
terraform import prisma-airs_red_team_target.app rest/00000000-0000-0000-0000-000000000000
terraform plan
```

Use real identifiers from your tenant. Import supports active objects; archived prompt sets and tombstoned model groups are rejected.

## Versioned profile ownership

The security-profile resource owns all revisions under a name. Refresh follows the highest numeric revision regardless of active status, including changes made outside Terraform. Policy updates retain the resource address and create a new UUID/revision. Renaming transfers Terraform to the new name and leaves old history in AIRS.

:::warning[Destroy scope]

Destroy deletes every revision under the currently managed name, including revisions predating import. Confirm this ownership boundary before importing a shared profile.

:::

## Secrets and unavailable values

An imported API key has a null key secret because the one-time secret cannot be retrieved. Imported targets have null values for unrecoverable credentials, payloads, request headers, and omitted native response keys. Configure desired values before editing connection settings.

`api_key`, `auth_code`, deployment-profile `profile_id` and `details`, and target authentication values are sensitive. Sensitivity hides ordinary terminal output but retains secrets in Terraform state and saved plans. Use an access-controlled backend and sensitive output declarations.

## Service deletion behavior

| Object | Result |
| --- | --- |
| Named security profile | Deletes all revisions of the managed name |
| API key | Deletes key; the service can cascade deletion to its customer app |
| Model Security group | Leaves a tombstone; Terraform treats it as absent |
| Custom prompt set | Archives the set; Terraform treats it as absent |
| Network Broker channel | Externally managed; target operations leave it untouched |

If an imported customer app disappears, refresh removes it from state. The next plan proposes creation, which the import-only resource cannot perform. Restore or establish the app externally and import it again.
