Status: ready-for-human

# 16: Restliche Stammdatenfelder

**What to build:** Die letzten sechs Felder der bestehenden Excel-Tabelle bekommen einen Ort in der App: **IBAN**, **Geschlecht**, **Google-Bewertung** und **Digital** am Mitglied, **Anmeldedatum** und **Anmeldegebühr** an der Mitgliedschaft. Damit hält die App jede Spalte der alten Tabelle — das ist die Voraussetzung dafür, dass die Excel danach wirklich abgeschaltet werden kann.

**Blocked by:** 13 (Beitrag individuell) — Anmeldedatum und Anmeldegebühr hängen an der Mitgliedschaft, die dort umstrukturiert wird

## Acceptance Criteria

- [x] Am Mitglied: IBAN, Geschlecht, Google-Bewertung, Digital
- [x] An der Mitgliedschaft: Anmeldedatum, Anmeldegebühr
- [x] Alle sechs sind im Anlegen- und im Bearbeiten-Formular vorhanden und optional
- [x] Die **IBAN** wird als reiner Text gespeichert: **keine Validierung**, keine Formatprüfung, keine SEPA-Erzeugung ([ADR-0006](../../../docs/adr/0006-rueckstand-statt-bezahlt-bis.md))
- [x] **Geschlecht** ist ein Freitextfeld, kein erzwungenes Dropdown — die Excel-Werteliste ist eine Eintipphilfe, keine Einschränkung
- [x] **Google-Bewertung** ist zweiwertig: hat bewertet / hat nicht bewertet
- [x] **Digital** wird als reiner Freitext geführt, **ohne Semantik im Modell**. Die Bedeutung der Spalte ist unbekannt; sie fährt mit, damit beim Import keine Daten verloren gehen. Keine Prüfung, kein Dropdown, kein Glossareintrag
- [x] Die **Anmeldegebühr** wird in Cent gespeichert und als Euro-Betrag eingegeben; sie ist ein historischer Wert und wird nicht neu berechnet
- [x] Das **Anmeldedatum** ist vom Eintritt getrennt und liegt in der Regel davor
- [x] Die Listenzeile bleibt lesbar: die neuen Felder erscheinen im Formular, in der Liste nur dort, wo sie dem Überblick dienen
- [x] Tests am Service-Seam: Anlegen und Bearbeiten mit allen Feldern, mit leeren Feldern, Euro-nach-Cent bei der Anmeldegebühr
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux — Linux geprüft, Windows steht aus

## Notes

Sobald die Bedeutung von `Digital` bekannt ist, ist das eine eigene, kleine Änderung — Feld umbenennen und einen Typ geben. Bis dahin ist Freitext bewusst die richtige Antwort: Daten behalten, ohne eine Bedeutung zu erfinden.

Entwicklungs-Datenbank vor dem ersten Start löschen.

## Comments

**Umgesetzt.** Die sechs Felder liegen an den Stellen, an denen sie fachlich hängen:

- Am Mitglied: `iban`, `geschlecht`, `google_bewertung`, `digital` — alle vier ungeprüft und freiwillig.
- An der Mitgliedschaft: `anmeldedatum` und `anmeldegebuehr_cents`, gebündelt als `service.Anmeldung`. Die beiden reisen zusammen wie die drei Felder der `Anschrift`: sie beschreiben denselben Vorgang und werden zusammen geändert. Dadurch drückt ein einzelner Zeiger im Patch beides aus — `nil` heißt „nicht angerührt", ein Zeiger auf die leere Anmeldung „nichts mehr erfasst".

Zwei Entscheidungen, die im Ticket nicht wörtlich standen:

- **Das Anmeldedatum bleibt beim Bearbeiten änderbar**, anders als Geburtsdatum und Eintritt. Es begrenzt keinen Zeitraum, gegen den Aus- und Wiedereintritt prüfen, sondern hält einen Vorgang fest — und ein falsch abgetippter Vorgang muss sich korrigieren lassen. Geprüft wird dabei *nicht*, dass es vor dem Eintritt liegt: „in der Regel davor" ist keine Regel.
- **Ein Wiedereintritt führt die Anmeldung nicht fort** — wie die Trainingsslots beginnt der neue Zeitraum leer. Die Werte des alten abzuschreiben hieße, eine Gebühr zu behaupten, die niemand gezahlt hat.

In der Liste steht von den sechs nur die **Google-Bewertung**, als Stern am Namen statt als siebte Spalte: die Frage „wen kann ich noch fragen" beantwortet man durchsehend, die übrigen fünf machten die Zeile nur breiter.

`CONTEXT.md` hat einen Eintrag **Anmeldegebühr** bekommen — der Begriff war unter *Beitrag* bereits als Abgrenzung genannt, aber nirgends definiert. `Digital` hat bewusst keinen bekommen.

**Geprüft:** `go test ./...` grün, `go vet` sauber, `wails build` unter Linux erfolgreich. Die Templates sind über einen temporären `httptest`-Durchlauf gerendert worden (Anlegen, Vorbefüllen, Leeren, Fehlerfälle) — der Testlauf selbst ist wieder entfernt, weil `app/` laut CLAUDE.md nicht unit-getestet wird.

**Offen für den Menschen:** `wails dev` und `wails build` unter Windows. Und: die Entwicklungs-Datenbank muss vor dem ersten Start gelöscht werden — das Schema benutzt `CREATE TABLE IF NOT EXISTS`, eine bestehende Datei behält also ihre alten Spalten und jede Abfrage scheitert dann an „no such column".
