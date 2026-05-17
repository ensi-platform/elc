# Phase 1: Stack and Distribution Refresh - Context

**Gathered:** 2026-05-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Эта фаза обновляет toolchain и прямые runtime-зависимости проекта до актуальных версий, удаляет встроенный self-update path и сохраняет обратную совместимость CLI/config behavior для существующих пользователей.

</domain>

<decisions>
## Implementation Decisions

### Dependency Targets
- **D-01:** Обновлять нужно прямые зависимости проекта до актуальных версий, подтвержденных по официальным интернет-источникам на дату обсуждения.
- **D-02:** Целевая линия Go для этой фазы — актуальная поддерживаемая ветка `1.26.x`.
- **D-03:** Для Phase 1 фиксируются следующие целевые зависимости:
  - `github.com/spf13/cobra` -> `v1.10.2`
  - `github.com/hashicorp/go-version` -> `v1.9.0`
  - `github.com/mattn/go-isatty` -> `v0.0.22`
  - `gopkg.in/yaml.v2` -> migration to `go.yaml.in/yaml/v3` (`v3.0.4` as current published v3 line)
- **D-04:** `github.com/golang/mock` не нужно мигрировать автоматически в этой фазе, если только это не окажется минимально необходимым для совместимости со свежим Go/tooling.

### Updater Removal
- **D-05:** Встроенную команду и runtime-механику self-update нужно удалить полностью, а не оставлять совместимый stub.
- **D-06:** Homebrew считается единственным поддерживаемым способом установки и обновления; поведение продукта и документация должны это явно отражать.

### Config Compatibility
- **D-07:** Поле `update_command` в `~/.elc.yaml` больше не используется продуктом.
- **D-08:** Наличие legacy-поля `update_command` не должно ломать загрузку или работу конфигурации; tolerant read сохраняется.
- **D-09:** Обязательной пользовательской миграции `~/.elc.yaml` в этой фазе быть не должно.

### Modernization Depth
- **D-10:** Делать только минимально необходимую модернизацию вокруг YAML/tooling, которая нужна для обновления стека и безопасной работы после него.
- **D-11:** Не включать в эту фазу более широкий hardening, redesign config loading или дополнительный security remediation вне прямой необходимости.

### the agent's Discretion
- Нужно ли мигрировать test/mocking tooling дальше текущего минимума, если сборка и тесты проходят без этого.
- Нужен ли мягкий backward-compatible parsing path для старых updater-related команд в CLI help/docs или их можно удалить полностью без alias/stub.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Planning
- `.planning/PROJECT.md` — продуктовый контекст, ограничения milestone и уже зафиксированные решения
- `.planning/REQUIREMENTS.md` — phase requirements `TOOL-01`, `TOOL-03`, `DIST-03`, `COMP-02`
- `.planning/ROADMAP.md` — phase goal, success criteria и границы scope
- `.planning/STATE.md` — текущее состояние milestone и накопленные concerns

### Codebase Maps
- `.planning/codebase/STACK.md` — текущий baseline стека, toolchain и прямых зависимостей
- `.planning/codebase/INTEGRATIONS.md` — updater path, Docker/Git/Homebrew-related внешние точки интеграции
- `.planning/codebase/CONCERNS.md` — известные проблемы вокруг legacy updater path, YAML/tooling debt и compatibility risks

### Research
- `.planning/research/STACK.md` — исследование целевого stack refresh и dependency upgrade path
- `.planning/research/SUMMARY.md` — краткие выводы по целевому стеку и основным migration pitfalls

### Code Hotspots
- `go.mod` — текущие прямые зависимости и версия Go
- `README.md` — уже задокументированный Homebrew install path
- `actions/general_actions.go` — текущая update command implementation
- `core/home-config.go` — `update_command`, default update command и чтение home config
- `.github/workflows/test.yml` — текущая pinned CI toolchain version

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `core/home-config.go`: централизованная модель `~/.elc.yaml`, через которую можно безопасно сохранить tolerant read старого поля без использования updater path
- `actions/general_actions.go`: изолированное место для удаления update-related action logic
- `go.mod` и `.github/workflows/test.yml`: явные точки обновления toolchain и direct dependencies
- `README.md`: уже отражает Homebrew как основной install path, что позволяет синхронизировать code behavior с документацией

### Established Patterns
- Конфигурация пользователя и workspace читается через YAML-модели в `core/*`, поэтому изменения в YAML library должны минимально менять существующие структуры и теги
- CLI построен через Cobra в `cmd/elc.go`, поэтому удаление updater-команд нужно делать без каскадной поломки остального command graph
- Build/test path централизован в `Makefile`, `gen.sh` и `.github/workflows/test.yml`

### Integration Points
- Обновление стека начинается с `go.mod`, `go.sum`, CI workflow и, возможно, генерации mock/tooling
- Удаление updater path затрагивает `actions/general_actions.go`, `core/home-config.go`, `cmd/elc.go`, `README.md` и, возможно, `doc/commands.md`
- YAML migration затронет минимум `core/home-config.go`, `core/workspace_config.go`, `core/pc.go` и связанные места чтения/записи конфигов

</code_context>

<specifics>
## Specific Ideas

- Искать актуальные версии зависимостей по интернет-источникам и закреплять именно прямые зависимости, а не пытаться вручную “освежать все подряд”.
- Legacy `update_command` должен быть безвредным мусором в конфиге: не использоваться, но и не ломать чтение старых файлов.
- Phase 1 должна остаться минимальной по изменениям: только то, что нужно для stack refresh и удаления updater path.

</specifics>

<deferred>
## Deferred Ideas

- Возможная миграция с `github.com/golang/mock` на `go.uber.org/mock`, если она не понадобится как минимально необходимая часть stack refresh
- Более строгая YAML schema validation и дополнительные user-facing diagnostics
- Остальные security и architecture concerns из `.planning/codebase/CONCERNS.md`, не связанные напрямую с updater removal и minimal stack refresh

</deferred>

---
*Phase: 1-Stack and Distribution Refresh*
*Context gathered: 2026-05-17*
