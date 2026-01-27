//! AWS SSM
//!
//!
//! ## Example configuration
//!
//! ```yaml
//! providers:
//!  ssm1:
//!    kind: ssm
//!    # options: ...
//! ```
//! ## Options
//!
//! See [`SSMOptions`]
//!
//!
#![allow(clippy::borrowed_box)]
use async_trait::async_trait;
use aws_config::{self, BehaviorVersion};
use aws_sdk_ssm as ssm;
use serde_derive::{Deserialize, Serialize};
use ssm::config::{Credentials, Region};
use ssm::{
    error::SdkError, operation::delete_parameter::DeleteParameterError, types::ParameterType,
};

use super::ProviderKind;
use crate::config::{PathMap, ProviderInfo, KV};
use crate::Provider;
use crate::{Error, Result};

fn handle_delete(e: SdkError<DeleteParameterError>, pm: &PathMap) -> Result<()> {
    match e.into_service_error() {
        DeleteParameterError::ParameterNotFound(_) => {
            // we're ok
            Ok(())
        }
        e => Err(Error::DeleteError {
            path: pm.path.to_string(),
            msg: e.to_string(),
        }),
    }
}

fn join_path(left: &str, right: &str) -> String {
    format!(
        "{}/{}",
        left.trim_end_matches('/'),
        right.trim_start_matches('/')
    )
}
/// # AWS SSM configuration
///
/// This holds the most commonly used and simplified configuration options for this provider. These
/// paramters can be used in the Teller YAML configuration.
///
/// For indepth description of each parameter see: [AWS SDK config](https://docs.rs/aws-config/latest/aws_config/struct.SdkConfig.html)
///
/// If you need an additional parameter from the AWS SDK included in our simplified configuration,
/// open an issue in Teller and request to add it.
///
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct SSMOptions {
    pub region: Option<String>,
    pub access_key_id: Option<String>,
    pub secret_access_key: Option<String>,
    pub endpoint_url: Option<String>,
}

pub struct SSM {
    pub name: String,
    pub client: ssm::Client,
}
impl SSM {
    #[must_use]
    pub fn with_client(name: &str, client: ssm::Client) -> Self {
        Self {
            name: name.to_string(),
            client,
        }
    }

    /// Create a new ssm provider
    ///
    /// # Errors
    ///
    /// This function will return an error if cannot create a provider
    pub async fn new(name: &str, opts: Option<serde_json::Value>) -> Result<Self> {
        let client = if let Some(opts) = opts {
            let opts: SSMOptions = serde_json::from_value(opts)?;

            let mut config = aws_config::defaults(BehaviorVersion::v2025_01_17());
            if let (Some(key), Some(secret)) = (opts.access_key_id, opts.secret_access_key) {
                config = config
                    .credentials_provider(Credentials::new(key, secret, None, None, "teller"));
            }
            if let Some(endpoint_url) = opts.endpoint_url {
                config = config.endpoint_url(endpoint_url);
            }
            if let Some(region) = opts.region {
                config = config.region(Region::new(region));
            }
            let ssmconf = ssm::config::Builder::from(&config.load().await).build();
            ssm::Client::from_conf(ssmconf)
        } else {
            let config = aws_config::load_defaults(BehaviorVersion::v2025_01_17()).await;
            let ssmconf = ssm::config::Builder::from(&config).build();
            ssm::Client::from_conf(ssmconf)
        };
        Ok(Self {
            client,
            name: name.to_string(),
        })
    }
}

#[async_trait]
impl Provider for SSM {
    fn kind(&self) -> ProviderInfo {
        ProviderInfo {
            kind: ProviderKind::SSM,
            name: self.name.clone(),
        }
    }

