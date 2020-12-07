# onboarding

## About this plugin
This plugin aims to quickly starts with Artifactory by creating the right repository structure and permissions

## Installation with JFrog CLI
Installing the latest version:

`$ jfrog plugin install onboarding`

Installing a specific version:

`$ jfrog plugin install onboarding@version`

Uninstalling a plugin

`$ jfrog plugin uninstall onboarding`

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
