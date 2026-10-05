# Описание параметров и команд
## Глобальные параметры

- `--debug` - выводит в консоль отладочную информацию
- `--dry-run` - подавляет выполнение реальных действий
- `--help`, `-h` - выводит справку по набранной команде
- `--workspace=NAME`, `-w NAME` - явно задать воркспейс для выполнения текущей команды, игнорируя выбранный или определённый автоматически
- `--branch=BRANCH`, `-b BRANCH` - работать с worktree-инстансом указанной ветки вместо основного clone компонента
- `--source=REF` - создать отсутствующую git-ветку от указанного ref (`branch` или `HEAD`); нужен вместе с `-b` / командой `worktree` 

## Параметры выбора сервиса

Многие команды позволяют указать один или несколько сервисов.  
Сделать это можно разными способами:
- ничего не указывая - сервис будет определён автоматически на основании того в какой папке вы находитесь. Если вы в папке worktree-инстанса, команды работают с этим инстансом (в том числе для worktree, созданных не через elc: компонент и ветка определяются через git)
- перечислив имена одного или нескольких сервисов как аргументы
- указав имя одного сервиса через флаг `-c NAME`, `--component=NAME` или `--svc=NAME`, используется когда нельзя использовать аргументы
- задав тэг через флаг `--tag=TAG`
- указав ветку через `--branch=BRANCH` / `-b BRANCH`, если нужно работать с worktree-инстансом, не находясь в его папке

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
После клонирования, если в воркспейсе для сервиса или шаблона задан `hooks.after_clone`
(или устаревший `after_clone_hook`), то он будет выполнен.  

Опции:
* `--no-hook` - не выполнять хук после клонирования
* `--tag=TAG` - запустить все сервисы помеченные тэгом

Примеры:
```
elc clone MY_SERVICE
elc clone --tag=frontend
```

## run-hook
```
elc run-hook <HOOK> [-- ARGS...]
elc rh <HOOK> [-- ARGS...]
```
Запустить именованный хук компонента из `workspace.yaml`
(`hooks.after_clone`, `hooks.worktree_create`, `hooks.worktree_remove`).

Алиас: `rh`.

Компонент берётся из текущей директории или через `-c`/`--component`.
Если вы в папке worktree (или указан `--branch`), хук выполняется в контексте
этого инстанса. Аргументы после `--` передаются скрипту хука.

Примеры:
```
elc run-hook after_clone
elc rh after_clone
elc run-hook worktree_create -- --env=staging
elc run-hook worktree_remove -c app1 -b feature/foo
```

## worktree
```
elc worktree <BRANCH> [-- HOOK_ARGS...]
elc worktree add <BRANCH> [-- HOOK_ARGS...]
```
Создать git worktree текущего или указанного (`-c`) компонента в папке
`$WORKTREES_PATH/<component>/<branch>`.

Алиас: `wt`.

Перед созданием выполняется `git fetch`; ошибка fetch прерывает создание
(чтобы `--source` не создал ветку при недоступном remote).
Если локальной ветки нет, а на `origin` есть одноимённая — прозрачно создаётся
локальная tracking-ветка с именем `<branch>`.
Если ветки нигде нет — ошибка; передайте `--source=<branch|HEAD>`, чтобы создать
новую ветку от указанного ref (`HEAD` / `head` — от текущего HEAD основного clone).

Путь задаётся переменной `WORKTREES_PATH` в блоке `variables` (можно переопределить в `env.yaml`).
Если переменная не задана, используется `${WORKSPACE_PATH}/worktrees` — рядом с `apps`.

Основной clone компонента должен уже существовать.

После создания, если для сервиса или шаблона задан `hooks.worktree_create`, он будет
выполнен. Аргументы после `--` передаются хуку как есть.

Опции:
* `--no-hook` - не выполнять хук после создания
* `--component=NAME`, `-c NAME` - компонент, если вы не в его папке
* `--source=REF` - создать отсутствующую ветку от ref (`main`, `HEAD`, …)

