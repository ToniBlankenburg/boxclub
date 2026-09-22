Status: ready-for-agent

# 06: Rechnung, Serienmail, Verein-Formular übersetzt

**What to build:** `templates/rechnung.html`, `templates/serienmail.html`,
`templates/verein.html` und die zugehörigen Handler
(`app/rechnung.go`, `app/verein.go`, `service/rechnung_pdf.go`). Die
erzeugte Rechnungs-PDF selbst (Briefkopf, Positionen, Fußzeile) bekommt
damit erstmals eine Sprache — bisher ist sie immer deutsch.

**Blocked by:** 01 (i18n-Infrastruktur)

## Acceptance Criteria

- [ ] Formulartexte in Rechnung, Serienmail und Verein übersetzt
- [ ] Die erzeugte Rechnungs-PDF trägt die zum Erstellzeitpunkt aktive
      Sprache (Betreff, Positionslabel, Steuerhinweis) — geklärt werden
      muss, ob das gewünscht ist oder die PDF unabhängig von der
      UI-Sprache deutsch bleiben soll (Rücksprache mit Nutzer nötig)
- [ ] `go test ./...` grün, `wails build` unter Linux grün
