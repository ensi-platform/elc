# Список доступных команд

## help
```
help
```
Получить справку по доступным командам.  
Можно использовать как подкоманду, так и опции `-h` и  `--help`.  

Примеры:
```
elc help
elc --help
elc -h
```
Можно использовать с любой другой подкомандой, чтобы получить справку по ней.
```
elc exec -h
elc stop --help
```

## workspace show
```
workspace show
```
Показать какой воркспейс сейчас выбран.

## workspace add
```
workspace add <NAME> <PATH>
```
Зарегистрировать воркспейс с именем `<NAME>` и путём до корня `<PATH>`.

## workspace select
```
workspace select <NAME>
```
Выбрать воркспейс с именем `<NAME>`.

## workspace list
```
workspace list
```
Показать список зарегистрированных воркспейсов.

## start
```
start [OPTIONS] [SERVICES]
```
Запустить текущий или указанный сервис.  
Опции:
* `--force` - запустить зависимости сервиса даже если сервис уже запущен
* `--mode=MODE` - режим запуска зависимостей сервиса

Примеры:
```
elc start
elc start other-service
elc start --mode=full
```

## stop
```
stop [SERVICES]
```
Остановить текущий или указанный сервис.  

Примеры:
```
elc stop
elc stop other-service
```
## destroy
```
destroy [SERVICES]
```
Остановить текущий сервис и удалить его контейнеры. Опционально можно передать список имён сервисов.  
Примеры:
```
elc destroy
elc destroy other-service-1 other-service-2
```
## restart
```
restart [OPTIONS] [SERVICES]
```
Перезапустить текущий сервис.  
Опционально можно передать список имён сервисов.  
Опции:
* `--hard` - пересоздать контейнер сервиса

Примеры:
```
elc restart
elc restart other-service
elc restart --hard
```

## exec
```
exec [OPTIONS] <SHELL-COMMAND>
```
Выполнить `<SHELL-COMMAND>` в контейнере текущего сервиса.  
Опции:
* `--svc` - указать другой сервис вместо текущего
* `--mode` - режим запуска сервиса
* `--uid` - идентификатор пользователя, по умолчанию использует `$(id -u)`

Примеры:
```
elc exec composer install
elc exec --svc=other-service npm run spectral
elc exec --uid=0 --svc=database psql -Upostgres
elc exec --mode=full php artisan import:stocks
```
Слово `exec` можно опустить - встретив неизвестную команду elc использует её как аргумент для неявного вызова exec.  
```
elc exec pwd === elc pwd
elc exec --svc=other-service composer install === elc --svc=other-service composer install
```
Команда exec номально обрабатывает проброс TTY, что позволяет запускать в контейнере интерактивные/цветные/TUI приложения.
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
* `--svc` - указать другой сервис вместо текущего
* `--mode` - режим запуска сервиса

Примеры:
```
elc compose logs -f app
elc compose build
elc compose --svc=other-service logs
```

## set-hooks
```
set-hooks <scripts-dir>
```
Сгенерировать скрипты для запуска git хуков.  
Смотрит на то какие скрипты лежат в папке `scripts-dir` и генерирует соответствующие скрипты в папке `.git/hooks`
используя имя подпапки в качестве имени хука: 
```
scripts-dir/pre-commit/ => .git/hooks/pre-commit
```

Примеры:
```
elc set-hooks .git_hooks
```

## vars
```
vars [SERVICE]
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
elc update --version=0.1.8
```
## version
```
version
```
Показать версию elc.
## wrap
```
wrap [OPTIONS] <SHELL-COMMAND>
```
Выполнить команду `<SHELL-COMMAND>` передав ей env переменные текущего или указанного сервиса.

Опции:
* `--svc` - указать другой сервис вместо текущего

Примеры:
```
elc wrap ./prepare-service.sh
elc wrap --svc=other-service ./prepare-service.sh
```
## fix-update-command
```
fix-update-command
```
Актуализировать команду обновления прописанную в ~/.elc.yaml
