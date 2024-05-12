//! In-memory Store
//!
//!
//! ## Example configuration
//!
//! ```yaml
//! providers:
//!  inmem1:
//!    kind: inmem
//!    options:
//!      key1: value
//!      key1: value
//! ```
//! ## Options
//!
//! The options to the inmem store are actually its initial data
//! representation and can be any `serde_json::Value` that can convert to
//! a `BTreeMap` (hashmap)
//!
use std::{collections::BTreeMap, sync::Mutex};

use async_trait::async_trait;

use super::ProviderKind;
use crate::{
    config::{PathMap, ProviderInfo, KV},
    Error, Provider, Result,
};

pub struct Inmem {
    store: Mutex<BTreeMap<String, BTreeMap<String, String>>>,
    name: String,
}

impl Inmem {
    /// Build from YAML
    ///
    /// # Errors
    ///
    /// This function will return an error if serialization fails
    pub fn from_yaml(name: &str, yaml: &str) -> Result<Self> {
        Ok(Self {
            store: Mutex::new(serde_yaml::from_str(yaml)?),
            name: name.to_string(),
        })
    }

    /// Create an inmem provider.
    /// `opts` is the memory map of the provider, which is a `BTreeMap<String, BTreeMap<String, String>>`,
    /// or in YAML:
    ///
    /// ```yaml
    /// production/foo/bar:
    ///     key: v
    ///     baz: bar
    /// ```
    ///
    /// # Errors
    ///
    /// This function will return an error if creation fails
    pub fn new(name: &str, opts: Option<serde_json::Value>) -> Result<Self> {
        Ok(if let Some(opts) = opts {
            Self {
                store: Mutex::new(serde_json::from_value(opts)?),
                name: name.to_string(),
            }
        } else {
            Self {
                store: Mutex::new(BTreeMap::default()),
                name: name.to_string(),
            }
        })
    }

    /// Returns the get state of this [`Inmem`].
    ///
    /// # Panics
    ///
    /// Panics if lock cannot be acquired
    pub fn get_state(&self) -> BTreeMap<String, BTreeMap<String, String>> {
        self.store
            .lock()
            .expect("inmem store failed getting a lock")
            .clone()
    }
}

#[async_trait]
impl Provider for Inmem {
    fn kind(&self) -> ProviderInfo {
        ProviderInfo {
            kind: ProviderKind::Inmem,
            name: self.name.clone(),
        }
    }

    #[allow(clippy::significant_drop_tightening)]
    async fn get(&self, pm: &PathMap) -> Result<Vec<KV>> {
        let store = self.store.lock().unwrap();
        let data = store.get(&pm.path).ok_or_else(|| Error::NotFound {
            path: pm.path.to_string(),
            msg: "not found".to_string(),
        })?;
        Ok(KV::from_data(data, pm, &self.kind()))
    }
    #[allow(clippy::significant_drop_tightening)]
    async fn put(&self, pm: &PathMap, kvs: &[KV]) -> Result<()> {
        let mut store = self.store.lock().unwrap();
        let mut data = store.get(&pm.path).cloned().unwrap_or_default();
        for kv in kvs {
            data.insert(kv.key.clone(), kv.value.clone());
        }
        store.insert(pm.path.clone(), data);
        Ok(())
    }

    async fn del(&self, pm: &PathMap) -> Result<()> {
        if pm.keys.is_empty() {
            self.store.lock().unwrap().remove(&pm.path);
        } else {
            let mut store = self.store.lock().unwrap();
            let mut data = store.get(&pm.path).cloned().unwrap_or_default();
            for key in pm.keys.keys() {
                data.remove(key);
            }
            store.insert(pm.path.clone(), data);
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use tokio::test;

    use crate::providers::test_utils;
    use crate::Provider;

    #[test]
    async fn sanity_test() {
        let p =
            Box::new(super::Inmem::new("test", None).unwrap()) as Box<dyn Provider + Send + Sync>;

        test_utils::ProviderTest::new(p).run().await;
    }
}
