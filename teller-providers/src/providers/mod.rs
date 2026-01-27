use std::{collections::HashMap, str::FromStr};

use lazy_static::lazy_static;
use serde::{Deserialize, Serialize};
use serde_variant::to_variant_name;
use strum::{EnumIter, IntoEnumIterator};

#[cfg(test)]
mod test_utils;

#[cfg(feature = "dotenv")]
pub mod dotenv;
pub mod inmem;

#[cfg(feature = "hashicorp_vault")]
pub mod hashicorp_vault;

#[cfg(feature = "ssm")]
pub mod ssm;

#[cfg(feature = "aws_secretsmanager")]
pub mod aws_secretsmanager;

#[cfg(feature = "google_secretmanager")]
pub mod google_secretmanager;

#[cfg(feature = "hashicorp_consul")]
pub mod hashicorp_consul;

#[cfg(feature = "etcd")]
pub mod etcd;

#[cfg(feature = "azure_keyvault")]
pub mod azure_keyvault;

lazy_static! {
    pub static ref PROVIDER_KINDS: String = {
        let providers: Vec<String> = ProviderKind::iter()
            .map(|provider| provider.to_string())
            .collect();
        providers.join(", ")
    };
}
#[derive(
    Serialize, Deserialize, Debug, Clone, Default, PartialOrd, Ord, PartialEq, Eq, EnumIter,
)]
pub enum ProviderKind {
    #[serde(rename = "inmem")]
    Inmem,

    #[default]
    #[cfg(feature = "dotenv")]
    #[serde(rename = "dotenv")]
    Dotenv,

    #[cfg(feature = "hashicorp_vault")]
    #[serde(rename = "hashicorp")]
    Hashicorp,

    #[cfg(feature = "hashicorp_consul")]
    #[serde(rename = "hashicorp_consul")]
    HashiCorpConsul,

    #[cfg(feature = "ssm")]
    #[serde(rename = "ssm")]
    SSM,

    #[cfg(feature = "aws_secretsmanager")]
    #[serde(rename = "aws_secretsmanager")]
    AWSSecretsManager,

    #[cfg(feature = "google_secretmanager")]
    #[serde(rename = "google_secretmanager")]
    GoogleSecretManager,

    #[cfg(feature = "etcd")]
    #[serde(rename = "etcd")]
    Etcd,

    #[cfg(feature = "azure_keyvault")]
    #[serde(rename = "azure_keyvault")]
    AzureKeyVault,
}

impl std::fmt::Display for ProviderKind {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        to_variant_name(self).expect("only enum supported").fmt(f)
    }
}

impl FromStr for ProviderKind {
    type Err = &'static str;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let providers = Self::iter()
            .map(|provider| (provider.to_string(), provider))
            .collect::<HashMap<String, Self>>();

        providers.get(input).map_or_else(
            || Err(&PROVIDER_KINDS as &'static str),
            |provider| Ok(provider.clone()),
        )
    }
}
