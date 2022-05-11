#  ELC - Ensi Local Ctl

[![Test](https://github.com/ensi-platform/elc/actions/workflows/test.yml/badge.svg)](https://github.com/ensi-platform/elc/actions/workflows/test.yml)

ELC - инструмент для развёртывания микросервисов на машине разработчика.

## Установка

```bash
curl -sSL https://raw.githubusercontent.com/ensi-platform/elc/master/get.sh | sudo bash

elc --help
```

## Сборка из исходников

Зависимости:
- go
- make

```bash
git clone git@github.com:ensi-platform/elc.git
cd elc

make
make install
```

## Использование

### Workspace

Workspace - это описание сервисов системы; файлы конфигурации, необходимые для запуска севрисов; данные сервисов.  
Workspace - это папка в которой находится файл workspace.yaml. 

Структура файла:

```yaml
name: elc-example-1                             # название воркспейса, используется для генерации названий контейнеров/доменов
elc_min_version: 0.1.4                          # минимальная версия elc необхоимая для запуска этого воркспейса
variables:                                      # глобальные переменные
  DEFAULT_APPS_ROOT: ${WORKSPACE_PATH}/apps
  APPS_ROOT: ${APPS_ROOT:-$DEFAULT_APPS_ROOT}
  NETWORK: ${NETWORK:-example}
  BASE_DOMAIN: ${BASE_DOMAIN:-example.127.0.0.1.nip.io}
  GROUP_ID: ${GROUP_ID:-1000}
  USER_ID: ${USER_ID:-1000}
  HOME_PATH: ${WORKSPACE_PATH}/home

templates:                                      # шаблоны сервисов
  fpm-8.1:                                      # название шаблона
    path: ${WORKSPACE_PATH}/templates/fpm-8.1   # путь до папки шаблона
    compose_file: ${TPL_PATH}/docker-compose.yml
    variables:                                  # переменные шаблона
      APP_IMAGE: fpm-8.1:latest
      BASE_IMAGE: php:8.1-fpm-alpine
      NGINX_IMAGE: nginx:1.19-alpine

services:                                       # список сервисв
  proxy:                                        # название сервиса
    path: ${WORKSPACE_PATH}/infra/proxy         # путь до папки сервиса (корень git репозитория)
    variables:
      APP_IMAGE: jwilder/nginx-proxy:latest

  app1:
    path: ${APPS_ROOT}/app1
    extends: fpm-8.1                            # использование шаблона
    dependencies:                               # зависимости сервиса (другие сервисы, которые надо запустить)
      app2: [default]                           # в режиме default надо запустить сервис app2
  app2:
    path: ${APPS_ROOT}/app2
    compose_file: ${SVC_PATH}/docker-compose.yml
    extends: fpm-8.1

modules:                                       # список модулей (пакетов, которые сами не могут быть запущены)
  package1:
    path: /path/to/package/on/host
    hosted_in: app1                            # название сервиса в контейнере которого надо выполнять команды для работы с пакетом
    exec_path: /path/to/package/in/container
```

### Основные понятия

**Сервис** - папка с docker-compose.yml файлом и дополнительными конфигами. В описании сервиса вы можете указать путь до папки,
путь до файла docker-compose.yml и список переменных, которые будут доступны в файле docker-compose.yml.

**Переменная** - может быть задана на уровне сервиса, на уровне шаблона, глобально или через файл env.yaml. При запуске серивса в файле docker-compose.yml
будут доступны все переменные в этой цепочке.  
В качестве значений переменных можно указывать другие переменные: `MY_VAR: ${MY_OTHER_VAR}`.  
Кроме того можно указывать значение по умолчанию, если переменная не определена: `MY_VAR: ${MY_OTHER_VAR:-default value}`.  
Значением по умолчанию может быть даже другая переменная: `MY_VAR: ${MY_OTHER_VAR:-$ANOTHER_VAR}`.  
Ссылаться можно только на переменные, которые определены выше текущей.

**Шаблон** - тоже что и сервис, только на него можно ссылаться из сервиса чтобы наследовать значения.

**Модуль** - папка с файлами, которые не являются самостоятельным сервисом, но могут быть примонтированы в контейнер сервиса.
Модуль нужен, когда вы хотите, находясь в в папке на хосте, запустить инструмент в контейнере. Для этого вы указываете сервис, чей контейнер использовать,
и путь внутри этого контейнера.  
Монтировать папку модуля в контейнер сервиса нужно самостоятельно через docker-compose.yml файл.

**Режим и Зависимости**
Зависимости сервиса - это другие сервисы, которые должны быть запущены перед тем как будет запущен сам сервис.  
Не всегда сервису необходимы все зависимости, поэтому для зависимостей можно указывать в каких режимах их запускать.  
Например в режиме dev сервису нужны database и proxy, а в режиме benchmark ещё нужны app2 и app3.  
По умолчанию используется режим `default`. Git-хуки выполняются в режиме `hook`.

## Возможности ELC

**Управление воркспейсами**

Перед тем как работать с сервисами воркспейса, воркспейс нужно зарегистрировать и выбрать.

```bash
elc workspace add project1 /path/to/workspace/dir/ # зарегистрировать воркспейс, если он первый, то он сразу выберется
elc workspace list                                 # посмотреть список зарегистрированных воркспейсов
elc workspace select project42                     # выбрать другой вореспейс
```

**Управление процессами**

```bash
elc start app1
elc stop app1 app2
elc destroy app1                                   # оставновить и удалить контейнеры сервиса
elc restart app1
elc restart --hard app1                            # удалить контейнеры сервиса и создать снова
```

Всё то же самое можно делать находясь в папке сервиса не указывая его название
```bash
elc start
elc stop
```

Можно указать режим запуска сервиса
```bash
elc start --mode=benchmark
```

**Выполнение команд в контейнере**

```bash
elc exec <command>
elc exec composer install
```

Вы можете войти в контейнер запусти в нём shell
```bash
host$ elc exec bash
app1# composer install
```

Команду exec можно опустить - все нераспознанные команды считаются аргументами для exec
```bash
elc composer install
elc bash
```

Можно выполнить команду в другом сервиса (не в текущей папке)
```bash
elc --svc=db psql
```

**Git хуки**

Часто для выполнения хуков гита нужна та же среда что и для работы сервиса, соответственно и хуки должны выполняться внутри контейнера.  
Для этого elc умеет генерировать хуки, которые будут запускать скрипты расположенные особым образом в репозитории сервиса.

```bash
elc set-hooks ./hooks-dir
```

Папка hooks-dir должна иметь следуюзую структуру:
```
./hooks-dir/
  ├── pre-commit
  │   ├── lint-openapi.sh
  │   ├── lint-php.sh
  │   └── php-cs-fixer.sh
  └── pre-push
      ├── composer-validate.sh
      ├── test-code.sh
      └── var-dump-checker.sh
```
Т.е. название подпапки - это название хука, а внутри сколько угодно скриптов, которые будут выполены при запуске хука. 

**Прочее**

Вы можете выполнить любую команду docker-compose в рамках текущего сервиса
```bash
elc compose <any docker-compose command>
elc compose logs -f app
```

Вы можете получить дамп всех переменных для текущего сервиса
```bash
elc vars
```

## License

Distributed under the MIT License. See [LICENSE.md](LICENSE.md).