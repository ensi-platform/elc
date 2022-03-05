#  ELC - orchestrator of development environments

[![Test](https://github.com/MadridianFox/ensi-local-ctl/actions/workflows/test.yml/badge.svg)](https://github.com/MadridianFox/ensi-local-ctl/actions/workflows/test.yml)

With ELC you can:
* start a couple of docker-compose projects with one command
* define dependencies across docker-compose projects
* use one docker-compose template for several services
* describe sets of services for different cases (testing, development, monitoring)
* use containerized development tools

## Installation

```bash
curl -sSL https://raw.githubusercontent.com/MadridianFox/ensi-local-ctl/master/get.sh | sudo bash

elc --help
```

## Build from source

Dependencies:
- go
- make

```bash
git clone git@github.com:MadridianFox/ensi-local-ctl.git
cd ensi-local-ctl

make
make install
```

## How to use

[Full documentation (ru)](https://greensight.atlassian.net/wiki/spaces/ENSI/pages/540246017/ELC)

Make a workspace config file, which contains:

**global variables**
```yaml
name: ensi
variables:
  NETWORK: ensi
  BASE_DOMAIN: ensi.127.0.0.1.nip.io
```
**docker compose templates**
```yaml
templates:
  - name: php80
    path: ${WORKSPACE_PATH}/templates/php8
    compose_file: ${TPL_PATH}/docker-compose.yml
    variables:
      BASE_IMAGE: php:8.0-fpm-alpine
      APP_IMAGE: php80:latest
      NGINX_IMAGE: nginx:1.19-alpine
```

**service definitions**
```yaml
services:
  - name: api
    extends: php80
    path: ${WORKSPACE_PATH}/apps/api
    variables:
      VAR1: ${VAR1:-default}
    dependencies:
      proxy:    [default]
      database: [default, hook]
```

Register workspace in elc:
```bash
$ elc workspace add ensi /path/to/workspace/
```

Start some services:

```bash
$ elc start api
```

Invoke some tool

```bash
$ cd /path/to/service/directory
$ elc composer install
```

## License

Copyright © 2022 Ivan Koryukov

Distributed under the MIT License. See [LICENSE.md](LICENSE.md).