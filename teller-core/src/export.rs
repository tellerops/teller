use std::collections::BTreeMap;
use std::str::FromStr;

use csv::WriterBuilder;
use lazy_static::lazy_static;
use serde_derive::{Deserialize, Serialize};
use serde_variant::to_variant_name;
use strum::EnumIter;
use strum::IntoEnumIterator;
use teller_providers::config::KV;

use crate::{Error, Result};

lazy_static! {
    pub static ref POSSIBLE_VALUES: String = {
        let providers: Vec<String> = Format::iter()
            .map(|provider| provider.to_string())
            .collect();
        providers.join(", ")
    };
}

#[derive(Serialize, Deserialize, Debug, Clone, EnumIter)]
pub enum Format {
    #[serde(rename = "csv")]
    CSV,
    #[serde(rename = "yaml")]
    YAML,
    #[serde(rename = "json")]
    JSON,
    #[serde(rename = "env")]
    ENV,
    #[serde(rename = "shell")]
    Shell,
}

impl std::fmt::Display for Format {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        to_variant_name(self).expect("only enum supported").fmt(f)
    }
}

impl FromStr for Format {
    type Err = &'static str;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let providers = Self::iter()
            .map(|provider| (provider.to_string(), provider))
            .collect::<BTreeMap<String, Self>>();

        providers.get(input).map_or_else(
            || Err(&POSSIBLE_VALUES as &'static str),
            |provider| Ok(provider.clone()),
        )
    }
}

impl Format {
    /// Export current format type to string
    ///
    /// # Errors
    ///
    pub fn export(&self, kvs: &[KV]) -> Result<String> {
        match self {
            Self::YAML => Ok(serde_yaml::to_string(&KV::to_data(kvs))?),
            Self::JSON => Ok(serde_json::to_string(&KV::to_data(kvs))?),
            Self::CSV => Self::export_csv(kvs),
            Self::ENV => Ok(Self::export_env(kvs)),
            Self::Shell => Ok(Self::export_shell(kvs)),
        }
    }

    fn export_shell(kvs: &[KV]) -> String {
        let mut out = String::new();
        out.push_str("#!/bin/sh\n");

        for kv in kvs {
            out.push_str(&format!("export {}='{}'\n", kv.key, kv.value));
        }
        out
    }

    fn export_env(kvs: &[KV]) -> String {
        let mut out = String::new();
        for kv in kvs {
            out.push_str(&format!("{}={}\n", kv.key, kv.value));
        }
        out
    }

    fn export_csv(kvs: &[KV]) -> Result<String> {
        let mut wtr = WriterBuilder::new().from_writer(vec![]);
        for kv in kvs {
            wtr.write_record(&[kv.key.clone(), kv.value.clone()])?;
        }
        Ok(String::from_utf8(
            wtr.into_inner()
                .map_err(Box::from)
                .map_err(Error::CSVInner)?,
        )?)
    }
}
