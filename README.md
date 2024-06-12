# GitHub Actions for Okteto

## Automate your development workflows using Github Actions and Okteto

GitHub Actions gives you the flexibility to build automated software development workflows. With GitHub Actions for Okteto you can create workflows to build, deploy and update your applications in [Okteto](https://okteto.com).
Follow [this tutorial](https://okteto.com/docs/cloud/preview-environments/preview-environments-github/) for a full preview environment configuration sample.

Try Okteto for free for 30 days, no credit card required. [Start your 30-day trial now](https://www.okteto.com/free-trial/)!

## Github Action for Creating a Preview environment in Okteto

You can use this action to create a preview environment in Okteto as part of your automated development workflow.

## Inputs

### `name`

**Required**  The name of the Okteto preview environment to create.

### `timeout`

The length of time to wait for completion. Values should contain a corresponding time unit e.g. 1s, 2m, 3h. If not specified it will use `5m`.

### `scope`

The scope of the Okteto preview environment to create. Allowed values are `personal` or `global` (defaults to `global`).

### `variables`

A list of variables to be used by the pipeline. If several variables are present, they should be separated by commas e.g. VAR1=VAL1,VAR2=VAL2,VAR3=VAL3.

### `file`

Relative path within the repository to the manifest file (default to okteto-pipeline.yaml or .okteto/okteto-pipeline.yaml).

### `branch`

The branch to use for the preview environment (defaults to the branch that triggered the action).

### `log-level`

Log level used. Supported values are: `debug`, `info`, `warn`, `error`. (defaults to `warn`)

## Environment Variables

If the `GITHUB_TOKEN` environment variable is set, the action will share the URL of the preview environment with the pull request that triggered the action.

## Example usage

This example runs the context action and then creates a preview environment.

```yaml
# File: .github/workflows/workflow.yml
on: [push]

name: example

jobs:

  devflow:
    runs-on: ubuntu-latest
    steps:
    - name: Context
      uses: okteto/context@latest
      with:
        url: https://okteto.example.com
        token: ${{ secrets.OKTETO_TOKEN }}

    - name: "Deploy the preview environment"
      uses: okteto/deploy-preview@latest
      with:
        name: dev-previews
        scope: global
```

## Advanced usage

 ### Custom Certification Authorities or Self-signed certificates

 You can specify a custom certificate authority or a self-signed certificate by setting the `OKTETO_CA_CERT` environment variable. When this variable is set, the action will install the certificate in the container, and then execute the action.

 Use this option if you're using a private Certificate Authority or a self-signed certificate in your [Okteto SH](https://www.okteto.com/docs/self-hosted/) instance.  We recommend that you store the certificate as an [encrypted secret](https://docs.github.com/en/actions/reference/encrypted-secrets), and that you define the environment variable for the entire job, instead of doing it on every step.


 ```yaml
 # File: .github/workflows/workflow.yml
 on: [push]

 name: example

 jobs:
   devflow:
     runs-on: ubuntu-latest
     env:
       OKTETO_CA_CERT: ${{ secrets.OKTETO_CA_CERT }}

     steps:
     - name: Context
       uses: okteto/context@latest
       with:
         url: https://okteto.example.com
         token: ${{ secrets.OKTETO_TOKEN }}

     - name: "Deploy the preview environment"
       uses: okteto/deploy-preview@latest
       with:
         name: dev-previews
         scope: global
         timeout: 15m
 ```
