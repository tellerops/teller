# Teller Roadmap

This document lists the new features being considered for the future and that are being experimented / worked on currently. We would like our contributors and users to know what features could come in the near future. We definitely welcome suggestions and ideas from everyone about the roadmap and features (feel free to open an [issue](https://github.com/SpectralOps/teller/issues)). 

## Currently worked on

### Additional use cases

* Secret value policy - allow for validating fetched secrets against a policy such as secret strength, raising a warning if policy is not met
* Secret enclave. Create a local secret engine to be stored co-located with code. This will become _just another provider_ to pick from 
* Zero-trust / last-mile encryption. Have Teller perform the last-mile encryption at the point of fetching secrets allowing for zero-trust secret management. (this may share implementation details with the secret enclave)


## Planned Features

### Additional use cases

* Push-only providers, where only write is supported. Typically CI providers only allow for secrets to be written-in but not read-from. We want to allow for a "push" use case where users could sync secrets from vaults into operational platforms such as CI and production. That means creating the concept of _write only providers_.
* Kubernetes secrets sidecar. Enable a seamless way to have secrets fetched just-in-time for processes as a sidecar.
* Native Jenkins plugin as secrets engine. Have Jenkins point to a secret store that's actually backed by Teller - which delegates to your favorite vaults.

### Additional Providers

**Read/Write**

* 1password  (via local tooling)
* Lastpass  (via local tooling)
* gopass



**Write only**

* Github
* Gitlab
* CircleCI
* Travis


