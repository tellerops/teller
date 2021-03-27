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
  region: env:AWS_REGION    # you can get env variables with the 'env:' prefix
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
`
