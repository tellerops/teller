//! Hashicorp Vault
//!
//!
//! ## Example configuration
//!
//! ```yaml
//! providers:
//!  vault1:
//!    kind: hashicorp_vault
//!    # options: ...
//! ```
//! ## Options
//!
//! See [`HashivaultOptions`] for more.
//!
#![allow(clippy::borrowed_box)]
use std::{
    collections::{BTreeMap, HashMap},
    env,
};

use async_trait::async_trait;
use serde_derive::{Deserialize, Serialize};
use vaultrs::{
    client::{VaultClient, VaultClientSettingsBuilder},
    error::ClientError,
    kv1, kv2,
};

use super::ProviderKind;
use crate::{
    config::{PathMap, ProviderInfo, KV},
    Error, Provider, Result,
};

/// # Hashicorp options
///
/// If no options provided at all, will take `VAULT_ADDR` and `VAULT_TOKEN` env variables.
/// If partial options provided, will only take what's provided.
///
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct HashivaultOptions {
    /// Vault address
    pub address: Option<String>,
    /// Vault token
    pub token: Option<String>,
    /// Vault namespace
    pub namespace: Option<String>,
}

pub struct Hashivault {
    pub client: VaultClient,
    pub name: String,
}

impl Hashivault {
    /// Create a new hashicorp vault
    ///
    /// # Errors
    ///
    /// This function will return an error if cannot create a provider
    pub fn new(name: &str, opts: Option<HashivaultOptions>) -> Result<Self> {
        let settings = if let Some(opts) = opts {
            let mut settings = VaultClientSettingsBuilder::default();

            if let Some(address) = opts.address {
                settings.address(address);
            }

            if let Some(token) = opts.token {
                settings.token(token);
            }

            if let Some(namespace) = opts.namespace {
                settings.set_namespace(namespace);
            }

            settings.build().map_err(Box::from)?
        } else {
            VaultClientSettingsBuilder::default()
                .address(env::var("VAULT_ADDR")?)
                .token(env::var("VAULT_TOKEN")?)
                .namespace(env::var("VAULT_NAMESPACE").ok())
                .build()
                .map_err(Box::from)?
        };

        let client = VaultClient::new(settings).map_err(Box::from)?;

        Ok(Self {
            client,
            name: name.to_string(),
        })
    }
}

fn parse_path(pm: &PathMap) -> Result<(&str, &str, &str)> {
    let (engine, full_path) = (pm.protocol.as_deref().unwrap_or("kv2"), pm.path.as_str());
    let (mount, path) = full_path.split_once('/').ok_or_else(|| {
        Error::Message(
            "path must have initial mount seperated by '/', e.g. `secret/foo`".to_string(),
        )
    })?;
    Ok((engine, mount, path))
}

fn xerr(pm: &PathMap, e: ClientError) -> Error {
    match e {
        ClientError::RestClientError { source } => match source {
            rustify::errors::ClientError::ServerResponseError { code, content } => {
                match (code, content.clone()) {
                    (404, Some(content))
                        if content.contains("Invalid path for a versioned K/V secrets") =>
                    {
                        Error::PathError(
                            pm.path.clone(),
                            "missing or incompatible protocol version".to_string(),
                        )
                    }
                    (404, _) => Error::NotFound {
                        path: pm.path.clone(),
                        msg: "not found".to_string(),
                    },
                    _ => Error::Message(format!("code: {code}, {content:?}")),
                }
            }
            _ => Error::Any(Box::from(source)),
        },
        ClientError::APIError {
            code: 404,
            errors: _,
        } => Error::NotFound {
            path: pm.path.clone(),
            msg: "not found".to_string(),
        },
        _ => Error::Any(Box::from(e)),
    }
}

async fn get_data(client: &VaultClient, pm: &PathMap) -> Result<BTreeMap<String, String>> {
    let (engine, mount, path) = parse_path(pm)?;
    let data = if engine == "kv2" {
        kv2::read(client, mount, path).await
    } else {
        kv1::get(client, mount, path).await
    }
    .map_err(|e| xerr(pm, e))?;

    Ok(data)
}

async fn get_data_or_empty(client: &VaultClient, pm: &PathMap) -> Result<BTreeMap<String, String>> {
    let data = match get_data(client, pm).await {
        Ok(data) => data,
        Err(Error::NotFound { path: _, msg: _ }) => BTreeMap::new(),
        Err(e) => return Err(e),
    };
    Ok(data)
}

