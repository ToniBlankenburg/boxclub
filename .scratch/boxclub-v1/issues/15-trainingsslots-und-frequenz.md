Status: ready-for-human

# 15: Trainingsslots und abgeleitete Trainingsfrequenz

**What to build:** Der Vereinsadmin ordnet einem Mitglied **null bis drei Trainingsslots** zu — die wöchentlichen Termine, an denen es trainiert. Die **Trainingsfrequenz** (1×, 2×, 3× pro Woche) wird aus der Anzahl dieser Slots abgelesen und nirgends gespeichert, damit beides nicht auseinanderlaufen kann. Der Frequenzfilter ersetzt den in Ticket 13 entfernten Beitragsklassen-Filter.

**Blocked by:** 13 (Beitrag individuell) — dieses Ticket baut auf der umstrukturierten Mitgliedschaft auf

## Acceptance Criteria

- [x] Trainingsslots hängen an der **Mitgliedschaft**, nicht am Mitglied; null bis drei pro Mitgliedschaft
- [x] Ein Slot ist ein Text aus Wochentag und Uhrzeit ("Samstag 10:30 Uhr"), frei eingebbar
- [x] Slots lassen sich im Formular hinzufügen und entfernen
- [x] Der Versuch, einen vierten Slot anzulegen, wird abgewiesen
- [x] Die Slots eines Mitglieds sind in der Liste sichtbar
- [x] Die **Trainingsfrequenz ergibt sich aus der Anzahl der Slots** und wird nicht gespeichert. Null Slots bedeuten "keine Frequenz", nicht "1×"
- [x] Filter über der Liste: alle / 1× / 2× / 3× pro Woche — und er lässt sich mit Rückstands- und Aktivitätsfilter kombinieren
- [x] Bei einem **Wiedereintritt** bleiben die Slots der alten Mitgliedschaft unberührt; die neue Mitgliedschaft startet ohne Slots
- [x] Tests am Service-Seam: Ableitung bei 0, 1, 2 und 3 Slots; Abweisen des vierten; Filterkombination; Wiedereintritt lässt alte Slots stehen
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux — Linux grün, **Windows steht aus** (siehe Kommentar)

## Notes

**Keine gepflegte Gruppen-Tabelle.** Die Werteliste der bestehenden Excel enthält Tippfehler (`Mitwoch`) und kombinierte Einträge (`Di - 19:30 - Sa - 10:30 Uhr`) — sie ist eine Eintipphilfe, kein sauberer Datensatz. Eine `gruppe`-Tabelle mit Verweis würde eine Pflegedisziplin behaupten, die es nicht gibt. Sie kommt, wenn "zeig mir alle im Samstag-Training" konkret gebraucht wird; siehe Out of Scope in der [Spec](../spec.md).

Entwicklungs-Datenbank vor dem ersten Start löschen.

## Comments

### Umgesetzt (Claude, 2026-09-11)

- `service/trainingsslot.go` hält Obergrenze (`MaxTrainingsslots = 3`), die abgeleitete `Trainingsfrequenz` samt ihrer Bezeichnung, den `Frequenzfilter` und das Normalisieren/Prüfen der Slot-Liste. Die Ableitung steht genau einmal in `trainingsfrequenzAus`; `Mitgliedschaft.Trainingsfrequenz()` und `Listeneintrag.Trainingsfrequenz()` rufen sie beide auf.
- Schema: neue Tabelle `trainingsslot(id, mitgliedschaft_id, bezeichnung)` mit Fremdschlüssel auf `mitgliedschaft` und `ON DELETE CASCADE`, dazu ein Index auf `mitgliedschaft_id`. Die Slots hängen damit an der Mitgliedschaft, nicht an der Person.
- Geschrieben wird die Slot-Liste einer Mitgliedschaft immer als Ganzes (löschen, neu einfügen): Slots haben außer ihrem Text nichts, woran sich ein einzelner wiedererkennen ließe. Hinzufügen und Entfernen sind für den Nutzer derselbe Vorgang. Leere Felder werden verworfen, umschließender Leerraum abgeschnitten.
- `Rejoin` kopiert weiterhin den Beitrag, aber ausdrücklich **keine** Slots: die alten galten für einen Zeitraum, der vorbei ist.
- Formular: drei feste Slot-Felder. Ein „Hinzufügen"-Knopf wäre ein Versprechen, das der Service zu Recht bricht — drei sind das Maximum, also stehen alle drei von vornherein da. Liste: neue Spalte „Training" mit den Slots und der abgeleiteten Frequenz im Tooltip; die aufklappenden Zeilen spannen jetzt `colspan="5"`.
- Filterleiste: neue Auswahlliste „Frequenz", kombinierbar mit Rückstands- und Aktivitätsfilter. Ausgewertet wird sie wie der Rückstand in Go, nicht in SQL ([ADR-0004](../../../docs/adr/0004-suche-und-filter-im-speicher.md)) — eine abgeleitete Größe steht in keiner Spalte.
- `CLAUDE.md` (drittes Schema-Element) und `CONTEXT.md` (Slots gehören zur Mitgliedschaft; null Slots sind keine Frequenz) nachgezogen.

**Keine Löschung der Entwicklungs-Datenbank nötig.** Anders als in den Notes angenommen ist die Schemaänderung rein additiv — es kommt eine Tabelle dazu, bestehende Spalten bleiben unberührt. Eine vorhandene `boxclub.db` läuft weiter; ihre Mitglieder haben dann schlicht keine Slots.

**Spannung mit [ADR-0005](../../../docs/adr/0005-beitrag-individuell-statt-beitragsklasse.md):** Dort steht „Die Trainingsfrequenz wird eine eigene, davon unabhängige Eigenschaft", und in den Alternativen „Die Frequenz gehört als eigene Eigenschaft **ans Mitglied**". Gemeint war dort Unabhängigkeit vom *Beitrag*, nicht eine gespeicherte Angabe — Spec („Abgeleitete Werte") und Glossar verlangen die Ableitung aus den Slots der Mitgliedschaft, und dieses Ticket setzt das um. Der Wortlaut von ADR-0005 liest sich inzwischen aber so, als sei die Frequenz ein eigenes Feld. Das ist nicht stillschweigend übergangen, sondern hier vermerkt; ob ein klarstellender Nachtrag oder eine eigene ADR daraus wird, entscheidet der Mensch.

**Offen:** `go test ./...`, `go vet` und `gofmt` sind sauber, `wails dev` und `wails build` unter **Linux** grün. **Windows ist auf dieser Maschine nicht prüfbar** und steht noch aus — deshalb `ready-for-human`.
