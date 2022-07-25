# End to End Tests

This package contains integration tests.

## Running
```sh
make e2e
```

## Create E2E test

You have two different options to create an E2E test in Teller.

### Simple
The simple and fast way is based on `yml` file configuration. All you need to do is create a *.yml file in [test folder](./tests/) with the following fields:

|         Field                 |           Description             |
|-------------------------------|------------------------------------
| `name`                        | E2E name.
| `command`                     | Command to execute. You need to keep `<name>`; this placeholder will replace with the `BINARY_PATH` binary value.
| `config_file_name`            | Configuration file name. If empty, the configuration file will not be created.
| `config_content`              | Configuration file content.
| `init_snapshot`               | List of files that were going to create before Teller binary is executed.
| `init_snapshot.path`          | Create file in path.
| `init_snapshot.file_name`     | File name.
| `init_snapshot.content`       | File content.
| `replace_stdout_content`      | Replace dynamic stdout content to static. for example, replace current timestemp to static text.
| `expected_snapshot`           | Compare the init_snapshot folder with the expected snapshot content. If empty, this compare check will be ignored.
| `expected_snapshot.path`      | Create file in path.
| `expected_snapshot.file_name` | File name.
| `expected_snapshot.content`   | File content.
| `expected_stdout`             | Check if Teller stdout equals this value. If empty, this check will be ignored.
| `expected_stderr`             | Check if Teller stderr equals this value. If empty, this check will be ignored.

### Advanced
In case the E2E `yml`not flexible enough you have the option to create a `*.go` file in [test folder](./tests/). the go file most to:
1. Implement [TestCaseDescriber](./register/interface.go) interface
2. Must to user `init` function and register to [AddSuite](./register/register.go)