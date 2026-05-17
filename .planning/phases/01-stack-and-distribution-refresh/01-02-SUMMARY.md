---
phase: 01-stack-and-distribution-refresh
plan: 02
subsystem: infra
tags: [cli, homebrew, config, docs]
requires:
  - phase: 01-01
    provides: refreshed toolchain and dependency baseline for final updater removal verification
provides:
  - Built-in updater command path removed
  - Legacy `update_command` tolerated but unused
  - Documentation aligned with Homebrew-only install/upgrade
affects: [phase-02, cli, docs, config]
tech-stack:
  added: []
  patterns: [legacy-config-tolerant-read]
key-files:
  created: []
  modified: [cmd/elc.go, core/home-config.go, doc/commands.md]
key-decisions:
  - "Removed updater commands entirely instead of leaving a compatibility stub."
  - "Kept `update_command` readable in home config while making it behaviorally inert."
patterns-established:
  - "Legacy config fields may remain readable without preserving their runtime semantics."
requirements-completed: [DIST-03, COMP-02]
duration: 30min
completed: 2026-05-17
---

# Phase 1: Stack and Distribution Refresh Summary

**Built-in updater removal with Homebrew-only command surface and harmless legacy `update_command` retention**

## Performance

- **Duration:** 30min
- **Started:** 2026-05-17
- **Completed:** 2026-05-17
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Removed updater and fix-updater command wiring from the CLI.
- Stopped seeding updater defaults while preserving tolerant reads of legacy `update_command` in `~/.elc.yaml`.
- Removed outdated updater documentation so the supported path is now Homebrew-only.

## Task Commits

No git commits were created during this run. Changes remain in the working tree.

## Files Created/Modified
- `cmd/elc.go` - removed updater command registrations and definitions
- `core/home-config.go` - preserved legacy field reads while removing updater default seeding
- `doc/commands.md` - removed updater command documentation
- `actions/workspace_actions_test.go` - kept tests aligned with the current YAML emitter and config behavior

## Decisions Made
- Removed the updater command entirely rather than leaving a shim that would point users to Homebrew.
- Preserved `update_command` as a legacy config field with `omitempty` so it no longer appears in new configs by default.

## Deviations from Plan

### Auto-fixed Issues

**1. Home config write behavior changed for new empty updater field**
- **Found during:** Task 2 verification
- **Issue:** Removing the default updater seed changed how newly written YAML is emitted.
- **Fix:** Added `omitempty` to `update_command` and kept tolerant reads for old configs.
- **Files modified:** `core/home-config.go`
- **Verification:** `go build ./...` and `go test ./...` passed
- **Committed in:** not committed

---

**Total deviations:** 1 auto-fixed
**Impact on plan:** The deviation was a direct compatibility adjustment and stayed within planned scope.

## Issues Encountered
- No additional code issues beyond the expected doc/config cleanup after removing the updater path.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- The obsolete updater path is fully removed.
- Phase 2 can focus on `JustStarted` and broader runtime/CLI verification without carrying legacy updater behavior forward.

---
*Phase: 01-stack-and-distribution-refresh*
*Completed: 2026-05-17*
