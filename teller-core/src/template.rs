use teller_providers::config::KV;
use tera::{from_value, to_value, Context, Result, Tera};

struct KeyFn {
    kvs: Vec<KV>,
}
impl tera::Function for KeyFn {
    fn call(
        &self,
        args: &std::collections::HashMap<String, tera::Value>,
    ) -> tera::Result<tera::Value> {
        args.get("name").map_or_else(
            || Err("cannot get parameter 'name'".into()),
            |val| {
                from_value::<String>(val.clone()).map_or_else(
                    |_| Err("cannot get parameter 'name'".into()),
                    |v| {
                        self.kvs
                            .iter()
                            .find(|kv| kv.key == v)
                            .and_then(|kv| to_value(&kv.value).ok())
                            .ok_or_else(|| "not found".into())
                    },
                )
            },
        )
    }
}

/// Render a template with access to KVs
///
/// # Errors
///
/// This function will return an error if rendering fails
pub fn render(template: &str, kvs: Vec<KV>) -> Result<String> {
    let mut tera = Tera::default();
    tera.register_function("key", KeyFn { kvs });
    let res = tera.render_str(template, &Context::new())?;
    Ok(res)
}

#[cfg(test)]
mod tests {
    use insta::assert_debug_snapshot;
    use teller_providers::{config::ProviderInfo, providers::ProviderKind};

    use super::*;

    #[test]
    fn render_template() {
        let kvs = &[KV::from_literal(
            "some/path",
            "k",
            "foobaz",
            ProviderInfo {
                kind: ProviderKind::Inmem,
                name: "test".to_string(),
            },
        )];
        assert_debug_snapshot!(render("hello {{ key(name='k') }}", kvs.to_vec()));
    }
}
