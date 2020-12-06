# onboarding

## About this plugin
This plugin aims to quickly starts with Artifactory by creating the right repository structure and permissions

## Installation with JFrog CLI
Installing the latest version:

`$ jfrog plugin install on-boarding`

Installing a specific version:

`$ jfrog plugin install on-boarding@version`

Uninstalling a plugin

`$ jfrog plugin uninstall on-boarding`

## Usage
### Commands
* create
    - Arguments:
        - pathToYamlFile - location of the config YAML file
    - Flags:
        - dry-run: Only outputs command to be executed **[Default: true]**
    - Example:
    ```
  $ jfrog onboarding create myconfig.yaml --dry-run=false
  ```

## Additional info
None.

## Release Notes
The release notes are available [here](RELEASE.md).
