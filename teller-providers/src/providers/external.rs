
//! external
//!
//!
//! ## Example configuration
//!
//! ```yaml
//! providers:
//!  my-external-provider:
//!    kind: external
//!    # options: ...
//! ```
//! ## Options
//!
//! See [`ExternalOptions`] for more.
//!

use async_trait::async_trait;
use serde_derive::{Deserialize, Serialize};
use which::which;
use std::str;

use super::ProviderKind;
use crate::{
    config::{PathMap, ProviderInfo, KV},
    Error, Provider, Result,
};


#[derive(Default, Serialize, Deserialize, Debug, Clone)]
pub struct ExternalOptions {
    /// bin extension
    pub extension: Option<String>,
    pub extra_arguments: Option<Vec<String>>,
}

#[derive(Clone)]
pub struct External {
    pub name: String,
    bin_path: String,
    opts: ExternalOptions,
}

impl External {
    /// Create a new external provider
    ///
    /// # Errors
    ///
    /// This function will return an error if cannot create a provider
    pub fn new(name: &str, opts: Option<ExternalOptions>) -> Result<Self> {
        let opts = opts.unwrap_or_default();

        let extension = opts
            .extension
            .as_ref()
            .ok_or_else(|| Error::Message("option 'extension' is required".to_string()))?;

        let bin_path = match which(format!("teller-provider-{}", extension)) {
            Ok(bin) => bin.to_str().unwrap().to_string(),
            Err(_) => return Err(Error::Message(format!("external provider 'teller-provider-{}' not on path", extension).to_string()))
        };

        Ok(Self {
            name: name.to_string(),
            bin_path: bin_path,
            opts,
        })
    }


}


#[async_trait]
impl Provider for External {
    fn kind(&self) -> ProviderInfo {
        ProviderInfo {
            kind: ProviderKind::External,
            name: self.name.clone(),
        }
    }

    async fn get(&self, pm: &PathMap) -> Result<Vec<KV>> {
        let mut res: Vec<KV> = Vec::new();
        for (from_key, to_key) in &pm.keys {
            //let full_from_key = self.full_key(&pm.path, from_key);
            let output = 
                self.prepare_command("get", &[&pm.path, from_key])?
                .output()?;
            let found_val = str::from_utf8(&output.stdout).unwrap();
            res.push(KV::from_value(found_val, from_key, to_key, pm, self.kind()));
        }

        if res.is_empty() {
            return Err(Error::NotFound {
                msg: "not found".to_string(),
                path: pm.path.clone(),
            });
        }

        Ok(res)
    }

    async fn put(&self, pm: &PathMap, kvs: &[KV]) -> Result<()> {
        for kv in kvs {
            //let full_from_key = self.full_key(&pm.path, &kv.key);
            let output = 
                self.prepare_command("put", &[&pm.path, &kv.key])?
                .output()?;

            if !output.status.success() {
                return Err(Error::PutError {
                    msg: format!("failed to put - {}", str::from_utf8(&output.stderr).unwrap()),
                    path: pm.path.clone(),
                });
            }
        }
        Ok(())
    }

    async fn del(&self, pm: &PathMap) -> Result<()> {
        let output = 
            self.prepare_command("del", &[&pm.path])?
            .output()?;

        if !output.status.success() {
            return Err(Error::PutError {
                msg: format!("failed to del - {}", str::from_utf8(&output.stderr).unwrap()),
                path: pm.path.clone(),
            });
        }
        Ok(())
    }
}

impl External {

    fn prepare_command(&self, action: &str, args: &[&str]) -> Result<std::process::Command> {
        let mut cmd = std::process::Command::new(self.bin_path.clone());
        cmd.arg(action);
        cmd.args(args);

        if let Some(extra_arguments) = &self.opts.extra_arguments {
            cmd.args(extra_arguments);
        }

        Ok(cmd)
    }

    //fn full_key(&self, path: &String, key: &String) -> String {
    //    return match Some(path.clone()) {
    //        Some(path) => format!("{}{}", path, key),
    //        None => key.clone(),
    //    };
    //}

}

#[cfg(test)]
mod tests {
    use tokio::test;

    use super::*;
    use crate::providers::test_utils;


    #[test]
    async fn sanity_test() {
        //use std::{collections::HashMap, env};

        //let mut env = HashMap::new();

        let opts = serde_json::json!({
            "extension": "some-bin",
        });

        let p: Box<dyn Provider + Send + Sync> = Box::new(
            super::External::new("external", Some(serde_json::from_value(opts).unwrap())).unwrap()
        ) as Box<dyn Provider + Send + Sync>;

        // fails, would need to mock? or compile a 'test' binary?
        test_utils::ProviderTest::new(p)
            .with_root_prefix("tmp/external/")
            .run()
            .await;

    }
}
