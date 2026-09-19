# 06: Dokumentablage: Bestätigung unverändert

**What to build:** Kein Code-Umbau — dieses Ticket dokumentiert und
bestätigt formal, dass die Dokumentablage (Vertrag) bei ihrer bestehenden
Statuskarte bleibt, nachdem Prototyp-Runde 4 das geprüft und bestätigt hat.

**Blocked by:** 01 (Fundament: Akzentfarbe, Navigation, Hauptbereich-Breite)

**Status:** ready-for-human

- [x] Dokumentablage zeigt weiterhin die bestehende Statuskarte mit
      Datei-Aktionen (Upload/Entfernen/Export), keine strukturelle Änderung
- [x] Trägt die neue Akzentfarbe aus Ticket 01 konsistent (nur Farbtoken,
      keine Struktur)
- [x] Verweis auf die Verdict-Datei bzw. den Prototyp-Branch
      `prototype/dokument-ablage` ist in den Ticket-Comments festgehalten,
      damit ein späterer Leser nicht erneut über eine Umgestaltung
      nachdenkt
- [x] `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei
- [x] Manueller Smoke-Test in `wails dev`: Dokumentablage an einer
      Mitgliedschaft mit und ohne abgelegten Vertrag ansehen

## Comments

### Kontext

Teil des UI-Refactorings, siehe [Spec](../spec.md). Prototyp-Runde 4
(Branch `prototype/dokument-ablage`, Commits `5fb7a4d`/`2c9469a`) hat drei
Varianten für `templates/dokument.html` verglichen und **Variante A (die
bestehende Statuskarte)** bestätigt — kein Fold nötig, `main` blieb dadurch
unverändert. Dieses Ticket macht diese Entscheidung im Tracker explizit,
statt sie nur in einer Verdict-Datei auf einem Branch stehen zu lassen.

### Verifikation (2026-09-19)

`templates/dokument.html` ist strukturell unverändert gegenüber Variante A
aus dem Prototyp: eine umrandete Statuskarte je Mitgliedschafts-Zeitraum mit
Textknöpfen für Exportieren/Entfernen/Hochladen bzw. Ersetzen. Der
Vollständigkeitsanker für die Entscheidung ist
`.scratch/prototyp-dokument-ablage-verdict.md` auf Commit `2c9469a`
(Branch `prototype/dokument-ablage`) — Verdict: „A — die bestehende
Statuskarte bleibt.“

Geprüft:

- Die drei Aktions-Buttons (Exportieren, Entfernen, Hochladen/Ersetzen)
  nutzen `focus:ring-2 focus:ring-[var(--akzent-ring-40)]` — dieselbe
  Akzentfarbe wie der Rest der App, keine feststehende Rot- oder
  Blau-Utility-Klasse.
- Die einzige verbliebene `red-*`-Stelle in der Datei ist die
  Fehlerliste (`border-red-700 bg-red-50 text-red-800`) — bewusst
  unverändert, siehe Spec-Ausnahme für Formular-Fehlerlisten.
- `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei.
- Smoke-Test per Wegwerf-`httptest` (`app.New` + `httptest.NewServer`,
  wie in Ticket 01 vorgesehen, da keine GUI-Umgebung zur Verfügung
  steht): ein Mitglied angelegt, `/api/mitglied/1/formular` einmal ohne
  und einmal nach Hochladen eines Test-PDFs über
  `/api/mitgliedschaft/{id}/vertrag` abgerufen. Beide Zustände zeigen die
  Statuskarte mit korrekt gesetzter Akzentfarbe und ohne Rot-Klassen
  außerhalb der Fehlerliste; die Datei wurde danach wieder entfernt.

Keine Code-Änderung nötig — das Ticket schließt als reine Verifikation,
wie schon Ticket 01.
