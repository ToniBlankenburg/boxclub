Status: ready-for-agent

# 22: Excel-Import auf den Terminkatalog umstellen

**What to build:** Der Import ordnet die Freitexte aus `Training - 1/2/3` **vorhandenen** Trainingsterminen zu und legt selbst keine an. Was er nicht zuordnen kann, geht in den Fehlerbericht aus Ticket 09 — dieselbe Stelle, an der heute schon unlesbare Beiträge und widersprüchliche Frequenzen landen.

**Blocked by:** 21
**Siehe:** [ADR-0008](../../../docs/adr/0008-trainingstermine-als-wochenplan.md)

## Acceptance Criteria

- [ ] Der Importer ordnet per Textvergleich zu: Groß-/Kleinschreibung und Leerraum sind egal, alles andere muss stimmen
- [ ] Ein **nicht** zuordenbarer Wert erzeugt einen Eintrag im Fehlerbericht mit Zeile, Spalte und dem gelesenen Text — und **legt keinen Termin an**
- [ ] Ein nicht zuordenbarer Wert verhindert **nicht** die Übernahme des Mitglieds; es kommt ohne diesen Termin herein, wie heute bei anderen Feldfehlern
- [ ] Der Wert `kein` bleibt, was er ist: kein Termin, keine Meldung (`importer.keinTraining`)
- [ ] Die bestehende Prüfung „angegebene Frequenz widerspricht der Anzahl der Termine" bleibt und zählt die **zugeordneten** Termine
- [ ] **Archivierte** Termine werden nicht zugeordnet — ein Import darf keine Zeiten vergeben, die es nicht mehr gibt. Der Wert geht in den Fehlerbericht
- [ ] Der Fehlerbericht sagt dem Admin, was zu tun ist: erst den Stundenplan pflegen, dann erneut importieren
- [ ] Tests in `importer/` und `service/import_test.go` auf den Katalog umgeschrieben
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

**Erwartungshaltung beim ersten Import:** die Excel enthält Schreibvarianten, die nicht treffen werden („Sa 10:30" gegen „Samstag 10:30"). Das ist eingepreist — eine Aufräumrunde nach dem ersten Lauf gehört zum Verfahren, nicht zu den Fehlern.

**Warum nicht automatisch anlegen:** ADR-0008 → Alternativen. Jede Schreibvariante würde ein Stundenplaneintrag, und aufgeräumt wird das nie.

**Die Import-Ansicht braucht einen Hinweis**, wenn der Stundenplan leer ist: sonst importiert jemand 200 Mitglieder und bekommt 400 Fehlermeldungen, ohne zu verstehen, warum.
