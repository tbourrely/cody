# Cody

Manage vscode-in-browser instances.

# Configuration

The tool will look for **cody.yml** files, under the home directory & current working directory.

Sample configuration : 

```yaml
# Expose the container on a port between 1 and 10 (chosen randomly)
ports:
    start: 1
    end: 10
auth_token: "vscodeauthtoken" # pay attention to only use letters / numbers 
```

# Dependencies

This tool is based on [gitpod openvscode-server](https://github.com/gitpod-io/openvscode-server/) implementation.