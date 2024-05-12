use std::process::exit;

use clap::Parser;
use eyre::Result;
use teller_cli::{cli, tracing};

#[tokio::main]
async fn main() -> Result<()> {
    let args = cli::Cli::parse();

    tracing(args.verbose);

    let resp = cli::run(&args).await?;

    if let Some(msg) = resp.message {
        println!("{msg}");
    }
    exit(resp.code);
}
