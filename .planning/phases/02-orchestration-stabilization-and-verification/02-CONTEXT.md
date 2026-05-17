# Phase 2: Orchestration Stabilization and Verification - Context

**Gathered:** 2026-05-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Эта фаза исправляет orchestration-логику вокруг `JustStarted` ровно настолько, чтобы в рамках одного command flow не происходил повторный запуск уже обработанных зависимостей на валидных acyclic workspace graphs. Одновременно фаза закрепляет минимально необходимую verification-проверку CLI сценариев и подтверждает, что проект собирается и тесты проходят на обновленном стеке.

</domain>

<decisions>
## Implementation Decisions

### Orchestration Scope
- **D-01:** В этой фазе нужно чинить только защиту от повторного запуска зависимостей в рамках одного command flow.
- **D-02:** Явную cycle detection с новой пользовательской ошибкой в scope Phase 2 включать не нужно.
- **D-03:** Исправление должно быть ориентировано на supported behavior для валидных acyclic workspace graphs, как и зафиксировано в roadmap и requirements.

### Restart Semantics
- **D-04:** Потеря флагов при `restart` не считается багом для этого milestone.
- **D-05:** После остановки контейнера `restart` должен просто запускать сервис заново без переноса дополнительных CLI-опций из исходного вызова.
- **D-06:** Phase 2 не должна расширяться на изменение публичной семантики `restart`, если это не потребуется для прохождения acceptance criteria.

### Verification Depth
- **D-07:** Для этой фазы достаточно минимального набора regression checks.
- **D-08:** Проверки должны покрывать только ключевые сценарии вокруг `start`, `restart` и shared dependencies, которые прямо подтверждают исправление `JustStarted` и отсутствие регрессий в поддерживаемых flow.
- **D-09:** Отдельная программа расширения test coverage, CLI wiring coverage или hardening test infrastructure в scope этой фазы не входит.

### Scope Guardrails
- **D-10:** Если по пути всплывут другие баги в тех же кодовых путях, исправлять их нужно только если они критически мешают достижению acceptance criteria этой фазы.
- **D-11:** Низкорисковые, но необязательные cleanup/fix работы без прямой связи с acceptance criteria нужно откладывать в будущие фазы или quick tasks.

### the agent's Discretion
- Как именно локально помечать компонент как уже обработанный в рамках одного command flow: через существующий `JustStarted` lifecycle или через эквивалентную минимальную механику без расширения публичного API.
- Нужен ли точечный automated test именно на shared dependency traversal, если это самый прямой и минимальный способ доказать отсутствие повторного запуска.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Planning
- `.planning/PROJECT.md` — milestone intent, ограничения по совместимости и уже зафиксированные решения
- `.planning/REQUIREMENTS.md` — phase requirements `TOOL-02`, `DIST-01`, `DIST-02`, `ORCH-01`, `ORCH-02`, `ORCH-03`, `COMP-01`, `COMP-03`
- `.planning/ROADMAP.md` — goal, success criteria и границы Phase 2
- `.planning/STATE.md` — текущее состояние milestone и накопленные decisions/concerns
- `.planning/phases/01-stack-and-distribution-refresh/01-01-SUMMARY.md` — итог toolchain/dependency refresh
- `.planning/phases/01-stack-and-distribution-refresh/01-02-SUMMARY.md` — итог удаления updater path и текущий runtime baseline

### Codebase Maps
- `.planning/codebase/ARCHITECTURE.md` — общий runtime flow CLI и orchestration structure
- `.planning/codebase/CONCERNS.md` — зафиксированные runtime bugs вокруг `JustStarted`, `restart`, coverage gaps и compatibility risks
- `.planning/codebase/TESTING.md` — текущее тестовое покрытие и его ограничения

### Code Hotspots
- `core/component.go` — `Start`, `startDependencies`, `Restart` и текущий `JustStarted` guard
- `actions/component_actions.go` — user-facing service actions и command-level orchestration entrypoints
- `cmd/elc.go` — CLI wiring для `start` / `restart` и связанных flags
- `actions/component_actions_test.go` — существующие behavior tests вокруг service actions
- `.github/workflows/test.yml` — CI path, который должен остаться зеленым на обновленном стеке

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `core/component.go`: уже содержит поле `JustStarted` и recursion path через `startDependencies`, поэтому фикс можно сделать локально в существующем orchestration code path без redesign command graph.
- `actions/component_actions_test.go`: уже содержит тесты на `RestartServiceAction`, что дает место для минимального regression coverage рядом с текущим поведением.
- `.github/workflows/test.yml`: после Phase 1 уже выровнен на новый Go baseline, поэтому Phase 2 может опираться на существующий CI path без нового toolchain refactor.

### Established Patterns
- Командные сценарии стартуют из `actions/*`, а затем делегируют runtime behavior в `core/*`; исправление нужно держать внутри этих текущих слоев, без вынесения нового orchestration framework.
- CLI compatibility в этом milestone понимается как сохранение поддерживаемых user flows, а не как исправление всех исторических странностей поведения.
- Milestone в целом ограничен минимально необходимыми изменениями; это относится и к Phase 2.

### Integration Points
- Фикс `JustStarted` затронет recursion path `Start -> startDependencies -> dep.Start`.
- Regression verification должна проходить через `go test ./...` и `go build ./...` на уже обновленном стеке.
- Если понадобится минимальный targeted test, он должен быть привязан к существующему поведению `start` / dependency traversal, а не открывать отдельную ветку test-hardening work.

</code_context>

<specifics>
## Specific Ideas

- Исправление должно предотвращать повторную обработку shared dependencies в рамках одного запуска команды, а не решать все возможные проблемы графа зависимостей.
- Семантику `restart` в этой фазе менять не нужно: после stop/destroy сервис просто стартует заново в своем обычном режиме.
- Проверки должны быть минимальными и доказательными: ровно столько, сколько нужно для подтверждения `JustStarted` fix и green build/test path.

</specifics>

<deferred>
## Deferred Ideas

- Явная dependency cycle detection с user-facing diagnostic
- Более широкое покрытие CLI flag propagation и command wiring
- Некритичные adjacent fixes в orchestration/runtime code paths, которые не нужны для acceptance criteria
- Более общий refactor orchestration layer, dependency planner или global runtime state

</deferred>

---
*Phase: 2-Orchestration Stabilization and Verification*
*Context gathered: 2026-05-17*
