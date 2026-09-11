Status: ready-for-agent

# 15: Trainingsslots und abgeleitete Trainingsfrequenz

**What to build:** Der Vereinsadmin ordnet einem Mitglied **null bis drei Trainingsslots** zu — die wöchentlichen Termine, an denen es trainiert. Die **Trainingsfrequenz** (1×, 2×, 3× pro Woche) wird aus der Anzahl dieser Slots abgelesen und nirgends gespeichert, damit beides nicht auseinanderlaufen kann. Der Frequenzfilter ersetzt den in Ticket 13 entfernten Beitragsklassen-Filter.

**Blocked by:** 13 (Beitrag individuell) — dieses Ticket baut auf der umstrukturierten Mitgliedschaft auf

## Acceptance Criteria

- [ ] Trainingsslots hängen an der **Mitgliedschaft**, nicht am Mitglied; null bis drei pro Mitgliedschaft
- [ ] Ein Slot ist ein Text aus Wochentag und Uhrzeit ("Samstag 10:30 Uhr"), frei eingebbar
- [ ] Slots lassen sich im Formular hinzufügen und entfernen
- [ ] Der Versuch, einen vierten Slot anzulegen, wird abgewiesen
- [ ] Die Slots eines Mitglieds sind in der Liste sichtbar
- [ ] Die **Trainingsfrequenz ergibt sich aus der Anzahl der Slots** und wird nicht gespeichert. Null Slots bedeuten "keine Frequenz", nicht "1×"
- [ ] Filter über der Liste: alle / 1× / 2× / 3× pro Woche — und er lässt sich mit Rückstands- und Aktivitätsfilter kombinieren
- [ ] Bei einem **Wiedereintritt** bleiben die Slots der alten Mitgliedschaft unberührt; die neue Mitgliedschaft startet ohne Slots
- [ ] Tests am Service-Seam: Ableitung bei 0, 1, 2 und 3 Slots; Abweisen des vierten; Filterkombination; Wiedereintritt lässt alte Slots stehen
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

**Keine gepflegte Gruppen-Tabelle.** Die Werteliste der bestehenden Excel enthält Tippfehler (`Mitwoch`) und kombinierte Einträge (`Di - 19:30 - Sa - 10:30 Uhr`) — sie ist eine Eintipphilfe, kein sauberer Datensatz. Eine `gruppe`-Tabelle mit Verweis würde eine Pflegedisziplin behaupten, die es nicht gibt. Sie kommt, wenn "zeig mir alle im Samstag-Training" konkret gebraucht wird; siehe Out of Scope in der [Spec](../spec.md).

Entwicklungs-Datenbank vor dem ersten Start löschen.
