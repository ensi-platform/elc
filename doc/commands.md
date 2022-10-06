# Описание параметров и команд
## Глобальные параметры

- `--debug` - выводит в консоль отладочную информацию
- `--dry-run` - подавляет выполнение реальных действий
- `--help`, `-h` - выводит справку по набранной команде
- `--workspace=NAME`, `-w NAME` - явно задать воркспейс для выполнения текущей команды, игнорируя выбранный или определённый автоматически 

## Параметры выбора сервиса

Многие команды позволяют указать один или несколько сервисов.  
Сделать это можно разными способами:
- ничего не указывая - сервис будет определён автоматически на основании того в какой папке вы находитесь
- перечислив имена одного или нескольких сервисов как аргументы
- указав имя одного сервиса через флаг `-c NAME`, `--component=NAME` или `--svc=NAME`, используется когда нельзя использовать аргументы
- задав тэг через флаг `--tag=TAG`

# Список доступных команд

## workspace show
```
workspace show
ws show
```
Показать какой воркспейс сейчас выбран.  
Текущий воркспейс прописан в файле ~/.elc.yaml.

## workspace add
```
workspace add <NAME> <PATH>
ws add <NAME> <PATH>
```
Зарегистрировать воркспейс с именем `<NAME>` и путём до корня `<PATH>`.  
Записывает данные в ~/.env.yaml.

## workspace select
```
workspace select <NAME>
ws select <NAME>
```
Выбрать воркспейс с именем `<NAME>`.  
Если вместо имени воркспейса написать `auto`, то текущий воркспейс будет вычисляться динамически, на основании того в какой
папке вы находитесь. Для того чтобы воркспейс мог быть обнаружен в таком режиме, ему нужно назначить путь командой `set-root`.

## workspace list
```
workspace list
ws ls
```
Показать список зарегистрированных воркспейсов.

## workspace set-root
```
workspace set-root <NAME> <PATH>
ws set-root <NAME> <PATH>
```
Задать корень воркспейса для автоматического определения в режиме `auto`.

## clone
```
elc clone [OPTIONS] [SERVICES]
```
Скачать код сервиса в предназначенную для него папку.  
Адрес git репозитория сервиса можно задать в `workspace.yaml`. В результате будет выполнен `git clone`.
После клонирования, если в воркспейсе для сервиса или шаблона задан `after_clone_hook`, то он будет выполнен.  

Опции:
* `--no-hook` - не выполнять хук после клонирования
* `--tag=TAG` - запустить все сервисы помеченные тэгом

Примеры:
```
elc clone MY_SERVICE
elc clone --tag=frontend
```

## start
```
start [OPTIONS] [SERVICES]
```
Запустить текущий или указанный сервис.  
Технически просто выполняет `docker compose up` вычислив все переменные и сформировав параметры запуска.
Перед запуском текущего сервиса рекурсивно запускает его зависимости для текущего режима.

Опции:
* `--force` - запустить зависимости сервиса даже если сервис уже запущен
* `--mode=MODE` - режим запуска зависимостей сервиса
* `--tag=TAG` - запустить все сервисы помеченные тэгом

Примеры:
```
elc start
elc start other-service
elc start --mode=full
elc start --tag=backend
```

## stop
```
stop [OPTIONS] [SERVICES]
```
Остановить текущий или указанный сервис.  
Технически просто выполняет `docker compose stop` вычислив все переменные. В отличие от `elc start` не работает с зависимостиями.  

Опции:
* `--all` - остановить все сервисы воркспейса
* `--tag=TAG` - остановить все сервисы c заданным тэгом

Примеры:
```
elc stop
elc stop other-service
elc start --tag=backend
```
## destroy
```
destroy [OPTIONS] [SERVICES]
```
Остановить текущий сервис и удалить его контейнеры. Опционально можно передать список имён сервисов.  
Аналог команды `docker compose down`.  

Опции:
* `--all` - остановить и удалить все сервисы воркспейса
* `--tag=TAG` - остановить и удалить все сервисы c заданным тэгом

Примеры:
```
elc destroy
elc destroy other-service-1 other-service-2
elc start --tag=backend
```
## restart
```
restart [OPTIONS] [SERVICES]
```
Перезапустить текущий сервис.  
Опционально можно передать список имён сервисов.

