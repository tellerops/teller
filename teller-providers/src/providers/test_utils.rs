use std::collections::{BTreeMap, HashMap};

use insta::{assert_debug_snapshot, with_settings};

use crate::config::{PathMap, KV};
use crate::{Error, Provider};

pub const ROOT_PATH_A: &str = "secret/development";
pub const ROOT_PATH_B: &str = "secret/multiple/app-1";
pub const ROOT_PATH_C: &str = "secret/multiple/app-2";
pub const ROOT_PATH_PAGING: &str = "secret/multiple/lotsakeys";
const PATH_A_KEY_1: &str = "db";
const PATH_A_KEY_2: &str = "log_level";
const PATH_A_KEY_3: &str = "app";
const PATH_A_VALUE_1: &str = "{\"DB_PASS\": \"1234\",\"DB_NAME\": \"FOO\"}";
const PATH_A_VALUE_2: &str = "DEBUG";
const PATH_A_VALUE_3: &str = "Teller";
const PATH_B_KEY_1: &str = "log_level";
const PATH_B_VALUE_1: &str = "DEBUG";
const PATH_C_KEY_1: &str = "foo";
const PATH_C_VALUE_1: &str = "bar";
const PATH_C_VALUE_1_UPDATE: &str = "baz";

pub struct ProviderTest {
    /// Adding the given prefix to all root path keys. you should use in case you want to change the root path key
    /// in case your provider is required a different path from the test case.
    /// In the all snapshots tests, the give value wan clean and you will not see it to aliment all the providers returns the sane response
    pub root_prefix: Option<String>,

    pub provider: Box<dyn Provider + Send + Sync>,
}

/// A struct representing a test suite for validating a Teller provider's functionality.
///
/// This struct encapsulates a set of test functions designed to validate the behavior of a Teller provider. It provides
/// methods for inserting, retrieving, updating, and deleting key-value pairs from the provider, allowing you to verify
/// the correctness of the provider's implementation.
///
/// # Example
/// ```
/// use tokio::test;
/// use crate::providers::test_utils;
/// use crate::Provider;
///
///
/// async fn sanity_test() {
///     let p = Box::new(super::Inmem::new(None).unwrap()) as Box<dyn Provider + Send + Sync>;
///
///     test_utils::ProviderTestBuilder::default().run(&p).await;
/// }
/// ```
impl ProviderTest {
    pub fn new(provider: Box<dyn Provider + Send + Sync>) -> Self {
        Self {
            root_prefix: None,
            provider,
        }
    }

    pub fn with_root_prefix(mut self, root_prefix: &str) -> Self {
        self.root_prefix = Some(root_prefix.to_string());
        self
    }

    pub async fn run(&self) {
        let path_tree = self.get_tree();

        self.validate_get_unexisting_key().await;
        self.validate_put(&path_tree).await;
        self.validate_get(&path_tree).await;
        self.validate_get_selective(&path_tree).await;
        self.validate_update().await;
        self.validate_delete().await;
        self.validate_delete_keys().await;
    }

    /// Returns a tree structure of test paths with associated key-value pairs.
    ///
    /// This function constructs a tree structure of test paths, where each path is associated with a vector of key-value pairs.
    /// The tree is represented as a `HashMap` where the keys are root paths, and the associated values are vectors of `KV` pairs.
    ///
    fn get_tree(&self) -> HashMap<&str, Vec<KV>> {
        HashMap::from([
            (
                ROOT_PATH_A,
                vec![
                    KV::from_literal(
                        "",
                        PATH_A_KEY_1,
                        PATH_A_VALUE_1,
                        self.provider.as_ref().kind(),
                    ),
                    KV::from_literal(
                        "",
                        PATH_A_KEY_2,
                        PATH_A_VALUE_2,
                        self.provider.as_ref().kind(),
                    ),
                    KV::from_literal(
                        "",
                        PATH_A_KEY_3,
                        PATH_A_VALUE_3,
                        self.provider.as_ref().kind(),
                    ),
                ],
            ),
            (
                ROOT_PATH_B,
                vec![KV::from_literal(
                    "",
                    PATH_B_KEY_1,
                    PATH_B_VALUE_1,
                    self.provider.as_ref().kind(),
                )],
            ),
            (
                ROOT_PATH_C,
                vec![KV::from_literal(
                    "",
                    PATH_C_KEY_1,
                    PATH_C_VALUE_1,
                    self.provider.as_ref().kind(),
                )],
            ),
            (
                ROOT_PATH_PAGING,
                (0..100)
                    .map(|idx| {
                        KV::from_literal(
                            "",
                            &format!("{PATH_C_KEY_1}_{idx}"),
                            PATH_C_VALUE_1,
                            self.provider.as_ref().kind(),
                        )
                    })
                    .collect::<Vec<_>>(),
            ),
        ])
    }

