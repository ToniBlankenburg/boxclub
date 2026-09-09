Status: ready-for-agent

# 11: Mac-Release-Build

**What to build:** Der erste ship-fähige `.app`-Bundle für macOS, den der Vereinsadmin auf seinem Mac starten kann. Enthält eine kleine Vorab-Entscheidung ("gelegentlicher Mac vs. `macos-latest`-GitHub-Actions-Runner") plus die eigentliche Umsetzung.

**Blocked by:** 06 (Suche & Filter), 07 (Aus-/Wiedereintritt), 08 (Beitragsklassen ansehen) — die letzten Kern-Feature-Tickets

## Acceptance Criteria

- [ ] Entscheidung dokumentiert (Kommentar in diesem Ticket): Mac-Build auf gelegentlich zugänglichem Mac vs. `macos-latest`-Runner auf GitHub Actions — mit Begründung
- [ ] `wails build` erzeugt ein `.app`-Bundle für macOS (arm64 oder universal)
- [ ] Das Bundle startet auf einem Mac ohne installierten Go-Compiler und ohne Node.js
- [ ] SQLite-Datei liegt an einem für macOS sinnvollen Ort (typischerweise `~/Library/Application Support/boxclub/`, nicht neben dem `.app`)
- [ ] Rauchtest im Bundle grün: Mitglied anlegen → in Liste sehen → `bezahlt_bis` setzen → Status wechselt auf grün
- [ ] Signierung mindestens als Ad-hoc-Signatur; volle Notarisierung (Apple Developer ID) ist v1.5, nicht v1

## Notes

Ist explizit **Ship-Vorbereitung**, kein Kern-Feature — daher am Ende der Blockerkette. Bewusst blockiert durch 06/07/08 statt "durch alle Tickets", weil 04, 05, 09, 10 orthogonal genug sind, dass Feature-Vollständigkeit durch die letzten drei Kern-Tickets ausreichend beschrieben ist.
