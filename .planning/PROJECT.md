# ELC - Ensi Local Ctl

## What This Is

ELC - это CLI-инструмент для локальной разработки, который помогает поднимать и обслуживать workspace из нескольких сервисов и модулей через единое описание в `workspace.yaml`. Он управляет запуском контейнеров, выполнением команд внутри сервисов, регистрацией и выбором workspace, а также генерацией git-хуков для запуска в контейнерной среде.

Milestone `v1.0 Technical Refresh` уже завершен: проект переведен на актуальный Go/tooling baseline, встроенный self-update удален в пользу Homebrew, а orchestration-ошибка вокруг `JustStarted` исправлена без поломки CLI и форматов конфигов. Следующий milestone должен определить, что делать дальше: hardening, coverage expansion, security cleanup или новый продуктовый scope.

## Core Value

Разработчик должен иметь предсказуемый и совместимый CLI для управления локальным workspace на Linux/macOS без ручной возни с запуском сервисов и контейнерных команд.

## Requirements

### Validated

- ✓ CLI умеет регистрировать, выбирать и автоматически определять workspace через `~/.elc.yaml` и `workspace.yaml` — existing
- ✓ CLI умеет запускать, останавливать, перезапускать и удалять сервисы workspace через `docker compose` — existing
- ✓ CLI умеет выполнять команды внутри контейнеров сервисов и модулей, включая работу из текущей директории или с явным выбором workspace/component — existing
- ✓ CLI поддерживает конфигурацию компонентов, шаблонов, переменных, зависимостей и режимов запуска через YAML-конфиги workspace — existing
- ✓ CLI умеет генерировать и использовать git-хуки, запускаемые в контейнерной среде сервиса — existing
- ✓ Проект обновлен до Go `1.26` и актуальных major versions ключевых зависимостей — v1.0
- ✓ Built-in self-update удален, а Homebrew закреплен как единственный поддерживаемый install/upgrade path — v1.0
- ✓ `JustStarted` исправлен так, чтобы shared dependencies не переобрабатывались повторно в одном command flow — v1.0
- ✓ Обратная совместимость CLI/config сохранена, а `go build ./...` и `go test ./...` проходят на обновленном стеке — v1.0

### Active

- [ ] Определить следующий milestone после технического refresh: hardening, coverage expansion, security cleanup или новые продуктовые задачи
- [ ] Решить, нужны ли explicit cycle detection и более широкая orchestration hardening-программа
- [ ] Приоритизировать оставшиеся concerns из `.planning/codebase/CONCERNS.md` и deferred requirements из milestone archive

### Out of Scope

- Новые пользовательские фичи — цель milestone не расширение возможностей продукта, а техническое обновление и стабилизация
- Крупный redesign архитектуры — можно делать точечные исправления, но не переписывать систему целиком
- Доведение test coverage до идеала — нужны достаточные тесты для безопасного обновления, но не отдельная программа тотального покрытия

## Current State

- **Shipped milestone:** `v1.0 Technical Refresh` on `2026-05-17`
- **Current baseline:** Go `1.26`, refreshed direct dependencies, YAML v3 compatibility layer, Homebrew-only distribution model
- **Runtime status:** `JustStarted` bug fixed for supported acyclic graphs; build and tests pass locally on the refreshed stack
- **Known deferred areas:** explicit cycle detection, broader runtime/CLI coverage, remaining hardening and audit concerns outside v1.0 scope

## Next Milestone Goals

- Выбрать следующий фокус: hardening / coverage / security cleanup / product work
- Сформировать новый `REQUIREMENTS.md` через `$gsd-new-milestone`
- Решить, какие deferred items из `v1.0` становятся активными, а какие остаются backlog

## Context

- Кодовая база brownfield: основной runtime написан на Go, CLI собран вокруг Cobra, конфигурация workspace хранится в YAML, а orchestration опирается на `docker compose`, `git` и локальное shell-окружение
- Анализ codebase уже выполнен и сохранен в `.planning/codebase/`
- В `STACK.md` зафиксирован исходный устаревший baseline, от которого проект уже ушел в `v1.0`: Go `1.26`, обновленные direct deps, YAML v3 compatibility layer
- В `CONCERNS.md` остаются deferred-проблемы после `v1.0`:
  - explicit cycle detection для dependency graphs
  - более широкое runtime / CLI wiring coverage
  - дополнительные архитектурные и security hardening-задачи вне scope технического refresh milestone
- README и runtime behavior теперь согласованы с Homebrew как единственным поддерживаемым путем установки и обновления

## Constraints

- **Compatibility**: Linux/macOS уровня 2022 года и выше — итоговый стек и бинарь должны оставаться пригодными для этих целевых систем
- **Distribution**: Homebrew-only update/install path — встроенный self-update должен быть удален, потому что поставка и обновление происходят через Homebrew
- **Backward Compatibility**: Нельзя ломать CLI, API поведения и формат конфигов — существующие workspace и сценарии использования должны продолжить работать
- **Scope**: Только техническое обновление и исправление очевидных проблем — без новых фич и без крупной архитектурной перестройки
- **Quality Gate**: Проект должен собираться и проходить тесты — это минимальный критерий завершения milestone

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Обновлять стек до последних мажорных версий | Цель milestone - привести проект к актуальному состоянию, а не ограничиться минорными апдейтами | ✓ Shipped in v1.0 |
| Полностью удалить встроенный self-update | Установка и обновление теперь идут через Homebrew; дублирующий updater больше не нужен и создает лишний риск | ✓ Shipped in v1.0 |
| Исправлять `JustStarted` без изменения публичного интерфейса | Нужно устранить очевидный баг в orchestration-логике, сохранив поведение CLI и формат конфигов | ✓ Shipped in v1.0 |
| Не включать в milestone новые фичи и крупный redesign | Иначе техническое обновление расползется по scope и потеряет прогнозируемость | ✓ Confirmed by v1.0 scope |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `$gsd-transition`):
1. Requirements invalidated? -> Move to Out of Scope with reason
2. Requirements validated? -> Move to Validated with phase reference
3. New requirements emerged? -> Add to Active
4. Decisions to log? -> Add to Key Decisions
5. "What This Is" still accurate? -> Update if drifted

**After each milestone** (via `$gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check - still the right priority?
3. Audit Out of Scope - reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-17 after v1.0 milestone*
