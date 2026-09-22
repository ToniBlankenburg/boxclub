Status: ready-for-agent

# 05: Excel-Import übersetzt

**What to build:** `templates/import.html`, `app/import.go` und der
Fehlerbericht aus `importer/` (siehe Ticket 09 der ursprünglichen Spec) —
inklusive der Zeilen- und Spaltenfehler, die der Importer je Excel-Zeile
meldet.

**Blocked by:** 01 (i18n-Infrastruktur)

## Acceptance Criteria

- [ ] Formular- und Berichtstexte übersetzt
- [ ] Fehlermeldungen des `ExcelImporter` übersetzt (Sprache muss vom
      Aufrufer übergeben werden — `importer/` kennt heute keine Sprache)
- [ ] `go test ./...` grün, `wails build` unter Linux grün