async fn put_data(
    client: &VaultClient,
    pm: &PathMap,
    data: &BTreeMap<String, String>,
) -> Result<()> {
    let (engine, mount, path) = parse_path(pm)?;
    if engine == "kv2" {
        kv2::set(client, mount, path, data)
            .await
            .map_err(|e| xerr(pm, e))?;
    } else {
        kv1::set(
            client,
            mount,
            path,
            &data
                .iter()
                .map(|(k, v)| (k.as_str(), v.as_str()))
                .collect::<HashMap<_, _>>(),
        )
        .await
        .map_err(|e| xerr(pm, e))?;
    };
    Ok(())
}

#[async_trait]
impl Provider for Hashivault {
    fn kind(&self) -> ProviderInfo {
        ProviderInfo {
            kind: ProviderKind::Hashicorp,
            name: self.name.clone(),
        }
    }

    async fn get(&self, pm: &PathMap) -> Result<Vec<KV>> {
        Ok(KV::from_data(
            &get_data(&self.client, pm).await.map_err(|e| match e {
                Error::NotFound { path, msg } => Error::NotFound { path, msg },
                _ => Error::GetError {
                    path: pm.path.to_string(),
                    msg: e.to_string(),
                },
            })?,
            pm,
            &self.kind(),
        ))
    }

    async fn put(&self, pm: &PathMap, kvs: &[KV]) -> Result<()> {
        let mut data = get_data_or_empty(&self.client, pm)
            .await
            .map_err(|e| Error::PutError {
                path: pm.path.to_string(),
                msg: e.to_string(),
            })?;
        for kv in kvs {
            data.insert(kv.key.clone(), kv.value.clone());
        }
        put_data(&self.client, pm, &data)
            .await
            .map_err(|e| Error::PutError {
                path: pm.path.to_string(),
                msg: e.to_string(),
            })?;
        Ok(())
    }

    async fn del(&self, pm: &PathMap) -> Result<()> {
        // if pm contains specific keys, we cannot delete the path,
        // deleting a complete path may drop everything under it (a path stores a dictionary of k/v)
        // we want to remove the keys from the secret object and re-write it into its path.
        if !pm.keys.is_empty() {
            let mut data =
                get_data_or_empty(&self.client, pm)
                    .await
                    .map_err(|e| Error::DeleteError {
                        path: pm.path.to_string(),
                        msg: e.to_string(),
                    })?;
            for key in pm.keys.keys() {
                data.remove(key);
            }
            put_data(&self.client, pm, &data)
                .await
                .map_err(|e| Error::DeleteError {
                    path: pm.path.to_string(),
                    msg: e.to_string(),
                })?;
            return Ok(());
        }

        // otherwise, delete the whole path
        let (engine, mount, path) = parse_path(pm)?;
        if engine == "kv2" {
            kv2::delete_latest(&self.client, mount, path)
                .await
                .map_err(|e| xerr(pm, e))
                .map_err(|e| Error::DeleteError {
                    path: pm.path.to_string(),
                    msg: e.to_string(),
                })?;
        } else {
            kv1::delete(&self.client, mount, path)
                .await
                .map_err(|e| xerr(pm, e))
                .map_err(|e| Error::DeleteError {
                    path: pm.path.to_string(),
                    msg: e.to_string(),
                })?;
        };
        Ok(())
    }
}

#[cfg(test)]
mod tests {

    use dockertest_server::servers::hashi::{VaultServer, VaultServerConfig};
    use dockertest_server::Test;

    use super::*;
    use crate::providers::test_utils;

    #[test]
    #[cfg(not(windows))]
    fn sanity_test() {
        use std::time::Duration;

        if env::var("RUNNER_OS").unwrap_or_default() == "macOS" {
            return;
        }

        let config = VaultServerConfig::builder()
            .version("1.8.2".into())
            .build()
            .unwrap();
        let mut test = Test::new();
        test.register(config);

        test.run(|instance| async move {
            let server: VaultServer = instance.server();

            let data = serde_json::json!({
                "address": server.external_url(),
                "token": server.token
            });

            // banner is not enough, we have to wait for the image to stabilize
            tokio::time::sleep(Duration::from_secs(2)).await;

            let p = Box::new(
                super::Hashivault::new(
                    "hashicorp_vault",
                    Some(serde_json::from_value(data).unwrap()),
                )
                .unwrap(),
            ) as Box<dyn Provider + Send + Sync>;

            test_utils::ProviderTest::new(p).run().await;
        });
    }
}
