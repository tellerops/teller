package pkg

var TellerFileTemplate = `
project: {{.Project}}
{{- if .Confirm }}
confirm: Are you sure you want to run on {{"{{stage}}"}}?
{{ end }}

# Set this if you want to carry over parent process' environment variables
# carry_env: true 


#
# Variables
#
# Feel free to add options here to be used as a variable throughout
# paths.
#
opts:
  region: env:AWS_REGION    # you can get env variables with the 'env:' prefix, for default values if env not found use comma. Example: env:AWS_REGION,{DEFAULT_VALUE}
  stage: development


#
# Providers
#
providers:
	
{{- if index .ProviderKeys "heroku" }}
  # requires an API key in: HEROKU_API_KEY (you can fetch yours from ~/.netrc)
  heroku:
  # sync a complete environment
    env_sync:
      path: drakula-demo

  # # pick and choose variables
  # env:
  #	  JVM_OPTS:
  #      path: drakula-demo
{{end}}

{{- if index .ProviderKeys "vercel" }}
  # requires an API token in: VERCEL_TOKEN 
  vercel:
	# sync a complete environment
    env_sync:
      path: drakula-demo

	# # pick and choose variables
	# env:
	#	  JVM_OPTS:
	#      path: drakula-demo
{{end}}

{{- if index .ProviderKeys "hashicorp_vault" }}
  # configure only from environment
  # https://github.com/hashicorp/vault/blob/api/v1.0.4/api/client.go#L28
  # this vars should not go through to the executing cmd
  hashicorp_vault:
    env_sync:
      path: secret/data/{{"{{stage}}"}}/billing/web/env
    env:
      SMTP_PASS:
        path: secret/data/{{"{{stage}}"}}/wordpress
        field: smtp
{{end}}

{{- if index .ProviderKeys "aws_secretsmanager" }}
  # configure only from environment
  # filter secret versioning by adding comma separating in path value (path: prod/foo/bar,<VERSION>).
  aws_secretsmanager:
    env_sync:
      path: prod/foo/bar
    env:
      FOO_BAR:
        path: prod/foo/bar
        field: SOME_KEY
{{end}}

{{- if index .ProviderKeys "aws_ssm" }}
  # configure only from environment
  aws_ssm:
    env:
      FOO_BAR:
        path: /prod/foobar
        decrypt: true
{{end}}

{{- if index .ProviderKeys "google_secretmanager" }}
  # GOOGLE_APPLICATION_CREDENTIALS=foobar.json
  # https://cloud.google.com/secret-manager/docs/reference/libraries#setting_up_authentication
  google_secretmanager:
    env:
      FOO_GOOG:
        # need to supply the relevant version (versions/1)
        path: projects/123/secrets/FOO_GOOG/versions/1
{{end}}

{{- if index .ProviderKeys "etcd" }}
  # Configure via environment:
  # ETCDCTL_ENDPOINTS
  # tls:
  # ETCDCTL_CA_FILE
  # ETCDCTL_CERT_FILE
  # ETCDCTL_KEY_FILE
  etcd:
    env_sync:
      path: /prod/foo
    env:
      ETC_DSN:
        path: /prod/foo/bar
{{end}}

{{- if index .ProviderKeys "consul" }}
  # Configure via environment:
  # CONSUL_HTTP_ADDR
  consul:
    env_sync:
      path: redis/config
    env:
      ETC_DSN:
        path: redis/config/foobar
{{end}}

{{- if index .ProviderKeys "dotenv" }}
  # you can mix and match many files
  dotenv:
    env_sync:
      path: ~/my-dot-env.env
    env:
      FOO_BAR:
        path: ~/my-dot-env.env
{{end}}

{{- if index .ProviderKeys "azure_keyvault" }}
  # you can mix and match many files
  azure_keyvault:
    env_sync:
      path: azure
    env:
      FOO_BAR:
        path: foobar
{{end}}

{{- if index .ProviderKeys "doppler" }}
  # set your doppler project with "doppler configure set project <my-project>"
  doppler:
    env_sync:
      path: prd
    env:
      FOO_BAR:
        path: prd
{{end}}

{{- if index .ProviderKeys "cyberark_conjur" }}
  # https://conjur.org
  # set CONJUR_AUTHN_LOGIN and CONJUR_AUTHN_API_KEY env vars
  # set .conjurrc file in user's home directory
  cyberark_conjur:
    env:
      FOO_BAR:
        path: /secrets/foo/bar
{{end}}

{{- if index .ProviderKeys "1password" }}
  # Configure via environment variables:
  # OP_CONNECT_HOST
  # OP_CONNECT_TOKEN
  1password:
    env_sync:
        path: # Key title
        source: # 1Password token gen include access to multiple vault. to get the secrets you must add and vaultUUID. the field is mandatory
    env:
      FOO_BAR:
        path: # Key title
        source: # 1Password token gen include access to multiple vault. to get the secrets you must add and vaultUUID. the field is mandatory
        field: # The secret field to get. notesPlain, {label key}, password etc.
{{end}}

{{- if index .ProviderKeys "gopass" }}
  # Override default configuration: https://github.com/gopasspw/gopass/blob/master/docs/config.md
  gopass:
    env_sync:
      path: foo
    env:
      ETC_DSN:
        path: foo/bar
{{end}}

{{- if index .ProviderKeys "lastpass" }}
  # Configure via environment variables:
  # LASTPASS_USERNAME
  # LASTPASS_PASSWORD

  lastpass:
    env_sync:
      path: # LastPass item ID
    env:
      ETC_DSN:
        path: # Lastpass item ID
        # field: by default taking password property. in case you want other property un-mark this line and set the lastpass property name.
{{end}}

{{- if index .ProviderKeys "cloudflare_workers_secrets" }}

  # Configure via environment variables for integration:
  # CLOUDFLARE_API_KEY: Your Cloudflare api key.
  # CLOUDFLARE_API_EMAIL: Your email associated with the api key.
  # CLOUDFLARE_ACCOUNT_ID: Your account ID.

  cloudflare_workers_secrets:
    env_sync:
      source: # Mandatory: script field
    env:
      script-value:
        path: foo-secret
        source: # Mandatory: script field
{{end}}

{{- if index .ProviderKeys "github" }}

  # Configure via environment variables for integration:
  # GITHUB_AUTH_TOKEN: GitHub token

  github:
    env_sync:
       path: owner/github-repo
    env:
      script-value:
        path: owner/github-repo

{{end}}

{{- if index .ProviderKeys "keypass" }}

  # Configure via environment variables for integration:
  # KEYPASS_PASSWORD: KeyPass password
  # KEYPASS_DB_PATH: Path to DB file

  keypass:
    env_sync:
      path: redis/config
      # source: Optional, all fields is the default. Supported fields: Notes, Title, Password, URL, UserName
    env:
      ETC_DSN:
        path: redis/config/foobar
        # source: Optional, Password is the default. Supported fields: Notes, Title, Password, URL, UserName

{{end}}

{{- if index .ProviderKeys "filesystem" }}

  filesystem:
    env_sync:
      path: redis/config
    env:
      ETC_DSN:
        path: redis/config/foobar

{{end}}

{{- if index .ProviderKeys "process_env" }}

  process_env:
    env:
      ETC_DSN:
        field: SOME_KEY # Optional, accesses the environment variable SOME_KEY and maps it to ETC_DSN

{{end}}

{{- if index .ProviderKeys "ansible_vault" }}

  # Configure via environment variables for integration:
  # ANSIBLE_VAULT_PASSPHRASE: Ansible Vault Password

  ansible_vault:
    env_sync:
       path: ansible/vars/vault_{{stage}}.yml

    env:
      KEY1:
        path: ansible/vars/vault_{{stage}}.yml
      NONEXIST_KEY:
        path: ansible/vars/vault_{{stage}}.yml

{{end}}

{{- if index .ProviderKeys "keeper_secretsmanager" }}

  # requires a configuration in: KSM_CONFIG=base64_config or file path KSM_CONFIG_FILE=ksm_config.json
  keeper_secretsmanager:
    env_sync:
      path: RECORD_UID
      # all non-empty fields are mapped by their labels, if empty then by field type, and index 1,2,...,N

    env:
      USER:
        path: RECORD_UID/field/login
        # use Keeper Notation to select individual field values
        # https://docs.keeper.io/secrets-manager/secrets-manager/about/keeper-notation

{{end}}
`
