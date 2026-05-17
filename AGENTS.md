<!-- GSD:project-start source:PROJECT.md -->
## Project

ELC - это CLI-инструмент для локальной разработки, который помогает поднимать и обслуживать workspace из нескольких сервисов и модулей через единое описание в `workspace.yaml`. Он управляет запуском контейнеров, выполнением команд внутри сервисов, регистрацией и выбором workspace, а также генерацией git-хуков для запуска в контейнерной среде.

Текущий milestone не добавляет новые пользовательские возможности. Его задача - технически оздоровить существующий продукт: обновить стек до последних мажорных версий, удалить встроенный self-update в пользу Homebrew и исправить уже известные дефекты без поломки CLI, API и формата конфигов.
<!-- GSD:project-end -->

<!-- GSD:stack-start source:STACK.md -->
## Technology Stack

Основной runtime написан на Go. CLI построен вокруг `github.com/spf13/cobra`, конфигурация workspace хранится в YAML, orchestration опирается на `docker compose`, `git` и локальное shell-окружение.
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

Следовать существующему разделению `cmd` -> `actions` -> `core`. Не ломать CLI/config compatibility без явного решения. Изменения в orchestration и config loading сопровождать regression checks.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

Система имеет layered CLI architecture: `main.go`/`cmd/elc.go` формируют command graph, `actions/*.go` оркестрируют операции, а `core/*.go` содержит runtime behavior, config loading и integration с host tools.
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

No project skills found. Add skills to any of: `.codex/skills/`, `.agents/skills/`, `.cursor/skills/`, or `.github/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `$gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `$gsd-debug` for investigation and bug fixing
- `$gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->

<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `$gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` - do not edit manually.
<!-- GSD:profile-end -->
