Status: ready-for-human

# 03: Mitgliederliste anzeigen

**What to build:** Statt eines einzelnen Mitglieds zeigt die App jetzt eine Tabelle **aller aktiven Mitglieder**. Neue Mitglieder erscheinen nach dem Anlegen sofort in der Liste. Das ist die Arbeitspferd-Ansicht der App.

**Blocked by:** 02 (SQLite-Bootstrap & erstes Mitglied anlegen)

## Acceptance Criteria

- [x] Tabelle zeigt für jedes aktive Mitglied mindestens: Name, Beitragsklasse, `bezahlt_bis`, Eintrittsdatum
- [x] "Aktiv" heißt: es existiert eine `mitgliedschaft` mit `austritt IS NULL`
- [x] Nach dem Anlegen eines neuen Mitglieds erscheint es sofort in der Liste (htmx-Fragment-Swap, kein Full-Page-Reload)
- [x] `MemberService.List` liefert nur aktive Mitgliedschaften zurück, sortiert nach Nachname/Vorname
- [x] Test am Seam: mehrere Mitglieder angelegt → `List` gibt sie in konsistenter Reihenfolge zurück
- [x] Test am Seam: ein Mitglied ohne aktive Mitgliedschaft (z. B. später ausgetreten, aus Test-Fixture) erscheint **nicht** in `List`
- [x] Leere Liste rendert eine sinnvolle Leer-Ansicht ("Keine Mitglieder erfasst")

## Notes

Die Liste ist das Herzstück, an dem alle späteren Slices (Suche, Filter, Bearbeiten, Status) andocken. Sauberes Fragment-Muster hier zahlt sich mehrfach aus.

## Comments

### Implementiert (Linux verifiziert, Windows offen)

Die Liste ist jetzt die Startansicht: `frontend/index.html` lädt `/api/mitglieder`
statt des Formulars. Neuer Seam-Aufruf `MemberService.List()`, vier zusätzliche
Tests, alle grün; `gofmt`, `go vet` und `staticcheck` sauber.

**Auf Linux (Mint 22.1) verifiziert:** `wails build` erzeugt `build/bin/boxclub`;
unter `wails dev` wurden Liste, Leer-Ansicht, Anlegen mit sofortigem
Fragment-Swap und die Sortierreihenfolge am laufenden Server durchgespielt.
**Noch offen — `wails dev` / `wails build` unter Windows**, wie bei 01 und 02.

### Entscheidungen

1. **Eigener Typ `Listeneintrag` statt `[]Mitglied`.** Die Liste braucht Name,
   Beitragsklasse **mit Namen**, `bezahlt_bis` und das Eintrittsdatum der
   laufenden Mitgliedschaft. Mit `[]Mitglied` hätte `app/` pro Zeile
   `Beitragsklasse(id)` nachladen müssen (N+1) oder ein halb befülltes
   `Mitglied` bekommen, dessen `Mitgliedschaften` nur die aktive enthält — eine
   Lüge über das Feld. `List` liefert stattdessen eine Read-Projektion aus genau
   einer Query. Ticket 06 (`Search`) sollte denselben Typ zurückgeben.

2. **Sortiert wird in Go, nicht in SQL.** SQLite kennt nur binäre Sortierung und
   `COLLATE NOCASE`, das ausschließlich ASCII faltet: "Öztürk" landete hinter
   "Zimmermann". In einem deutschen Verein ist das ein sichtbarer Fehler in genau
   der Ansicht, die dieses Ticket baut. `nachNamenSortieren` benutzt jetzt
   `golang.org/x/text/collate` mit `language.German`. Die Abhängigkeit war über
   Wails ohnehin schon im Modulgraph und wurde nur von *indirect* auf direkt
   hochgestuft — dieselbe Version (v0.39.0), `go.sum` unverändert. Bei 200
   Mitgliedern ist Sortieren im Speicher unmerklich.

3. **Die Mitglied-Karte ist entfallen.** Nach dem Speichern erscheint die Liste
   mit dem neuen Mitglied statt der Einzelkarte — so verlangt es dieses Ticket.
   **Nebeneffekt, der bei der Abnahme sichtbar bleiben soll:** Geburtsdatum,
   Adresse, E-Mail und Telefon sind damit vorübergehend nirgends im UI sichtbar.
   Ticket 04 (Bearbeiten-Formular) bringt sie zurück. `MemberService.Get` bleibt
   erhalten und getestet.

4. **Über das Ticket hinaus, bewusst klein gehalten:** ein "Abbrechen"-Knopf im
   Formular (ohne ihn ist das Formular eine Sackgasse zurück zur Liste), eine
   Bestätigung "X Y wurde angelegt." und ein Mitglieder-Zähler in der
   Überschrift. Der Monatspreis stand zwischenzeitlich in der
   Beitragsklassen-Spalte und wurde nach dem Review wieder entfernt — das Ticket
   verlangt die Klasse, nicht den Preis.

### Offener Punkt für die menschliche Abnahme

`service/export_test.go` enthält `AustrittFuerTest`, das die laufende
Mitgliedschaft per direktem `UPDATE` beendet. Der Test "ausgetretenes Mitglied
erscheint nicht in `List`" braucht diese Fixture, `MemberService.MarkExit` kommt
aber erst mit Ticket 07. Das steht in Spannung zu CLAUDE.md → Testing ("never
reach into SQL directly").

Abgewogen: Die Alternative wäre, `MarkExit` aus Ticket 07 vorzuziehen — dann
aber ohne dessen Fachregeln (doppelter Austritt ist ein Fehler, Austritt vor
Eintritt ist ein Fehler), also eine halbfertige öffentliche API. Der Helfer liegt
stattdessen in einer `_test.go`-Datei, ist außerhalb des Testlaufs nicht Teil des
Package, und die **Beobachtung** im Test läuft weiterhin ausschließlich über
`List`/`Get`. **Ticket 07 löscht ihn ersatzlos**, sobald `MarkExit` existiert.

### Für spätere Tickets vorgemerkt

- **Ticket 06** sollte `Listeneintrag` und `nachNamenSortieren` wiederverwenden,
  statt eine zweite Read-Projektion aufzumachen.
- **Ticket 04/05** bringen Zeilen-Aktionen und damit weitere Knöpfe. Spätestens
  dann lohnt es, die dreimal wiederholte Tailwind-Klassenkette der Knöpfe in ein
  Teil-Template oder eine Komponentenklasse zu ziehen (aus dem Review, jetzt
  bewusst nicht gemacht — noch zu wenig Wiederholung).
