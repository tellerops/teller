use std::fs;

fn prep_data_for_mutating_tests() {
    fs::write("tests/cmd/put.in/new.env", "EMPTY=true\n").expect("writing a fixture file");
    fs::write(
        "tests/cmd/delete.in/new.env",
        "EMPTY=true\nDELETE_ME=true\n",
    )
    .expect("writing a fixture file");
    fs::write("tests/cmd/copy.in/target.env", "TARGET_ONLY=true\n")
        .expect("writing a fixture file");
}
#[test]
fn cli_tests() {
    fs::write("tests/cmd/scan.in/git-hidden-file", "happy").expect("writing a fixture file");
    prep_data_for_mutating_tests();

    let c = trycmd::TestCases::new();
    c.case("tests/cmd/*.trycmd");
    #[cfg(windows)]
    c.skip("tests/cmd/run.trycmd");

    c.run();
    prep_data_for_mutating_tests();
}
