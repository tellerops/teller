use std::collections::HashMap;
use std::path::{Path, PathBuf};

use dialoguer::{theme::ColorfulTheme, Confirm, MultiSelect};
use strum::IntoEnumIterator;
use teller_providers::providers::ProviderKind;

pub type Result<T, E = Error> = std::result::Result<T, E>;

#[derive(thiserror::Error, Debug)]
pub enum Error {
    #[error("Provider: {0} not exists")]
    ProviderNotFound(String),

    #[error(transparent)]
    Prompt(#[from] dialoguer::Error),

    #[error("Config file already exists")]
    ConfigurationAlreadyExists,

    #[error("Invalid prompt selection")]
    InvalidSelection,
}

pub struct AppConfig {
    file_path: Option<PathBuf>,
    providers: Option<Vec<ProviderKind>>,
    pub override_file: bool,
}

pub struct Results {
    pub providers: Vec<ProviderKind>,
}

/// Creating wizard flow for crating Teller configuration file
impl AppConfig {
    #[must_use]
    pub const fn new(override_file: bool) -> Self {
        Self {
            file_path: None,
            providers: None,
            override_file,
        }
    }

    pub fn with_file_validation(&mut self, file_path: &Path) -> &mut Self {
        self.file_path = Some(file_path.to_path_buf());
        self
    }

    pub fn with_providers(&mut self, providers: Vec<ProviderKind>) -> &mut Self {
        self.providers = Some(providers);
        self
    }

    /// Start wizard flow
    ///
    /// # Errors
    /// this function return an errors when from `Error` options
    pub fn start(&self) -> Result<Results> {
        if let Some(file_path) = &self.file_path {
            if file_path.exists()
                && !self.override_file
                && !Self::confirm_override_file(file_path.as_path())?
            {
                return Err(Error::ConfigurationAlreadyExists {});
            }
        }

        let providers = match &self.providers {
            Some(providers) => providers.clone(),
            None => Self::select_providers()?,
        };
        Ok(Results { providers })
    }

    fn confirm_override_file(file_path: &Path) -> Result<bool> {
        Ok(Confirm::with_theme(&ColorfulTheme::default())
            .with_prompt(format!(
                "Teller config {:?} already exists. Do you want to override the configuration \
                 with new settings",
                file_path.display()
            ))
            .interact()?)
    }

    /// Prompt provider selection
    ///
    /// # Errors
    /// When has a problem with prompt selection
    fn select_providers() -> Result<Vec<ProviderKind>> {
        let providers = ProviderKind::iter()
            .map(|provider| (provider.to_string(), provider))
            .collect::<HashMap<String, ProviderKind>>();

        let names = &providers
            .keys()
            .map(std::string::String::as_str)
            .collect::<Vec<_>>();

        let selected_providers = MultiSelect::with_theme(&ColorfulTheme::default())
            .with_prompt("Select your secret providers")
            .items(names)
            .report(false)
            .interact()?;

        let mut selected = vec![];
        for selection in selected_providers {
            let Some(provider_name) = names.get(selection) else {
                return Err(Error::InvalidSelection);
            };

            match providers.get(*provider_name) {
                Some(p) => selected.push(p.clone()),
                _ => return Err(Error::ProviderNotFound((*provider_name).to_string())),
            };
        }

        Ok(selected)
    }
}
