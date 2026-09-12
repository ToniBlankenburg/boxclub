Status: ready-for-agent

# 26: Dashboard mit Monatssoll und Mitgliederzahlen

**What to build:** Ein eigener Menüpunkt mit den Zahlen, die Trainer und Admin im Blick haben wollen: wie viele Mitglieder, und was diesen Monat eingezogen wird.

**Blocked by:** —
**Siehe:** [ADR-0009](../../../docs/adr/0009-rechnungen-und-monatssoll-ohne-zahlungsmodell.md), `CONTEXT.md` → Monatssoll

## Acceptance Criteria

- [ ] Eigener Menüpunkt; die **Mitgliederliste bleibt Startseite**
- [ ] **Monatssoll** des laufenden Monats: Summe der Beiträge aktiver Mitgliedschaften **einschließlich** derer in der Kündigungsfrist, **ohne** ruhende, **ohne** noch nicht begonnene
- [ ] Ruhende getrennt daneben ausgewiesen, mit Anzahl und Betrag („davon ruhend gestellt: 3 · 135 €") — sonst ist die Summe unerklärt kleiner als die Mitgliederzahl vermuten lässt
- [ ] Mitgliederzahlen nach Status: aktiv, in Kündigungsfrist, neu (Eintritt steht bevor), ruhend, ausgetreten
- [ ] Anzahl der Mitglieder **im Rückstand**, verlinkt auf die Liste mit gesetztem Rückstandsfilter — die Zahl ist nur nützlich, wenn man von ihr aus weiterarbeiten kann
- [ ] Neu eingetreten und ausgetreten **in diesem Monat**
- [ ] **Kein Verlaufsdiagramm** und keine Zeitreihe
- [ ] Die Seite heißt die Summe **Monatssoll**, nirgends „Einnahmen" oder „Umsatz"
- [ ] Service-Tests für die Berechnung, insbesondere die Ränder: ruhend zählt nicht, Kündigungsfrist zählt, Eintritt am Ersten des Folgemonats zählt nicht, Austritt heute zählt nicht
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

**Es ist ein Soll, kein Ist.** Was tatsächlich einging, weiß die Bank; v1 kennt keine Zahlungen. Der Name ist die Gegenmaßnahme gegen die Fehllesung und darf in der Oberfläche nicht wegvereinfacht werden.

**Warum kein Verlauf:** die Datenbank führt je Mitgliedschaft **einen** Beitrag, den heutigen. Eine Kurve über zwölf Monate wäre zwölfmal die heutige Summe, nur nach Ein- und Austritten gefiltert — ein Diagramm, das aussieht wie eine Messung und eine Hochrechnung ist. Wenn der Verlauf einmal wirklich gebraucht wird, braucht es zuerst eine datierte Beitragshistorie, und das ist ein Umbau am Mitgliedschaftsmodell.

**Rechnet nur, speichert nichts.** Das Dashboard braucht kein neues Feld und keine neue Tabelle; es liest Beitrag, Status und `ruhend`.
