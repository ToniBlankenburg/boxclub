# ADR-0004: Suche und Filter arbeiten im Speicher, nicht in SQL

**Status:** Accepted
**Datum:** 2026-09-10

## Kontext

Die Spec legt für Ticket 06 fest: *"`MemberService.Search(query, filter)` implementiert
Suche über SQL `LIKE` gegen Vorname, Nachname, E-Mail, Telefon."* Das Ticket verlangt
zusätzlich, dass die Suche **case-insensitive** trifft und dass ein Query mit
**Umlauten** die passenden Mitglieder findet.

Beides zusammen geht mit `LIKE` nicht auf:

- SQLite faltet Groß-/Kleinschreibung ausschließlich im ASCII-Bereich — weder `LIKE`
  noch `LOWER()` noch `COLLATE NOCASE` kennen "ü" und "Ü" als denselben Buchstaben.
  `WHERE nachname LIKE '%öztürk%'` fände "Öztürk" also nicht. In einem deutschen
  Verein ist das kein Randfall.
- Ein Suchbegriff aus dem Suchfeld kann `%` und `_` enthalten. In `LIKE` sind das
  Platzhalter; sie müssten mit `ESCAPE` entschärft werden, sonst sucht der Nutzer
  etwas anderes, als er tippt.

An derselben Grenze stand das Projekt schon einmal: die Liste wird seit Ticket 03
in Go sortiert (`nachNamenSortieren`), weil SQLite keine deutsche Kollation kennt.
Die Sortierung geht dabei weiter als die Suche — `collate.German` behandelt "Ö"
wie "O", die Suche faltet nur Groß- und Kleinschreibung.

## Entscheidung

`Search` liest die Zeilen einmal aus der Datenbank und wendet Suchbegriff,
Zahlungsstatus- und Beitragsklassen-Filter danach in Go an:

- Der Suchbegriff trifft per `strings.Contains` gegen die klein geschriebenen
  Felder Vorname, Nachname, E-Mail, Telefon. `strings.ToLower` faltet Unicode
  und damit auch Umlaute: "öztürk", "Öztürk" und "ÖZTÜRK" sind derselbe Begriff.
- Der Zahlungsstatus ist ohnehin ein abgeleiteter Wert (`bezahlt_bis` gegen heute)
  und existiert als Spalte gar nicht — er lässt sich nur dort filtern, wo er
  auch berechnet wird.
- Der Beitragsklassen-Filter ist ein reiner Gleichheitsvergleich.

**In SQL bleibt allein die Aktivität.** Ob ein Mitglied aktiv oder ehemalig ist,
entscheidet, *welche* Mitgliedschaft die Zeile überhaupt bekommt — das ist Teil
des Verbunds und keine nachgelagerte Auswahl.

Die Abgrenzung aus dem Ticket ("kein Client-Side-Filtern") bleibt gewahrt: gesucht
und gefiltert wird vollständig im `MemberService`. Die htmx-Oberfläche bekommt ein
fertiges Ergebnis und siebt nichts nach.

## Konsequenzen

**Gut:**

- Die Suche verhält sich so, wie ein deutscher Nutzer es erwartet — "öztürk",
  "Öztürk" und "ÖZTÜRK" finden dasselbe Mitglied.
- Sonderzeichen im Suchfeld sind harmlos; es gibt keine Platzhalter zu escapen
  und keine dynamisch zusammengesetzte `WHERE`-Klausel.
- Such- und Filterlogik liegt als gewöhnlicher Go-Code vor: am Seam testbar,
  ohne SQL-Verhalten mitzuprüfen.

**Schlecht:**

- Jede Suche liest alle Mitglieder mit laufender (bzw. zusätzlich beendeter)
  Mitgliedschaft. Bei den 50–200 Mitgliedern dieses Vereins ist das im
  Millisekundenbereich; die Sortierung in Go tut ohnehin schon dasselbe.
- Gefaltet wird nur die Groß-/Kleinschreibung, nicht die Diakritik: "Ozturk"
  findet "Öztürk" nicht. Das ist eine bewusste Grenze — wer den Namen kennt,
  tippt ihn mit Umlaut; wer ihn nicht kennt, sucht ohnehin über E-Mail oder
  Telefon. Soll die Suche das doch können, gehört die Faltung neben
  `nachNamenSortieren` in dieselbe Ecke des Service.
- Die Entscheidung skaliert nicht beliebig. Sollte der Bestand je fünfstellig
  werden, müsste die Suche zurück in die Datenbank — dann aber mit einer
  Volltext-Lösung (FTS5 samt Unicode-Tokenizer), nicht mit `LIKE`.

## Alternativen

- **`LIKE` mit `LOWER()`** — scheitert an der ASCII-Faltung, siehe oben.
- **Umlaut-Ersetzung in SQL** (`REPLACE(REPLACE(...)))`) — verteilt deutsche
  Sprachregeln über SQL-Strings und deckt nur die Fälle ab, an die man gedacht hat.
- **SQLite mit ICU-Erweiterung** — bräuchte CGo und widerspricht damit der
  Wahl von `modernc.org/sqlite` aus ADR-0001.
- **FTS5-Volltextindex** — löst das Problem, kostet aber eine zweite Tabelle samt
  Triggern, um sie synchron zu halten. Für Teiltreffer mitten im Wort
  ("9876543" in einer Telefonnummer) ist ein Wort-Index zudem das falsche Werkzeug.

## Nachtrag (2026-09-11, Ticket 12)

Zwei der drei Filter aus der Entscheidung gibt es nicht mehr, und damit sind zwei
der Begründungen hinfällig:

- Der **Beitragsklassen-Filter** ist mit [ADR-0005](0005-beitrag-individuell-statt-beitragsklasse.md)
  entfallen — Beiträge sind individuell vereinbart, es gibt keine Klassen.
- Der **Zahlungsstatus-Filter** ist mit [ADR-0006](0006-rueckstand-statt-bezahlt-bis.md)
  durch den **Rückstandsfilter** ersetzt. Das Argument oben ("ist ohnehin ein
  abgeleiteter Wert und existiert als Spalte gar nicht") trägt für ihn **nicht**:
  `mitglied.rueckstand` ist eine echte Spalte und ließe sich in SQL filtern.

Die Entscheidung bleibt trotzdem, jetzt aber allein aus dem ersten Grund: der
**Suchbegriff** muss in Go ausgewertet werden, weil `LIKE` die Umlaut-Faltung
nicht kann. Suche und Filter zusammen in einem Durchlauf zu halten ist einfacher
als eine Abfrage, die einen Teil in SQL und den Rest in Go erledigt — bei 200
Mitgliedern kostet es nichts.

Dazu gekommen ist die **Mitglieds-ID** als Suchfeld. Sie ist der eine Sonderfall:
sie trifft **genau**, nicht als Teilzeichenkette. Als Teiltreffer brächte "7" die
7, die 17, die 27 und die 70er zurück, und die Nummer wäre als Sprungmarke zu
einer bekannten Zeile gerade nicht mehr zu gebrauchen.
