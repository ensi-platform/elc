# Phase 1: Stack and Distribution Refresh - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-17
**Phase:** 1-stack-and-distribution-refresh
**Areas discussed:** Точная целевая линия обновления стека, Поведение CLI после удаления self-update, Обратная совместимость `~/.elc.yaml`, Глубина YAML/tooling-модернизации

---

## Точная целевая линия обновления стека

| Option | Description | Selected |
|--------|-------------|----------|
| Обновить только часть прямых runtime dependency | Обновить лишь минимальный набор библиотек вокруг активного кода | |
| Обновить все прямые зависимости до актуальных версий | Взять текущие upstream версии для всех прямых зависимостей | ✓ |
| Расширить scope на transitive/test ecosystem | Обновлять также окружающий tooling сверх минимума | |

**User's choice:** Нужно обновить прямые зависимости до актуальных версий и найти актуальные в интернете.
**Notes:** По результатам обсуждения и internet lookup как target зафиксированы Go `1.26.x`, `cobra` `v1.10.2`, `go-version` `v1.9.0`, `go-isatty` `v0.0.22`, migration на `go.yaml.in/yaml/v3` (`v3.0.4` published v3 line). `golang/mock` не трогать автоматически без необходимости.

---

## Поведение CLI после удаления self-update

| Option | Description | Selected |
|--------|-------------|----------|
| Оставить совместимый stub | Команда сохраняется, но печатает сообщение про Homebrew | |
| Удалить команду и механику полностью | Update path больше не является частью поддерживаемого продукта | ✓ |
| Частично скрыть, но не удалять | Снизить видимость, оставив legacy behavior | |

**User's choice:** Полностью удали команду удаления.
**Notes:** Интерпретировано как полное удаление встроенной update-команды и runtime-механики self-update.

---

## Обратная совместимость `~/.elc.yaml`

| Option | Description | Selected |
|--------|-------------|----------|
| Удалять/переписывать legacy поле автоматически | Конфиг будет модифицироваться при чтении/сохранении | |
| Игнорировать поле, но не ломать старые конфиги | Tolerant read сохраняется, runtime его больше не использует | ✓ |
| Жестко ошибаться на старом поле | Пользователь должен мигрировать конфиг вручную | |

**User's choice:** Просто перестань использовать поле в конфиге, но пусть не ломает ничего своим присутствием.
**Notes:** Поле `update_command` остается harmless legacy data.

---

## Глубина YAML/tooling-модернизации

| Option | Description | Selected |
|--------|-------------|----------|
| Минимально необходимое | Только то, что нужно для stack refresh и безопасной совместимости | ✓ |
| Средняя модернизация | Включить часть cleanup/hardening рядом с YAML/tooling | |
| Широкая модернизация | Вынести заметную часть tech debt в эту фазу | |

**User's choice:** Обнови минимально необходимое.
**Notes:** Дополнительный hardening и большой cleanup сознательно отложены.

---

## the agent's Discretion

- Оставить `github.com/golang/mock` как есть, если это не блокирует обновленный Go/toolchain.
- Решить на уровне планирования, нужны ли дополнительные локальные test/tooling-adaptations для прохождения CI.

## Deferred Ideas

- Полная миграция mock ecosystem на `go.uber.org/mock`
- Строгая YAML schema validation
- Остальные security/architecture concerns вне непосредственного scope Phase 1
