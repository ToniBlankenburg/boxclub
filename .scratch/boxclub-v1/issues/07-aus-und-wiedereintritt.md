Status: ready-for-agent

# 07: Aus- und Wiedereintritt

**What to build:** Ein Mitglied kann mit einem Datum als ausgetreten markiert werden — es verschwindet aus der Standardansicht, der Datensatz bleibt aber erhalten. Ein ehemaliges Mitglied kann wieder eintreten; dabei entsteht eine **neue** `Mitgliedschaft`-Zeile, nicht ein neues `Mitglied`.

**Blocked by:** 03 (Mitgliederliste)

## Acceptance Criteria

- [ ] Aktion "Austritt eintragen" fragt nach einem Austrittsdatum und setzt `austritt` auf der aktiven Mitgliedschaft
- [ ] Nach dem Austritt verschwindet das Mitglied aus der Standardansicht "nur aktive", bleibt aber sichtbar unter dem Filter "auch ehemalige"
- [ ] Aktion "Wiedereintritt" ist auf ehemaligen Mitgliedern verfügbar (nicht auf aktiven); sie fragt nach einem Eintrittsdatum und legt eine **neue** `mitgliedschaft`-Zeile an (`austritt` NULL)
- [ ] Bei Wiedereintritt bleibt der `mitglied`-Datensatz derselbe — Stammdaten, ID, Historie
- [ ] `MemberService.MarkExit(id, datum)` und `MemberService.Rejoin(id, datum)` am Seam
- [ ] Test am Seam: Aus- gefolgt von Wiedereintritt → `mitglied` unverändert, zwei `mitgliedschaft`-Zeilen in der DB
- [ ] Test am Seam: doppelter `MarkExit` auf dieselbe aktive Mitgliedschaft ist ein Fehler
- [ ] Test am Seam: `Rejoin` auf ein bereits aktives Mitglied ist ein Fehler
- [ ] Test am Seam: `MarkExit` mit Datum vor Eintrittsdatum ist ein Fehler
