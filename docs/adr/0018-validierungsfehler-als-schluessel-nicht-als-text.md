# ADR-0018: `service.ValidierungsFehler` trägt Schlüssel, `app/` übersetzt sie

**Status:** Accepted
**Datum:** 2026-09-22

## Kontext

ADR-0017 hat entschieden, dass `service/` keine Sprache kennt, und die Frage
offengelassen, wie seine Fehlermeldungen dann übersetzt werden, ohne dass
`service/` einen `i18n`-Import bekommt (Ticket 07 der Mehrsprachigkeits-Spec).

`service/` meldet Regelverstöße auf zwei Wegen:

- **Sentinel-Fehler** (`ErrNichtGefunden`, `ErrNichtAktiv`, `ErrBereitsAktiv`,
  `ErrKeinVertrag`) — das war schon vor diesem Ticket sprachunabhängig:
  `app/` erkennt sie über `errors.Is` und wählt selbst einen Katalogschlüssel
  (siehe `App.veralteteAnsichtOderFehler`, `App.nichtGefundenOderFehler`).
  Hier ändert sich nichts.
- **`*service.ValidierungsFehler`** mit `Meldungen []string` — das war der
  eigentliche Bruch: die Meldungen kamen als fertige deutsche Sätze aus
  `service/` (`"Vorname darf nicht leer sein."`, gebaut mit `fmt.Sprintf` für
  Fälle, die einen Namen oder eine Zahl nennen) und landeten unverändert in
  den Formularen. Rund zwanzig Konstruktionsstellen über acht Dateien
  (`member_service.go`, `kuendigung.go`, `mitgliedschaft_trainingstermin.go`,
  `trainingstermin.go`, `dokument.go`, `vereinslogo.go`, `rechnung.go`,
  `beitrag.go`, `import.go`) — dazu `app/`s eigener Umweg in
  `validierungsMeldungen`, der eine dieser Meldungen schon einmal ungeprüft in
  einen eigenen `i18n.Text`-Aufruf einsetzte (`rechnung.fehler_position_praefix`)
  und damit deutschen Text in eine sonst englische Meldung schmuggelte.

## Entscheidung

**`ValidierungsFehler.Meldungen` wird `[]service.Meldung`**, mit

```go
type Meldung struct {
    Schluessel string
    Args       []any
}
```

statt fertigem Text. `service/` baut weiterhin an jeder Stelle, an der es
heute schon einen Regelverstoß erkennt, eine `Meldung` — mit demselben
`Schluessel`-Namensraum wie der Rest des Katalogs (`validierung.<bereich>.<grund>`,
etwa `validierung.mitglied.vorname_leer` oder `validierung.termin.archiviert`)
und denselben `Args`, die heute schon in die `fmt.Sprintf`-Aufrufe gingen
(ein Dateiname, eine Positionsnummer, `MaxTrainingstermine`). Übersetzt wird
das erst in `app/`, mit derselben `i18n.Text(sprache, schluessel, args...)`,
die auch für jeden anderen Text der Oberfläche gilt — ein neuer Helfer
`App.uebersetzeMeldungen` an der Stelle, an der bisher `validierung.Meldungen`
direkt ins Formular ging.

**Nur Validierungsfehler, keine internen `fmt.Errorf`.** Die zahlreichen
`fmt.Errorf("transaktion starten: %w", err)`-artigen Meldungen in `service/`
bleiben unübersetzt und laufen weiterhin in `app.fehlerAntwort` (Status 500,
`err.Error()` roh ausgegeben). Sie sind kein Ergebnis einer abgelehnten
Eingabe, sondern eine Auskunft über einen Programmierfehler oder eine kaputte
Datenbank — Diagnosetext für den Fall, dass etwas schiefgelaufen ist, das gar
nicht hätte passieren sollen, nicht Oberfläche, die ein Nutzer im
Normalbetrieb sieht. Sie mit demselben Schlüsselaufwand zu versehen wie die
rund vierzig Validierungsmeldungen stünde in keinem Verhältnis zum Nutzen: es
gibt Dutzende solcher Stellen, jede so selten wie ein Datenbankfehler, und
"transaction failed" ist für einen englischsprachigen Nutzer nicht
verständlicher als "transaktion starten fehlgeschlagen" — beides ist ein
Diagnosetext, keine Handlungsanweisung.

