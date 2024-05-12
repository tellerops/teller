use std::{collections::HashMap, path::Path, process::Output};

// use crate::{Error, Result};
// use teller_providers::errors::{Error, Result};
use crate::{Error, Result};
pub struct Opts<'a> {
    pub pwd: &'a Path,
    pub capture: bool,
    pub sh: bool,
    pub reset_env: bool,
}

const ENV_OK: &[&str] = &[
    "USER",
    "HOME",
    "PATH",
    "TMPDIR",
    "SHELL",
    "SSH_AUTH_SOCK",
    "LANG",
    "LC_ALL",
    "TEMPDIR",
    "TERM",
    "COLORTERM",
    "LOGNAME",
];

/// Run a command
///
/// # Errors
///
/// This function will return an error if running command fails
pub fn cmd(cmdstr: &str, env_kvs: &[(String, String)], opts: &Opts<'_>) -> Result<Output> {
    let words = if opts.sh {
        shell_command_argv(cmdstr.into())
    } else {
        shell_words::split(cmdstr)?.iter().map(Into::into).collect()
    };
    cmd_slice(
        words
            .iter()
            .map(String::as_str)
            .collect::<Vec<_>>()
            .as_slice(),
        env_kvs,
        opts,
    )
}

fn cmd_slice(words: &[&str], env_kvs: &[(String, String)], opts: &Opts<'_>) -> Result<Output> {
    // env handling
    let mut env_map: HashMap<_, _> = if opts.reset_env {
        std::env::vars()
            .filter(|(k, _)| ENV_OK.contains(&k.as_str()))
            .collect()
    } else {
        std::env::vars().collect()
    };

    for (k, v) in env_kvs {
        env_map.insert(k.clone(), v.clone());
    }

    // no shell
    let (first, rest) = words
        .split_first()
        .ok_or_else(|| Error::Message("command has not enough arguments".to_string()))?;

    let mut expr = duct::cmd(Path::new(first), rest)
        .dir(opts.pwd)
        .full_env(&env_map);

    if opts.capture {
        expr = expr.stdout_capture();
    }

    Ok(expr.run()?)
}

#[cfg(unix)]
fn shell_command_argv(command: String) -> Vec<String> {
    use std::env;

    let shell = env::var("SHELL").unwrap_or_else(|_| "/bin/sh".into());
    vec![shell, "-c".into(), command]
}

#[cfg(windows)]
fn shell_command_argv(command: String) -> Vec<String> {
    let comspec = std::env::var_os("COMSPEC")
        .and_then(|s| s.into_string().ok())
        .unwrap_or_else(|| "cmd.exe".into());
    vec![comspec, "/C".into(), command]
}

#[cfg(test)]
mod tests {
    use std::path::Path;

    use insta::assert_debug_snapshot;
    use teller_providers::config::ProviderInfo;
    use teller_providers::config::KV;
    use teller_providers::providers::ProviderKind;

    use super::cmd;
    use super::Opts;

    #[test]
    #[cfg(not(windows))]
    fn run_echo() {
        let out = cmd(
            "echo $MY_VAR",
            &std::iter::once(&KV::from_literal(
                "/foo/bar",
                "MY_VAR",
                "shazam",
                ProviderInfo {
                    kind: ProviderKind::Inmem,
                    name: "test".to_string(),
                },
            ))
            .map(|kv| (kv.key.clone(), kv.value.clone()))
            .collect::<Vec<_>>(),
            &Opts {
                pwd: Path::new("."),
                capture: true,
                reset_env: true,
                sh: true,
            },
        )
        .unwrap();
        let s = String::from_utf8_lossy(&out.stdout[..]);
        assert_debug_snapshot!(s);
    }

    #[ignore]
    #[test]
    fn env_reset() {
        let out = cmd(
            "/usr/bin/env",
            &std::iter::once(&KV::from_literal(
                "/foo/bar",
                "MY_VAR",
                "shazam",
                ProviderInfo {
                    kind: ProviderKind::Inmem,
                    name: "test".to_string(),
                },
            ))
            .map(|kv| (kv.key.clone(), kv.value.clone()))
            .collect::<Vec<_>>(),
            &Opts {
                pwd: Path::new("."),
                capture: true,
                reset_env: false, // <-- notice this!
                sh: false,
            },
        )
        .unwrap();
        let stdout = String::from_utf8_lossy(&out.stdout[..]).to_string();

        // dirty secret here
        assert!(stdout.contains("GITHUB_TOKEN="));

        let out = cmd(
            "/usr/bin/env",
            &std::iter::once(&KV::from_literal(
                "/foo/bar",
                "MY_VAR",
                "shazam",
                ProviderInfo {
                    kind: ProviderKind::Inmem,
                    name: "test".to_string(),
                },
            ))
            .map(|kv| (kv.key.clone(), kv.value.clone()))
            .collect::<Vec<_>>(),
            &Opts {
                pwd: Path::new("."),
                capture: true,
                reset_env: true, // <-- reset env
                sh: false,
            },
        )
        .unwrap();
        let stdout = String::from_utf8_lossy(&out.stdout[..]).to_string();

        assert!(stdout.contains("USER="));
        assert!(stdout.contains("PATH="));
        // no secret here!
        assert!(!stdout.contains("GITHUB_TOKEN="));
    }
}
