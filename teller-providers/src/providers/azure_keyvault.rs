/// Azure Key Vault Provider
///
/// Uses Azure Key Vault as a secrets engine.
///
/// ```yaml
/// providers:
///   azurekv:
///     kind: azure_keyvault
///     options:
///       vault_name: "your-vault-name" # Provide the name, not the full URL
/// ```
///
/// **Authentication:** Uses [`DefaultAzureCredential`](https://docs.rs/azure_identity/latest/azure_identity/struct.DefaultAzureCredential.html) which supports various methods like environment variables, managed identity, Azure CLI, etc.
///
/// **Options:**
///
/// *   `vault_name` (required): The name of your Azure Key Vault instance (e.g., `my-vault`). The URL will be constructed as `https://<vault_name>.vault.azure.net/`.
use async_trait::async_trait;
use azure_identity::DefaultAzureCredentialBuilder; // Use builder pattern
use azure_security_keyvault::SecretClient;

use crate::providers::ProviderKind;
use crate::{
    config::{PathMap, ProviderInfo, KV},
    Error, Provider, Result,
}; // Import ProviderInfo // Import ProviderKind for ProviderInfo

// Add Debug derive for test panic message
#[derive(Debug)]
pub struct AzureKeyVault {
    client: SecretClient,
    name: String,
    kind: ProviderKind, // Store kind for ProviderInfo
}

impl AzureKeyVault {
    pub async fn new(name: &str, vault_name: &str) -> Result<Self> {
        // Handle the Result from the builder
        let credential = DefaultAzureCredentialBuilder::new().build().map_err(|e| {
            Error::CreateProviderError(format!(
                "Failed to build Azure credential for '{}': {}",
                name, e
            ))
        })?;
        let credential_arc = std::sync::Arc::new(credential);

        let client = SecretClient::new(
            &format!("https://{}.vault.azure.net/", vault_name),
            credential_arc,
        )
        .map_err(|e| {
            Error::CreateProviderError(format!(
                "Failed to create Azure Key Vault client for '{}': {}",
                name, e
            ))
        })?;
        Ok(Self {
            client,
            name: name.to_string(),
            kind: ProviderKind::AzureKeyVault,
        })
    }
}

#[async_trait]
impl Provider for AzureKeyVault {
    // Return ProviderInfo struct
    fn kind(&self) -> ProviderInfo {
        ProviderInfo {
            kind: self.kind.clone(),
            name: self.name.clone(),
        }
    }

    async fn get(&self, pm: &PathMap) -> Result<Vec<KV>> {
        let secret_name = &pm.path;
        let secret = self
            .client
            .get(secret_name)
            .await
            .map_err(|e| Error::GetError {
                path: secret_name.clone(),
                msg: format!("Provider '{}': {}", self.name, e), // Format details into msg
            })?;

        let value = secret.value;
        // Pass &str to from_value
        Ok(vec![KV::from_value(
            &value,
            secret_name,
            secret_name,
            pm,
            self.kind(),
        )])
    }

    // Mark pm as unused
    async fn put(&self, _pm: &PathMap, kvs: &[KV]) -> Result<()> {
        for kv in kvs {
            self.client
                .set(&kv.key, &kv.value)
                .await
                .map_err(|e| Error::PutError {
                    path: kv.key.clone(),
                    msg: format!("Provider '{}': {}", self.name, e),
                })?;
        }
        Ok(())
    }

    // Rename method to del
    async fn del(&self, pm: &PathMap) -> Result<()> {
        // The delete operation returns a DeletedSecretBundle, we just need to check for error
        self.client
            .delete(&pm.path)
            .await
            .map_err(|e| Error::DeleteError {
                path: pm.path.clone(),
                msg: format!("Provider '{}': {}", self.name, e), // Format details into msg
            })?;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use chrono::Utc;
    use tokio;

    use super::*;
    use crate::config::{MetaInfo, PathInfo, PathMap, Sensitivity, KV};
    // Helper to create a provider instance using the constant vault name
    async fn create_provider() -> AzureKeyVault {
        AzureKeyVault::new("test-akv", &"tellertestkv")
            .await
            .expect(
                "Failed to create AzureKeyVault provider for testing. Ensure TEST_VAULT_NAME \
                 constant is set correctly.",
            )
    }

    #[tokio::test]
    #[ignore]
    async fn test_azure_keyvault_put_get_delete() {
        let provider = create_provider().await;
        let provider_info = provider.kind(); // Get provider info
        let test_key = format!("teller-test-secret-{}", Utc::now().timestamp_millis());
        let test_value = "test-value";

        // Use correct PathMap fields
        let pm = PathMap {
            id: "test-map".to_string(), // Add id
            path: test_key.clone(),
            keys: Default::default(),       // Add keys
            sensitivity: Sensitivity::None, // Add sensitivity
            // other fields default
            ..Default::default()
        };

        // Use correct KV fields and types
        let kv = KV {
            key: test_key.clone(),
            value: test_value.to_string(), // Value should be String
            from_key: test_key.clone(),
            path: Some(PathInfo {
                id: pm.id.clone(),
                path: pm.path.clone(),
            }),
            provider: Some(provider_info.clone()),
            meta: Some(MetaInfo {
                sensitivity: Sensitivity::High, // Example sensitivity
                ..Default::default()
            }),
        };

        // PUT
        provider
            .put(&pm, &[kv.clone()])
            .await
            .expect("Failed to put secret");
        println!("Successfully put secret: {}", test_key);

        // GET
        let fetched_kvs = provider.get(&pm).await.expect("Failed to get secret");
        assert_eq!(fetched_kvs.len(), 1);
        assert_eq!(fetched_kvs[0].key, test_key);
        assert_eq!(fetched_kvs[0].value, test_value); // Compare String with &str
        println!("Successfully got secret: {}", test_key);

        // DELETE (call del)
        provider.del(&pm).await.expect("Failed to delete secret");
        println!("Successfully deleted secret: {}", test_key);

        // Verify deletion (GET should fail)
        let get_result = provider.get(&pm).await;
        assert!(get_result.is_err());
        // Check for the correct error variant
        if let Err(Error::GetError { path, .. }) = get_result {
            assert_eq!(path, test_key);
            println!(
                "Verified secret deletion (get failed as expected): {}",
                test_key
            );
        } else {
            panic!("Expected GetError after deletion, but got {:?}", get_result);
        }
    }

    #[tokio::test]
    #[ignore]
    async fn test_azure_keyvault_get_nonexistent() {
        let provider = create_provider().await;
        let test_key = "teller-nonexistent-secret";
        // Use correct PathMap fields
        let pm = PathMap {
            id: "test-nonexistent".to_string(),
            path: test_key.to_string(),
            ..Default::default()
        };

        let get_result = provider.get(&pm).await;
        assert!(get_result.is_err());
        // Check for the correct error variant
        if let Err(Error::GetError { path, .. }) = get_result {
            assert_eq!(path, test_key);
        } else {
            panic!(
                "Expected GetError for non-existent secret, but got {:?}",
                get_result
            );
        }
    }
}
