Status: ready-for-agent

# 03: Dashboard übersetzt

**What to build:** `templates/dashboard.html` und `app/dashboard.go`
(Monatssoll, Mitgliederzahlen, MoneyMoney-Export-Hinweise).

**Blocked by:** 01 (i18n-Infrastruktur)

## Acceptance Criteria

- [ ] Alle sichtbaren Texte übersetzt, inklusive der Warnhinweise zum
      MoneyMoney-Export (ADR-0016)
- [ ] `go test ./...` grün, `wails build` unter Linux grün
