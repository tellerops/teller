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
<br/>
<hr/>
</p>

<p align="center">
<img src="https://github.com/spectralops/teller/actions/workflows/ci.yml/badge.svg"/>

</p>

# Teller - the open-source universal secret manager for developers


Never leave your terminal to use secrets while developing, testing, and building your apps.

Instead of custom scripts, tokens in your `.zshrc` files, visible `EXPORT`s in your bash history, misplaced `.env.production` files and more around your workstation -- just use `teller` and connect it to any vault, key store, or cloud service you like (Teller support Hashicorp Vault, AWS Secrets Manager, Google Secret Manager, and many more).

You can use Teller to tidy your own environment or for your team as a process and best practice.

![](media/providers.png)



## Quick Start with `teller` (or `tlr`)

You can install `teller` with homebrew:

```
$ brew tap spectralops/tap && brew install teller
````

You can now use `teller` or `tlr` (if you like shortcuts!) in your terminal.



![](media/teller.gif)

`teller` will pull variables from your various cloud providers, vaults and others, and will populate your current working session (in various ways!, see more below) so you can work safely and much more productively.


`teller` needs a tellerfile. This is a `.teller.yml` file that lives in your repo, or one that you point teller to with `teller -c your-conf.yml`.
## Using a Github Action

For those using Github Action, you can have a 1-click experience of installing Teller in your CI:

```yaml
      - name: Setup Teller
        uses: spectralops/setup-teller@v1
      - name: Run a Teller task (show, scan, run, etc.)
        run: teller run [args]
```

For more, check our [setup teller action](https://github.com/marketplace/actions/setup-teller) on the marketplace.


## Create your configuration

Run `teller new` and follow the wizard, pick the providers you like and it will generate a `.teller.yml` for you.


Alternatively, you can use the following minimal template or [view a full example](.teller.example.yml):

```yaml
project: project_name
opts:
  stage: development

# remove if you don't like the prompt
confirm: Are you sure you want to run in {{stage}}?

providers:
  # uses environment vars to configure
  # https://github.com/hashicorp/vault/blob/api/v1.0.4/api/client.go#L28
  hashicorp_vault:
    env_sync:
      path: secret/data/{{stage}}/services/billing

  # this will fuse vars with the below .env file
  # use if you'd like to grab secrets from outside of the project tree
  dotenv:
    env_sync:
      path: ~/billing.env.{{stage}}
```

Now you can just run processes with:

```
$ teller run node src/server.js
Service is up.
Loaded configuration: Mailgun, SMTP
Port: 5050
```

Behind the scenes: `teller` fetched the correct variables, placed those (and _just_ those) in `ENV` for the `node` process to use.

# Features

## :running: Running subprocesses

Manually exporting and setting up environment variables for running a process with demo-like / production-like set up?

Got bitten by using `.env.production` and exposing it in the local project itself?


Using `teller` and a `.teller.yml` file that exposes nothing to the prying eyes, you can work fluently and seamlessly with zero risk, also no need for quotes:

```
$ teller run -- your-process arg1 arg2... --switch1 ...
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

## :warning:  Scan for secrets

Teller can help you fight secret sprawl and hard coded secrets, as well as be the best productivity tool for working with your vault.

It can also integrate into your CI and serve as a shift-left security tool for your DevSecOps pipeline.

Look for your vault-kept secrets in your code by running:

```
$ teller scan
```

You can run it as a linter in your CI like so:

```
run: teller scan --silent
```

It will break your build if it finds something (returns exit code `1`).


Use Teller for productively and securely running your processes and you get this for free -- nothing to configure. If you have data that you're bringing that you're sure isn't sensitive, flag it in your `teller.yml`:

```
dotenv:
  env:
    FOO:
      path: ~/my-dot-env.env
      severity: none # will skip scanning. possible values: high | medium | low | none
```

By default we treat all entries as sensitive, with value `high`.


## :recycle: Redact secrets from process outputs, logs, and files

You can use `teller` as a redaction tool across your infrastructure, and run processes while redacting their output as well as clean up logs and live tails of logs.

Run a process and redact its output in real time:

```
$ teller run --redact -- your-process arg1 arg2
```

Pipe any process output, tail or logs into teller to redact those, live:

```
$ cat some.log | teller redact
```

It should also work with `tail -f`:

```
$ tail -f /var/log/apache.log | teller redact
```


Finally, if you've got some files you want to redact, you can do that too:

```
$ teller redact --in dirty.csv --out clean.csv
```

If you omit `--in` Teller will take `stdin`, and if you omit `--out` Teller will output to `stdout`.

## :beetle: Detect secrets and value drift

