# Requirements: ELC - Ensi Local Ctl

**Defined:** 2026-05-17
**Core Value:** Разработчик должен иметь предсказуемый и совместимый CLI для управления локальным workspace на Linux/macOS без ручной возни с запуском сервисов и контейнерных команд

## v1 Requirements

Requirements for the current technical-refresh milestone.

### Toolchain

- [ ] **TOOL-01**: Maintainer can build ELC from source using an актуальный поддерживаемый major release Go toolchain
- [ ] **TOOL-02**: CI uses the updated Go toolchain and passes the existing automated test suite on the refreshed dependency graph
- [ ] **TOOL-03**: ELC uses updated major versions of its key dependencies with no unresolved build-time incompatibilities

### Distribution

- [ ] **DIST-01**: User installs ELC through Homebrew as the supported installation path on Linux/macOS
- [ ] **DIST-02**: User upgrades ELC through Homebrew without relying on any built-in self-update mechanism
- [ ] **DIST-03**: ELC no longer executes a configurable shell-based self-update command from home config during normal product usage

### Orchestration

- [ ] **ORCH-01**: Starting a component does not repeatedly restart dependencies already started in the same command flow
- [ ] **ORCH-02**: Dependency traversal handles the `JustStarted` guard consistently so shared dependencies are not redundantly reprocessed
- [ ] **ORCH-03**: The fix for dependency startup guarding does not break existing start/restart flows for valid acyclic workspace graphs

### Compatibility

- [ ] **COMP-01**: Existing CLI commands and flags continue to behave compatibly for the supported workflows in this repository
- [ ] **COMP-02**: Existing `workspace.yaml`, `env.yaml`, and `~/.elc.yaml` formats remain readable without requiring user migration for this milestone
- [ ] **COMP-03**: The refreshed project remains compatible with supported Linux/macOS environments from 2022 and newer

## v2 Requirements

### Deferred Hardening

- **HARD-01**: Config loading validates YAML schemas strictly and reports user-facing diagnostics for invalid shapes
- **HARD-02**: Generated git hooks quote paths robustly and have dedicated tests for edge cases
- **HARD-03**: Remaining security and architecture concerns from `.planning/codebase/CONCERNS.md` are prioritized and addressed beyond this milestone

### Testing Expansion

- **TEST-01**: Core runtime logic gains broader regression coverage beyond the minimum needed for this milestone
- **TEST-02**: CLI wiring and security-sensitive execution paths receive dedicated coverage improvements

## Out of Scope

| Feature | Reason |
|---------|--------|
| New user-facing features | Этот milestone посвящен обновлению стека и исправлению известных проблем |
| Large-scale architectural redesign | Слишком расширяет scope относительно цели технической стабилизации |
| Full test coverage initiative | Нужна достаточная регрессия, а не отдельная программа максимального покрытия |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| TOOL-01 | Phase 1 | Complete |
| TOOL-02 | Phase 2 | Pending |
| TOOL-03 | Phase 1 | Complete |
| DIST-01 | Phase 2 | Pending |
| DIST-02 | Phase 2 | Pending |
| DIST-03 | Phase 1 | Complete |
| ORCH-01 | Phase 2 | Pending |
| ORCH-02 | Phase 2 | Pending |
| ORCH-03 | Phase 2 | Pending |
| COMP-01 | Phase 2 | Pending |
| COMP-02 | Phase 1 | Complete |
| COMP-03 | Phase 2 | Pending |

**Coverage:**
- v1 requirements: 12 total
- Mapped to phases: 12
- Unmapped: 0

---
*Requirements defined: 2026-05-17*
*Last updated: 2026-05-17 after Phase 1 execution*
