//! AWS Secret Manager
//!
//!
//! ## Example configuration
//!
//! ```yaml
//! providers:
//!  aws1:
//!    kind: aws_secretsmanager
//!    # options: ...
//! ```
//! ## Options
//!
//! See [`AWSSecretsManagerOptions`]
//!
//!
#![allow(clippy::borrowed_box)]

use std::collections::BTreeMap;

use async_trait::async_trait;
use aws_config::{self, BehaviorVersion};
use aws_sdk_secretsmanager as secretsmanager;
use secretsmanager::config::{Credentials, Region};
use secretsmanager::operation::get_secret_value::GetSecretValueError;
use secretsmanager::{error::SdkError, operation::delete_secret::DeleteSecretError};
use serde_derive::{Deserialize, Serialize};

use super::ProviderKind;
use crate::config::ProviderInfo;
use crate::{
    config::{PathMap, KV},
    Error, Provider, Result,
};

fn handle_get_err(
    mode: &Mode,
    e: SdkError<GetSecretValueError>,
    pm: &PathMap,
) -> Result<Option<String>> {
    match e.into_service_error() {
        GetSecretValueError::ResourceNotFoundException(_) => {
            if mode == &Mode::Get {
                Err(Error::NotFound {
                    path: pm.path.to_string(),
                    msg: "not found".to_string(),
                })
            } else {
                // we're ok
                Ok(None)
            }
        }
        e => {
            if e.to_string().contains("marked deleted") {
                Err(Error::NotFound {
                    path: pm.path.to_string(),
                    msg: "not found".to_string(),
                })
            } else {
                Err(Error::GetError {
                    path: pm.path.to_string(),
                    msg: e.to_string(),
                })
            }
        }
    }
}

fn handle_del_err(e: SdkError<DeleteSecretError>, pm: &PathMap) -> Result<()> {
    match e.into_service_error() {
        DeleteSecretError::ResourceNotFoundException(_) => {
            // we're ok
            Ok(())
        }
        e => Err(Error::DeleteError {
            path: pm.path.to_string(),
            msg: e.to_string(),
        }),
    }
}

#[derive(PartialEq)]
enum Mode {
    Get,
    Put,
    Del,
}

///
/// # AWS provider configuration
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
pub struct AWSSecretsManagerOptions {
    pub region: Option<String>,
    pub access_key_id: Option<String>,
    pub secret_access_key: Option<String>,
    pub endpoint_url: Option<String>,
}

pub struct AWSSecretsManager {
    pub client: secretsmanager::Client,
    pub name: String,
}

impl AWSSecretsManager {
    #[must_use]
    pub fn with_client(name: &str, client: secretsmanager::Client) -> Self {
        Self {
            client,
            name: name.to_string(),
        }
    }
    /// Create a new secretsmanager provider
    ///
    /// # Errors
    ///
    /// This function will return an error if cannot create a provider
    pub async fn new(name: &str, opts: Option<AWSSecretsManagerOptions>) -> Result<Self> {
        let client = if let Some(opts) = opts {
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
            let ssmconf = secretsmanager::config::Builder::from(&config.load().await).build();
            secretsmanager::Client::from_conf(ssmconf)
        } else {
            let config = aws_config::load_defaults(BehaviorVersion::v2025_01_17()).await;
            let ssmconf = secretsmanager::config::Builder::from(&config).build();
            secretsmanager::Client::from_conf(ssmconf)
        };
        Ok(Self {
            client,
            name: name.to_string(),
        })
    }
}

async fn get_data(
    mode: &Mode,
    client: &secretsmanager::Client,
    pm: &PathMap,
) -> Result<Option<BTreeMap<String, String>>> {
    let resp = client
        .get_secret_value()
        .secret_id(&pm.path)
        .send()
        .await
        .map_or_else(
            |e| handle_get_err(mode, e, pm),
            |res| Ok(res.secret_string().map(std::string::ToString::to_string)),
        )?;

    if let Some(raw_string) = resp {
        Ok(Some(serde_json::from_str::<BTreeMap<String, String>>(
            &raw_string,
        )?))
    } else {
        Ok(None)
    }
}

async fn put_data(
    client: &secretsmanager::Client,
    pm: &PathMap,
    data: &BTreeMap<String, String>,
) -> Result<()> {
    if client
        .get_secret_value()
        .secret_id(&pm.path)
        .send()
        .await
        .is_ok()
    {
        client
            .put_secret_value()
            .set_secret_id(Some(pm.path.clone()))
            .secret_string(serde_json::to_string(&data)?)
            .send()
            .await
            .map_err(|e| Error::PutError {
                msg: e.to_string(),
                path: pm.path.clone(),
            })?;
    } else {
        client
            .create_secret()
            .set_name(Some(pm.path.clone()))
            .secret_string(serde_json::to_string(&data)?)
            .send()
            .await
            .map_err(|e| Error::PutError {
                msg: e.to_string(),
                path: pm.path.clone(),
            })?;
    };

    Ok(())
}

#[async_trait]
impl Provider for AWSSecretsManager {
    fn kind(&self) -> ProviderInfo {
        ProviderInfo {
            kind: ProviderKind::AWSSecretsManager,
            name: self.name.clone(),
        }
    }

    async fn get(&self, pm: &PathMap) -> Result<Vec<KV>> {
        get_data(&Mode::Get, &self.client, pm).await?.map_or_else(
            || Ok(vec![]),
            |data| Ok(KV::from_data(&data, pm, &self.kind())),
        )
    }

    async fn put(&self, pm: &PathMap, kvs: &[KV]) -> Result<()> {
        let mut data = get_data(&Mode::Put, &self.client, pm)
            .await?
            .unwrap_or_default();
        for kv in kvs {
            data.insert(kv.key.clone(), kv.value.clone());
        }
        put_data(&self.client, pm, &data).await
    }

    async fn del(&self, pm: &PathMap) -> Result<()> {
        if pm.keys.is_empty() {
            self.client
                .delete_secret()
                .secret_id(&pm.path)
                .send()
                .await
                .map_or_else(|e| handle_del_err(e, pm), |_| Ok(()))?;
        } else {
            let mut data = get_data(&Mode::Del, &self.client, pm)
                .await?
                .unwrap_or_default();
            for k in pm.keys.keys() {
                data.remove(k);
            }
            put_data(&self.client, pm, &data).await?;
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

    use crate::{providers::test_utils, Provider};

    #[test]
    #[cfg(not(windows))]
    fn sanity_test() {
        if env::var("RUNNER_OS").unwrap_or_default() == "macOS" {
            return;
        }

        let env: HashMap<_, _> = vec![(
            "SERVICES".to_string(),
            "iam,sts,ssm,kms,secretsmanager".to_string(),
        )]
        .into_iter()
        .collect();
        let config = LocalStackServerConfig::builder()
            .env(env)
            .port(4561)
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
                "endpoint_url": server.external_url()
            });

            let p = Box::new(
                super::AWSSecretsManager::new(
                    "aws_secretsmanager",
                    Some(serde_json::from_value(data).unwrap()),
                )
                .await
                .unwrap(),
            ) as Box<dyn Provider + Send + Sync>;

            test_utils::ProviderTest::new(p).run().await;
        });
    }
}
