//! Google Secret Manager
//!
//!
//! ## Example configuration
//!
//! ```yaml
//! providers:
//!  gsm1:
//!    kind: google_secretmanager
//!    # options: ...
//! ```
//! ## Options
//!
//! Authentication uses Application Default Credentials (ADC) via the official
//! Google Cloud SDK (`google-cloud-auth`). ADC is resolved automatically and
//! supports, out of the box:
//!
//! * `GOOGLE_APPLICATION_CREDENTIALS` (service-account key or user creds)
//! * `$HOME/.config/gcloud/application_default_credentials.json` — i.e. the
//!   `authorized_user` credentials written by `gcloud auth application-default login`
//! * GCE/GKE/Cloud Run metadata server
//! * impersonated / external accounts
//!
//! If you need specific configuration options for this provider, please request via opening an issue.
//!
use async_trait::async_trait;
use google_cloud_gax::paginator::ItemPaginator as _;
use google_cloud_secretmanager_v1::{
    client::SecretManagerService,
    model::{replication::Automatic, Replication, Secret, SecretPayload},
};

use super::ProviderKind;
use crate::{
    config::{PathMap, ProviderInfo, KV},
    Error, Provider, Result,
};

#[async_trait]
pub trait GSM {
    async fn list(&self, name: &str) -> Result<Vec<(String, String)>>;
    async fn get(&self, name: &str) -> Result<Option<String>>;
    async fn put(&self, name: &str, value: &str) -> Result<()>;
    async fn del(&self, name: &str) -> Result<()>;
}

pub struct GSMClient {
    client: SecretManagerService,
}

impl GSMClient {
    /// Create a GSM client
    ///
    /// # Errors
    /// Fails if cannot create the client
    pub async fn new() -> Result<Self> {
        let client = SecretManagerService::builder()
            .build()
            .await
            .map_err(Box::from)?;
        Ok(Self { client })
    }
}

#[async_trait]
impl GSM for GSMClient {
    async fn list(&self, name: &str) -> Result<Vec<(String, String)>> {
        let mut out = Vec::new();
        let mut items = self
            .client
            .list_secrets()
            .set_parent(name.to_string())
            .by_item();

        while let Some(secret) = items
            .next()
            .await
            .transpose()
            .map_err(|e| Error::ListError {
                path: name.to_string(),
                msg: e.to_string(),
            })?
        {
            let secret_name = secret.name;
            if let Some(value) = self.get(&secret_name).await? {
                out.push((secret_name, value));
            }
        }
        Ok(out)
    }

    async fn get(&self, name: &str) -> Result<Option<String>> {
        let resource = if name.contains("/versions") {
            name.to_string()
        } else {
            format!("{name}/versions/latest")
        };

        // A failed access (missing secret or no permission) yields `None`
        // rather than an error, so callers can skip it.
        match self
            .client
            .access_secret_version()
            .set_name(resource)
            .send()
            .await
        {
            Ok(resp) => Ok(resp
                .payload
                .map(|p| String::from_utf8_lossy(p.data.as_ref()).to_string())),
            Err(_) => Ok(None),
        }
    }

    async fn put(&self, name: &str, value: &str) -> Result<()> {
        // Create the secret first if it does not exist yet (HTTP 404).
        if let Err(err) = self
            .client
            .get_secret()
            .set_name(name.to_string())
            .send()
            .await
        {
            if err.http_status_code() == Some(404) {
                if let Some((project, secret_id)) = name.split_once("/secrets/") {
                    self.client
                        .create_secret()
                        .set_parent(project.to_string())
                        .set_secret_id(secret_id.to_string())
                        .set_secret(
                            Secret::new().set_replication(
                                Replication::new().set_automatic(Automatic::new()),
                            ),
                        )
                        .send()
                        .await
                        .map_err(|e| Error::PutError {
                            path: name.to_string(),
                            msg: e.to_string(),
                        })?;
                }
            }
        }

        // Add the value as a new secret version.
        self.client
            .add_secret_version()
            .set_parent(name.to_string())
            .set_payload(SecretPayload::new().set_data(value.as_bytes().to_vec()))
            .send()
            .await
            .map_err(|e| Error::PutError {
                path: name.to_string(),
                msg: e.to_string(),
            })?;

        Ok(())
    }

