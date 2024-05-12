pub mod config;
pub mod exec;
pub mod export;
mod io;
pub mod redact;
pub mod scan;
pub mod teller;
pub mod template;

use std::string::FromUtf8Error;

#[derive(thiserror::Error, Debug)]
pub enum Error {
    #[error("{0}")]
    Message(String),

    #[error(transparent)]
    Shellwords(#[from] shell_words::ParseError),

    #[error(transparent)]
    IO(#[from] std::io::Error),

    #[error(transparent)]
    Provider(#[from] teller_providers::Error),

    #[error(transparent)]
    Handlebars(Box<dyn std::error::Error + Send + Sync>),

    #[error(transparent)]
    Json(#[from] serde_json::Error),

    #[error(transparent)]
    YAML(#[from] serde_yaml::Error),

    #[error(transparent)]
    CSV(#[from] csv::Error),

    #[error(transparent)]
    CSVInner(Box<dyn std::error::Error + Send + Sync>),

    #[error(transparent)]
    Tera(#[from] tera::Error),

    #[error(transparent)]
    Utf(#[from] FromUtf8Error),
}
pub type Result<T, E = Error> = std::result::Result<T, E>;
