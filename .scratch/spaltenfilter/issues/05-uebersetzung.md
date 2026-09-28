# 05: Übersetzung der neuen Beschriftungen ins Englische

**What to build:** Alle in den Tickets 01–04 neu eingeführten,
nutzersichtbaren Beschriftungen (Zahnrad-Menü-Inhalte, Filterlabel je Spalte,
Fallback-Element für ausgeblendete Spalten, Reset-Button) bekommen englische
Übersetzungen im bestehenden `i18n`-Katalog, nach dem etablierten
`i18n.Text(sprache, "...")`-Muster. Die App bleibt dadurch vollständig
zweisprachig, wie es das Mehrsprachigkeits-Feature vorsieht.

**Blocked by:** 01, 02, 03, 04 (braucht den vollständigen, finalen Satz an
neuen Beschriftungen)

**Status:** ready-for-agent

- [ ] Jede in 01–04 neu eingeführte Beschriftung hat einen deutschen und
      einen englischen Katalog-Eintrag
- [ ] Keine hartkodierten deutschen Strings in den neuen Template-/Go-Stellen
      übrig
- [ ] Manueller Smoke-Test: Umschalten auf Englisch zeigt alle neuen Zahnrad-
      Menüs, Filterlabel, das Fallback-Element und den Reset-Button korrekt
      übersetzt
- [ ] `wails dev` und `wails build` laufen weiterhin auf Windows und Linux
