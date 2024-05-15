//! `dotenv` Provider
//!
//!
//! ## Example configuration
//!
//! ```yaml
//! providers:
//!  dotenv1:
//!    kind: dotenv
//!    # options: ...
//! ```
//! ## Options
//!
//! See [`DotEnvOptions`]
//!
//!
#![allow(clippy::borrowed_box)]
use std::fs::File;
use std::io::prelude::*;
use std::{
    collections::{BTreeMap, HashMap},
    io,
    path::Path,
};

use async_trait::async_trait;
use dotenvy::{self};
use fs_err as fs;
use serde_derive::{Deserialize, Serialize};

use super::ProviderKind;
use crate::config::ProviderInfo;
use crate::{
    config::{PathMap, KV},
    Error, Provider, Result,
};

#[derive(PartialEq)]
enum Mode {
    Get,
    Put,
    Del,
}

#[derive(Serialize, Deserialize, Debug, Clone, Default)]
pub struct DotEnvOptions {
    /// create a file if did not exist, when writing new data to provider
    pub create_on_put: bool,
}

pub struct Dotenv {
    pub name: String,
    opts: DotEnvOptions,
}
impl Dotenv {
    /// Create a new provider
    ///
    /// # Errors
    ///
    /// This function will return an error if cannot create a provider
    pub fn new(name: &str, opts: Option<DotEnvOptions>) -> Result<Self> {
        let opts = opts.unwrap_or_default();

        Ok(Self {
            name: name.to_string(),
            opts,
        })
    }
}

fn load(path: &Path, mode: &Mode) -> Result<BTreeMap<String, String>> {
    let content = fs::File::open(path)?;
    let mut env = BTreeMap::new();

    if mode == &Mode::Get {
        let metadata = content.metadata().map_err(|e| Error::GetError {
            path: format!("{path:?}"),
            msg: format!("could not get file metadata. err: {e:?}"),
        })?;

        if metadata.len() == 0 {
            return Err(Error::NotFound {
                path: format!("{path:?}"),
                msg: "file is empty".to_string(),
            });
        }
    }

    for res in dotenvy::Iter::new(&content) {
        let (k, v) = res.map_err(|e| Error::GetError {
            path: format!("{path:?}"),
            msg: e.to_string(),
        })?;
        env.insert(k, v);
    }

    Ok(env)
}
// poor man's serialization, loses original comments and formatting
fn save(path: &Path, data: &BTreeMap<String, String>) -> Result<String> {
    let mut out = String::new();
    for (k, v) in data {
        let maybe_json: serde_json::Result<HashMap<String, serde_json::Value>> =
            serde_json::from_str(v);

        let json_value = if maybe_json.is_ok() {
            serde_json::to_string(&v).map(Some).unwrap_or_default()
        } else {
            None
        };

        let value = json_value.unwrap_or_else(|| v.to_string());
        if value.chars().any(char::is_whitespace) {
            out.push_str(&format!("{k}=\"{value}\"\n"));
        } else {
            out.push_str(&format!("{k}={value}\n"));
        }
    }

    fs::write(path, &out)?;
    Ok(out)
}

#[async_trait]
impl Provider for Dotenv {
    fn kind(&self) -> ProviderInfo {
        ProviderInfo {
            kind: ProviderKind::Dotenv,
            name: self.name.clone(),
        }
    }

    async fn get(&self, pm: &PathMap) -> Result<Vec<KV>> {
        let data = load(Path::new(&pm.path), &Mode::Get)?;
        Ok(KV::from_data(&data, pm, &self.kind()))
    }

    async fn put(&self, pm: &PathMap, kvs: &[KV]) -> Result<()> {
        // Create file if not exists + add the option to set is as false
        self.load_modify_save(
            pm,
            |data| {
                for kv in kvs {
                    data.insert(kv.key.to_string(), kv.value.to_string());
                }
            },
            &Mode::Put,
        )?;
        Ok(())
    }

    async fn del(&self, pm: &PathMap) -> Result<()> {
        self.load_modify_save(
            pm,
            |data| {
                if pm.keys.is_empty() {
                    data.clear();
                } else {
                    for k in pm.keys.keys() {
                        if data.contains_key(k) {
                            data.remove(k);
                        }
                    }
                }
            },
            &Mode::Del,
        )?;
        Ok(())
    }
}
impl Dotenv {
    fn load_modify_save<F>(&self, pm: &PathMap, modify: F, mode: &Mode) -> Result<()>
    where
        F: Fn(&mut BTreeMap<String, String>),
    {
        if mode == &Mode::Put && self.opts.create_on_put {
            Self::create_empty_file(&pm.path).map_err(|e| Error::GetError {
                path: format!("{:?}", pm.path),
                msg: format!("could not create file: {:?}. err: {e:?}", pm.path),
            })?;
        }
        let file = Path::new(&pm.path);
        let mut data = load(file, mode)?;
        modify(&mut data);
        save(file, &data)?;
        Ok(())
    }

    fn create_empty_file(path: &str) -> io::Result<()> {
        if let Some(parent_dir) = Path::new(path).parent() {
            std::fs::create_dir_all(parent_dir)?;
        }
        let mut file = File::create(path)?;
        file.write_all(b"")?;

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use tokio::test;

    use super::*;
    use crate::providers::test_utils;

    #[test]
    async fn sanity_test() {
        let opts = serde_json::json!({
            "create_on_put": true,
        });

        let p: Box<dyn Provider + Send + Sync> = Box::new(
            super::Dotenv::new("dotenv", Some(serde_json::from_value(opts).unwrap())).unwrap(),
        ) as Box<dyn Provider + Send + Sync>;

        test_utils::ProviderTest::new(p)
            .with_root_prefix("tmp/dotenv/")
            .run()
            .await;
    }
}
