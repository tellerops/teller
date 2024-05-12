use std::fs;

use eyre::Result;
use teller_core::config::{Config, RenderTemplate};
use teller_providers::providers;

use super::Response;
use crate::{cli::NewArgs, wizard};

pub const CMD_NAME: &str = "new";

/// Create a new teller configuration
///
/// # Errors
///
/// This function will return an error if operation fails
#[allow(clippy::future_not_send)]
pub fn run(args: &NewArgs) -> Result<Response> {
    let providers: Vec<providers::ProviderKind> = args.providers.clone();

    let file = {
        let mut file_path = args.filename.clone();
        let ext = file_path
            .extension()
            .and_then(std::ffi::OsStr::to_str)
            .unwrap_or("");

        if ext != "yaml" || ext != "yml" {
            file_path.set_extension("yml");
        }

        file_path
    };

    let w = {
        let mut wizard = wizard::AppConfig::new(args.force);

        if !args.std {
            wizard.with_file_validation(file.as_path());
        }

        if !providers.is_empty() {
            wizard.with_providers(providers);
        }
        wizard
    };
    let results = match w.start() {
        Ok(r) => r,
        Err(e) => match e {
            wizard::Error::ProviderNotFound(_)
            | wizard::Error::Prompt(_)
            | wizard::Error::InvalidSelection => return Err(eyre::Error::new(e)),
            wizard::Error::ConfigurationAlreadyExists => return Response::ok(),
        },
    };

    let template = Config::render_template(&RenderTemplate {
        providers: results.providers,
    })?;

    if args.std {
        Response::ok_with_message(template)
    } else {
        if let Some(folder) = &file.parent() {
            fs::create_dir_all(folder)?;
        }
        fs::write(&file, template)?;
        Response::ok_with_message(format!("Configuration saved in: {:?}", file.display()))
    }
}