    fn get_key_path(&self, root_path: &str) -> String {
        self.root_prefix.as_ref().map_or_else(
            || root_path.to_string(),
            |prefix| format!("{prefix}{root_path}"),
        )
    }

    /// Validates that a Teller provider implementation returns an error when attempting to retrieve a non-existent key path.
    ///
    /// This function checks the behavior of the provided `Provider` implementation when trying to access an invalid key path
    /// by constructing a key path from a given root path and an invalid path segment. It then awaits the result and asserts
    /// that it is an error.
    async fn validate_get_unexisting_key(&self) {
        let res = self
            .provider
            .as_ref()
            .get(&PathMap::from_path(&format!("{ROOT_PATH_A}/invalid-path")))
            .await;
        println!("validate_get_unexisting_key: {res:?}");
        assert!(res.is_err());
    }

    /// Validate that Teller provider successfully put all the tests tree map into the provider.
    async fn validate_put(&self, path_tree: &HashMap<&str, Vec<KV>>) {
        for (root_path, keys) in path_tree {
            let path_map = PathMap::from_path(&self.get_key_path(root_path));
            let res = self.provider.as_ref().put(&path_map, keys).await;
            println!("validate_put: {res:?}");
            assert!(res.is_ok());
            assert_debug_snapshot!(format!("[put-{}]", root_path.replace('/', "_"),), res);
        }
    }

    /// Validates that a Teller provider successfully stores a tree map of key-value pairs.
    ///
    /// This function tests the behavior of a Teller provider by attempting to put a collection of key-value
    /// into the provider and the function returns successfully response.
    /// Each put response function snapshot the function response struct
    async fn validate_get(&self, path_tree: &HashMap<&str, Vec<KV>>) {
        for root_path in path_tree.keys() {
            let res = self
                .provider
                .as_ref()
                .get(&PathMap::from_path(&self.get_key_path(root_path)))
                .await;

            println!("validate_get: {res:?}");
            assert!(res.is_ok());

            let mut res = res.unwrap();
            res.sort_by(|a: &KV, b| a.value.cmp(&b.value));
            res.sort_by(|a: &KV, b| a.key.cmp(&b.key));

            with_settings!({filters => vec![
                (format!("{:?}", self.provider.as_ref().kind().kind).as_str(), "PROVIDER_KIND"),
                (format!("{:?}", self.provider.as_ref().kind().name).as_str(), "PROVIDER_NAME"),
                (format!("\".*{ROOT_PATH_A}").as_str(), format!("\"{ROOT_PATH_A}").as_str()),
                (format!("\".*{ROOT_PATH_B}").as_str(), format!("\"{ROOT_PATH_B}").as_str()),
                (format!("\".*{ROOT_PATH_C}").as_str(), format!("\"{ROOT_PATH_C}").as_str()),
                (format!("\".*{ROOT_PATH_PAGING}").as_str(), format!("\"{ROOT_PATH_PAGING}").as_str()),
            ]}, {
                assert_debug_snapshot!(
                format!("[get-{}]", root_path.replace('/', "_"),),
                res
            );
            });
        }
    }

    #[allow(unused_variables)]
    async fn validate_get_selective(&self, path_tree: &HashMap<&str, Vec<KV>>) {
        let mut selective_pm = PathMap::from_path(&self.get_key_path(ROOT_PATH_A));
        selective_pm
            .keys
            .insert(PATH_A_KEY_1.to_string(), PATH_A_KEY_1.to_string());

        let res = self.provider.as_ref().get(&selective_pm).await;

        println!("validate_get_selective: {res:?}");
        assert!(res.is_ok());

        with_settings!({filters => vec![
                    (format!("{:?}", self.provider.as_ref().kind().kind).as_str(), "PROVIDER_KIND"),
                    (format!("{:?}", self.provider.as_ref().kind().name).as_str(), "PROVIDER_NAME"),
                    (format!("\".*{ROOT_PATH_A}").as_str(), format!("\"{ROOT_PATH_A}").as_str()),
                    (format!("\".*{ROOT_PATH_B}").as_str(), format!("\"{ROOT_PATH_B}").as_str()),
                    (format!("\".*{ROOT_PATH_C}").as_str(), format!("\"{ROOT_PATH_C}").as_str()),
                    (format!("\".*{ROOT_PATH_PAGING}").as_str(), format!("\"{ROOT_PATH_PAGING}").as_str()),
                ]}, {
                    assert_debug_snapshot!(
                    format!("[get-selective-{}]", PATH_A_KEY_1.replace('/', "_"),),
                    res
                );
        });
    }

