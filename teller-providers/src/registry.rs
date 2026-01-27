use std::collections::{BTreeMap, HashMap};

use crate::providers::ProviderKind;
use crate::Result;
use crate::{config::ProviderCfg, Provider};

pub struct Registry {
    providers: HashMap<String, Box<dyn Provider + Sync + Send>>,
}

impl Registry {
    /// Create a registry from config
    ///
    /// # Errors
    ///
    /// This function will return an error if any provider loading failed
    pub async fn new(providers: &BTreeMap<String, ProviderCfg>) -> Result<Self> {
        let mut loaded_providers = HashMap::new();
        for (k, provider) in providers {
            let provider: Box<dyn Provider + Sync + Send> = match provider.kind {
                ProviderKind::Inmem => Box::new(crate::providers::inmem::Inmem::new(
                    k,
                    provider.options.clone(),
                )?),

                #[cfg(feature = "dotenv")]
                ProviderKind::Dotenv => Box::new(crate::providers::dotenv::Dotenv::new(
                    k,
                    provider
                        .options
                        .clone()
                        .map(serde_json::from_value)
                        .transpose()?,
                )?),
                #[cfg(feature = "hashicorp_vault")]
                ProviderKind::Hashicorp => {
                    Box::new(crate::providers::hashicorp_vault::Hashivault::new(
                        k,
                        provider
                            .options
                            .clone()
                            .map(serde_json::from_value)
                            .transpose()?,
                    )?)
                }
                #[cfg(feature = "ssm")]
                ProviderKind::SSM => {
                    Box::new(crate::providers::ssm::SSM::new(k, provider.options.clone()).await?)
                }
                #[cfg(feature = "aws_secretsmanager")]
                ProviderKind::AWSSecretsManager => Box::new(
                    crate::providers::aws_secretsmanager::AWSSecretsManager::new(
                        k,
                        provider
                            .options
                            .clone()
                            .map(serde_json::from_value)
                            .transpose()?,
                    )
                    .await?,
                ),
                #[cfg(feature = "google_secretmanager")]
                ProviderKind::GoogleSecretManager => Box::new(
                    crate::providers::google_secretmanager::GoogleSecretManager::new(
                        k,
                        Box::new(crate::providers::google_secretmanager::GSMClient::new().await?)
                            as Box<dyn crate::providers::google_secretmanager::GSM + Send + Sync>,
                    ),
                ),
                #[cfg(feature = "hashicorp_consul")]
                ProviderKind::HashiCorpConsul => {
                    Box::new(crate::providers::hashicorp_consul::HashiCorpConsul::new(
                        k,
                        provider
                            .options
                            .clone()
                            .map(serde_json::from_value)
                            .transpose()?,
                    )?)
                }
                #[cfg(feature = "etcd")]
                ProviderKind::Etcd => Box::new(
                    crate::providers::etcd::Etcd::new(
                        k,
                        provider
                            .options
                            .clone()
                            .map(serde_json::from_value)
                            .transpose()?,
                    )
                    .await?,
                ),

                #[cfg(feature = "azure_keyvault")]
                ProviderKind::AzureKeyVault => Box::new(
                    crate::providers::etcd::Etcd::new(
                        k,
                        provider
                            .options
                            .clone()
                            .map(serde_json::from_value)
                            .transpose()?,
                    )
                    .await?,
                ),
            };
            loaded_providers.insert(k.clone(), provider);
        }
        Ok(Self {
            providers: loaded_providers,
        })
    }
    #[must_use]
    #[allow(clippy::borrowed_box)]
    pub fn get(&self, name: &str) -> Option<&Box<dyn Provider + Sync + Send>> {
        self.providers.get(name)
    }
}
