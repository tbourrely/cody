# Cody

Manage vscode-in-browser instances.

# Requirements

* Docker

# Usage

```bash
# start an instance, the used name will be the directory name
cody start

# stop an instance
cody stop <instance name>

# show instance url
cody url <instance name>

# list active instances
cody list
```

# Configuration

The tool will look for **cody.yml** files, under the home directory & current working directory.

Sample configuration : 

```yaml
# Expose the container on a port between 1 and 10 (chosen randomly)
ports:
    start: 1
    end: 10
auth_token: "vscodeauthtoken" # pay attention to only use letters / numbers 
extensions:
  - vscodevim.vim
  - redhat.ansible
  - ...
```

# Dependencies

This tool is based on [gitpod openvscode-server](https://github.com/gitpod-io/openvscode-server/) implementation.
