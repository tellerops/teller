pub mod config;
pub mod providers;
pub mod registry;

use async_trait::async_trait;

use crate::config::{PathMap, ProviderInfo, KV};

#[async_trait]
pub trait Provider {
    fn kind(&self) -> ProviderInfo;
    /// Get a mapping
    ///
    /// # Errors
    ///
    /// ...
    async fn get(&self, pm: &PathMap) -> Result<Vec<KV>>;
    /// Put a mapping
    ///
    /// # Errors
    ///
    /// ...
    async fn put(&self, pm: &PathMap, kvs: &[KV]) -> Result<()>;
    /// Delete a mapping
    ///
    /// # Errors
    ///
    /// ...
    async fn del(&self, pm: &PathMap) -> Result<()>;
}
#[derive(thiserror::Error, Debug)]
pub enum Error {
    #[error("{0}")]
    Message(String),

    #[error("{0}: {1}")]
    PathError(String, String),

    #[error(transparent)]
    IO(#[from] std::io::Error),

    #[error(transparent)]
    Env(#[from] std::env::VarError),

    #[error(transparent)]
    Any(#[from] Box<dyn std::error::Error + Send + Sync>),

    #[error(transparent)]
    Json(#[from] serde_json::Error),

    #[error(transparent)]
    YAML(#[from] serde_yaml::Error),

    #[error("NOT FOUND {path}: {msg}")]
    NotFound { path: String, msg: String },

    #[error("GET {path}: {msg}")]
    GetError { path: String, msg: String },

    #[error("DEL {path}: {msg}")]
    DeleteError { path: String, msg: String },

    #[error("PUT {path}: {msg}")]
    PutError { path: String, msg: String },

    #[error("LIST {path}: {msg}")]
    ListError { path: String, msg: String },

    #[error("{0}")]
    CreateProviderError(String),
}

pub type Result<T, E = Error> = std::result::Result<T, E>;
