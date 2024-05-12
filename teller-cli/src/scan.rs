use comfy_table::presets::NOTHING;
use comfy_table::{Cell, Table};
use eyre::Result;
use teller_core::{scan, teller::Teller};

use crate::cli::ScanArgs;
use crate::Response;

fn hide_chars(s: &str) -> String {
    let mut result = String::new();
    let chars_to_display = s.chars().take(2).collect::<String>();
    let asterisks = "*".repeat(3);
    result.push_str(&chars_to_display);
    result.push_str(&asterisks);
    result
}

/// Scan a folder for secrets fetched from providers
///
/// # Errors
///
/// This function will return an error if the operation fails
#[allow(clippy::future_not_send)]
pub async fn run(teller: &Teller, args: &ScanArgs) -> Result<Response> {
    let opts = scan::Opts {
        include_all: args.all,
        include_binary: args.binary,
    };

    let kvs = teller.collect().await?;
    let res = teller.scan(&args.root, &kvs, &opts)?;
    let count = res.len();
    eprintln!("scanning for {} item(s) in {}", kvs.len(), args.root);
    if args.json {
        println!("{}", serde_json::to_string_pretty(&res)?);
    } else {
        let mut table = Table::new();
        table.load_preset(NOTHING);
        for m in res {
            let pos = m.position.unwrap_or((0, 0));
            table.add_row(vec![
                Cell::new(format!("{}:{}", pos.0, pos.1)),
                Cell::new(m.path.to_string_lossy()),
                Cell::new(hide_chars(&m.query.value)),
                Cell::new(
                    m.query
                        .provider
                        .map_or_else(|| "n/a".to_string(), |p| p.kind.to_string())
                        .to_string(),
                ),
                Cell::new(m.query.path.map_or_else(|| "n/a".to_string(), |p| p.path)),
            ]);
        }
        println!("{table}");
    }
    eprintln!("found {count} result(s)");

    if args.error_if_found && count > 0 {
        Response::fail()
    } else {
        Response::ok()
    }
}
