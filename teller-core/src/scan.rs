use std::fs;

use aho_corasick::AhoCorasick;
use ignore::WalkBuilder;
use teller_providers::config::KV;
use unicode_width::UnicodeWidthStr;

use crate::{config::Match, io::is_binary_file, Error, Result};

#[derive(Debug, Clone, Default)]
pub struct Opts {
    pub include_all: bool,
    pub include_binary: bool,
}

/// (ln, col), 1 based (not zero based)
fn get_visual_position(text: &[u8], byte_position: usize) -> Option<(usize, usize)> {
    if byte_position >= text.len() || text.is_empty() {
        return None;
    }

    let lines = text
        .iter()
        .take(byte_position)
        .filter(|c| **c == b'\n')
        .count();
    let last_ln_start = text
        .iter()
        .take(byte_position)
        .rposition(|c| *c == b'\n')
        .unwrap_or(0);

    let len = UnicodeWidthStr::width(
        String::from_utf8_lossy(&text[last_ln_start..byte_position]).as_ref(),
    );

    // index starts from 1 for both
    Some((lines + 1, len + 1))
}

// aho
// offset into row/col visual https://github.com/zkat/miette/blob/f4d056e1ffeb9a0bf36e2a6501365bd7e00db22d/src/handlers/graphical.rs#L619
// match repr
///
/// # Errors
///
/// TODO
#[allow(clippy::module_name_repetitions)]
pub fn scan_root(root: &str, kvs: &[KV], opts: &Opts) -> Result<Vec<Match>> {
    let patterns = kvs.iter().map(|kv| kv.value.as_str()).collect::<Vec<_>>();
    let finder = AhoCorasick::new(patterns).map_err(|e| Error::Message(e.to_string()))?;

    let mut wb = WalkBuilder::new(root);

    let mut matches = vec![];
    for entry in wb
        .ignore(!opts.include_all)
        .git_ignore(!opts.include_all)
        .hidden(opts.include_all)
        .build()
        .filter_map(Result::ok)
        .filter(|ent| ent.path().is_file())
    {
        let path = entry.path();
        if is_binary_file(path)? && !opts.include_binary {
            continue;
        }

        let content = String::from_utf8_lossy(&fs::read(path)?).to_string();
        let bytes = content.as_bytes();

        finder.find_iter(&content).for_each(|aho_match| {
            matches.push(Match {
                path: path.to_path_buf(),
                query: kvs[aho_match.pattern()].clone(),
                position: get_visual_position(bytes, aho_match.start()),
                offset: aho_match.start(),
            });
        });
    }

    matches.sort();
    Ok(matches)
}

#[cfg(test)]
mod tests {
    use std::{
        fs,
        path::{Path, PathBuf},
    };

    use insta::assert_debug_snapshot;
    use teller_providers::{
        config::{ProviderInfo, KV},
        providers::ProviderKind,
    };

    use super::*;
    use crate::scan;

    fn normalize_path_separators(path: &Path) -> PathBuf {
        let path_str = path.to_string_lossy().replace('\\', "/");
        PathBuf::from(path_str)
    }
    fn normalize_matches(ms: &[Match]) -> Vec<Match> {
        ms.iter()
            .map(|m| Match {
                path: normalize_path_separators(&m.path),
                ..m.clone()
            })
            .collect::<Vec<_>>()
    }

    #[test]
    fn test_position() {
        assert_eq!(get_visual_position(b"", 4), None);
        assert_eq!(get_visual_position(b"", 1), None);
        assert_eq!(get_visual_position(b"", 0), None);
        assert_eq!(get_visual_position(b"a", 1), None);

        assert_eq!(get_visual_position(b"abcde\nfghi", 8), Some((2, 3)));
        assert_eq!(get_visual_position(b"abcde\r\nfghi", 8), Some((2, 2)));

        let text = r#" 100% ❯ j teller-rs
    /Users/jondot/spikes/teller-rs
    (base)
    ~/spikes/teller-rs on  master [!?] via 🦀 v1.73.0-nightly
     100% ❯ code .
    (base)
    ~/spikes/teller-rs on  master [!?] via 🦀 v1.73.0-nightly
     100% ❯ [WARN] - (starship::utils): Executing command "/opt/homebrew/bin/git" timed out.
    (base)
    ~/spikes/teller-rs on  master [!?] via 🦀 v1.73.0-nightly
     100% ❯ open /Users/jondot/Movies
    (base)
    ~/spikes/teller-rs on  master [!?] via 🦀 v1.73.0-nightly
     100% ❯"#;
        let position = get_visual_position(text.as_bytes(), 438);
        assert_eq!(position, Some((11, 19)));
    }

    #[test]
    fn test_scan() {
        let provider = ProviderInfo {
            kind: ProviderKind::Inmem,
            name: "test".to_string(),
        };
        let kvs = vec![
            KV::from_literal("/some/path", "key1", "hashicorp", provider.clone()),
            KV::from_literal("/some/path", "key1", "dont-find-me", provider.clone()),
            KV::from_literal("/some/path", "key1", "trooper123", provider.clone()),
            KV::from_literal("/some/path", "key1", "pass1", provider.clone()),
            KV::from_literal("/some/path", "key1", "nested111", provider),
        ];

        let res = scan_root("fixtures", &kvs[..], &scan::Opts::default());
        assert_debug_snapshot!(normalize_matches(&res.unwrap()));

        let res = scan_root(
            "fixtures",
            &kvs[..],
            &scan::Opts {
                include_binary: true,
                include_all: false,
            },
        );
        assert_debug_snapshot!(normalize_matches(&res.unwrap()));

        fs::write("fixtures/git-ignored-file", "trooper123").expect("cannot write file");

        let res = scan_root(
            "fixtures",
            &kvs[..],
            &scan::Opts {
                include_binary: false,
                include_all: true,
            },
        );
        assert_debug_snapshot!(normalize_matches(&res.unwrap()));
    }
}
