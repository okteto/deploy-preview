# GitHub Actions for Okteto Cloud

## Automate your development workflows using Github Actions and Okteto Cloud

GitHub Actions gives you the flexibility to build an automated software development workflows. With GitHub Actions for Okteto Cloud you can create workflows to build, deploy and update your applications in [Okteto Cloud](https://cloud.okteto.com). Follow [this tutorial](https://okteto.com/docs/cloud/preview-environments/preview-environments-github/) for a full preview environment configuration sample.

Get started today with a [free Okteto Cloud account](https://cloud.okteto.com)!

## Github Action for Creating a Preview environment in Okteto Cloud

You can use this action to create a preview environment in Okteto Cloud as part of your automated development workflow.

## Inputs

### `name`

**Required**  The name of the Okteto preview environment to create.

> Remember that the preview environment name must have your github ID as a suffix.

### `timeout`

The length of time to wait for completion. Values should contain a corresponding time unit e.g. 1s, 2m, 3h. If not specified it will use `5m`.

### `scope`

The scope of the Okteto preview environment to create.

> Available scopes are `personal` and `global` (defaults to `personal`). To create a preview environment with [global scope](https://okteto.com/docs/cloud/preview-environments/preview-environments-github/#preview-environments-for-okteto-enterprise-users) it is necessary to have administrator permissions. Global preview environments are accessible by all cluster members.

### `variables`

A list of variables to be used by the pipeline. If several variables are present, they should be separated by commas e.g. VAR1=VAL1,VAR2=VAL2,VAR3=VAL3.

### `filename`

Relative path within the repository to the manifest file (default to okteto-pipeline.yaml or .okteto/okteto-pipeline.yaml).

## Environment Variables

If the `GITHUB_TOKEN` environment variable is set, the action will share the URL of the preview environment with the pull request that triggered the action.

## Example usage

This example runs the login action and then creates a preview environment.

```yaml
# File: .github/workflows/workflow.yml
on: [push]

name: example

jobs:

  devflow:
    runs-on: ubuntu-latest
    steps:
    - uses: okteto/login@latest
      with:
        token: ${{ secrets.OKTETO_TOKEN }}

    - name: "Deploy the preview environment"
      uses: okteto/deploy-preview@latest
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        name: dev-previews-cindylopez
```

## Advanced usage

 ### Custom Certification Authorities or Self-signed certificates

 You can specify a custom certificate authority or a self-signed certificate by setting the `OKTETO_CA_CERT` environment variable. When this variable is set, the action will install the certificate in the container, and then execute the action.

 Use this option if you're using a private Certificate Authority or a self-signed certificate in your [Okteto Enterprise](http://okteto.com/enterprise) instance.  We recommend that you store the certificate as an [encrypted secret](https://docs.github.com/en/actions/reference/encrypted-secrets), and that you define the environment variable for the entire job, instead of doing it on every step.


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
     - name: "Deploy the preview environment"
       uses: okteto/deploy-preview@latest
       env:
         OKTETO_URL: https://cloud.okteto.com
         OKTETO_TOKEN: ${{ secrets.OKTETO_TOKEN }}
         GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
       with:
         name: dev-previews-cindylopez
         scope: global
         timeout: 15m
 ```
