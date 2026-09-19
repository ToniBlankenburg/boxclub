# 07: Freigabe: Ticket 11 (Mac-Release-Build) entsperren

**What to build:** Bestätigt, dass alle UI-Refactoring-Tickets (01–06)
abgeschlossen sind, führt eine app-weite Abschlussprüfung durch und
aktualisiert den Blocker-Vermerk in `boxclub-v1/issues/11-mac-release-build.md`,
damit dieses Ticket starten darf.

**Blocked by:** 02 (Dashboard: Kennzahlenleiste), 03 (Mitglied-Formular:
Gruppenboxen), 04 (Trainingstermin-, Rechnung- & Vereinsdaten-Formulare:
Gruppenboxen), 05 (Excel-Import-Bericht: Kennzahlenleiste), 06
(Dokumentablage: Bestätigung unverändert) — 01 transitiv über alle fünf

**Status:** ready-for-human

- [x] Alle Tickets 01–06 stehen auf `ready-for-human` oder besser
- [x] `go build ./...`, `go vet ./...`, `go test ./...` und `wails build`
      laufen app-weit fehlerfrei
- [x] Vollständiger manueller Durchlauf aller 6 Bereiche in `wails dev` ohne
      visuelle Inkonsistenz (z. B. vergessene Rot-Reste außerhalb der zwei
      bewusst roten Warnstellen)
- [x] `boxclub-v1/issues/11-mac-release-build.md`: Blocker-Vermerk zum
      UI-Refactoring ist aktualisiert (z. B. "erledigt, siehe
      `.scratch/ui-refactoring/`" statt offen)

## Comments

### Kontext

Schließt das UI-Refactoring formal ab, siehe [Spec](../spec.md). Kein
eigener Code-Umbau — reine Verifikation und Freigabe.

### Verifikation (2026-09-19)

- Tickets 01–06 stehen alle auf `ready-for-human` (geprüft per Status-Zeile
  in jeder Ticket-Datei).
- `go build ./...`, `go vet ./...`, `go test ./...` laufen app-weit
  fehlerfrei. `wails build` (Ziel: linux/amd64, das verfügbare
  Entwicklungs-Target) erzeugt erfolgreich `build/bin/boxclub` — macOS/
  Windows-Targets sind laut CLAUDE.md ohnehin nicht Teil der
  Entwicklungsumgebung.
- Grep über `templates/`, `frontend/index.html` und `frontend/src/style.css`
  nach `red-*`-Klassen findet nur die bereits einzeln begründeten Stellen:
  die fünf Formular-Fehlerlisten (`mitglied_formular.html`,
  `trainingstermine.html`, `rechnung.html`, `verein.html`, `dokument.html`),
  das Rückstand-Kennzeichen samt Inline-Fehlermeldungen in
  `mitglieder_liste.html`, sowie dessen zwei Ableger auf Zahlenebene
  (`dashboard.html` „Im Rückstand“, `import.html` „Gescheitert“ — beide
  bereits in Ticket 02/05 als Warnfarbe akzeptiert). Keine unbeachtete
  Rot-Stelle gefunden.
- Ohne GUI-Umgebung wurde der Durchlauf aller sechs Bereiche per
  Wegwerf-`httptest` geprüft (wie schon bei Ticket 01/05/06): jede der
  sechs `/api/...`-Routen liefert genau einen aktiven
  `aria-current="page"`-Navigationseintrag, keine Doppelung und kein
  fehlender Bereich. Testdatei danach wieder entfernt.
- `boxclub-v1/issues/11-mac-release-build.md` ist aktualisiert: der
  UI-Refactoring-Blocker ist als erledigt vermerkt, das Ticket ist
  entsperrt.

Keine Code-Änderung nötig — das UI-Refactoring schließt hiermit formal ab.
