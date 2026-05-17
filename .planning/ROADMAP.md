# Roadmap: ELC - Ensi Local Ctl

## Milestones

- 🚧 **v1.0 Technical Refresh** — обновление стека, удаление встроенного self-update, исправление `JustStarted`, стабилизация сборки и тестов

## Phases

### 🚧 v1.0 Technical Refresh

- [x] Phase 1: Stack and Distribution Refresh (2/2 plans) — completed 2026-05-17
- [x] Phase 2: Orchestration Stabilization and Verification (1/1 plans) — completed 2026-05-17

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Stack and Distribution Refresh | v1.0 | 2/2 | Complete | 2026-05-17 |
| 2. Orchestration Stabilization and Verification | v1.0 | 1/1 | Complete | 2026-05-17 |

## Phase Details

### Phase 1: Stack and Distribution Refresh
**Goal:** Обновить toolchain и ключевые зависимости до актуальных major versions, убрать встроенный self-update и сохранить совместимость конфигов и пользовательских сценариев.

**Requirements:** `TOOL-01`, `TOOL-03`, `DIST-03`, `COMP-02`

**Success Criteria:**
1. `go.mod`, dependency graph и связанные build paths переведены на актуальный поддерживаемый стек без build-time incompatibilities.
2. Встроенный shell-based self-update path удален из поддерживаемого runtime behavior и больше не используется в обычном продукте.
3. Чтение существующих `workspace.yaml`, `env.yaml` и `~/.elc.yaml` остается совместимым без обязательной миграции пользователей.
4. Документация и локальные runtime assumptions согласованы с Homebrew как единственным поддерживаемым каналом установки/обновления.

### Phase 2: Orchestration Stabilization and Verification
**Goal:** Исправить dependency orchestration вокруг `JustStarted`, закрепить обратную совместимость CLI и довести проект до стабильной сборки и проходящих тестов на обновленном стеке.

**Requirements:** `TOOL-02`, `DIST-01`, `DIST-02`, `ORCH-01`, `ORCH-02`, `ORCH-03`, `COMP-01`, `COMP-03`

**Success Criteria:**
1. Логика dependency startup использует `JustStarted` согласованно и не переобрабатывает уже запущенные зависимости в рамках одного command flow.
2. Исправление не ломает существующие supported CLI scenarios для start/restart и related flows на валидных acyclic workspace graphs.
3. Сборка проекта и существующий automated test suite проходят на обновленном стеке и в обновленном CI/toolchain path.
4. Пользовательская модель установки и обновления через Homebrew отражена в поведении и документации продукта без зависимостей от встроенного updater.

## Coverage

- v1 requirements: 12
- phases: 2
- mapped requirements: 12/12
- unmapped requirements: 0

---
*Roadmap created: 2026-05-17*
