use std::collections::HashMap;

use dockertest_server::{
    servers::{
        cloud::{LocalStackServer, LocalStackServerConfig},
        hashi::{VaultServer, VaultServerConfig},
    },
    Test,
};
use fs_err as fs;
use insta::assert_debug_snapshot;
use teller_core::config::Config;
use teller_providers::{
    config::{ProviderInfo, KV},
    providers::ProviderKind,
    registry::Registry,
};

fn build_providers() -> Test {
    let mut test = Test::new();
    test.register(
        VaultServerConfig::builder()
            .port(9200)
            .version("1.8.2".into())
            .build()
            .unwrap(),
    );
    test.register(
        LocalStackServerConfig::builder()
            .env(
                vec![(
                    "SERVICES".to_string(),
                    "iam,sts,ssm,kms,secretsmanager".to_string(),
                )]
                .into_iter()
                .collect(),
            )
            .port(4561)
            .version("2.0.2".into())
            .build()
            .unwrap(),
    );
    test
}

#[test]
#[cfg(not(windows))]
fn providers_smoke_test() {
    use std::{env, time::Duration};

    if env::var("RUNNER_OS").unwrap_or_default() == "macOS" {
        return;
    }

    let test = build_providers();

    test.run(|instance| async move {
        let user = "linus";
        let password = "torvalds123";
        let vault_server: VaultServer = instance.server();
        // banner is not enough for vault, we have to wait for the image to stabilize
        tokio::time::sleep(Duration::from_secs(2)).await;

        let localstack_server: LocalStackServer = instance.server();

        fs::write(
            "fixtures/flow.env",
            r"
FOO=bar
HELLO=world
",
        )
        .unwrap();
        let mut vars = HashMap::new();
        vars.insert("address".to_string(), vault_server.external_url());
        vars.insert("token".to_string(), vault_server.token);
        vars.insert("endpoint_url".to_string(), localstack_server.external_url());

        let config = Config::with_vars(
            &fs::read_to_string("fixtures/flow_test.yml").unwrap(),
            &vars,
        )
        .unwrap();
        let registry = Registry::new(&config.providers).await.unwrap();

        // (1) start: put on hashi
        let hashi = registry.get("hashi_1").unwrap();
        let hashi_pm0 = &config.providers.get("hashi_1").unwrap().maps[0];
        hashi
            .put(
                hashi_pm0,
                &[
                    KV::from_value(
                        user,
                        "USER",
                        "USER",
                        hashi_pm0,
                        ProviderInfo {
                            kind: ProviderKind::Hashicorp,
                            name: "test".to_string(),
                        },
                    ),
                    KV::from_value(
                        password,
                        "PASSWORD",
                        "PASSWORD",
                        hashi_pm0,
                        ProviderInfo {
                            kind: ProviderKind::Hashicorp,
                            name: "test".to_string(),
                        },
                    ),
                ],
            )
            .await
            .unwrap();
        let res = hashi.get(hashi_pm0).await;
        assert_debug_snapshot!("flow-test-hashi-0", res);

        // (2) push results into secretsmanager
        let kvs = res.unwrap();

        let smgr = registry.get("sm_1").unwrap();
        let smgr_pm0 = &config.providers.get("sm_1").unwrap().maps[0];
        smgr.put(smgr_pm0, &kvs[..]).await.unwrap();
        let res = smgr.get(smgr_pm0).await;
        assert_debug_snapshot!("flow-test-smgr-0", res);

        // (3) push results into ssm - remember it has custom key mapping so
        // check that in snapshots (USER -> USER_NAME, and drops the pass)
        let kvs = res.unwrap();

        let ssm = registry.get("ssm_1").unwrap();
        let ssm_pm0 = &config.providers.get("ssm_1").unwrap().maps[0];
        ssm.put(ssm_pm0, &kvs[..]).await.unwrap();
        let res = ssm.get(ssm_pm0).await;
        assert_debug_snapshot!("flow-test-ssm-0", res);

        // (4) lastly, write it into dotenv file, read it back, and snapshot it
        let kvs = res.unwrap();

        let dot = registry.get("dot_1").unwrap();
        let dot_pm0 = &config.providers.get("dot_1").unwrap().maps[0];
        dot.put(dot_pm0, &kvs[..]).await.unwrap();
        let res = dot.get(dot_pm0).await;
        assert_debug_snapshot!("flow-test-dot-0", res);
    });
}
