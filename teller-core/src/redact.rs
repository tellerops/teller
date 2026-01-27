use std::{
    borrow::Cow,
    io::{BufRead, Write},
};

// use crate::{Result, KV};
use teller_providers::config::KV;

pub struct Redactor {}

impl Redactor {
    #[must_use]
    pub const fn new() -> Self {
        Self {}
    }

    /// Redact a reader into writer
    ///
    /// # Errors
    ///
    /// This function will return an error if IO fails
    pub fn redact<R: BufRead, W: Write>(
        &self,
        reader: R,
        mut writer: W,
        kvs: &[KV],
    ) -> std::io::Result<()> {
        for line in reader.lines().map_while(Result::ok) {
            let redacted = self.redact_string(line.as_str(), kvs);
            writer.write_all(redacted.as_bytes())?;
            writer.write_all(b"\n")?; // TODO: support crlf for windows
            writer.flush()?;
        }
        Ok(())
    }

    #[must_use]
    pub fn redact_string<'a>(&'a self, message: &'a str, kvs: &[KV]) -> Cow<'a, str> {
        if self.has_match(message, kvs) {
            let mut redacted = message.to_string();
            for kv in kvs {
                // only replace values with at least 2 chars
                if kv.value.len() >= 2 {
                    redacted = redacted.replace(
                        &kv.value,
                        kv.meta
                            .as_ref()
                            .and_then(|m| m.redact_with.as_ref())
                            .map_or("[REDACTED]", |s| s.as_str()),
                    );
                }
            }
            Cow::Owned(redacted)
        } else {
            Cow::Borrowed(message)
        }
    }

    #[must_use]
    pub fn has_match<'a>(&'a self, message: &'a str, kvs: &[KV]) -> bool {
        kvs.iter().any(|kv| message.contains(&kv.value))
    }
}

impl Default for Redactor {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use std::io::{BufReader, BufWriter};

    use stringreader::StringReader;
    use teller_providers::{config::ProviderInfo, providers::ProviderKind};

    use super::*;

    #[test]
    fn redact_none() {
        let data = "foobar\nfoobaz\n";
        let mut reader = BufReader::new(StringReader::new(data));
        let mut writer = BufWriter::new(Vec::new());
        let redactor = Redactor {};

        redactor.redact(&mut reader, &mut writer, &[]).unwrap();
        let s = String::from_utf8(writer.into_inner().unwrap()).unwrap();
        assert_eq!(s, "foobar\nfoobaz\n");
    }

    #[test]
    fn redact_some() {
        let data = "foobar\nfoobaz\n";
        let mut reader = BufReader::new(StringReader::new(data));
        let mut writer = BufWriter::new(Vec::new());
        let redactor = Redactor {};

        redactor
            .redact(
                &mut reader,
                &mut writer,
                &[KV::from_literal(
                    "some/path",
                    "k",
                    "foobaz",
                    ProviderInfo {
                        kind: ProviderKind::Inmem,
                        name: "test".to_string(),
                    },
                )],
            )
            .unwrap();
        let s = String::from_utf8(writer.into_inner().unwrap()).unwrap();
        assert_eq!(s, "foobar\n[REDACTED]\n");
    }
}
