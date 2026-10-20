# 02: Kind-Status

**What to build:** Ein Mitglied ist *Kind*, wenn es am heutigen Tag jünger als 18 ist. Der Status wird bei jedem Lesen aus dem Geburtsdatum abgeleitet und nie gespeichert (wie die Trainingsfrequenz). Hat das Mitglied kein Geburtsdatum, zählt ein von Hand im Formular gesetztes Kennzeichen an der Person. Der Status ist über die Service-API lesbar; eine Spalte oder ein Filter in der Mitgliederliste kommt nicht dazu.

**Blocked by:** None (can start immediately)

**Status:** done

- [ ] Mit Geburtsdatum: Kind genau bis zum Tag vor dem 18. Geburtstag (Stichtagstests um den Geburtstag, Schaltjahr 29.02. eingeschlossen)
- [ ] Ohne Geburtsdatum: Kind entspricht dem Handkennzeichen
- [ ] Mit Geburtsdatum wird ein gesetztes Handkennzeichen ignoriert
- [ ] Das Formular bietet das Handkennzeichen nur an, solange kein Geburtsdatum gesetzt ist (Beschriftung in Deutsch und Englisch)
- [ ] Tests über die Service-API; `go test ./...` grün