    async fn get(&self, pm: &PathMap) -> Result<Vec<KV>> {
        let mut out = Vec::new();
        if pm.keys.is_empty() {
            // get parameters by path, auto paginate, sends multiple requests
            let resp = self
                .client
                .get_parameters_by_path()
                .path(&pm.path)
                .with_decryption(pm.decrypt)
                .into_paginator()
                .send()
                .collect::<std::result::Result<Vec<_>, _>>()
                .await
                .map_err(|e| Error::GetError {
                    msg: e.to_string(),
                    path: pm.path.clone(),
                })?;

            // sematics: total pages empty or *first page* empty is a 404
            if resp.is_empty()
                || resp
                    .first()
                    .and_then(|params| params.parameters.as_ref())
                    .is_some_and(Vec::is_empty)
            {
                return Err(Error::NotFound {
                    msg: "not found".to_string(),
                    path: pm.path.clone(),
                });
            }

            for params in resp {
                for p in params.parameters.unwrap_or_default() {
                    let ssm_key = p.name().unwrap_or_default();
                    if !ssm_key.starts_with(&pm.path) {
                        return Err(Error::GetError {
                            path: pm.path.clone(),
                            msg: format!("{ssm_key} is not contained in root path"),
                        });
                    }

                    let relative_key = ssm_key
                        .strip_prefix(&pm.path)
                        .map_or(ssm_key, |k| k.trim_start_matches('/'));

                    out.push(KV::from_value(
                        p.value().unwrap_or_default(),
                        relative_key,
                        relative_key,
                        pm,
                        self.kind(),
                    ));
                }
            }
        } else {
            for (k, v) in &pm.keys {
                let resp = self
                    .client
                    .get_parameter()
                    .name(join_path(&pm.path, k))
                    .with_decryption(pm.decrypt)
                    .send()
                    .await
                    .map_err(|e| Error::GetError {
                        msg: e.to_string(),
                        path: pm.path.clone(),
                    })?;
                let param = resp.parameter();
                if let Some(p) = param {
                    out.push(KV::from_value(
                        p.value().unwrap_or_default(),
                        k,
                        v,
                        pm,
                        self.kind(),
                    ));
                }
            }
        }

        Ok(out)
    }

    async fn put(&self, pm: &PathMap, kvs: &[KV]) -> Result<()> {
        for kv in kvs {
            // proper separator sensitive concat
            let path = format!("{}/{}", pm.path, kv.key);
            self.client
                .put_parameter()
                .name(&path)
                .value(&kv.value)
                .overwrite(true)
                .r#type(ParameterType::String)
                .send()
                .await
                .map_err(|e| Error::PutError {
                    msg: e.to_string(),
                    path,
                })?;
        }
        Ok(())
    }

    async fn del(&self, pm: &PathMap) -> Result<()> {
        let paths = if pm.keys.is_empty() {
            let kvs = self.get(pm).await?;
            kvs.iter()
                .map(|kv| join_path(&pm.path, &kv.key))
                .collect::<Vec<_>>()
        } else {
            pm.keys
                .keys()
                .map(|k| join_path(&pm.path, k))
                .collect::<Vec<_>>()
        };

        for path in paths {
            let res = self.client.delete_parameter().name(path).send().await;
            res.map_or_else(|e| handle_delete(e, pm), |_| Ok(()))?;
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use std::collections::HashMap;
    use std::env;

    use dockertest_server::servers::cloud::LocalStackServer;
    use dockertest_server::servers::cloud::LocalStackServerConfig;
    use dockertest_server::Test;

    use super::*;
    use crate::providers::test_utils;

    #[test]
    #[cfg(not(windows))]
    fn sanity_test() {
        if env::var("RUNNER_OS").unwrap_or_default() == "macOS" {
            return;
        }

        let env: HashMap<_, _> = vec![("SERVICES".to_string(), "iam,sts,ssm,kms".to_string())]
            .into_iter()
            .collect();
        let config = LocalStackServerConfig::builder()
            .env(env)
            .port(4551)
            .version("2.0.2".into())
            .build()
            .unwrap();
        let mut test = Test::new();
        test.register(config);

        test.run(|instance| async move {
            let server: LocalStackServer = instance.server();
            let data = serde_json::json!({
                "region": "us-east-1",
                "access_key_id": "stub",
                "secret_access_key": "stub",
                "provider_name": "faked",
                "endpoint_url": server.external_url(),
            });

            let p = Box::new(super::SSM::new("ssm", Some(data)).await.unwrap())
                as Box<dyn Provider + Send + Sync>;

            test_utils::ProviderTest::new(p)
                .with_root_prefix("/")
                .run()
                .await;
        });
    }
}
