use std::{
    env,
    path::{Path, PathBuf},
};

use clap::{Args, Parser, Subcommand, ValueEnum};
use eyre::{eyre, OptionExt};
use teller_core::{exec, export, teller::Teller};
use teller_providers::{config::KV, providers::ProviderKind};

use crate::{
    io::{self, or_stdin, or_stdout},
    new, scan, Response,
};

#[derive(Debug, Clone, Parser)] // requires `derive` feature
#[command(name = "teller")]
#[command(about = "A multi provider secret management tool", version, long_about = None)]
pub struct Cli {
    /// Path to your teller.yml config
    #[arg(short, long)]
    pub config: Option<String>,

    /// Path to your teller.yml config
    #[arg(long)]
    pub verbose: bool,

    /// A teller command
    #[command(subcommand)]
    pub command: Commands,
}
#[derive(Debug, Clone, Subcommand)]
pub enum Commands {
    /// Run a command
    Run {
        /// Reset environment variables before running
        #[arg(short, long)]
        reset: bool,
        /// Run command as shell command
        #[arg(short, long)]
        shell: bool,
        /// The command to run
        #[arg(value_name = "COMMAND", raw = true)]
        command: Vec<String>,
    },

    /// Scan files
    Scan(ScanArgs),
    /// Export key-secret pairs to a specified format
    Export {
        /// The format to export to
        #[arg(value_enum, index = 1)]
        format: Format,
    },
    /// Redact text using fetched secrets
    Redact {
        /// Input file (stdin if none given)
        #[arg(name = "in", short, long)]
        in_file: Option<String>,
        /// Output file (stdout if none given)
        #[arg(short, long)]
        out: Option<String>,
    },

    /// Render a key-value aware template
    Template {
        /// Input template (stdin if none given)
        #[arg(name = "in", short, long)]
        in_file: Option<String>,
        /// Output destination (stdout if none given)
        #[arg(short, long)]
        out: Option<String>,
    },

    /// Export compatible with ENV
    Env {},

    /// Print all currently accessible data
    Show {},

    /// Export as source-able shell script
    Sh {},

    /// Create a new Teller configuration
    New(NewArgs),

    /// Put new key-values onto a list of providers on a specified path
    Put {
        #[arg(long, short)]
        map_id: String,

        #[arg(long, value_delimiter = ',')]
        providers: Vec<String>,

        #[clap(value_parser = parse_key_val::<String,String>)]
        kvs: Vec<(String, String)>,
    },

    /// Delete specific keys or complete paths
    Delete {
        #[arg(long, short)]
        map_id: String,

        #[arg(long, value_delimiter = ',')]
        providers: Vec<String>,

        keys: Vec<String>,
    },
    Copy {
        #[arg(long, short)]
        from: String,

        #[arg(long, short, value_delimiter = ',')]
        to: Vec<String>,

        #[arg(long, short)]
        replace: bool,
    },
}

fn parse_key_val<T, U>(
    s: &str,
) -> std::result::Result<(T, U), Box<dyn std::error::Error + Send + Sync>>
where
    T: std::str::FromStr,
    T::Err: std::error::Error + Send + Sync + 'static,
    U: std::str::FromStr,
    U::Err: std::error::Error + Send + Sync + 'static,
{
    let pos = s
        .find('=')
        .ok_or_else(|| format!("invalid KEY=value: no `=` found in `{s}`"))?;
    Ok((s[..pos].parse()?, s[pos + 1..].parse()?))
}

#[derive(Debug, Copy, Clone, PartialEq, Eq, PartialOrd, Ord, ValueEnum)]
pub enum Format {
    /// Export as CSV
    CSV,
    /// Export as YAML
    YAML,
    /// Export as JSON
    JSON,
    /// Export as env variables
    ENV,
}

#[allow(clippy::struct_excessive_bools)]
#[derive(Debug, Clone, Args)] // requires `derive` feature
pub struct ScanArgs {
    /// Root folder to scan recursively
    #[arg(short, long, default_value = ".")]
    pub root: String,
    /// Include hidden and ignored files
    #[arg(short, long)]
    pub all: bool,
    /// Returns exit code 1 if has finding
    #[arg(long)]
    pub error_if_found: bool,
    /// Include binary files
    #[arg(short, long)]
    pub binary: bool,
    /// Output matches as JSON
    #[arg(short, long)]
    pub json: bool,
}

const DEFAULT_FILE_PATH: &str = ".teller.yml";

#[derive(Debug, Clone, Args)]
pub struct NewArgs {
    /// Stuff to add
    #[arg(short, long, conflicts_with = "std", default_value=DEFAULT_FILE_PATH)]
    pub filename: PathBuf,

    /// Print configuration to the STDOUT
    #[arg(long)]
    pub std: bool,

    /// Force teller configuration file if exists
    #[arg(long)]
    pub force: bool,

    #[arg(long, value_delimiter = ',')]
    pub providers: Vec<ProviderKind>,
}

fn find_file_upwards(start_dir: &Path, config_filename: &str) -> eyre::Result<Option<PathBuf>> {
    let mut current_dir = start_dir;

    loop {
        let config_path = current_dir.join(config_filename);

        // Check if the configuration file exists at the current path
        if config_path.exists() {
            return Ok(Some(config_path));
        }

        // Move to the parent directory
        match current_dir.parent() {
            Some(parent) => current_dir = parent,
            None => return Ok(None), // No parent means we've reached the root
        }
    }
}

