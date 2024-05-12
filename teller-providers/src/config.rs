use std::cmp::Ordering;
use std::collections::BTreeMap;

use serde_derive::{Deserialize, Serialize};

use crate::providers::ProviderKind;

fn is_default<T: Default + PartialEq>(t: &T) -> bool {
    t == &T::default()
}

#[derive(Serialize, Deserialize, Debug, Clone, Default)]
pub struct ProviderCfg {
    #[serde(rename = "kind")]
    pub kind: ProviderKind,
    #[serde(rename = "options", skip_serializing_if = "Option::is_none")]
    pub options: Option<serde_json::Value>,
    #[serde(rename = "name", skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    pub maps: Vec<PathMap>,
}

#[derive(Serialize, Deserialize, Debug, Clone, Default, Eq, PartialEq)]
pub enum Sensitivity {
    #[default]
    None,
    Low,
    Medium,
    High,
    Critical,
}

#[derive(Serialize, Deserialize, Debug, Clone, Default, Eq, PartialEq)]
pub struct ProviderInfo {
    pub kind: ProviderKind,
    pub name: String,
}
#[derive(Serialize, Deserialize, Debug, Clone, Default, Eq, PartialEq)]
pub struct PathInfo {
    pub id: String,
    pub path: String,
}

#[derive(Serialize, Deserialize, Debug, Clone, Default, Eq, PartialEq)]
pub struct MetaInfo {
    pub sensitivity: Sensitivity,
    pub redact_with: Option<String>,
    pub source: Option<String>,
    pub sink: Option<String>,
}
#[derive(Serialize, Deserialize, Debug, Clone, Default, Eq, PartialEq)]
pub struct KV {
    pub value: String,
    pub key: String, // mapped-to key
    pub from_key: String,
    pub path: Option<PathInfo>, // always toplevel
    pub provider: Option<ProviderInfo>,
    pub meta: Option<MetaInfo>,
}

impl PartialOrd for KV {
    fn partial_cmp(&self, other: &Self) -> Option<Ordering> {
        Some(self.cmp(other))
    }
}

impl Ord for KV {
    fn cmp(&self, other: &Self) -> Ordering {
        if let (Some(provider), Some(other_provider)) = (&self.provider, &other.provider) {
            let provider_cmp = provider.kind.cmp(&other_provider.kind);
            if provider_cmp != Ordering::Equal {
                return provider_cmp;
            }
        }

        self.key.cmp(&other.key)
    }
}

impl KV {
    #[must_use]
    pub fn to_data(kvs: &[Self]) -> BTreeMap<String, String> {
        let mut data = BTreeMap::new();
        for kv in kvs {
            data.insert(kv.key.clone(), kv.value.clone());
        }
        data
    }

    #[must_use]
    pub fn from_data(
        data: &BTreeMap<String, String>,
        pm: &PathMap,
        provider: &ProviderInfo,
    ) -> Vec<Self> {
        // map all of the data found
        if pm.keys.is_empty() {
            data.iter()
                .map(|(k, v)| Self::from_value(v, k, k, pm, provider.clone()))
                .collect::<Vec<_>>()
        } else {
            // selectively map only keys from pathmap
            pm.keys
                .iter()
                .filter_map(|(from_key, to_key)| {
                    data.get(from_key).map(|found_val| {
                        Self::from_value(found_val, from_key, to_key, pm, provider.clone())
                    })
                })
                .collect::<Vec<_>>()
        }
    }
    #[must_use]
    pub fn from_value(
        found_val: &str,
        from_key: &str,
        to_key: &str,
        pm: &PathMap,
        provider: ProviderInfo,
    ) -> Self {
        Self {
            value: found_val.to_string(),
            key: to_key.to_string(),
            from_key: from_key.to_string(),
            path: Some(PathInfo {
                path: pm.path.clone(),
                id: pm.id.to_string(),
            }),
            provider: Some(provider),
            meta: Some(MetaInfo {
                sensitivity: pm.sensitivity.clone(),
                redact_with: pm.redact_with.clone(),
                source: pm.source.clone(),
                sink: pm.sink.clone(),
            }),
        }
    }
    #[must_use]
    pub fn from_literal(path: &str, key: &str, value: &str, provider: ProviderInfo) -> Self {
        Self {
            value: value.to_string(),
            key: key.to_string(),
            from_key: key.to_string(),
            path: Some(PathInfo {
                id: path.to_string(),
                path: path.to_string(),
            }),
            provider: Some(provider),
            ..Default::default()
        }
    }

    /// represents a KV without any source (e.g. created manually by a user, pending insert to
    /// one of the providers)
    #[must_use]
    pub fn from_kv(key: &str, value: &str) -> Self {
        Self {
            value: value.to_string(),
            key: key.to_string(),
            from_key: key.to_string(),
            ..Default::default()
        }
    }
}

#[derive(Serialize, Deserialize, Debug, Clone, Default)]
pub struct PathMap {
    pub id: String,
    #[serde(rename = "protocol", skip_serializing_if = "Option::is_none")]
    pub protocol: Option<String>,
    #[serde(rename = "path")]
    pub path: String,
    #[serde(default, rename = "keys", skip_serializing_if = "is_default")]
    pub keys: BTreeMap<String, String>,
    #[serde(default, rename = "decrypt", skip_serializing_if = "is_default")]
    pub decrypt: bool,
    #[serde(default, rename = "sensitivity", skip_serializing_if = "is_default")]
    pub sensitivity: Sensitivity,
    #[serde(
        default,
        rename = "redact_with",
        skip_serializing_if = "Option::is_none"
    )]
    pub redact_with: Option<String>,
    #[serde(default, rename = "source", skip_serializing_if = "Option::is_none")]
    pub source: Option<String>,
    #[serde(default, rename = "sink", skip_serializing_if = "Option::is_none")]
    pub sink: Option<String>,
    // ignore population if optional + we got error
    #[serde(default, rename = "optional", skip_serializing_if = "is_default")]
    pub optional: bool,
}

impl PathMap {
    #[must_use]
    pub fn from_path(path: &str) -> Self {
        Self {
            path: path.to_string(),
            ..Default::default()
        }
    }
}
