# How to add a new provider

Adding a new Teller provider is very easy, but you still need to know where to start. We summarize the steps very shortly to make your life easier

## Provider implementation

1. Copy the file [example.go](../examples/providers/example.go) from [examples/providers/example.go](../examples/providers/example.go) and make sure to implement all the required behaviors. The [example.go](../examples/providers/example.go) file is a skeleton for adding a new provider, it contains stubs for an interface which declares the required functionality that any provider must have.
2. After copying to the providers dir, remove the comment from the `init` function to register your provider

```go
 // func init() {
// 	metaInto := core.MetaInfo{
// 		Description:    "ProviderName",
// 		Name:           "provider_name",
// 		Authentication: "If you have the Consul CLI working and configured, there's no special action to take.\nConfiguration is environment based, as defined by client standard. See variables [here](https://github.com/hashicorp/consul/blob/master/api/api.go#L28).",
// 		ConfigTemplate: `
//   provider:
//     env:
//       KEY_EAXMPLE:
//         path: pathToKey
// `,
// 		Ops: core.OpMatrix{Get: true, GetMapping: true, Put: true, PutMapping: true},
// 	}
// 	RegisterProvider(metaInto, NewExample)
// }
```

Set `Description, Name, and Authentication`, as well as `OpMatrix` that descibes the action this provider supports, based on your implementation.

3. Add a provider template configuration in path: [pkg/wizard_template.go](../pkg/wizard_template.go). This will be used to auto-generate a configuration.

```go
{{- if index .ProviderKeys "example" }}
  # Add here some authentication requirements, like a token that should be in the user's environment.
  example:
    env_sync:
       path: redis/config
    env:
      ETC_DSN:
        path: redis/config/foobar

{{end}}
```

You're done! :rocket:

### Verify your work:

Run the command `go run main.go new` and run through the flow in the wizard.
Ensure that you see your provider in the `Select your secret providers` question.

After the `teller.yml` file is created, run the command `go run main.go yaml`, you should see the message :

```sh
FATA[0000] could not load all variables from the given existing providers  error="provider \"Example\" does not implement write yet"
```

This means that you configured the provider successfully and are ready to implement the functions in it.

### Notes

- Since each provider uses some kind of system behind it (e.g. Hashicorp Vault provider connects to the Hashicorp Vault itself) try to wrap the access to the backend or system with your own abstract client-provider with an interface. It will help you to test your provider easier.
- Use provider logger for better visibility when an error occurs.
- Add the new provider to provider mapping in [README.md](../README.md#remapping-provider-variables).

### Adding third-party packages

We `vendor` our dependencies and push them to the repo. This creates an immutable, independent build, that's also free from risks of fetching unknown code in CI/release time.

After adding your packages to import in your provider file, run the commands:

```sh
$ go mod tidy
$ go mod vendor
```

## Adding tests

Create an `example_test.go` file in [pkg/providers](../pkg/providers) folder.

In case you warp the client-provider with an interface you can run a mock generator with the [mock](https://github.com/golang/mock) framework and add this command to the [Makefile](../Makefile)

```sh
mockgen -source pkg/providers/example.go -destination pkg/providers/mock_providers/example_mock.go
```

Test guidelines:

- Create a `TestExample` function and call [AssertProvider](../pkg/providers/helpers_test.go) for testing main functionality.
- Create a `TestExampleFailures` for testing error handling.
- You can also add more tests for testing private functions.
- Run `make lint` to validate linting.
- Run `make test` for make sure that all the test pass.

```

```
