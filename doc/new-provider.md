# How to add a new provider

Adding a new Teller provider is very easy, but you still need to know where to start. we summarize the steps very shortly to make your life easier  


## Provider implementation

1. Copy the file [example.go](pkg/providers/example.go) from path [pkg/providers/example.go](../pkg/providers/example.go) and make sure to implement all the required behaviors.
The [example.go](../pkg/providers/example.go) file is a skeleton for adding new provider, it contains stubs for an interface which declares the required functionality that any provider must have.
2. Go to [pkg/providers.go](../pkg/providers.go) and add your provider name to `ProviderHumanToMachine` hashMap (this function maps between provider display name and internal name).
```go
func (p *BuiltinProviders) ProviderHumanToMachine() map[string]string {
    return map[string]string{
        "Heroku":  "heroku",
        ...
        "Example": "example"
    }
}
```
3. Add your provider to switch case providers in [pkg/providers.go](../pkg/providers.go) file (this function returns the provider handler by provider key).
```go
func (p *BuiltinProviders) GetProvider(name string) (core.Provider, error) {
    logger := logging.GetRoot().WithField("provider_name", name)
    switch name {
    case "heroku":
        return providers.NewHeroku(logger)
    case "example":
        return providers.NewExample(logger)
    default:
        return nil, fmt.Errorf("provider '%s' does not exist", name)
    }
}
```
4. Add provider template configuration in path: [pkg/wizard_template.go](../pkg/wizard_template.go)
```go
{{- if index .ProviderKeys "example" }}
  # Add here some authentication required, like list of environment variables
  example:
    env_sync:
       path: redis/config
    env:
      ETC_DSN:
        path: redis/config/foobar

{{end}}
```

You done! :rocket:

### Verify the steps:
run the command `go run main.go new` and flow the wizard. 
Ensure that you see your provider in the `Select your secret providers` question.

After `.teller` file is created, run the command `go run main.go yaml`, you should see the message :
```sh
FATA[0000] could not load all variables from the given existing providers  error="provider \"Example\" does not implement write yet"
```
This means that you configure the provider successfully and ready to implement the function.

### Notes
* Try to wrap the client-provider with an interface. it will help you to test your provider easier.
* Use provider logger for better visibility when an error occurs.
* Add the new provider to provider mapping in [README.md](../README.md#remapping-provider-variables).


### Add third-party packages
We `vendor` our dependencies and push them to the repo. This creates an immutable, independent build, that's also free from risks of fetching unknown code in CI/release time.

After adding your packages to import in your provider file, run the command:
```sh
$ go mod tidy
$ go mod vendor
```

## Add tests 

Create a `example_test.go` file in [pkg/providers](../pkg/providers) folder.

In case you warp the client-provider with an interface you needs to run a moc generator with [mock](https://github.com/golang/mock) framework and add ths command to [Makefile](../Makefile)
```sh
mockgen -source pkg/providers/example.go -destination pkg/providers/mock_providers/example_mock.go
```

Tests:
* Create `TestExample` function and call [AssertProvider](../pkg/providers/helpers_test.go) for testing main functionality.
* Create `TestExampleFailures` for testing errors handling.
* You can also add more tests for testing private functions.
* Run `make lint` to validate linting.
* Run `make test` for make sure that all the test pass.