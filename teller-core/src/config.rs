use std::cmp::Ordering;
use std::path::PathBuf;
use std::{
    collections::{BTreeMap, HashMap},
    path::Path,
};

use fs_err as fs;
use serde_derive::{Deserialize, Serialize};
use teller_providers::config::{PathMap, ProviderCfg, KV};
use teller_providers::providers::ProviderKind;
use tera::{Context, Tera};

use crate::Result;

#[derive(Serialize, Deserialize, Debug, Clone, Default)]
pub struct Config {
    pub providers: BTreeMap<String, ProviderCfg>,
}

#[derive(Serialize)]
pub struct RenderTemplate {
    pub providers: Vec<ProviderKind>,
}

fn apply_eqeq(config: &mut Config) {
    config.providers.iter_mut().for_each(|(_name, provider)| {
        provider.maps.iter_mut().for_each(|pm| {
            pm.keys.iter_mut().for_each(|(k, v)| {
                // THINK: replace with:
                // 1. templating: {{id}} (identity), {{snake_case}} (snake case it)
                // 2. other symbols: == id, ^^ capitalize, snake case __ lower snake case
                if v == "==" {
                    v.clone_from(k);
                }
            });
        });
    });
}

impl Config {
    /// Config from text
    ///
    /// # Errors
    ///
    /// This function will return an error if serialization fails
    pub fn with_vars(text: &str, vars: &HashMap<String, String>) -> Result<Self> {
        let rendered_text = Tera::one_off(text, &Context::from_serialize(vars)?, false)?;
        let mut config: Self = serde_yaml::from_str(&rendered_text)?;

        apply_eqeq(&mut config);

        Ok(config)
    }

    /// Config from text
    ///
    /// # Errors
    ///
    /// This function will return an error if serialization fails
    pub fn from_text(text: &str) -> Result<Self> {
        Self::with_vars(text, &HashMap::new())
    }

    /// Config from file
    ///
    /// # Errors
    ///
    /// This function will return an error if IO fails
    pub fn from_path(path: &Path) -> Result<Self> {
        Self::from_text(&fs::read_to_string(path)?)
    }

    /// Create configuration template file
    ///
    /// # Errors
    /// When could not convert config to string
    pub fn render_template(data: &RenderTemplate) -> Result<String> {
        let res: BTreeMap<String, ProviderCfg> = data
            .providers
            .iter()
            .map(|p| {
                (
                    format!("{p}_1"),
                    ProviderCfg {
                        kind: p.clone(),
                        maps: vec![PathMap::from_path("example/dev")],
                        ..ProviderCfg::default()
                    },
                )
            })
            .collect();

        let config = Self { providers: res };

        let a: String = serde_yaml::to_string(&config)?;
        Ok(a)
    }
}

#[derive(Debug, Clone, Serialize, Eq, PartialEq)]
pub struct Match {
    pub path: PathBuf,
    pub position: Option<(usize, usize)>,
    pub offset: usize,
    pub query: KV,
}

impl PartialOrd for Match {
    fn partial_cmp(&self, other: &Self) -> Option<Ordering> {
        Some(self.cmp(other))
    }
}

impl Ord for Match {
    fn cmp(&self, other: &Self) -> Ordering {
        let query_cmp = self.query.cmp(&other.query);

        if query_cmp != Ordering::Equal {
            return query_cmp;
        }

        self.offset.cmp(&other.offset)
    }
}

#[cfg(test)]
mod tests {
    use insta::assert_yaml_snapshot;

    use super::*;
    #[test]
    fn load_config() {
        std::env::set_var("TEST_LOAD_1", "DEV");
        let config = Config::from_path(Path::new("fixtures/config.yml")).unwrap();
        assert_eq!(config.providers.len(), 2);
        assert_yaml_snapshot!(config);
    }

    #[test]
    fn can_render_template_config() {
        let data = RenderTemplate {
            providers: vec![ProviderKind::Inmem, ProviderKind::Dotenv],
        };

        let config = Config::render_template(&data).unwrap();
        assert_yaml_snapshot!(config);
    }
}
