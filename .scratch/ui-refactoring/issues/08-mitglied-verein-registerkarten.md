# 08: Mitglied & Verein: Registerkarten statt Stapelung

**What to build:** Die Mitglied-Bearbeiten-Seite (Stammdaten / Verträge /
Rechnung, drei eigenständige Formulare) und das Verein-Formular (Anschrift &
Kontakt / Bankverbindung / Rechnungstext, drei Gruppenboxen in einem
Formular) zeigen ihre Bereiche als Registerkarten statt gestapelt
untereinander.

**Blocked by:** 01 (Fundament: Akzentfarbe, Navigation, Hauptbereich-Breite),
03 (Mitglied-Formular: Gruppenboxen), 04 (Restformulare: Gruppenboxen)

**Status:** ready-for-human

- [x] Mitglied-Bearbeiten zeigt drei Registerkarten (Stammdaten / Verträge /
      Rechnung); die vier Gruppenboxen aus Ticket 03 bleiben innerhalb von
      Stammdaten unverändert zusammen
- [x] Mitglied-Anlegen zeigt weiterhin keine Registerkarten (keine
      Mitgliedschaft, an der Verträge/Rechnung hängen könnten) — identisch
      zum bisherigen Verhalten
- [x] Verein zeigt drei Registerkarten (Anschrift & Kontakt / Bankverbindung /
      Rechnungstext), Feldinhalt unverändert
- [x] Keine Feld- oder Validierungsänderung an einem der beiden Formulare
- [x] Die eigenständige Rechnung (`GET /api/rechnung`) und das
      Trainingstermin-Formular bleiben bewusst bei Gruppenboxen ohne
      Registerkarten (siehe ADR-0011)
- [x] `go build ./...`, `go vet ./...`, `go test ./...` und `wails build`
      laufen fehlerfrei
- [x] Manueller Smoke-Test in `wails dev`: nicht möglich in dieser
      Umgebung (keine GUI) — stattdessen per Wegwerf-`httptest` gegen die
      echten Handler geprüft, wie in Ticket 03/04 vorgemacht

## Comments

### Kontext

Teil des UI-Refactorings, siehe [Spec](../spec.md). Anlass war eine
Nutzer-Rückmeldung: die Mitglied-Bearbeiten-Seite reihte Stammdaten,
Verträge und Rechnung als drei eigenständige Formulare untereinander — zu
viel Inhalt auf einer Seite. Ein Grillen-Protokoll hat den Geltungsbereich
geklärt (nur Mitglied und Verein, nicht Trainingstermin oder die
eigenständige Rechnung — Begründung in ADR-0011) und die Aufteilung je Seite
festgelegt. Eine vierte Prototyp-Runde (Branch `prototype/registerkarten`)
hat drei Varianten (Reiter oben/Unterstrich, Pillen, seitliche Reiter)
verglichen; Verdict siehe `.scratch/prototyp-registerkarten-verdict.md` auf
dem Prototyp-Branch — gewählt: seitliche Reiter mit Trennlinie und getöntem
Hintergrund für die Navigationsspalte.

### Umsetzung (2026-09-18)

- `frontend/src/registerkarten.js` (neu): ein einziger delegierter
  Klick-Handler, der `[data-rk-tab]`/`[data-rk-panel]` innerhalb einer
  `[data-registerkarten]`-Gruppe umschaltet (`aria-selected` +
  `hidden`-Attribut). Funktioniert nach jedem htmx-Austausch ohne erneute
  Bindung, wie die Spaltenwahl der Mitgliederliste.
- Die Optik steckt vollständig in Tailwind-Klassen der Templates
  (`aria-selected:`-Varianten) — keine neue CSS-Datei, kein neues
  Custom-CSS, konsistent mit dem Rest der Codebase.
- `templates/mitglied_formular.html`: das bisherige `mitglied-formular`
  wickelt jetzt beim Bearbeiten in die Registerkarten-Struktur; das
  eigentliche Formular ist in ein neues `mitglied-stammdaten-formular`
  ausgelagert, damit es sowohl im Reiter (Bearbeiten) als auch unverändert
  ohne Reiter (Anlegen) gerendert werden kann.
- `templates/verein.html`: die drei Gruppenboxen liegen jetzt in
  Registerkarten-Panels innerhalb desselben `<form>` — ein einziges Formular
  bleibt es, nur mit umschaltbarer Sichtbarkeit; unproblematisch, weil kein
  Feld Pflicht ist (siehe ADR-0011 zur eigenständigen Rechnung, wo das
  anders liegt).
- `templates/dokument.html` und `templates/rechnung.html`: das feste bzw.
  bedingte `mt-6` der eingebetteten Verträge-/Rechnung-Blöcke entfernt — es
  stammte aus der alten Stapel-Anordnung und wäre in der eigenen
  Registerkarte ein unnötiger oberer Abstand gewesen.
- Verifiziert per Wegwerf-httptest (`app.New` + `a.Handler().ServeHTTP`,
  danach wieder entfernt): `GET /api/mitglied/{id}/formular` zeigt alle drei
  Reiter- und Panel-Marker sowie den aktiven Standard-Reiter; `GET
  /api/mitglied/formular` (Anlegen) zeigt keine Registerkarten-Struktur;
  `GET /api/verein` zeigt alle drei Reiter- und Panel-Marker.
- `go build ./...`, `go vet ./...`, `go test ./...` und `wails build`
  (Linux) laufen fehlerfrei.
