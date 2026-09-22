Status: ready-for-agent

# 02: Mitgliederliste, Filterleiste, Mitglied-Formular übersetzt

**What to build:** Alle Texte in `templates/mitglieder_liste.html`,
`templates/mitglied_formular.html` und die zugehörigen `Beschriftung`-Literale
in `app/app.go` (Spaltenköpfe, Filteroptionen, Feldbeschriftungen) wandern in
den Katalog, inklusive der Fachbegriffe (Mitglied → Member, Rückstand →
Arrears, Anmeldegebühr → Registration fee, Kündigung → Cancellation, Eintritt
→ Join date, Austritt → Exit date, Ruhend → Suspended, Trainingsfrequenz →
Training frequency, …).

**Blocked by:** 01 (i18n-Infrastruktur)

## Acceptance Criteria

- [ ] Spaltenköpfe, Filteroptionen (`spalte*`, `filteroption`-Listen),
      Formularbeschriftungen übersetzt
- [ ] Meldungstexte (`meldung{Text: "..."}`) an den betroffenen Stellen
      übersetzt
- [ ] `go test ./...` grün, `wails build` unter Linux grün