Опции:
* `--hard` - пересоздать контейнер сервиса
* `--tag=TAG` - пересоздать все сервисы c заданным тэгом

Примеры:
```
elc restart
elc restart other-service
elc restart --hard
elc start --tag=backend
```

## exec
```
exec [OPTIONS] <SHELL-COMMAND>
```
Выполнить `<SHELL-COMMAND>` в контейнере текущего сервиса.  
Выполняет `docker compose exec` предварительно запустив сервис и его зависимости.

Опции:
* `--mode` - режим запуска сервиса
* `--component=NAME`, - указать другой сервис вместо текущего
* `--uid` - идентификатор пользователя, по умолчанию использует переменную воркспейса USER_ID
* `--no-tty` - не выделять псевдо-TTY

Примеры:
```
elc exec composer install
elc exec --component=other-service npm run spectral
elc exec --uid=0 --component=database psql -Upostgres
elc exec --mode=full php artisan import:stocks
```
Слово `exec` можно опустить - встретив неизвестную команду elc использует её как аргумент для неявного вызова exec.  
```
elc exec pwd === elc pwd
elc exec --component=other-service composer install === elc --component=other-service composer install
```
Команда exec номально обрабатывает проброс TTY, что позволяет запускать в контейнере интерактивные/цветные/TUI приложения.
Если нужно принудительно отказаться от TTY, например при запуске команды в скриптах, есть опция `--no-tty`.
```
elc grep --color=always -r Request .
elc htop
```
Отдельно стоит отметить возможность запуска sh/bash как способ зайти внутрь контейнера.
```
elc bash
```

## compose
```
compose [OPTIONS] <DOCKER-COMPOSE-COMMAND>
```
Выполнить `<DOCKER-COMPOSE-COMMAND>` в рамках docker-compose проекта текущего сервиса.  
Опции:
* `--component=NAME` - указать другой сервис вместо текущего

Примеры:
```
elc compose logs -f app
elc compose build
elc compose --component=other-service logs
```

## set-hooks
```
elc set-hooks <SCRIPTS_DIR>
```
Сгенерировать скрипты для запуска git хуков.  
Смотрит на то какие скрипты лежат в папке `SCRIPTS_DIR` и генерирует соответствующие скрипты в папке `.git/hooks`
используя имя подпапки в качестве имени хука: 
```
scripts-dir/pre-commit/ => .git/hooks/pre-commit
```

Примеры:
```
elc set-hooks .git_hooks
```

В результате для каждой папки будет создан скрипт хука со следующим содержимым:
```
#!/bin/bash
echo "Run hook via ELC"

elc exec --mode=hook --no-tty scripts-dir/pre-commit/my-script-1.sh
elc exec --mode=hook --no-tty scripts-dir/pre-commit/my-script-2.sh
```

## vars
```
elc vars [SERVICE]
```
Показать переменные текущего или указанного сервиса.

Примеры:
```
elc vars
elc vars other-service
```

## update
```
update [OPTIONS]
```
Обновить elc или переключить на конкретную версию.  
По умолчанию скачивает самую свежую версию elc и переключается на неё.

Опции:
* `--version=VERSION` - версия, на которую нужно переключиться

Примеры:
```
elc update
elc update --version=v0.1.8
```

## wrap
```
wrap [OPTIONS] <SHELL-COMMAND>
```
Выполнить команду `<SHELL-COMMAND>` передав ей env переменные текущего или указанного сервиса.

Опции:
* `--component=NAME` - указать другой сервис вместо текущего

Примеры:
```
elc wrap ./prepare-service.sh
elc wrap --component=other-service ./prepare-service.sh
```

## fix-update-command
```
elc fix-update-command
```
Актуализировать shell-команду обновления прописанную в ~/.elc.yaml.  
Shell-команда обновления сохраняется в ~/.elc.yaml в момент создания этого файла при первом запуске elc.  
Если по какой-то причине эта команда изменися, например из-за переименования репозитория, то ~/.elc.yaml нужно редактировать.  
Чтобы не делать это вручную, существует команда `elc fix-update-command`. 
