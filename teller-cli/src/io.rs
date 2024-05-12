use std::io::{self, BufRead, BufReader, BufWriter, Write};

use eyre::Result;
use fs_err::File;
use teller_providers::config::KV;

/// Read from a file or stdin
///
/// # Errors
///
/// This function will return an error if IO fails
pub fn or_stdin(file: Option<String>) -> Result<Box<dyn BufRead>> {
    let out: Box<dyn BufRead> = file.map_or_else(
        || Ok(Box::new(BufReader::new(io::stdin())) as Box<dyn BufRead>),
        |file_path| File::open(file_path).map(|f| Box::new(BufReader::new(f)) as Box<dyn BufRead>),
    )?;

    Ok(out)
}

/// Write to a file or stdout
///
/// # Errors
///
/// This function will return an error if IO fails
pub fn or_stdout(file: Option<String>) -> Result<Box<dyn Write>> {
    let out = file.map_or_else(
        || Ok(Box::new(BufWriter::new(std::io::stdout())) as Box<dyn Write>),
        |file_path| File::open(file_path).map(|f| Box::new(BufWriter::new(f)) as Box<dyn Write>),
    )?;
    Ok(out)
}

pub fn print_kvs(kvs: &[KV]) {
    for kv in kvs {
        println!(
            "[{}]: {} = {}***",
            kv.provider
                .as_ref()
                .map_or_else(|| "n/a".to_string(), |p| format!("{} ({})", p.name, p.kind)),
            kv.key,
            kv.value.get(0..2).unwrap_or_default()
        );
    }
}
