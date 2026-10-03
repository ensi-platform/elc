#  ELC - Ensi Local Ctl

[![Test](https://github.com/ensi-platform/elc/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/ensi-platform/elc/actions/workflows/test.yml)

ELC - инструмент для развёртывания микросервисов на машине разработчика, целью которого является запуск всех необходимых для разработки
программ в контейнере.  

Особенности:  
- позволяет описать конфигурацию запуска всех компонентов системы в одном месте, упрощая начало работы с проектом до клонирования одного репозитория
- упрощает конфигурирование, вводя свой набор переменных, часть из которых рассчитывается на лету
- сокращает количество и размер команд, необходимых для запуска проекта
- позволяет запускать .git хуки в контейнере

## Установка (Linux, WSL, Mac)

```shell
brew tap ensi-platform/elc
brew install elc
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
elc_min_version: 0.2.3                          # минимальная версия elc необхоимая для запуска этого воркспейса
variables:                                      # глобальные переменные
  DEFAULT_APPS_ROOT: ${WORKSPACE_PATH}/apps
  APPS_ROOT: ${APPS_ROOT:-$DEFAULT_APPS_ROOT}
  DEFAULT_WORKTREES_PATH: ${WORKSPACE_PATH}/worktrees
  WORKTREES_PATH: ${WORKTREES_PATH:-$DEFAULT_WORKTREES_PATH}
  NETWORK: ${NETWORK:-example}
  BASE_DOMAIN: ${BASE_DOMAIN:-example.127.0.0.1.nip.io}
  GROUP_ID: ${GROUP_ID:-1000}
  USER_ID: ${USER_ID:-1000}
  HOME_PATH: ${WORKSPACE_PATH}/home

templates:                                      # шаблоны сервисов
  fpm-8.5:                                      # название шаблона
    path: ${WORKSPACE_PATH}/templates/fpm-8.5   # путь до папки шаблона
    compose_file: ${TPL_PATH}/docker-compose.yml
    hooks:
      after_clone: ${TPL_PATH}/hooks/after-clone.sh
      worktree_create: ${TPL_PATH}/hooks/after-worktree.sh
      worktree_remove: ${TPL_PATH}/hooks/before-worktree-remove.sh
    variables:                                  # переменные шаблона
      BASE_IMAGE: php:8.5-fpm-alpine
      NGINX_IMAGE: nginx:1.28-alpine

services:                                       # список сервисов
  proxy:                                        # название сервиса
    path: ${WORKSPACE_PATH}/infra/proxy         # путь до папки сервиса (корень git репозитория)
    variables:
      APP_IMAGE: nginxproxy/nginx-proxy:1.11-alpine

  app1:
    path: ${APPS_ROOT}/app1
    extends: fpm-8.5                            # использование шаблона
    repository: git@github.com:example/app1.git
    tags:
      - frontend
    dependencies:                               # зависимости сервиса (другие сервисы, которые надо запустить)
      app2: [default]                           # в режиме default надо запустить сервис app2
  app2:
    path: ${APPS_ROOT}/app2
    compose_file: ${SVC_PATH}/docker-compose.yml
    tags:
      - backend
    extends: fpm-8.5

  configs:                                      # host-only: без запуска контейнеров
    path: ${WORKSPACE_PATH}/configs
    repository: git@github.com:example/configs.git
    compose_file: null                          # явно отключает compose / runtime-команды

modules:                                       # список модулей (пакетов, которые сами не могут быть запущены)
  package1:
    path: /path/to/package/on/host
    hosted_in: app1                            # название сервиса в контейнере которого надо выполнять команды для работы с пакетом
    exec_path: /path/to/package/in/container
```

### Основные понятия

**Сервис** - папка с docker-compose.yml файлом и дополнительными конфигами. В описании сервиса вы можете указать путь до папки,
путь до файла docker-compose.yml и список переменных, которые будут доступны в файле docker-compose.yml.
Если `compose_file` не указан, по умолчанию используется `${SVC_PATH}/docker-compose.yml` (или compose шаблона при `extends`).

**Host-only компонент** - репозиторий без собственного runtime (например, набор конфигов). Задаётся через `compose_file: null`.
Доступны `clone`, `wrap`, `launch`, `worktree`, `vars`, git hooks; недоступны `start`/`stop`/`restart`/`compose`/`exec`/`run`.

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

**Тэги**
Многие команды можно применить сразу к нескольким сервисам. Чтобы обозначить какой-то часто используемый набор сервисов,
можно назначить им одинаковый тэг и в дальшейгем, вместо перечисления названий сервисов в команде можно использовать флаг `--tag=<my-tag>`.

## Возможности ELC

[Список всех команд](/doc/commands.md)

**Управление воркспейсами**

Перед тем как работать с сервисами воркспейса, воркспейс нужно зарегистрировать.

```bash
elc workspace add project1 /path/to/project1/workspace
elc workspace set-root project1 /path/to/project1
```

Далее есть два варианта работы с воркспейсами. Первый - включить режим автоматического определения воркспейса на основании
того в какой папке вы находитесь.
```
elc workspace select auto
```
Второй вариант - это явно выбрать один воркспейс
```
elc workspace select project1
```

Кроме того, вы всегда можете указать в каком воркспейсе выполнить действие указав опцию `--workspace=project1`.

**Управление процессами**

```bash
elc start app1
elc destroy app1                                   # оставновить и удалить контейнеры сервиса
elc restart app1
elc restart --hard app1                            # удалить контейнеры сервиса и создать снова
```

Всё то же самое можно делать находясь в папке сервиса не указывая его название
```bash
elc start
elc stop
```

Если нужно параллельно работать над другой веткой сервиса, создайте git worktree.
Инстанс появится в `$WORKTREES_PATH/<component>/<branch>` (по умолчанию
`${WORKSPACE_PATH}/worktrees`, рядом с `apps`) и получит отдельные
`APP_NAME` / `COMPOSE_PROJECT_NAME`, чтобы контейнеры и hostname не пересекались
с основным clone.

```bash
elc worktree feature/foo
elc wt feature/foo --source=HEAD
elc worktree feature/foo --source=main
elc start -b feature/foo
elc launch -c app1 -b feature/foo --source=HEAD code .
```

Команды `start`, `stop`, `exec` и остальные определяют инстанс по текущей папке.
Работают и worktree, созданные не через elc: компонент находится по пути основного
репозитория (`git rev-parse`), ветка — по текущему HEAD.
Если вы не в папке worktree, укажите ветку флагом `--branch` / `-b`.
Отсутствующий worktree для существующей ветки создаётся автоматически (только по схеме `$WORKTREES_PATH/...`).
Новую git-ветку можно создать только с `--source=<branch|HEAD>`.
`elc launch` запускает программу на хосте в папке инстанса.

```bash
cd worktrees/app1/feature/foo
elc start
elc exec bash
elc worktree remove
```

Список инстансов:

```bash
elc worktree list
elc wt ls
elc wt ls -c app1
```

Можно указать сразу несколько сервисов перечислив их имена или используя тэг
```bash
elc start app1 app2 app3
elc start --tag=core-services
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

Можно выполнить команду в другом сервисе (не в текущей папке)
```bash
elc --component=db psql
elc -c db psql
```
Или даже в другом воркспейсе
```bash
elc --workspace=project2 --component=db psql
elc -w project2 -c db psql
```

**Git хуки**

Часто для выполнения хуков гита нужна та же среда что и для работы сервиса, соответственно и хуки должны выполняться внутри контейнера.  
Для этого elc умеет генерировать хуки, которые будут запускать скрипты расположенные особым образом в репозитории сервиса.

```bash
elc set-hooks ./hooks-dir
elc set-hooks --native .githooks
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

С `--native` elc не генерирует обёртки, а выставляет `git config core.hooksPath` на указанную папку
(ожидаются обычные git-хуки: файлы `pre-commit`, `pre-push` и т.д. прямо в этой папке).

**Прочее**

Вы можете выполнить любую команду docker-compose в рамках текущего сервиса
```bash
elc compose <any docker-compose command>
elc compose logs -f app
```

## License

Distributed under the MIT License. See [LICENSE.md](LICENSE.md).