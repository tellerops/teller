use anyhow::{self, Context};
/// main
///
/// # Errors
///
/// This function will return an error
pub fn main() -> anyhow::Result<()> {
    use clap::{AppSettings, Arg, Command};
    let cli = Command::new("xtask")
        .setting(AppSettings::SubcommandRequiredElseHelp)
        .subcommand(
            Command::new("coverage").arg(
                Arg::new("dev")
                    .short('d')
                    .long("dev")
                    .help("generate an html report")
                    .takes_value(false),
            ),
        )
        .subcommand(Command::new("vars"))
        .subcommand(Command::new("ci"))
        .subcommand(Command::new("powerset"))
        .subcommand(
            Command::new("bloat-deps").arg(
                Arg::new("package")
                    .short('p')
                    .long("package")
                    .help("package to build")
                    .required(true)
                    .takes_value(true),
            ),
        )
        .subcommand(
            Command::new("bloat-time").arg(
                Arg::new("package")
                    .short('p')
                    .long("package")
                    .help("package to build")
                    .required(true)
                    .takes_value(true),
            ),
        )
        .subcommand(Command::new("docs"));
    let matches = cli.get_matches();

    let root = xtaskops::ops::root_dir();
    let res = match matches.subcommand() {
        Some(("coverage", sm)) => xtaskops::tasks::coverage(sm.is_present("dev")),
        Some(("vars", _)) => {
            println!("root: {root:?}");
            Ok(())
        }
        Some(("ci", _)) => xtaskops::tasks::ci(),
        Some(("docs", _)) => xtaskops::tasks::docs(),
        Some(("powerset", _)) => xtaskops::tasks::powerset(),
        Some(("bloat-deps", sm)) => xtaskops::tasks::bloat_deps(
            sm.get_one::<String>("package")
                .context("please provide a package with -p")?,
        ),
        Some(("bloat-time", sm)) => xtaskops::tasks::bloat_time(
            sm.get_one::<String>("package")
                .context("please provide a package with -p")?,
        ),
        _ => unreachable!("unreachable branch"),
    };
    res
}
