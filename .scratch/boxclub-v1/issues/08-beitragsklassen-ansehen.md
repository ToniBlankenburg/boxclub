Status: ready-for-agent

# 08: Beitragsklassen ansehen

**What to build:** Eine eigene Ansicht listet die verfügbaren **Beitragsklassen** mit Namen und Monatspreis. In v1 read-only — Pflege der Klassen selbst ist bewusst nicht enthalten.

**Blocked by:** 02 (SQLite-Bootstrap)

## Acceptance Criteria

- [ ] Ein Navigationseintrag "Beitragsklassen" öffnet eine schlichte Tabelle
- [ ] Jede Klasse wird mit Name und Preis pro Monat angezeigt (Preis lesbar formatiert, z. B. "60,00 €")
- [ ] `MemberService.ListBeitragsklassen` liefert alle Klassen (unabhängig davon, ob Mitglieder zugeordnet sind)
- [ ] Test am Seam: nach dem Seed sind genau die zwei erwarteten Klassen mit den definierten Preisen sichtbar

## Notes

Klein aus Absicht — die Ansicht ist eher eine Referenz für den Admin und liefert später den natürlichen Andockpunkt, wenn Klassen editierbar werden sollen (nicht v1).
