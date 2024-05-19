//! etcd
//!
//!
//! ## Example configuration
//!
//! ```yaml
//! providers:
//!  etcd1:
//!    kind: etcd
//!    # options: ...
//! ```
//! ## Options
//!
//! See [`EtcdOptions`] for more.
//!

use async_trait::async_trait;
use etcd_client::{Client, ConnectOptions, DeleteOptions, GetOptions};
use serde_derive::{Deserialize, Serialize};
use tokio::sync::Mutex;

use super::ProviderKind;
use crate::{
    config::{PathMap, ProviderInfo, KV},
    Error, Provider, Result,
};

/// Etcd Options
///
#[allow(clippy::module_name_repetitions)]
#[derive(Default, Serialize, Deserialize, Debug, Clone)]
pub struct EtcdOptions {
    /// Etcd address.
    pub address: Option<String>,

    /// An optional 'user' field, which includes a user name and password separated by `:`.
    ///
    /// Example: `joe:rootpass`
    ///
    /// see more: [on this github thread](https://github.com/etcd-io/etcd/issues/10338#issuecomment-448280528)
    pub user: Option<String>,
}

pub struct Etcd {
    pub client: Mutex<Client>,
    pub name: String,
}

fn to_err(_pm: &PathMap, err: etcd_client::Error) -> Error {
    Error::Any(Box::new(err))
}

impl Etcd {
    /// Create a new etcd provider
    ///
    /// # Errors
    ///
    /// This function will return an error if cannot create a provider
    pub async fn new(name: &str, opts: Option<EtcdOptions>) -> Result<Self> {
        let opts = opts.unwrap_or_default();

        let address = opts
            .address
            .as_ref()
            .ok_or_else(|| Error::Message("address not present.".to_string()))?;

        let user = opts
            .user
            .or_else(|| std::env::var("ETCDCTL_USER").ok())
            .map(|user| {
                user.split_once(':')
                    .ok_or_else(|| {
                        Error::Message("user field should be in `user:pwd` format".to_string())
                    })
                    .map(|(user, pwd)| (user.to_string(), pwd.to_string()))
            })
            .transpose()?;

        let mut connect_opts = ConnectOptions::new();
        if let Some(user) = user {
            connect_opts = connect_opts.with_user(user.0, user.1);
        }

        Ok(Self {
            client: Mutex::new(
                Client::connect([address], Some(connect_opts))
                    .await
                    .map_err(|err| Error::CreateProviderError(err.to_string()))?,
            ),
            name: name.to_string(),
        })
    }
}

#[async_trait]
impl Provider for Etcd {
    fn kind(&self) -> ProviderInfo {
        ProviderInfo {
            kind: ProviderKind::Etcd,
            name: self.name.clone(),
        }
    }

    async fn get(&self, pm: &PathMap) -> Result<Vec<KV>> {
        let mut client = self.client.lock().await.kv_client();

        let res = if pm.keys.is_empty() {
            client
                .get(pm.path.as_str(), Some(GetOptions::new().with_prefix()))
                .await
                .map_err(|err| to_err(pm, err))?
                .kvs()
                .to_vec()
        } else {
            let mut res = Vec::new();
            for key in pm.keys.keys() {
                let fetched = client
                    .get(format!("{}/{}", pm.path.as_str(), key), None)
                    .await
                    .map_err(|err| to_err(pm, err))?
                    .kvs()
                    .to_vec();
                res.extend(fetched);
            }
            res
        };

        drop(client);

        if res.is_empty() {
            return Err(Error::NotFound {
                msg: "not found".to_string(),
                path: pm.path.clone(),
            });
        }

        let mut results = vec![];
        for kv_pair in res {
            let key = kv_pair.key_str().map_err(|err| to_err(pm, err))?;

            // strip path pref
            let key = key
                .strip_prefix(&pm.path)
                .map_or(key, |s| s.trim_start_matches('/'));

            let val = kv_pair.value_str().map_err(|err| to_err(pm, err))?;

            results.push(KV::from_value(val, key, key, pm, self.kind()));
        }

        Ok(results)
    }

    async fn put(&self, pm: &PathMap, kvs: &[KV]) -> Result<()> {
        let mut client = self.client.lock().await.kv_client();
        for kv in kvs {
            client
                .put(
                    format!("{}/{}", pm.path, kv.key).as_str(),
                    kv.value.as_bytes().to_vec(),
                    None,
                )
                .await
                .map_err(|e| to_err(pm, e))?;
        }
        drop(client);

        Ok(())
    }

    async fn del(&self, pm: &PathMap) -> Result<()> {
        let mut client = self.client.lock().await.kv_client();
        if pm.keys.is_empty() {
            client
                .delete(
                    pm.path.as_str(),
                    Some(DeleteOptions::default().with_prefix()),
                )
                .await
                .map_err(|err| to_err(pm, err))?;
        } else {
            for key in pm.keys.keys().map(|kv| format!("{}/{kv}", &pm.path)) {
                client
                    .delete(key, None)
                    .await
                    .map_err(|err| to_err(pm, err))?;
            }
        };
        drop(client);

        Ok(())
    }
}

#[cfg(test)]
mod tests {

    use super::*;
    use crate::providers::test_utils;

    const PORT: u32 = 2379;

    #[test_log::test]
    #[cfg(not(windows))]
    fn sanity_test() {
        use std::{collections::HashMap, env};

        use dockertest::{waitfor, Composition, DockerTest, Image};

        if env::var("RUNNER_OS").unwrap_or_default() == "macOS" {
            return;
        }
        let mut test = DockerTest::new();
        let wait = Box::new(waitfor::MessageWait {
            message: "serving client traffic insecurely".to_string(),
            source: waitfor::MessageSource::Stderr,
            timeout: 20,
        });

        let mut env = HashMap::new();

        env.insert("ALLOW_NONE_AUTHENTICATION".to_string(), "yes".to_string());

        #[cfg(target_arch = "aarch64")]
        env.insert("ETCD_UNSUPPORTED_ARCH".to_string(), "arm64".to_string());

        #[cfg(target_arch = "aarch64")]
        let image_name = "bitnami/etcd";
        #[cfg(not(target_arch = "aarch64"))]
        let image_name = "bitnami/etcd";

        let image = Image::with_repository(image_name)
            .pull_policy(dockertest::PullPolicy::IfNotPresent)
            .source(dockertest::Source::DockerHub);
        let mut etcd_container = Composition::with_image(image)
            .with_container_name("etcd-server")
            .with_env(env)
            .with_wait_for(wait);
        etcd_container.port_map(PORT, PORT);

        test.add_composition(etcd_container);

        test.run(|ops| async move {
            let _instance = ops.handle("etcd-server");
            let address = format!("localhost:{PORT}");
            // banner is not enough, we have to wait for the image to stabilize

            let p = Box::new(
                super::Etcd::new(
                    "etcd",
                    Some(EtcdOptions {
                        address: Some(address),
                        user: None,
                    }),
                )
                .await
                .unwrap(),
            ) as Box<dyn Provider + Send + Sync>;

            test_utils::ProviderTest::new(p).run().await;
        });
    }
}
