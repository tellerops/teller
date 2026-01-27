use std::collections::BTreeMap;
use std::io::{BufRead, Write};
use std::path::Path;
use std::process::Output;

use teller_providers::config::PathMap;
use teller_providers::Provider;
// use csv::WriterBuilder;
use teller_providers::{config::KV, registry::Registry, Result as ProviderResult};

use crate::redact::Redactor;
use crate::template;
use crate::{
    config::{Config, Match},
    exec, export, scan, Error, Result,
};

pub struct Teller {
    registry: Registry,
    config: Config,
}

impl Teller {
    /// Build from config
    ///
    /// # Errors
    ///
    /// This function will return an error if loading fails
    pub async fn from_config(config: &Config) -> teller_providers::Result<Self> {
        let registry = Registry::new(&config.providers).await?;
        Ok(Self {
            registry,
            config: config.clone(),
        })
    }

    /// Build from YAML
    ///
    /// # Errors
    ///
    /// This function will return an error if loading fails
    pub async fn from_yaml(file: &Path) -> Result<Self> {
        let config = Config::from_path(file)?;
        Self::from_config(&config).await.map_err(Error::Provider)
    }
    /// Collects kvs from all provider maps in the current configuration
    ///
    /// # Errors
    ///
    /// This function will return an error if IO fails
    pub async fn collect(&self) -> ProviderResult<Vec<KV>> {
        let mut res = Vec::new();
        for (name, providercfg) in &self.config.providers {
            if let Some(provider) = self.registry.get(name) {
                for pm in &providercfg.maps {
                    let kvs = provider.get(pm).await?;
                    res.push(kvs);
                }
            }
        }
        Ok(res.into_iter().flatten().collect::<Vec<_>>())
    }
    /// Put a list of KVs into a list of providers, on a specified path
    ///
    /// # Errors
    ///
    /// This function will return an error if put fails
    pub async fn put(&self, kvs: &[KV], map_id: &str, providers: &[String]) -> Result<()> {
        // a target provider has to have the specified path id
        for provider_name in providers {
            let (provider, pm) = self.get_pathmap_on_provider(map_id, provider_name)?;
            provider.put(pm, kvs).await?;
        }
        Ok(())
    }

    /// Delete a list of keys or a complete path for every provider in the list
    ///
    /// # Errors
    ///
    /// This function will return an error if delete fails
    pub async fn delete(&self, keys: &[String], map_id: &str, providers: &[String]) -> Result<()> {
        // a target provider has to have the specified path id
        for provider_name in providers {
            let (provider, pm) = self.get_pathmap_on_provider(map_id, provider_name)?;
            // 1. if keys is empty, use the default pathmap
            // 2. otherwise, create a new pathmap, with a subset of keys
            if keys.is_empty() {
                provider.del(pm).await?;
            } else {
                let mut subset_keys = BTreeMap::new();
                for key in keys {
                    subset_keys.insert(key.clone(), key.clone());
                }
                let mut new_pm = pm.clone();
                new_pm.keys = subset_keys;
                provider.del(&new_pm).await?;
            }
        }
        Ok(())
    }
    /// Get a provider and pathmap from configuration and registry
    ///
    /// # Errors
    ///
    /// This function will return an error if operation fails
    #[allow(clippy::borrowed_box)]
    pub fn get_pathmap_on_provider(
        &self,
        map_id: &str,
        provider_name: &String,
    ) -> Result<(&Box<dyn Provider + Send + Sync>, &PathMap)> {
        let pconf = self.config.providers.get(provider_name).ok_or_else(|| {
            Error::Message(format!(
                "cannot find provider '{provider_name}' path configuration"
            ))
        })?;
        let pm = pconf.maps.iter().find(|m| m.id == map_id).ok_or_else(|| {
            Error::Message(format!(
                "cannot find path id '{map_id}' in provider '{provider_name}'"
            ))
        })?;
        let provider = self.registry.get(provider_name).ok_or_else(|| {
            Error::Message(format!("cannot get initialized provider '{provider_name}'"))
        })?;
        Ok((provider, pm))
    }
    /// Run an external command with provider based environment variables
    ///
    /// # Errors
    ///
    /// This function will return an error if command fails
    pub async fn run<'a>(&self, cmd: &[&str], opts: &exec::Opts<'a>) -> Result<Output> {
        let cmd = shell_words::join(cmd);
        let kvs = self.collect().await?;
        let res = exec::cmd(
            cmd.as_str(),
            &kvs.iter()
                .map(|kv| (kv.key.clone(), kv.value.clone()))
                .collect::<Vec<_>>()[..],
            opts,
        )?;
        Ok(res)
    }

    /// Redact streams
    ///
    /// # Errors
    ///
    /// This function will return an error if Is or collecting keys fails
    #[allow(clippy::future_not_send)]
    pub async fn redact<R: BufRead, W: Write>(&self, reader: R, writer: W) -> Result<()> {
        let kvs = self.collect().await?;
        let redactor = Redactor::new();
        redactor.redact(reader, writer, kvs.as_slice())?;
        Ok(())
    }

    /// Populate a custom template with KVs
    ///
    /// # Errors
    ///
    /// This function will return an error if template rendering fails
    pub async fn template(&self, template: &str) -> Result<String> {
        let kvs = self.collect().await?;
        let out = template::render(template, kvs)?; // consumes kvs
        Ok(out)
    }

    /// Export KV data
    ///
    /// # Errors
    ///
    /// This function will return an error if export fails
    pub async fn export(&self, format: &export::Format) -> Result<String> {
        let kvs = self.collect().await?;
        format.export(&kvs)
    }

    /// Scan a folder recursively for secrets or values
    ///
    /// # Errors
    ///
    /// This function will return an error if IO fails
    pub fn scan(&self, root: &str, kvs: &[KV], opts: &scan::Opts) -> Result<Vec<Match>> {
        scan::scan_root(root, kvs, opts)
    }

    /// Copy from provider to target provider.
    /// Note: `replace` will first delete data at target, then copy.
    ///
    /// # Errors
    ///
    /// This function will return an error if copy fails
    pub async fn copy(
        &self,
        from_provider: &str,
        from_map_id: &str,
        to_provider: &str,
        to_map_id: &str,
        replace: bool,
    ) -> Result<()> {
        // XXX fix &str, &String params
        let (from_provider, from_pm) =
            self.get_pathmap_on_provider(from_map_id, &from_provider.to_string())?;
        let data = from_provider.get(from_pm).await?;

        let (to_provider, to_pm) =
            self.get_pathmap_on_provider(to_map_id, &to_provider.to_string())?;

        if replace {
            to_provider.del(to_pm).await?;
        }
        to_provider.put(to_pm, &data).await?;
        Ok(())
    }
}
