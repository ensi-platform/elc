# Milestones

## Shipped

### v1.0 Technical Refresh

- **Date:** 2026-05-17
- **Phases:** 2
- **Plans:** 3
- **Tasks:** 6
- **Archive:** [v1.0-ROADMAP.md](/Users/ivankoryukov/Work/ensi/runtime/elc-go/.planning/milestones/v1.0-ROADMAP.md:1)

Delivered:
- Обновлен Go baseline до `1.26` и освежены прямые runtime dependencies.
- Конфигурационный слой переведен на YAML v3 с compatibility-focused адаптацией.
- Built-in updater path удален; Homebrew зафиксирован как единственный поддерживаемый путь установки и обновления.
- `JustStarted` стал эффективным guard для shared dependency traversal в одном command flow.
- `go build ./...` и `go test ./...` проходят на обновленном стеке.
