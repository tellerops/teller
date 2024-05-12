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
use consulrs::{
    api::kv::requests::{
        DeleteKeyRequestBuilder, ReadKeyRequestBuilder, ReadKeysRequestBuilder,
        SetKeyRequestBuilder,
    },
    client::{ConsulClient, ConsulClientSettingsBuilder},
    error::ClientError,
    kv as ConsulKV,
};
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
            message: _,
        } => Error::NotFound {
            path: pm.path.clone(),
            msg: "not found".to_string(),
        },
        _ => Error::Any(Box::from(e)),
    }
}

pub struct HashiCorpConsul {
    pub client: ConsulClient,
    opts: HashiCorpConsulOptions,
    pub name: String,
}

impl HashiCorpConsul {
    #[must_use]
    pub fn with_client(name: &str, client: ConsulClient) -> Self {
        Self {
            client,
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

        let settings = ConsulClientSettingsBuilder::default()
            .address(address)
            .token(token)
            .build()
            .map_err(Box::from)?;

        let client = ConsulClient::new(settings).map_err(Box::from)?;

        Ok(Self {
            client,
            opts,
            name: name.to_string(),
        })
    }
}

impl HashiCorpConsul {
    fn prepare_get_builder_request(&self) -> ReadKeyRequestBuilder {
        let mut opts: ReadKeyRequestBuilder = ReadKeyRequestBuilder::default();
        if let Some(dc) = self.opts.dc.as_ref() {
            opts.dc(dc.to_string());
        }
        opts.recurse(true);
        opts
    }

    fn prepare_put_builder_request(&self) -> SetKeyRequestBuilder {
        let mut opts: SetKeyRequestBuilder = SetKeyRequestBuilder::default();
        if let Some(dc) = self.opts.dc.as_ref() {
            opts.dc(dc.to_string());
        }
        opts
    }

    fn prepare_delete_builder_request(&self) -> DeleteKeyRequestBuilder {
        let mut opts: DeleteKeyRequestBuilder = DeleteKeyRequestBuilder::default();
        if let Some(dc) = self.opts.dc.as_ref() {
            opts.dc(dc.to_string());
        }
        opts.recurse(false);
        opts
    }

    fn prepare_keys_builder_request(&self) -> ReadKeysRequestBuilder {
        let mut opts: ReadKeysRequestBuilder = ReadKeysRequestBuilder::default();
        if let Some(dc) = self.opts.dc.as_ref() {
            opts.dc(dc.to_string());
        }
        opts.recurse(false);
        opts
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
        let res = ConsulKV::read(
            &self.client,
            &pm.path,
            Some(&mut self.prepare_get_builder_request()),
        )
        .await
        .map_err(|e| xerr(pm, e))?;

        let mut results = vec![];
        for kv_pair in res.response {
            let kv_value = kv_pair.value.ok_or_else(|| Error::NotFound {
                path: pm.path.to_string(),
                msg: "value not found".to_string(),
            })?;

            let val: String = kv_value.try_into().map_err(|e| Error::GetError {
                path: pm.path.to_string(),
                msg: format!("could not decode Base64 value. err: {e:?}"),
            })?;

            let (_, key) = kv_pair.key.rsplit_once('/').unwrap_or(("", &kv_pair.key));

            results.push(KV::from_value(&val, key, key, pm, self.kind()));
        }

        Ok(results)
    }

    async fn put(&self, pm: &PathMap, kvs: &[KV]) -> Result<()> {
        for kv in kvs {
            ConsulKV::set(
                &self.client,
                &format!("{}/{}", pm.path, kv.key),
                kv.value.as_bytes(),
                Some(&mut self.prepare_put_builder_request()),
            )
            .await
            .map_err(|e| Error::PutError {
                path: pm.path.to_string(),
                msg: format!(
                    "could not put value in key {}. err: {:?}",
                    kv.key.as_str(),
                    e
                ),
            })?;
        }
        Ok(())
    }

    async fn del(&self, pm: &PathMap) -> Result<()> {
        let keys = if pm.keys.is_empty() {
            ConsulKV::keys(
                &self.client,
                pm.path.as_str(),
                Some(&mut self.prepare_keys_builder_request()),
            )
            .await
            .map_err(|e| Error::DeleteError {
                path: pm.path.to_string(),
                msg: format!(
                    "could not get keys in path: {}. err: {:?}",
                    pm.path.as_str(),
                    e
                ),
            })?
            .response
        } else {
            pm.keys
                .keys()
                .map(|kv| format!("{}/{kv}", &pm.path))
                .collect::<Vec<_>>()
        };

        for key in keys {
            ConsulKV::delete(
                &self.client,
                key.as_str(),
                Some(&mut self.prepare_delete_builder_request()),
            )
            .await
            .map_err(|e| Error::DeleteError {
                path: pm.path.to_string(),
                msg: format!("could not delete key: {}. err: {:?}", pm.path.as_str(), e),
            })?;
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