    /// Validates the update operation on a Teller provider after inserting a tree structure.
    ///
    /// This function is intended to be executed after running the `validate_get` function to insert a tree structure into the
    /// Teller provider. It performs the following steps:
    ///
    /// 1. Updates a specific key-value pair with new values and expects a successful response from the provider.
    /// 2. Verifies that the updated value is reflected correctly in the provider.
    ///
    /// The function utilizes the provided `provider` to perform the update operation and snapshots the results for verification.
    async fn validate_update(&self) {
        // ** Update value and expected success response.
        let update_keys = vec![KV::from_literal(
            "",
            PATH_C_KEY_1,
            PATH_C_VALUE_1_UPDATE,
            self.provider.as_ref().kind(),
        )];
        let update_res = self
            .provider
            .as_ref()
            .put(
                &PathMap::from_path(&self.get_key_path(ROOT_PATH_C)),
                &update_keys,
            )
            .await;
        assert!(update_res.is_ok());
        assert_debug_snapshot!(
            format!("[put-update-{}]", ROOT_PATH_C.replace('/', "_"),),
            update_res
        );

        // ** Read the updated value and make sure that the value is updated.
        let read_after_update_res = self
            .provider
            .as_ref()
            .get(&PathMap::from_path(&self.get_key_path(ROOT_PATH_C)))
            .await;
        assert!(read_after_update_res.is_ok());

        with_settings!({filters => vec![
                (format!("{:?}", self.provider.as_ref().kind().kind).as_str(), "PROVIDER_KIND"),
                (format!("{:?}", self.provider.as_ref().kind().name).as_str(), "PROVIDER_NAME"),
                (format!("\".*{ROOT_PATH_A}").as_str(), format!("\"{ROOT_PATH_A}").as_str()),
                (format!("\".*{ROOT_PATH_B}").as_str(), format!("\"{ROOT_PATH_B}").as_str()),
                (format!("\".*{ROOT_PATH_C}").as_str(), format!("\"{ROOT_PATH_C}").as_str()),
            ]}, {
                assert_debug_snapshot!(
            format!("[get-after-update-{}]", ROOT_PATH_C.replace('/', "_"),),
            read_after_update_res
        );
            });
    }

    /// Validates the deletion of a key from a Teller provider.
    ///
    /// This function tests the behavior of a Teller provider by performing the following steps:
    ///
    /// 1. Deletes a specific key from the provider using the `del` method.
    /// 2. Verifies that the deletion operation is successful by checking the result for success.
    /// 3. Attempts to retrieve the deleted key and confirms that it returns an error, indicating that the key no longer exists.
    async fn validate_delete(&self) {
        let delete_res = self
            .provider
            .as_ref()
            .del(&PathMap::from_path(&self.get_key_path(ROOT_PATH_B)))
            .await;
        assert!(delete_res.is_ok());
        assert_debug_snapshot!(
            format!("[del-{}]", ROOT_PATH_B.replace('/', "_"),),
            delete_res
        );

        // ** Validate deletion key.

        let get_del_res = self
            .provider
            .as_ref()
            .get(&PathMap::from_path(&self.get_key_path(ROOT_PATH_B)))
            .await;

        assert!(matches!(
            get_del_res,
            Err(Error::NotFound { path: _, msg: _ })
        ));
        assert!(get_del_res.is_err());
    }

    /// Validates deletion specifics key from a Teller provider.
    ///
    /// This function tests the behavior of a Teller provider by performing the following steps:
    ///
    /// 1. Delete specifics keys from a path by using adding keys list.
    /// 2. Verifies that the deletion operation is successful by run `get`
    ///    again on the same path and expecting that not see the deletion keys
    async fn validate_delete_keys(&self) {
        let mut path_path = PathMap::from_path(&self.get_key_path(ROOT_PATH_A));

        path_path.keys = BTreeMap::from([
            (PATH_A_KEY_2.to_string(), String::new()),
            (PATH_A_KEY_3.to_string(), String::new()),
        ]);

        let delete_keys_res = self.provider.as_ref().del(&path_path).await;
        println!("validate_delete_keys: {delete_keys_res:#?}");
        assert!(delete_keys_res.is_ok());
        assert_debug_snapshot!(
            format!("[del-keys-{}]", ROOT_PATH_A.replace('/', "_")),
            delete_keys_res
        );

        let get_del_res = self
            .provider
            .as_ref()
            .get(&PathMap::from_path(&self.get_key_path(ROOT_PATH_A)))
            .await;

        with_settings!({filters => vec![
                (format!("{:?}", self.provider.as_ref().kind().kind).as_str(), "PROVIDER_KIND"),
                (format!("{:?}", self.provider.as_ref().kind().name).as_str(), "PROVIDER_NAME"),
                (format!("\".*{ROOT_PATH_A}").as_str(), format!("\"{ROOT_PATH_A}").as_str()),
            ]}, {
                assert_debug_snapshot!(
            format!("[get-del-keys-{}]", ROOT_PATH_A.replace('/', "_")),
            get_del_res
        );
            });
    }
}
