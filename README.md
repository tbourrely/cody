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
```