Примеры:
```
elc worktree feature/foo
elc wt feature/foo
elc worktree feature/foo --source=HEAD
elc worktree feature/foo --source=main
elc worktree feature/foo -- --env=staging
elc worktree add feature/foo -c app1 -- --env=staging
```

## worktree list
```
elc worktree list
elc worktree ls
elc wt ls
```
Показать список worktree-инстансов воркспейса.  
Строки вывода: `<component>\t<branch>\t<path>`.

Опции:
* `--component=NAME`, `-c NAME` - только worktrees указанного компонента

Примеры:
```
elc worktree list
elc wt ls -c app1
```

## worktree remove
```
elc worktree remove [BRANCH]
elc worktree rm [BRANCH]
elc wt rm [BRANCH]
```
Остановить контейнеры инстанса и удалить git worktree. Саму git-ветку не удаляет.

Перед удалением, если задан `hooks.worktree_remove`, он выполняется в контексте
worktree-инстанса (пока папка ещё существует). Затем `compose down` и `git worktree remove`.

Ветку можно не указывать, если вы находитесь в папке инстанса. Иначе передайте имя
ветки аргументом или через `--branch`.

Опции:
* `--force` - принудительно удалить worktree и игнорировать ошибки hook/`compose down`
* `--no-hook` - не выполнять `hooks.worktree_remove`

Примеры:
```
elc worktree remove
elc worktree remove feature/foo
elc worktree rm feature/foo -c app1
elc wt rm feature/foo --no-hook
```

## launch
```
elc launch [OPTIONS] <COMMAND>
```
Запустить `<COMMAND>` на хосте в папке компонента (основной clone или worktree).
Передаёт env переменные компонента, как `wrap`, но меняет рабочую директорию на `SVC_PATH`.

Если `--branch` не задан — используется основной репозиторий.
Если `--branch` задан (или вы в папке worktree) — команда выполняется в worktree;
отсутствующий worktree создаётся автоматически (fetch + resolve remote branch).
Новая git-ветка создаётся только с `--source=<branch|HEAD>`.

Опции:
* `--component=NAME`, `-c NAME` - компонент
* `--branch=BRANCH`, `-b BRANCH` - worktree-инстанс ветки
* `--source=REF` - создать отсутствующую ветку от ref

Примеры:
```
elc launch code .
elc launch -c app1 -b feature/foo code .
elc launch -b feature/foo --source=HEAD cursor .
```

## start
```
start [OPTIONS] [SERVICES]
```
Запустить текущий или указанный сервис.  
Технически просто выполняет `docker compose up` вычислив все переменные и сформировав параметры запуска.
Перед запуском текущего сервиса рекурсивно запускает его зависимости для текущего режима.

Для host-only компонентов (`compose_file: null`) команда недоступна.

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

## run
```
exec run [OPTIONS] <SHELL-COMMAND>
```
Выполнить `<SHELL-COMMAND>` в отдельном контейнере текущего сервиса.  
Выполняет `docker compose run` не запуская ни сам сервис, ни его зависимости.

Опции:
* `--component=NAME`, - указать другой сервис вместо текущего
* `--uid` - идентификатор пользователя, по умолчанию использует переменную воркспейса USER_ID
* `--no-tty` - не выделять псевдо-TTY

Примеры:
```
elc run composer install
elc run --component=other-service npm run spectral
elc run --uid=0 --component=database psql -Upostgres
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
elc set-hooks [OPTIONS] <SCRIPTS_DIR>
```
Сгенерировать скрипты для запуска git хуков.  
Смотрит на то какие скрипты лежат в папке `SCRIPTS_DIR` и генерирует соответствующие скрипты в папке `.git/hooks`
используя имя подпапки в качестве имени хука: 
```
scripts-dir/pre-commit/ => .git/hooks/pre-commit
```

Опции:
* `--native` - вместо генерации обёрток в `.git/hooks` прописать в `.git/config` `core.hooksPath=<SCRIPTS_DIR>`.
  В этом режиме git ожидает нативные хуки прямо в `SCRIPTS_DIR` (файлы `pre-commit`, `pre-push`, …).

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
