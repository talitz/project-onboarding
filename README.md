# onboarding

## About this plugin
This plugin aims to quickly starts with Artifactory by creating the right repository structure and permissions

## Installation with JFrog CLI
Since this plugin is currently not included in [JFrog CLI Plugins Registry](https://github.com/jfrog/jfrog-cli-plugins-reg), it needs to be built and installed manually. Follow these steps to install and use this plugin with JFrog CLI.
1. Make sure JFrog CLI is installed on you machine by running ```jfrog```. If it is not installed, [install](https://jfrog.com/getcli/) it.
2. Create a directory named ```plugins``` under ```~/.jfrog/``` if it does not exist already.
3. Clone this repository.
4. CD into the root directory of the cloned project.
5. Run ```go build``` to create the binary in the current directory.
6. Copy the binary into the ```~/.jfrog/plugins``` directory.

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