You can detect _secret drift_ by comparing values from different providers against each other. It might be that you want to pin a set of keys in different providers to always be the same value; when they aren't -- that means you have a drift.

For this, you first need to label values as `source` and couple with the appropriate sink as `sink` (use same label for both to couple them). Then, source keys will be compared against other keys in your configuration:

```yaml
providers:
  dotenv:
    env_sync:
      path: ~/my-dot-env.env
      source: s1
  dotenv2:
    kind: dotenv
    env_sync:
      path: ~/other-dot-env.env
      sink: s1
```

And run

```
$ teller drift dotenv dotenv2 -c your-config.yml
```

![](https://user-images.githubusercontent.com/83390/117453797-07512380-af4e-11eb-949e-cc875e854fad.png)



## :scroll: Populate templates

Have a kickstarter project you want to populate quickly with some variables (not secrets though!)?

Have a production project that just _has_ to have a file to read that contains your variables?

You can use `teller` to inject variables into your own templates (based on [go templates](https://golang.org/pkg/text/template/)).

With this template:

```go
Hello, {{.Teller.EnvByKey "FOO_BAR" "default-value" }}!
```

Run:

```
$ teller template my-template.tmpl out.txt
```

Will get you, assuming `FOO_BAR=Spock`:

```
Hello, Spock!
```

## :white_check_mark: Prompts and options

There are a few options that you can use:

* __carry_env__ - carry the environment from the parent process into the child process. By default we isolate the child process from the parent process. (default: _false_)

* __confirm__ - an interactive question to prompt the user before taking action (such as running a process). (default: _empty_)

* __opts__ - a dict for our own variable/setting substitution mechanism. For example:

```
opts:
  region: env:AWS_REGION
  stage: qa
```

And now you can use paths like `/{{stage}}/{{region}}/billing-svc` where ever you want (this templating is available for the __confirm__ question too).

If you prefix a value with `env:` it will get pulled from your current environment.

# Providers


For each provider, there are a few points to understand:

* Sync - full sync support. Can we provide a path to a whole environment and have it synced (all keys, all values). Some of the providers support this and some don't.
* Key format - some of the providers expect a path-like key, some env-var like, and some don't care. We'll specify for each.

## General provider configuration

We use the following general structure to specify sync mapping for all providers:

```yaml
# you can use either `env_sync` or `env` or both
env_sync:
  path: ... # path to mapping
  remap:
    PROVIDER_VAR1: VAR3 # Maps PROVIDER_VAR1 to local env var VAR3
env:
  VAR1:
    path: ... # path to value or mapping
    field: <key> # optional: use if path contains a k/v dict
    decrypt: true | false # optional: use if provider supports encryption at the value side
    severity: high | medium | low | none # optional: used for secret scanning, default is high. 'none' means not a secret 
    redact_with: "**XXX**" # optional: used as a placeholder swapping the secret with it. default is "**REDACTED**"
  VAR2:
    path: ...
```

### Remapping Provider Variables

Providers which support syncing a list of keys and values can be remapped to different environment variable keys. Typically, when teller syncs paths from `env_sync`, the key returned from the provider is directly mapped to the environment variable key. In some cases it might be necessary to have the provider key mapped to a different variable without changing the provider settings. This can be useful when using `env_sync` for [Hashicorp Vault Dynamic Database credentials](https://www.vaultproject.io/docs/secrets/databases):

```yaml
env_sync:
  path: database/roles/my-role
  remap:
    username: PGUSER
    password: PGPASSWORD
```

After remapping, the local environment variable `PGUSER` will contain the provider value for `username` and `PGPASSWORD` will contain the provider value for `password`.

## Hashicorp Vault

### Authentication

If you have the Vault CLI configured and working, there's no special action to take.

Configuration is environment based, as defined by client standard. See variables [here](https://github.com/hashicorp/vault/blob/api/v1.0.4/api/client.go#L28).

### Features

* Sync - `yes`
* Mapping - `yes`
* Key format - path based, usually starts with `secret/data/`, and more generically `[engine name]/data`

### Example Config

```yaml
hashicorp_vault:
  env_sync:
    path: secret/data/demo/billing/web/env
  env:
    SMTP_PASS:
      path: secret/data/demo/wordpress
      field: smtp
```

## Consul

### Authentication

If you have the Consul CLI working and configured, there's no special action to take.

Configuration is environment based, as defined by client standard. See variables [here](https://github.com/hashicorp/consul/blob/master/api/api.go#L28).

### Features

* Sync - `yes`
* Mapping - `yes`
* Key format 
  * `env_sync` - path based, we use the last segment as the variable name
  * `env` - any string, no special requirement

### Example Config

```yaml
consul:
  env_sync:
    path: ops/config
  env:
    SLACK_HOOK:
      path: ops/config/slack
```


## Heroku

### Authentication

Requires an API key populated in your environment in: `HEROKU_API_KEY` (you can fetch it from your ~/.netrc).

### Features

* Sync - `yes`
* Mapping - `yes`
* Key format 
  * `env_sync` - name of your Heroku app
  * `env` - the actual env variable name in your Heroku settings

### Example Config

```yaml
heroku:
  env_sync:
    path: my-app-dev
  env:
    MG_KEY:
      path: my-app-dev
```

## Etcd

### Authentication

If you have `etcdctl` already working there's no special action to take.

We follow how `etcdctl` takes its authentication settings. These environment variables need to be populated

* `ETCDCTL_ENDPOINTS`

For TLS:

* `ETCDCTL_CA_FILE`
* `ETCDCTL_CERT_FILE`
* `ETCDCTL_KEY_FILE`


### Features

* Sync - `yes`
* Mapping - `yes`
* Key format 
  * `env_sync` - path based
  * `env` - path based


### Example Config

```yaml
etcd:
  env_sync:
    path: /prod/billing-svc
  env:
    MG_KEY:
      path: /prod/billing-svc/vars/mg
```

## AWS Secrets Manager

### Authentication

Your standard `AWS_DEFAULT_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` need to be populated in your environment

### Features

* Sync - `yes`
* Mapping - `yes`
* Key format 
  * `env_sync` - path based
  * `env` - path based


### Example Config

```yaml
aws_secretsmanager:
  env_sync:
    path: /prod/billing-svc
  env:
    MG_KEY:
      path: /prod/billing-svc/vars/mg
```

## AWS Paramstore


### Authentication

Your standard `AWS_DEFAULT_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` need to be populated in your environment

### Features

* Sync - `no`
* Mapping - `yes`
* Key format 
  * `env` - path based
  * `decrypt` - available in this provider, will use KMS automatically


### Example Config

```yaml
aws_ssm:
  env:
    FOO_BAR:
      path: /prod/billing-svc/vars
      decrypt: true
```

## Google Secret Manager
### Authentication

You should populate `GOOGLE_APPLICATION_CREDENTIALS=account.json` in your environment to your relevant `account.json` that you get from Google.


### Features

* Sync - `no`
* Mapping - `yes`
* Key format 
  * `env` - path based, needs to include a version
  * `decrypt` - available in this provider, will use KMS automatically


### Example Config

```yaml
google_secretmanager:
  env:
    MG_KEY:
      # need to supply the relevant version (versions/1)
      path: projects/44882/secrets/MG_KEY/versions/1
```

## .ENV (dotenv)

### Authentication

No need. You'll be pointing to a one or more `.env` files on your disk.

### Features

* Sync - `yes`
* Mapping - `yes`
* Key format 
  * `env` - env key like


### Example Config

You can mix and match any number of files, sitting anywhere on your drive.

```yaml
dotenv:
  env_sync:
    path: ~/my-dot-env.env
  env:
    MG_KEY:
      path: ~/my-dot-env.env
```

## Doppler

### Authentication

Install the [doppler cli][dopplercli] then run `doppler login`. You'll also need to configure your desired "project" for any given directory using `doppler configure`. Alternatively, you can set a global project by running `doppler configure set project <my-project>` from your home directory.

### Features

* Sync - `yes`
* Mapping - `yes`
* Key format
  * `env` - env key like

### Example Config

```yaml
doppler:
  env_sync:
    path: prd
  env:
    MG_KEY:
      path: prd
      field: OTHER_MG_KEY # (optional)
```

[dopplercli]: https://docs.doppler.com/docs/cli

# Security Model

## Project Practices

* We `vendor` our dependencies and push them to the repo. This creates an immutable, independent build, that's also free from risks of fetching unknown code in CI/release time.

## Providers

For every provider, we are federating all authentication and authorization concern to the system of origin. In other words, if for example you connect to your organization's Hashicorp Vault, we assume you already have a secure way to do that, which is "blessed" by the organization.

In addition, we don't offer any way to specify connection details to these systems in writing (in configuration files or other), and all connection details, to all providers, should be supplied via environment variables.

That allows us to keep two important points:

1. Don't undermine the user's security model and threat modeling for the sake of productivity (security AND productivity CAN be attained)
2. Don't encourage the user to do what we're here for -- save secrets and sensitive details from being forgotten in various places.




### Thanks:

To all [Contributors](https://github.com/spectralops/teller/graphs/contributors) - you make this happen, thanks!


# Copyright

Copyright (c) 2021 [@jondot](http://twitter.com/jondot). See [LICENSE](LICENSE.txt) for further details.
