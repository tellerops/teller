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
//! Uses default GSM options location strategy (by order):
//!
//! * Use `GOOGLE_APPLICATION_CREDENTIALS`
//! * Try `$HOME/.config/gcloud/application_default_credentials.json`
//!
//! If you need specific configuration options for this provider, please request via opening an issue.
//!
use async_trait::async_trait;
use google_secretmanager1::{
    api::{AddSecretVersionRequest, Automatic, Replication, Secret, SecretPayload},
    hyper::{self, client::HttpConnector},
    hyper_rustls::{self, HttpsConnector},
    oauth2::{
        self,
        authenticator::{ApplicationDefaultCredentialsTypes, Authenticator},
        ApplicationDefaultCredentialsAuthenticator, ApplicationDefaultCredentialsFlowOpts,
    },
    SecretManager,
};

use super::ProviderKind;
use crate::{
    config::{PathMap, ProviderInfo, KV},
    Error, Provider, Result,
};

#[async_trait]
pub trait GSM {
    fn get_hub(&self) -> Option<&SecretManager<HttpsConnector<HttpConnector>>>;
    async fn list(&self, name: &str) -> Result<Vec<(String, String)>>;
    async fn get(&self, name: &str) -> Result<Option<String>>;
    async fn put(&self, name: &str, value: &str) -> Result<()>;
    async fn del(&self, name: &str) -> Result<()>;
}

pub struct GSMClient {
    hub: SecretManager<HttpsConnector<HttpConnector>>,
}

impl GSMClient {
    /// Create a GSM client
    ///
    /// # Errors
    /// Fails if cannot create the client
    pub async fn new() -> Result<Self> {
        let authenticator = resolve_auth().await.map_err(Box::from)?;

        let hub = SecretManager::new(
            hyper::Client::builder().build(
                hyper_rustls::HttpsConnectorBuilder::new()
                    .with_native_roots()
                    .https_or_http()
                    .enable_http1()
                    .enable_http2()
                    .build(),
            ),
            authenticator,
        );
        Ok(Self { hub })
    }
}

#[async_trait]
impl GSM for GSMClient {
    fn get_hub(&self) -> Option<&SecretManager<HttpsConnector<HttpConnector>>> {
        Some(&self.hub)
    }

    async fn list(&self, name: &str) -> Result<Vec<(String, String)>> {
        let hub = self.get_hub().expect("hub");

        let (_, secret) = hub
            .projects()
            .secrets_list(name)
            .doit()
            .await
            .map_err(|e| Error::ListError {
                path: name.to_string(),
                msg: e.to_string(),
            })?;

        let mut out = Vec::new();
        if let Some(secrets) = secret.secrets {
            for secret in &secrets {
                let secret_name = secret
                    .name
                    .as_ref()
                    .expect("secretmanager API should output a secret resource name");

                if let Some(value) = self.get(secret_name).await? {
                    out.push((secret_name.clone(), value));
                }
            }
        }
        Ok(out)
    }

    async fn get(&self, name: &str) -> Result<Option<String>> {
        let hub = self.get_hub().expect("hub");
        let resource = if name.contains("/versions") {
            name.to_string()
        } else {
            format!("{name}/versions/latest")
        };

        let maybe_secret = hub
            .projects()
            .secrets_versions_access(&resource)
            .doit()
            .await
            .ok();

        if let Some((_, secret)) = maybe_secret {
            let payload = secret
                .payload
                .ok_or_else(|| Error::Message(format!("no secret payload found in {resource}")))?;

            Ok(payload
                .data
                .map(|d| String::from_utf8_lossy(&d).to_string()))
        } else {
            Ok(None)
        }
    }

    async fn put(&self, name: &str, value: &str) -> Result<()> {
        let hub = self.get_hub().expect("hub");

        let res = hub.projects().secrets_get(name).doit().await;

        // attempt adding a secret if missing
        if let Err(err) = res {
            let repr = err.to_string();

            // F-you google cloud. resorting to checking a
            // string representation of the error instead of navigating
            // the horrible error story you give us.
            if repr.contains("\"NOT_FOUND\"") {
                if let Some((project, secret_id)) = name.split_once("/secrets/") {
                    hub.projects()
                        .secrets_create(
                            Secret {
                                replication: Some(Replication {
                                    automatic: Some(Automatic::default()),
                                    user_managed: None,
                                }),
                                ..Secret::default()
                            },
                            project,
                        )
                        .secret_id(secret_id)
                        .doit()
                        .await
                        .map_err(|e| Error::PutError {
                            path: name.to_string(),
                            msg: e.to_string(),
                        })?;
                }
            }
        }

        // add value under a secret version
        hub.projects()
            .secrets_add_version(
                AddSecretVersionRequest {
                    payload: Some(SecretPayload {
                        data: Some(value.as_bytes().to_vec()),
                        data_crc32c: Some(i64::from(crc32c::crc32c(value.as_bytes()))),
                    }),
                },
                name,
            )
            .doit()
            .await
            .map_err(|e| Error::PutError {
                path: name.to_string(),
                msg: e.to_string(),
            })?;

        Ok(())
    }

    async fn del(&self, name: &str) -> Result<()> {
        let hub = self.get_hub().expect("hub");

        // only delete if exists
        if let Ok(Some(_)) = self.get(name).await {
            hub.projects()
                .secrets_delete(name)
                .doit()
                .await
                .map_err(|e| Error::DeleteError {
                    path: name.to_string(),
                    msg: e.to_string(),
                })?;
        }

        Ok(())
    }
}

async fn resolve_auth() -> Result<Authenticator<oauth2::hyper_rustls::HttpsConnector<HttpConnector>>>
{
    //
    // try SA creds (via env, GOOGLE_APPLICATION_CREDENTIALS)
    //
    let service_auth = match ApplicationDefaultCredentialsAuthenticator::builder(
        ApplicationDefaultCredentialsFlowOpts::default(),
    )
    .await
    {
        ApplicationDefaultCredentialsTypes::ServiceAccount(auth) => {
            Ok(auth.build().await.map_err(Box::from)?)
        }
        ApplicationDefaultCredentialsTypes::InstanceMetadata(_) => Err(Error::Message(
            "expected sa detail, found instance metadata".to_string(),
        )),
    };
    if service_auth.is_ok() {
        return service_auth;
    }

    //
    // try user creds
    //
    let creds = home::home_dir()
        .ok_or_else(|| Error::Message("cannot find home dir".to_string()))?
        .join(".config/gcloud/application_default_credentials.json");

    let user_secret = oauth2::read_authorized_user_secret(creds)
        .await
        .map_err(Box::from)?;

    Ok(oauth2::AuthorizedUserAuthenticator::builder(user_secret)
        .build()
        .await
        .map_err(Box::from)?)
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
                path: pm.path.to_string(),
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
    use google_secretmanager1::hyper::client::HttpConnector;
    use google_secretmanager1::hyper_rustls::HttpsConnector;
    use google_secretmanager1::SecretManager;

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
        fn get_hub(&self) -> Option<&SecretManager<HttpsConnector<HttpConnector>>> {
            None
        }

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
