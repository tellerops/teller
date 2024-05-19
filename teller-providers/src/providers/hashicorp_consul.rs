//! Hashicorp Consul
//!
//!
//! ## Example configuration
//!
//! ```yaml
//! providers:
//!  consul1:
//!    kind: hashicorp_consul
//!    # options: ...
//! ```
//! ## Options
//!
//! See [`HashiCorpConsulOptions`] for more.
//!
#![allow(clippy::borrowed_box)]
use std::env;

use async_trait::async_trait;
use rs_consul::{Consul, ConsulError};
use serde_derive::{Deserialize, Serialize};

use super::ProviderKind;
use crate::{
    config::{PathMap, ProviderInfo, KV},
    Error, Provider, Result,
};

#[derive(Default, Serialize, Deserialize, Debug, Clone)]
pub struct HashiCorpConsulOptions {
    /// Consul address. if is None, search address from `CONSUL_HTTP_ADDR`
    pub address: Option<String>,
    /// Consul token. if is None, search address from `CONSUL_HTTP_TOKEN`
    pub token: Option<String>,
    /// Specifies the datacenter to query.
    pub dc: Option<String>,
}

fn to_err(pm: &PathMap, e: ConsulError) -> Error {
    match e {
        ConsulError::UnexpectedResponseCode(hyper::http::StatusCode::NOT_FOUND, _) => {
            Error::NotFound {
                path: pm.path.clone(),
                msg: "not found".to_string(),
            }
        }
        _ => Error::Any(Box::from(e)),
    }
}

pub struct HashiCorpConsul {
    pub consul: Consul,
    opts: HashiCorpConsulOptions,
    pub name: String,
}

impl HashiCorpConsul {
    #[must_use]
    pub fn with_client(name: &str, client: Consul) -> Self {
        Self {
            consul: client,
            opts: HashiCorpConsulOptions::default(),
            name: name.to_string(),
        }
    }

    /// Create a new hashicorp Consul
    ///
    /// # Errors
    ///
    /// This function will return an error if cannot create a provider
    pub fn new(name: &str, opts: Option<HashiCorpConsulOptions>) -> Result<Self> {
        let opts = opts.unwrap_or_default();

        let address = opts
            .address
            .as_ref()
            .map_or_else(
                || env::var("CONSUL_HTTP_ADDR"),
                |address| Ok(address.to_string()),
            )
            .map_err(|_| Error::Message("Consul address not present.".to_string()))?;

        let token = opts
            .token
            .as_ref()
            .map_or_else(
                || env::var("CONSUL_HTTP_TOKEN"),
                |token| Ok(token.to_string()),
            )
            .unwrap_or_default();

        Ok(Self {
            consul: Consul::new(rs_consul::Config {
                address,
                token: Some(token),
                #[allow(clippy::default_trait_access)]
                hyper_builder: Default::default(),
            }),
            opts,
            name: name.to_string(),
        })
    }
}

#[async_trait]
impl Provider for HashiCorpConsul {
    fn kind(&self) -> ProviderInfo {
        ProviderInfo {
            kind: ProviderKind::HashiCorpConsul,
            name: self.name.clone(),
        }
    }

    async fn get(&self, pm: &PathMap) -> Result<Vec<KV>> {
        let res = self
            .consul
            .read_key(rs_consul::ReadKeyRequest {
                key: &pm.path,
                datacenter: &self.opts.dc.clone().unwrap_or_default(),
                recurse: false,
                ..Default::default()
            })
            .await
            .map_err(|e| to_err(pm, e))?;

        let mut results = vec![];
        for kv_pair in res {
            let val = kv_pair.value.ok_or_else(|| Error::NotFound {
                path: pm.path.to_string(),
                msg: "value not found".to_string(),
            })?;

            let (_, key) = kv_pair.key.rsplit_once('/').unwrap_or(("", &kv_pair.key));

            // take all or slice the requested keys
            if pm.keys.is_empty() || pm.keys.contains_key(key) {
                results.push(KV::from_value(&val, key, key, pm, self.kind()));
            }
        }

        Ok(results)
    }

    async fn put(&self, pm: &PathMap, kvs: &[KV]) -> Result<()> {
        for kv in kvs {
            self.consul
                .create_or_update_key(
                    rs_consul::CreateOrUpdateKeyRequest {
                        key: &format!("{}/{}", pm.path, kv.key),
                        datacenter: &self.opts.dc.clone().unwrap_or_default(),
                        ..Default::default()
                    },
                    kv.value.as_bytes().to_vec(),
                )
                .await
                .map_err(|e| to_err(pm, e))?;
        }
        Ok(())
    }

    async fn del(&self, pm: &PathMap) -> Result<()> {
        let keys = if pm.keys.is_empty() {
            self.consul
                .read_key(rs_consul::ReadKeyRequest {
                    key: &pm.path,
                    datacenter: &self.opts.dc.clone().unwrap_or_default(),
                    recurse: true,
                    ..Default::default()
                })
                .await
                .map_err(|e| to_err(pm, e))?
                .iter()
                .map(|resp| resp.key.clone())
                .collect::<Vec<_>>()
        } else {
            pm.keys
                .keys()
                .map(|kv| format!("{}/{kv}", &pm.path))
                .collect::<Vec<_>>()
        };

        for key in keys {
            self.consul
                .delete_key(rs_consul::DeleteKeyRequest {
                    key: &key,
                    datacenter: &self.opts.dc.clone().unwrap_or_default(),
                    ..Default::default()
                })
                .await
                .map_err(|e| to_err(pm, e))?;
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {

    use dockertest_server::servers::hashi::{ConsulServer, ConsulServerConfig};
    use dockertest_server::Test;

    use super::*;
    use crate::providers::test_utils;

    const PORT: u32 = 8501;

    #[test]
    #[cfg(not(windows))]
    fn sanity_test() {
        use std::time::Duration;

        if env::var("RUNNER_OS").unwrap_or_default() == "macOS" {
            return;
        }

        let config = ConsulServerConfig::builder()
            .port(PORT)
            .version("1.15.4".into())
            .build()
            .unwrap();
        let mut test = Test::new();
        test.register(config);

        test.run(|instance| async move {
            let server: ConsulServer = instance.server();

            let data = serde_json::json!({
                "address": server.external_url(),
            });

            // banner is not enough, we have to wait for the image to stabilize
            tokio::time::sleep(Duration::from_secs(2)).await;

            let p = Box::new(
                super::HashiCorpConsul::new(
                    "hashicorp_consul",
                    Some(serde_json::from_value(data).unwrap()),
                )
                .unwrap(),
            ) as Box<dyn Provider + Send + Sync>;

            test_utils::ProviderTest::new(p).run().await;
        });
    }
}
