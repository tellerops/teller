<p align="center">
<br/>
<br/>
<br/>
   <img src="media/teller-logo.png" width="288"/>
<br/>
<br/>
</p>

<p align="center">
<b>:computer: Never leave your terminal for secrets</b>
<br/>
<b>:pager: Create easy and clean workflows for working with cloud environments</b>
<br/>
<b>:mag_right: Scan for secrets and fight secret sprawl</b>
<hr/>
</p>

<p align="center">
<img src="https://github.com/tellerops/teller/actions/workflows/build.yml/badge.svg"/>

</p>

# Teller - the open-source universal secret manager for developers

Never leave your terminal to use secrets while developing, testing, and building your apps.

Instead of custom scripts, tokens in your `.zshrc` files, visible `EXPORT`s in your bash history, misplaced `.env.production` files and more around your workstation -- just use `teller` and connect it to any vault, key store, or cloud service you like (Teller support Hashicorp Vault, AWS Secrets Manager, Google Secret Manager, and many more).

You can use Teller to tidy your own environment or for your team as a process and best practice.

![](media/providers.png)

## Quick Start with `teller`

**Download a binary**
Grab a binary from [releases](https://github.com/tellerops/teller/releases)

**Build from source**
Using this method will allow you to eye-ball the source code, review it, and build a copy yourself.

This will install the binary locally on your machine:

```bash
$ cd teller-cli
$ cargo install --path .
```

**Create a new configuration**

```
$ teller new
? Select your secret providers ›
⬚ hashicorp_consul
⬚ aws_secretsmanager
⬚ ssm
⬚ dotenv
⬚ hashicorp
⬚ google_secretmanager
```

Then, edit the newly created `.teller.yml` to set the maps and keys that you need for your providers.

## A look at `teller.yml`
The teller YAML describes your providers and within each provider a `map` that describes:

* What is the root path to fetch key-values from
* For each such map, its unique `id` which will serve you for operations later
* For each map, an optional specific key name mapping - you can rename keys that you will fetch from the source provider

Here's an example configuration file. Note that it also include templating constructs -- such as fetching environment variables while loading the configuration:

```yaml
providers:
  hashi_1:
    kind: hashicorp
    maps:
      - id: test-load
        path: /{{ get_env(name="TEST_LOAD_1", default="test") }}/users/user1
        # if empty, map everything
        # == means map to same key name
        # otherwise key on left becomes right
        # in the future: key_transform: camelize, snake_case for automapping the keys
        keys:
          GITHUB_TOKEN: ==
          mg: FOO_BAR
  dot_1:
    kind: dotenv
    maps:
      - id: stg
        path: VAR_{{ get_env(name="STAGE", default="development") }}

```

You can now address these providers as `hashi_1` or `dot_1`. Teller pulls the specified data from all providers by default.


# Features

## :running: Running subprocesses

Manually exporting and setting up environment variables for running a process with demo-like / production-like set up?

Got bitten by using `.env.production` and exposing it in the local project itself?

Using `teller` and a `.teller.yml` file that exposes nothing to the prying eyes, you can work fluently and seamlessly with zero risk, also no need for quotes:

```
$ teller run --reset --shell -- node index.js
```

## :mag_right: Inspecting variables

This will output the current variables `teller` picks up. Only first 2 letters will be shown from each, of course.

```
$ teller show
```

## :tv: Local shell population

Hardcoding secrets into your shell scripts and dotfiles?

In some cases it makes sense to eval variables into your current shell. For example in your `.zshrc` it makes much more sense to use `teller`, and not hardcode all those into the `.zshrc` file itself.

In this case, this is what you should add:

```
eval "$(teller sh)"
```

## :whale: Easy Docker environment

Tired of grabbing all kinds of variables, setting those up, and worried about these appearing in your shell history as well?

Use this one liner from now on:

```
$ docker run --rm -it --env-file <(teller env) alpine sh
```

## :warning: Scan for secrets

Teller can help you fight secret sprawl and hard coded secrets, as well as be the best productivity tool for working with your vault.

It can also integrate into your CI and serve as a shift-left security tool for your DevSecOps pipeline.

Look for your vault-kept secrets in your code by running:

```bash
$ teller scan
```

You can run it as a linter in your CI like so:

```yaml
run: teller scan --error-if-found
```

It will break your build if it finds something (returns exit code `1`).

You can also export results as JSON with `--json` and scan binary files with `-b`.

## :recycle: Redact secrets from process outputs, logs, and files

You can use `teller` as a redaction tool across your infrastructure, and run processes while redacting their output as well as clean up logs and live tails of logs.

Pipe any process output, tail or logs into teller to redact those, live:

```
$ cat some.log | teller redact
```

It should also work with `tail -f`:

```
$ tail -f /var/log/apache.log | teller redact
```

Finally, if you've got some files you want to redact, you can do that too:

```bash
$ teller redact --in dirty.csv --out clean.csv
```

If you omit `--in` Teller will take `stdin`, and if you omit `--out` Teller will output to `stdout`.


## :scroll: Populate templates

You can populate custom templates:

```bash
$ teller template --in config-templ.t
```

Template format is [Tera](https://keats.github.io/tera) which is very similar to liquid or handlebars.

Here is an example template:

```yaml
production_var: {{ key(name="PRINT_NAME")}}
production_mood: {{ key(name="PRINT_MOOD")}}
```

## :arrows_counterclockwise: Copy/sync data between providers

In cases where you want to sync between providers, you can do that with `teller copy`.

**Specific mapping key sync**

You can use the `<provider name>/<map id>` format to copy a mapping from a provider to another provider:

```bash
$ teller copy --from source/dev --to target/prod,<...>
```

In this simplistic example, we use the following configuration file

```yaml
providers:
  dot1:
    kind: dotenv
    maps:
      - id: one
        path: one.env
  dot2:
    kind: dotenv
    maps:
      - id: two
        path: two.env
```

This will:

1. Grab all mapped values from source mapping
2. For each target provider, find the matching mapping, and copy the values from source into it


By default copying will **update** target mapping (upsert data), if you want to replace you can use `--replace`.

## :bike: Write and multi-write to providers

Teller providers supporting _write_ use cases which allow writing values _into_ providers.

Remember, for this feature it still revolves around definitions in your `teller.yml` file:

```bash
$ teller put --providers new --map-id one NEW_VAR=s33kret
```

In this example, this configuration is being used:

```yaml
providers:
  new:
    kind: dotenv
    maps:
      - id: one
        path: new.env
```

A few notes:

- Values are key-value pair in the format: `key=value` and you can specify multiple pairs at once
- When you're specifying a literal sensitive value, make sure to use an ENV variable so that nothing sensitive is recorded in your history
- The flag `--providers` lets you push to one or more providers at once

## :x: Delete and multi-delete from providers

Teller providers support _deleting_ values _from_ providers.

```bash
$ teller delete --providers new --map-id one DELETE_ME
```

A few notes:

- You can specify multiple keys to delete, for example:
- The flag `--providers` lets you push to one or more providers at once


## `YAML` Export in YAML format

XXX TODO: rewrite how the command export works

You can export in a YAML format, suitable for [GCloud](https://cloud.google.com/functions/docs/env-var):

```
$ teller export yaml
```

Example format:

```yaml
FOO: "1"
KEY: VALUE
```

## `JSON` Export in JSON format


You can export in a JSON format, suitable for piping through `jq` or other workflows:

```
$ teller export json
```

Example format:

```json
{
  "FOO": "1"
}
```

# Providers

You can get a list of the providers and their described configuration values [in the documentation](https://docs.rs/teller-providers/latest/teller_providers/providers/index.html).

### Azure Key Vault

Uses Azure Key Vault as a secrets engine.

```yaml
providers:
  azurekv:
    kind: azure_keyvault
    options:
      vault_name: "your-vault-name"
```

**Authentication:** Uses [`DefaultAzureCredential`](https://docs.rs/azure_identity/latest/azure_identity/struct.DefaultAzureCredential.html) which supports various methods like environment variables, managed identity, Azure CLI, etc.

**Options:**

*   `vault_name` (required): The name of your Azure Key Vault instance (e.g., `my-vault`). The URL will be constructed as `https://<vault_name>.vault.azure.net/`.

### Dotenv

Uses `.env` files as a secrets engine.

```yaml
providers:
  dotenv:
    kind: dotenv
    maps:
      - id: development
        path: .env
```

### Google Secret Manager

Uses Google Secret Manager as a secrets engine.

```yaml
providers:
  google:
    kind: google_secretmanager
    project_id: "your-gcp-project-id"
```

### HashiCorp Consul

Uses HashiCorp Consul as a secrets engine.

```yaml
providers:
  consul:
    kind: hashicorp_consul
    address: "127.0.0.1:8500" # Address of your Consul agent
    scheme: "http" # or "https" if using HTTPS
```

### HashiCorp Vault

Uses HashiCorp Vault as a secrets engine.

```yaml
providers:
  vault:
    kind: hashicorp
    address: "http://127.0.0.1:8200" # Address of your Vault server
    token: "your-vault-token" # Optional, if not using agent or other auth methods
```

### AWS Secrets Manager

Uses AWS Secrets Manager as a secrets engine.

```yaml
providers:
  aws:
    kind: aws_secretsmanager
    region: "us-west-2" # Your AWS region
```

### SSM Parameter Store

Uses AWS SSM Parameter Store as a secrets engine.

```yaml
providers:
  ssm:
    kind: ssm
    region: "us-west-2" # Your AWS region
```

### Testing check list:

* [ ] **docker on windows**: if you have a container based test that uses Docker, make sure to exclude it on Windows using `#[cfg(not(windows))]`

* [ ] **resource semantics**: while building providers, align with the semantics of _empty_ and _not found_ as two different semantics: if a provider supports an explicit "not found" semantic (404, NotFound, etc.), use `Error::NotFound`. Otherwise when a provider signals a "not found" semantic as an empty data bag, return an empty `KV[]` (i.e. do not translate a sematic of "empty" into "not found").

### Testing

Testing is done with:

```
$ cargo test --all --all-features
```

And requires Docker (or equivalent) on your machine.

### Thanks:

To all [Contributors](https://github.com/spectralops/teller/graphs/contributors) - you make this happen, thanks!

### Code of conduct

Teller follows [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md)

# Copyright

Copyright (c) 2024 [@jondot](http://twitter.com/jondot). See [LICENSE](LICENSE.txt) for further details.
