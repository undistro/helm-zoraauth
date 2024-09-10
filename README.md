# Helm zoraauth Plugin

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/undistro/helm-zoraauth)](https://goreportcard.com/report/github.com/undistro/helm-zoraauth)
[![Release](https://img.shields.io/github/release/undistro/helm-zoraauth.svg?style=flat-square)](https://github.com/undistro/helm-zoraauth/releases/latest)
[![Build Status](https://github.com/undistro/helm-zoraauth/workflows/build-test/badge.svg)](https://github.com/undistro/helm-zoraauth/actions?workflow=build-test)

`zoraauth` is a Helm v3 plugin for handling the [OAuth2.0 Device Authorization Grant](https://oauth.net/2/device-flow/) process, creating a values file for use during `helm install`.

## Install

Based on the version in `plugin.yaml`, release binary will be downloaded from GitHub:

```console
$ helm plugin install https://github.com/undistro/helm-zoraauth
Downloading and installing helm-zoraauth v0.1.0 ...
https://github.com/undistro/helm-zoraauth/releases/download/v0.1.0/helm-zoraauth_0.1.0_darwin_amd64.tar.gz
Installed plugin: zoraauth
```

### For Windows (using WSL)

Helm's plugin install hook system relies on `/bin/sh`, regardless of the operating system present. Windows users can work around this by using Helm under [WSL](https://docs.microsoft.com/en-us/windows/wsl/install-win10).
```
$ wget https://get.helm.sh/helm-v3.0.0-linux-amd64.tar.gz
$ tar xzf helm-v3.0.0-linux-amd64.tar.gz
$ ./linux-amd64/helm plugin install https://github.com/undistro/helm-zoraauth
```

## Usage

Handle the OAuth 2.0 Device Authorization Grant process, creating a values.yaml file containing details of the `access_token`, `refresh_token` and `token_type`.

```console
$ helm zoraauth [flags]

Flags:
      --audience string    OAuth audience
      --client-id string   OAuth client ID
      --domain string      OAuth domain (e.g. Auth0 domain)
      --output string      Output file for tokens in YAML format (default "tokens.yaml")
```

Example Output:

```console
$ helm zoraauth --audience "<audience>"  --client-id="<client id>" --domain="<domain>"

Initiating Device Authorization Flow...
Please visit https://<domain>/activate and enter code: ABCD-EFGH, or visit: https://<domain>/activate?user_code=ABCD-EFGH
Tokens saved to tokens.yaml
```

The output file will take the form

```yaml
zoraauth:
  access_token: <access token>
  refresh_token: <refresh token>
  token_type: <token type>
```

## Developer (From Source) Install

If you would like to handle the build yourself, this is the recommended way to do it.

You must first have [Go v1.22+](http://golang.org) installed, and then you run:

```console
$ mkdir -p ${GOPATH}/src/github.com
$ cd $_
$ git clone git@github.com:undistro/helm-zoraauth.git
$ cd helm-zoraauth
$ make
$ export HELM_LINTER_PLUGIN_NO_INSTALL_HOOK=true
$ helm plugin install <your_path>/helm-zoraauth
```

That last command will use the binary that you built.

## Notes

The structure of this repository is based on the [helm-mapkubeapis](https://github.com/helm/helm-mapkubeapis) repository.
