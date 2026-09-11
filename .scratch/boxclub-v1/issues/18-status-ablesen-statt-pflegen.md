Status: ready-for-agent

# 18: Status ablesen statt pflegen

**What to build:** Der Vereinsadmin sieht pro Mitglied dessen **Status** — *Neu*, *Aktiv*, *In Kündigungsfrist* oder *Ausgetreten* — ohne ihn zu pflegen: er wird aus den Datumsfeldern abgelesen. Damit kann er nicht mehr im Widerspruch zu den Daten stehen, wie es in der handgepflegten Excel regelmäßig passiert. Gleichzeitig ändert sich, was "aktiv" bedeutet: ein Mitglied in der Kündigungsfrist bleibt in der Standardansicht, weil es weiter trainiert und weiter zahlt.

**Blocked by:** 17 (Kündigung und ruhend) — der Status *In Kündigungsfrist* braucht das dort angelegte Kündigungsdatum

## Acceptance Criteria

- [ ] Der Status wird **abgeleitet und nirgends gespeichert**:
  - *Neu* — Eintritt liegt in der Zukunft
  - *Aktiv* — Eintritt erreicht, Austritt nicht erreicht
  - *In Kündigungsfrist* — Kündigungsdatum gesetzt, Austritt liegt in der Zukunft
  - *Ausgetreten* — Austritt erreicht
- [ ] *Ruhend* wird **zusätzlich** angezeigt, nicht anstelle des Status — es ist ein Merkmal, kein Lebenszyklus-Zustand
- [ ] **"Aktiv" ändert seine Definition**: bisher "kein Austrittsdatum gesetzt", künftig "Austritt nicht erreicht"
- [ ] Der Aktivitätsfilter zieht nach: ein Mitglied mit **zukünftigem** Austrittsdatum erscheint in der Standardansicht "nur aktive"; erst ab dem Austrittstag verschwindet es dort
- [ ] Der Status ist pro Zeile in der Mitgliederliste sichtbar
- [ ] Randfälle getestet: Eintritt genau heute, Austritt genau heute, Kündigungsdatum ohne Austritt, Austritt in der Zukunft, mehrere Mitgliedschaften nacheinander
- [ ] Die Excel-Werte `Mitglied` und `Aktiv` ergeben denselben Status; `Stillgelegt` und `Inakiv` sind beide *ruhend* und kein eigener Status
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

Hier bereinigt sich die aktive Liste künftig **von selbst**: niemand muss ein Mitglied "auf ausgetreten setzen", das Austrittsdatum erledigt es am Stichtag. Das ist der eigentliche Gewinn gegenüber der Excel, in der der Status eine Spalte war, die jemand von Hand nachziehen musste.
