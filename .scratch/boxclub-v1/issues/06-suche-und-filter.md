Status: ready-for-human

# 06: Suche & Filter

**What to build:** Über der Mitgliederliste liegen ein Suchfeld und Filter-Steuerungen. Suche trifft auf Name, E-Mail oder Telefon. Filter grenzen nach Zahlungsstatus, Beitragsklasse und Aktivität ein. Standardansicht zeigt nur aktive Mitglieder.

**Blocked by:** 03 (Mitgliederliste), 05 (Zahlungsstatus — für den Status-Filter)

## Acceptance Criteria

- [x] Suchfeld findet Mitglieder mit Teiltreffer (case-insensitive) in Vorname, Nachname, E-Mail oder Telefonnummer
- [x] Filter **Zahlungsstatus**: alle / bezahlt / nicht bezahlt
- [x] Filter **Beitragsklasse**: alle / einzelne Klasse (dynamisch aus der DB gelistet)
- [x] Filter **Aktivität**: nur aktive (Standard beim Öffnen) / auch ehemalige
- [x] Suche und Filter kombinieren sich serverseitig in `MemberService.Search(query, filter)` — kein Client-Side-Filtern
- [x] Reset-Aktion setzt Suchfeld und alle Filter auf Standardwerte zurück
- [x] Ergebnisliste aktualisiert sich per htmx-Fragment bei jeder Eingabe oder Filter-Änderung
- [x] Test am Seam für jede einzelne Filterdimension isoliert
- [x] Test am Seam für kombinierte Filter (z. B. "aktive Erwachsen 2×/Woche mit unbezahltem Status")
- [x] Test am Seam: leerer Query mit Filtern angewendet ≙ Filter-only
- [x] Test am Seam: Query mit Sonderzeichen (Umlaute, Bindestriche) findet die passenden Mitglieder

## Comments

**Umgesetzt.** Seam: `MemberService.Search(query, service.Suchfilter)` in
[service/member_service.go](../../../service/member_service.go), Tests in
[service/search_test.go](../../../service/search_test.go) (11 Testfunktionen, jede
Filterdimension einzeln plus Kombinationen). `List()` ist jetzt nur noch
`Search("", Suchfilter{})` — die Standardansicht und "Filter zurücksetzen" sind
damit derselbe Zustand, nicht zwei Definitionen davon.

Vier Entscheidungen, die über das Ticket hinausgingen:

1. **Kein SQL `LIKE`.** Die Spec sah es vor, es trägt aber nicht: SQLite faltet
   Groß-/Kleinschreibung nur im ASCII-Bereich, "öztürk" fände "Öztürk" also nicht.
   Suchbegriff, Zahlungsstatus und Beitragsklasse werden deshalb in Go
   ausgewertet; in SQL bleibt allein die Aktivität, weil sie darüber entscheidet,
   welche Mitgliedschaft eine Zeile überhaupt bekommt. Begründung in
   [ADR-0004](../../../docs/adr/0004-suche-und-filter-im-speicher.md), Spec ist
   entsprechend annotiert.
2. **"Nicht bezahlt" schließt "nicht gesetzt" aus.** Ein Mitglied ohne jede
   Zahlungsangabe darf nicht im Mahn-Filter landen — sonst mahnt der Verein
   jemanden, über den er nichts weiß (CONTEXT.md → Statusanzeige). Eigener Test.
3. **`Eintrag(id)` liefert jetzt auch die Zeile eines Ausgetretenen.** Vorher war
   das ErrNichtGefunden, weil Ausgetretene nie in der Liste standen. Seit sie es
   tun, hätte jede Aktion in einer solchen Zeile — etwa die Schaltfläche "Ändern"
   für `bezahlt_bis` — die Ansicht mit "Dieses Mitglied gibt es nicht mehr."
   abgeräumt. Der Test dazu ist mitgewandert.
4. **`Listeneintrag.Austritt`** ist neu. Ohne ihn wäre in der Ansicht "auch
   Ehemalige" nicht erkennbar, wer ehemalig ist; die Zeile zeigt jetzt eine
   Markierung "ausgetreten <Datum>". Gesetzt ist das Feld nur in Ergebnissen, die
   Ehemalige einschließen.

Zwei Punkte für später:

- **Filter überleben eine Bearbeitung nicht.** Nach Anlegen, Speichern oder einer
  Zahlungsänderung kehrt die Ansicht in den Standard zurück. Das ist bewusst so
  (das geänderte Mitglied soll sichtbar sein und nicht hinter einem Filter
  verschwinden), aber wenn es beim Benutzen stört, wäre das Weiterreichen der
  Filterwerte durch die Formulare der nächste Schritt.
- **Der Klassenfilter listet nur aktive Beitragsklassen.** Würde eine Klasse je
  deaktiviert, während noch Mitglieder ihr zugeordnet sind, ließe sie sich nicht
  mehr auswählen — am ehesten spürbar unter "auch Ehemalige". Heute gibt es keinen
  Weg, eine Klasse zu deaktivieren, deshalb ist der Fall noch nicht erreichbar.
- **Eine Zahlungsänderung unter aktivem Statusfilter lässt die Zeile stehen.**
  Wer unter "Nicht bezahlt" jemanden auf bezahlt setzt, sieht die Zeile mit grüner
  Markierung weiter — sie fällt erst beim nächsten Filterlauf heraus, und die
  Anzahl darüber stimmt bis dahin nicht. Das ist die Folge davon, dass die
  Zahlungspflege nur ihre eigene Zeile zurückgibt (Ticket 05) und damit denselben
  Ursprung wie der Punkt darüber.
- **Die Suche kennt nur einen Begriff.** "anna b" findet nichts, "berger" schon.
  Ein Aufteilen an Leerzeichen (jeder Teil muss irgendwo treffen) wäre eine kleine
  Erweiterung — hier bewusst nicht gemacht, weil das Ticket Teiltreffer pro Feld
  verlangt.

`go test ./...` grün, `wails build -tags webkit2_41` auf Linux grün (Tailwind
findet die neuen Klassen aus `templates/`). Windows-Build steht noch aus.
Die Handler sind wie vereinbart nicht unit-getestet; geprüft wurden sie über einen
Wegwerf-`httptest`-Lauf gegen alle neuen Routen.