**`ValidierungsFehler.Error()`** verkettet jetzt die Schlüssel selbst
(`strings.Join(schluessel, "; ")`) statt der früheren deutschen Sätze — für
Logs und den seltenen Fall, dass ein `*ValidierungsFehler` doch bei
`fehlerAntwort` landet, ohne dass ein Aufrufer ihn vorher über `errors.As`
abgefangen hat. Jeder Aufrufer, der die Meldungen tatsächlich zeigt, liest
weiterhin `.Meldungen` direkt.

**`app/dokument.go` bleibt außerhalb des Katalogs.** Die Datei wurde von
keinem der Tickets 01–06 berührt (kein `i18n`-Import, keine Katalogschlüssel)
— sie komplett zu übersetzen wäre ein eigenes Ticket wie 02–06, nicht Teil
von Ticket 07. Der `ValidierungsFehler`, den `VertragAblegen` liefert, wird
trotzdem korrekt über `App.uebersetzeMeldungen` aufgelöst, damit die Datei
weiterhin kompiliert und die Meldung nicht als rohe Schlüssel erscheint;
Formularbeschriftungen und die beiden `errors.Is`-Zweige in
`vertragNichtGefundenOderFehler` bleiben deutsches Literal, wie der Rest der
Datei, bis ein eigenes Ticket sie aufnimmt.

## Betrachtete Alternativen

### Fehlercode statt Schlüssel-Struct

Ein `int`- oder eigener `enum`-Fehlercode statt eines String-Schlüssels, von
`app/` über eine `switch`-Anweisung in Text übersetzt. **Contra:** der Katalog
adressiert jeden anderen Text der Oberfläche schon über String-Schlüssel
(`i18n.Text(sprache, schluessel, args...)`); ein zweites Adressierungsschema
nur für Validierungsmeldungen wäre ein Bruch ohne Gewinn — beide Wege sind
gleich typo-anfällig (ein Code so wie ein Schlüssel bricht beim Tippfehler
lautlos auf den deutschen Rückfall zurück, siehe ADR-0017), aber der
Code bräuchte zusätzlich eine `switch`-Anweisung pro Aufrufer statt eines
einzigen `i18n.Text`-Aufrufs.

### Service liefert direkt zwei Meldungen (de/en)

`ValidierungsFehler` behält fertigen Text, aber für beide Sprachen zugleich.
**Contra:** genau die Abhängigkeit, die ADR-0002/0017 ausschließen wollen —
`service/` müsste bei jeder neuen Sprache mitwachsen, und die Übersetzung läge
an zwanzig verstreuten Stellen statt an einer (`i18n/en.go`).

## Konsequenzen

- Jeder neue Regelverstoß in `service/` braucht ab jetzt einen Schlüssel im
  Katalog (beide Sprachen) statt eines Satzes im Code — ein Schritt mehr,
  aber derselbe, den jeder andere Text der Oberfläche schon geht.
- Tests, die bisher den deutschen Wortlaut einer Meldung prüften
  (`validierung.Meldungen[0] == "Vorname darf nicht leer sein."`), prüfen
  jetzt den Schlüssel und ggf. die Argumente
  (`meldung.Schluessel == "validierung.mitglied.vorname_leer"`) — der
  eigentlich stabile Vertrag war ohnehin nie der deutsche Satz, sondern der
  Grund, den er benennt.
- `megabyte()` (Dateigrößen in Meldungen wie `validierung.dokument.zu_gross`)
  formatiert weiterhin mit deutschem Dezimalkomma, unabhängig von der
  Anzeigesprache — eine Zahlenformatierung, kein Übersetzungsfall im Sinne
  dieses ADRs, und ein Zwischenstand, den die Größenangabe schon vor
  Mehrsprachigkeit hatte. Nachzuziehen wäre ein eigener, kleiner Schnitt.