async fn load_teller(config: Option<String>) -> eyre::Result<Teller> {
    let config_arg = if let Some(config) = config {
        config
    } else {
        find_file_upwards(env::current_dir()?.as_path(), DEFAULT_FILE_PATH)?
            .ok_or_eyre("cannot find configuration from current folder and up to root")?
            .to_string_lossy()
            .to_string()
    };

    let config_path = Path::new(&config_arg);
    let teller = Teller::from_yaml(config_path).await?;
    Ok(teller)
}

/// Run the CLI logic
///
/// # Errors
///
/// This function will return an error if operation fails
#[allow(clippy::future_not_send)]
#[allow(clippy::too_many_lines)]
pub async fn run(args: &Cli) -> eyre::Result<Response> {
    match args.command.clone() {
        Commands::Run {
            reset,
            shell,
            command,
        } => {
            let teller = load_teller(args.config.clone()).await?;
            let pwd = std::env::current_dir()?;
            let opts = exec::Opts {
                pwd: pwd.as_path(),
                sh: shell,
                reset_env: reset,
                capture: false,
            };
            teller
                .run(
                    command
                        .iter()
                        .map(String::as_str)
                        .collect::<Vec<_>>()
                        .as_slice(),
                    &opts,
                )
                .await?;
            Response::ok()
        }
        Commands::Scan(cmdargs) => {
            let teller = load_teller(args.config.clone()).await?;
            scan::run(&teller, &cmdargs).await
        }
        Commands::Export { format } => {
            let teller_format = match format {
                Format::CSV => export::Format::CSV,
                Format::YAML => export::Format::YAML,
                Format::JSON => export::Format::JSON,
                Format::ENV => export::Format::ENV,
            };
            let teller = load_teller(args.config.clone()).await?;
            let out = teller.export(&teller_format).await?;
            Response::ok_with_message(out)
        }
        Commands::Redact { in_file, out } => {
            let teller = load_teller(args.config.clone()).await?;
            teller
                .redact(&mut or_stdin(in_file)?, &mut or_stdout(out)?)
                .await?;
            Response::ok()
        }
        Commands::Template { in_file, out } => {
            let mut input = String::new();
            or_stdin(in_file)?.read_to_string(&mut input)?;
            let teller = load_teller(args.config.clone()).await?;
            let rendered = teller.template(&input).await?;
            let mut out = or_stdout(out)?;
            out.write_all(rendered.as_bytes())?;
            out.flush()?;
            Response::ok()
        }
        Commands::Env {} => {
            let teller = load_teller(args.config.clone()).await?;
            let out = teller.export(&export::Format::ENV).await?;
            Response::ok_with_message(out)
        }
        Commands::New(new_args) => new::run(&new_args),
        Commands::Show {} => {
            let teller = load_teller(args.config.clone()).await?;
            let kvs = teller.collect().await?;
            io::print_kvs(&kvs);
            Response::ok()
        }
        Commands::Sh {} => {
            let teller = load_teller(args.config.clone()).await?;
            let out = teller.export(&export::Format::Shell).await?;
            Response::ok_with_message(out)
        }
        Commands::Put {
            kvs,
            map_id,
            providers,
        } => {
            let kvs = kvs
                .iter()
                .map(|(k, v)| KV::from_kv(k, v))
                .collect::<Vec<_>>();
            let teller = load_teller(args.config.clone()).await?;
            teller
                .put(kvs.as_slice(), map_id.as_str(), providers.as_slice())
                .await?;
            Response::ok()
        }
        Commands::Delete {
            map_id,
            providers,
            keys,
        } => {
            let teller = load_teller(args.config.clone()).await?;
            teller
                .delete(keys.as_slice(), &map_id, providers.as_slice())
                .await?;
            Response::ok()
        }
        Commands::Copy { from, to, replace } => {
            // a copy report should state how many keys were copied and to where.
            // invent a new kvrl (key-value resource location) format: kvurl://dotenv/?meta
            // <provider>/<map-id> like server/resource-path
            // <provider>?path=varbatim/path/to/location request specific path overriding resource routing
            //
            // dotenv/map-id -> foo/map-id: copied 4 key(s).
            // dotenv/map-id -> f/map-id: copied 4 key(s).
            // copied 4 key(s) [in replace mode] from `dotenv:path-id` to `foo:path-id`, `bar:path-id`
            let teller = load_teller(args.config.clone()).await?;
            let (from_provider, from_map_id) = from.split_once('/').ok_or_else(|| {
                eyre!(
                    "cannot parse '--from': '{}', did you format it as: '<provider name>/<map \
                     id>' ?",
                    from
                )
            })?;
            for to_provider in to {
                let (to_provider, to_map_id) = to_provider.split_once('/').ok_or_else(|| {
                    eyre!(
                        "cannot parse '--to': '{}', did you format it as: '<provider name>/<map \
                         id>' ?",
                        to_provider
                    )
                })?;
                teller
                    .copy(from_provider, from_map_id, to_provider, to_map_id, replace)
                    .await?;
            }

            Response::ok()
        }
    }
}