    async fn del(&self, name: &str) -> Result<()> {
        // only delete if exists
        if let Ok(Some(_)) = self.get(name).await {
            self.client
                .delete_secret()
                .set_name(name.to_string())
                .send()
                .await
                .map_err(|e| Error::DeleteError {
                    path: name.to_string(),
                    msg: e.to_string(),
                })?;
        }

        Ok(())
    }
}

pub struct GoogleSecretManager {
    client: Box<dyn GSM + Send + Sync>,
    pub name: String,
}

impl GoogleSecretManager {
    #[must_use]
    pub fn new(name: &str, client: Box<dyn GSM + Send + Sync>) -> Self {
        Self {
            client,
            name: name.to_string(),
        }
    }
}

#[async_trait]
impl Provider for GoogleSecretManager {
    fn kind(&self) -> ProviderInfo {
        ProviderInfo {
            kind: ProviderKind::GoogleSecretManager,
            name: self.name.clone(),
        }
    }

    async fn get(&self, pm: &PathMap) -> Result<Vec<KV>> {
        let mut out = Vec::new();
        if pm.keys.is_empty() {
            // get parameters by path
            // ("projects/1xxx34/secrets/DSN4", "foobar")
            let values = self.client.list(&pm.path).await?;

            for (resource, v) in values {
                // projects/123/secrets/FOOBAR -> FOOBAR
                //                     ^-<--<--< rsplit
                if let Some((_, key)) = resource.rsplit_once('/') {
                    out.push(KV::from_value(&v, key, key, pm, self.kind()));
                }
            }
        } else {
            for (k, v) in &pm.keys {
                let resp = self
                    .client
                    .get(&format!("{}/secrets/{}", pm.path, k))
                    .await?;
                if let Some(val) = resp {
                    out.push(KV::from_value(&val, k, v, pm, self.kind()));
                }
            }
        }

        if out.is_empty() {
            return Err(Error::NotFound {
                path: pm.path.clone(),
                msg: "path not found".to_string(),
            });
        }
        Ok(out)
    }

    async fn put(&self, pm: &PathMap, kvs: &[KV]) -> Result<()> {
        for kv in kvs {
            self.client
                .put(&format!("{}/secrets/{}", pm.path, kv.key), &kv.value)
                .await?;
        }
        Ok(())
    }

    async fn del(&self, pm: &PathMap) -> Result<()> {
        if pm.keys.is_empty() {
            let values = self.client.list(&pm.path).await?;

            for (resource, _) in values {
                self.client.del(&resource).await?;
            }
        } else {
            for k in pm.keys.keys() {
                self.client
                    .del(&format!("{}/secrets/{}", pm.path, k))
                    .await?;
            }
        }
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use std::{
        collections::BTreeMap,
        sync::{Arc, Mutex},
    };

    use async_trait::async_trait;

    use crate::{
        providers::{google_secretmanager::GSM, test_utils},
        Provider, Result,
    };

    struct MockClient {
        data: Arc<Mutex<BTreeMap<String, String>>>,
    }

    impl MockClient {
        /// Create a GSM client
        ///
        /// # Errors
        /// Fails if cannot create the client
        pub fn new() -> Self {
            Self {
                data: Arc::new(Mutex::new(BTreeMap::new())),
            }
        }
    }

    #[async_trait]
    impl GSM for MockClient {
        async fn list(&self, name: &str) -> Result<Vec<(String, String)>> {
            Ok(self
                .data
                .lock()
                .unwrap()
                .iter()
                .filter(|(k, _)| k.starts_with(name))
                .map(|(k, v)| (k.clone(), v.clone()))
                .collect::<Vec<_>>())
        }

        async fn get(&self, name: &str) -> Result<Option<String>> {
            Ok(self.data.lock().unwrap().get(name).cloned())
        }

        async fn put(&self, name: &str, value: &str) -> Result<()> {
            self.data
                .lock()
                .unwrap()
                .insert(name.to_string(), value.to_string());
            Ok(())
        }

        async fn del(&self, name: &str) -> Result<()> {
            self.data.lock().unwrap().remove(name);
            Ok(())
        }
    }

    #[tokio::test]
    async fn sanity_test() {
        let mock_client = MockClient::new();
        let p = Box::new(super::GoogleSecretManager::new(
            "test",
            Box::new(mock_client) as Box<dyn GSM + Send + Sync>,
        )) as Box<dyn Provider + Send + Sync>;

        test_utils::ProviderTest::new(p).run().await;
    }
}
