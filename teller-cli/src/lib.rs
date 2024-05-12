pub mod cli;
pub mod io;
pub mod new;
pub mod scan;
pub mod wizard;
use eyre::Result;
use tracing::level_filters::LevelFilter;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter, Registry};
#[allow(clippy::module_name_repetitions)]
pub struct Response {
    pub code: exitcode::ExitCode,
    pub message: Option<String>,
}
impl Response {
    #[allow(clippy::missing_const_for_fn)]
    #[allow(clippy::unnecessary_wraps)]
    fn fail() -> Result<Self> {
        Ok(Self {
            code: 1,
            message: None,
        })
    }
    #[allow(clippy::missing_const_for_fn)]
    #[allow(clippy::unnecessary_wraps)]
    fn ok() -> Result<Self> {
        Ok(Self {
            code: exitcode::OK,
            message: None,
        })
    }

    #[allow(clippy::missing_const_for_fn)]
    #[allow(clippy::unnecessary_wraps)]
    fn ok_with_message(message: String) -> Result<Self> {
        Ok(Self {
            code: exitcode::OK,
            message: Some(message),
        })
    }
}

pub fn tracing(verbose: bool) {
    let level = if verbose {
        LevelFilter::INFO
    } else {
        LevelFilter::OFF
    };
    Registry::default()
        .with(tracing_tree::HierarchicalLayer::new(2))
        .with(
            EnvFilter::builder()
                .with_default_directive(level.into())
                .with_env_var("LOG")
                .from_env_lossy(),
        )
        .init();
}
