Status: ready-for-agent

# Spec: Spaltenfilter am Zahnrad statt Filterleiste

## Problem Statement

Die Mitgliederliste filtert heute über eine separate, ausklappbare
Filterleiste (ADR-0015): ein Trichter-Button öffnet ein Panel mit vier
Auswahlfeldern (Rückstand, Frequenz, Geschlecht, Termin) plus einer
Ehemalige-Checkbox, daneben ein eigenes "Spalten"-Menü zum Ausblenden von
Spalten. Wer wissen will, was gerade filtert, muss das Panel öffnen; wer
nach einer bestimmten Spalte filtern will, muss erst erraten, in welchem
der beiden Menüs (Filter oder Spalten) das zugehörige Feld steckt. Die
Zuordnung "dieser Filter gehört zu dieser Spalte" ist nirgends sichtbar.
Außerdem fehlen für mehrere Spalten (Status, Ruhend, Beitrag, Eintritt)
bisher überhaupt Filtermöglichkeiten.

## Solution

Jede Spaltenüberschrift der Mitgliederliste, die heute schon ausblendbar
ist (alle außer Nr. und Name), bekommt ein eigenes Zahnrad-Icon. Ein Klick
öffnet ein kleines Menü direkt unter dieser Spalte mit genau den Filtern,
die zu ihr gehören, plus der Aktion "Spalte ausblenden". Das alte
Filterpanel (Trichter-Button) und das alte "Spalten"-Menü entfallen
vollständig zugunsten dieser pro Spalte verteilten Zahnräder. Ein
zusätzlicher Reset-Icon-Button setzt alle Filter auf einmal zurück. Ist
eine Spalte ausgeblendet, verschwindet ihr Zahnrad mit ihr — ein neues
Fallback-Element am Ende der Kopfzeile listet ausgeblendete Spalten zum
Wiedereinblenden auf. Die globale Volltextsuche bleibt unverändert
bestehen, unabhängig von den Spaltenfiltern.

Ungenutzte Filter aus dem alten Panel wandern eins zu eins in die passende
Spalte (Rückstand → Rückstand, Frequenz und Termin → Training, Geschlecht →
Name, da Geschlecht keine eigene Spalte hat). Neu hinzu kommen Filter für
Status, Ruhend, Beitrag und Eintritt. Die bisherige Ehemalige-Checkbox
entfällt: ihre Funktion übernimmt der neue Status-Filter über den Wert
"Ausgetreten".

## User Stories

1. Als Vereinsadmin will ich an jeder ausblendbaren Spaltenüberschrift ein
   Zahnrad-Icon sehen, damit ich sofort weiß, wo ich für diese Spalte
   filtern oder sie ausblenden kann.
2. Als Vereinsadmin will ich durch Klick auf das Zahnrad ein Menü direkt
   unter der jeweiligen Spalte öffnen, damit der Zusammenhang zwischen
   Filter und Spalte offensichtlich ist.
3. Als Vereinsadmin will ich, dass sich ein offenes Zahnrad-Menü schließt,
   sobald ich ein anderes öffne, damit sich nie mehrere Menüs überlappen.
4. Als Vereinsadmin will ich am Zahnrad-Icon selbst erkennen, ob auf dieser
   Spalte gerade gefiltert wird (z. B. ausgefüllt statt Umriss), damit
   aktive Filter nicht unsichtbar im Hintergrund weiterlaufen.
5. Als Vereinsadmin will ich die Rückstand-Spalte nach "alle" / "im
   Rückstand" / "in Ordnung" filtern können, genau wie bisher im
   Filterpanel.
6. Als Vereinsadmin will ich die Training-Spalte nach Trainingsfrequenz
   (1×/2×/3×/alle) filtern können, genau wie bisher im Filterpanel.
7. Als Vereinsadmin will ich die Training-Spalte zusätzlich nach einem
   bestimmten Trainingstermin aus dem Stundenplan filtern können, genau
   wie bisher im Filterpanel.
8. Als Vereinsadmin will ich Frequenz- und Termin-Filter im selben
   Zahnrad-Menü der Training-Spalte vorfinden, weil beide sich auf dieselbe
   Spalte beziehen.
9. Als Vereinsadmin will ich über das Zahnrad der Name-Spalte nach
   Geschlecht filtern können, obwohl Geschlecht keine eigene Spalte in der
   Liste hat.
10. Als Vereinsadmin will ich die Status-Spalte nach Neu, Aktiv, In
    Kündigungsfrist, Ausgetreten oder allen filtern können, damit ich
    gezielt eine Lebenszyklus-Phase sehen kann.
11. Als Vereinsadmin will ich, dass die Standardansicht (kein Status-Filter
    aktiv) weiterhin keine ausgetretenen Mitglieder zeigt, damit sich das
    gewohnte Verhalten der alten Ehemalige-Checkbox nicht unbemerkt ändert.
12. Als Vereinsadmin will ich Ausgetretene nur sehen, wenn ich aktiv
    "Ausgetreten" im Status-Filter wähle, damit sie nicht versehentlich die
    Liste überfüllen.
13. Als Vereinsadmin will ich im selben Zahnrad-Menü der Status-Spalte
    zusätzlich nach Ruhend (alle / ruhend / laufend) filtern können, als
    eigenständiges zweites Filterfeld, weil Ruhend laut Datenmodell kein
    eigener Status ist, sondern eine Fahne neben dem Lebenszyklus.
14. Als Vereinsadmin will ich die Beitrag-Spalte über einen
    Doppel-Schieberegler (Minimum und Maximum) eingrenzen können, damit ich
    z. B. alle Mitglieder mit besonders niedrigem oder hohem Beitrag finde.
15. Als Vereinsadmin will ich, dass sich die Grenzen dieses Schiebereglers
    automatisch an die tatsächlich vorkommenden Beitragswerte anpassen,
    damit der Regler nicht großteils leerläuft oder echte Werte abschneidet.
16. Als Vereinsadmin will ich, dass ein Beitrag von 0 € im Schieberegler
    ein gültiger, wählbarer Wert ist (kein "kein Filter"-Sonderfall), weil
    0 € laut Datenmodell ein gültiger Beitrag ist.
17. Als Vereinsadmin will ich die Eintritt-Spalte über ein Von-Bis-
    Datumspaar eingrenzen können, damit ich z. B. alle Neuzugänge eines
    bestimmten Zeitraums finde.
18. Als Vereinsadmin will ich, dass Von und Bis beim Eintritt-Filter beide
    optional und unabhängig voneinander nutzbar sind (nur Von, nur Bis,
    oder beide), damit ich auch offene Zeiträume eingrenzen kann.
19. Als Vereinsadmin will ich, dass die Anschrift-Spalte weiterhin keinen
    Filter anbietet (ihr Zahnrad enthält nur "Spalte ausblenden"), weil das
    explizit außerhalb des Zuschnitts liegt.
20. Als Vereinsadmin will ich, dass Nr. und Name weiterhin nicht ausblendbar
    sind und deshalb kein Zahnrad tragen, genau wie heute.
21. Als Vereinsadmin will ich über "Spalte ausblenden" im Zahnrad-Menü eine
    Spalte verschwinden lassen, genau wie heute über das alte
    "Spalten"-Menü.
22. Als Vereinsadmin will ich, nachdem ich eine Spalte ausgeblendet habe,
    über ein Fallback-Element am Ende der Kopfzeile eine Liste der
    ausgeblendeten Spalten sehen und einzelne davon wieder einblenden
    können, damit das Ausblenden keine Sackgasse ist.
23. Als Vereinsadmin will ich, dass ausgeblendete Spalten wie bisher den
    Neustart der App überleben (localStorage), weil das eine reine
    Ansichtsvorliebe an diesem Rechner ist.
24. Als Vereinsadmin will ich über einen eigenen Reset-Icon-Button in der
    Werkzeugleiste alle aktiven Filter auf einmal zurücksetzen können,
    ohne dafür jedes Zahnrad einzeln öffnen zu müssen.
25. Als Vereinsadmin will ich, dass Filterwerte (anders als Spaltensichtbarkeit)
    beim Verlassen und Wiederöffnen der Mitgliederliste zurückgesetzt
    werden, genau wie heute schon beim alten Filterpanel.
26. Als Vereinsadmin will ich, dass die globale Suche im Suchfeld unverändert
    über alle Felder sucht, unabhängig davon, was ich in den Spaltenfiltern
    eingestellt habe.
27. Als Vereinsadmin will ich, dass Klicks auf den Spaltentitel weiterhin
    sortieren (bestehendes Verhalten), unabhängig vom neuen Zahnrad-Icon
    daneben.
28. Als Vereinsadmin will ich, dass alle neuen Beschriftungen (Filterlabel,
    Menüeinträge, Fallback-Element) sowohl auf Deutsch als auch auf
    Englisch angezeigt werden, weil die App laut Mehrsprachigkeits-Feature
    komplett zweisprachig ist.
29. Als Entwickler will ich, dass `Suchfilter` weiterhin mit seinem Nullwert
    exakt die heutige Standardansicht ergibt (keine Ausgetretenen, keine
    sonstigen Einschränkungen), damit `List()` unverändert funktioniert.
30. Als Entwickler will ich, dass die Filterung weiterhin vollständig in
    `service.MemberService.Search` in Go entschieden wird (ADR-0004), auch
    für die vier neuen Filterdimensionen, statt SQL-Bedingungen
    einzuführen.

## Implementation Decisions

- **Seam:** Einziger Test-Seam ist `service.MemberService.Search` über
  `Suchfilter`. `app/` (HTTP-Handler, Query-Parameter) und die Templates
  bleiben unverändert reine Adapter und werden wie bisher nur manuell
  smoke-getestet (CLAUDE.md-Konvention).
- **`Suchfilter`-Struktur:** `AuchEhemalige bool` entfällt als öffentliches
  Feld. An seine Stelle tritt `Status Statusfilter` (neuer Typ, Werte:
  alle/Neu/Aktiv/InKündigungsfrist/Ausgetreten, Nullwert = alle). `Search`
  leitet intern ab, ob ausgetretene Mitgliedschaften überhaupt geladen
  werden müssen (`eintraegeLesen(...)`): nur wenn `Status ==
  StatusfilterAusgetreten`. Bei jedem anderen Statuswert (inklusive
  "alle") bleibt das Ladeverhalten wie heute ohne Ehemalige — damit ändert
  sich die Standardansicht nicht.
- **Neuer Filtertyp `Ruhendfilter`:** analog zu `Rueckstandsfilter`
  (alle/ruhend/laufend als eigener int-Enum mit `trifft`-Methode),
  eigenständiges Feld `Ruhend Ruhendfilter` in `Suchfilter`, unabhängig vom
  Statusfilter geprüft.
- **Beitrag-Filter:** zwei neue optionale Felder in `Suchfilter`,
  `BeitragVonCents *int64` und `BeitragBisCents *int64` (Pointer statt
  Sentinel-Wert, weil 0 € ein gültiger Beitrag ist und nicht als "kein
  Filter" missverstanden werden darf — dasselbe Pointer-Muster wie bei den
  bestehenden nullbaren Mitgliedschaftsfeldern). `nil` grenzt nicht ein.
  Umrechnung von Euro-Eingabe des Reglers nach Cents über das bestehende
  `service.BeitragAusEuro`.
- **Eintritt-Filter:** zwei neue optionale Felder `EintrittVon *time.Time`,
  `EintrittBis *time.Time` in `Suchfilter`, `nil` grenzt jeweils nicht ein,
  Vergleich einschließend (Randtage zählen mit, wie beim bestehenden
  Status-Konzept).
- **Geschlecht-Filter:** Feld und Verhalten unverändert, nur die Herkunft
  im UI ändert sich (Zahnrad der Name-Spalte statt Filterpanel).
- **Query-Parameter (`app/app.go`, `suchEingabe`):** `ehemalige` entfällt
  ersatzlos (keine Rückwärtskompatibilität nötig, da nur transienter
  URL-Zustand einzelner htmx-Requests, nirgends gespeichert). Neu:
  `status`, `ruhend`, `beitragVon`, `beitragBis`, `eintrittVon`,
  `eintrittBis`. Bestehende Parameter (`q`, `rueckstand`, `frequenz`,
  `geschlecht`, `termin`, `sort`, `richtung`) bleiben unverändert.
- **Template-Struktur:** Die bestehende Filterleiste (`mitglieder-filterleiste`,
  Trichter-Button, Filterpanel, "Spalten"-Menü) entfällt. Jeder
  ausblendbare Spaltenkopf (`spaltenkopf`-Template) bekommt zusätzlich zum
  bestehenden Sortier-Klick ein Zahnrad-Icon, das ein kleines,
  spaltenspezifisches Menü öffnet (Filtersteuerelemente je nach Spalte +
  "Spalte ausblenden"). Öffnen/Schließen folgt demselben client-seitigen
  Muster wie das bisherige Filterpanel bzw. `<details>` wie das bisherige
  Spalten-Menü — konkrete Wahl bleibt der Umsetzung überlassen, solange
  nur ein Menü gleichzeitig offen ist.
- **Zahnrad-Zustand:** Icon zeigt einen aktiven Zustand (z. B. gefüllt
  statt Umriss), wenn mindestens einer der Filter dieser Spalte vom
  Nullwert abweicht — serverseitig berechnet, analog zu
  `AktiveFilterAnzahl()` heute, nur pro Spalte statt einmal gesamt.
- **Fallback für ausgeblendete Spalten:** neues, permanent sichtbares
  Element am Ende der Kopfzeile (z. B. "+"-Icon), das die Liste
  ausgeblendeter Spalten zum Wiedereinblenden zeigt. Läuft wie die
  bestehende Spaltensichtbarkeit vollständig client-seitig über
  `frontend/src/main.js` (localStorage, kein Server-Zugriff).
- **Reset-Button:** neuer, eigenständiger Icon-Button in der Werkzeugleiste
  (Ersatz für den alten Trichter-Button), setzt alle serverseitigen Filter
  auf `Suchfilter{}` zurück (entspricht einem Klick, der alle
  Query-Parameter außer `q` entfernt).
- **i18n:** alle neuen Beschriftungen (Filterlabel, Menüeinträge,
  Fallback-Element, Reset-Button) als neue Schlüssel in beiden Katalogen
  (Deutsch/Englisch) nach bestehendem `i18n.Text(sprache, "...")`-Muster.
- **ADR:** Diese Entscheidung löst ADR-0015 ab (und ändert den
  Filterpanel-Teil von ADR-0014-Nachfolgern) — wird als eigene ADR vor der
  Umsetzung festgehalten, da schwer umkehrbar (betrifft Template,
  `app/app.go`, `service/member_service.go`, `frontend/src/main.js`
  gleichzeitig) und Ergebnis einer echten Abwägung (Panel/Seitenleiste vs.
  pro Spalte verteilt).

## Testing Decisions

- Tests laufen ausschließlich gegen `service.MemberService.Search` über
  eine echte temporäre SQLite-Datenbank, nach dem Muster in
  `service/member_service_test.go` — keine Mocks, kein Zugriff auf HTTP
  oder Templates.
- Zu testende externe Verhaltensweisen (nicht Implementierungsdetails):
  - `Suchfilter{}` (Nullwert) liefert weiterhin exakt dieselbe Menge wie
    `List()` — keine Ausgetretenen, keine sonstige Einschränkung.
  - `Statusfilter` grenzt korrekt auf Neu/Aktiv/InKündigungsfrist ein,
    jeweils ohne Ausgetretene zusätzlich zu laden.
  - `Statusfilter = Ausgetreten` liefert ausschließlich ausgetretene
    Mitglieder (und lädt sie erst dazu) — Ersatz für den bisherigen
    `AuchEhemalige`-Test.
  - `Ruhendfilter` grenzt unabhängig vom Statusfilter korrekt ein
    (Kombination Status=Aktiv + Ruhend=ruhend liefert nur aktive UND
    ruhende Mitglieder).
  - `BeitragVonCents`/`BeitragBisCents`: Grenzfälle 0 €, nur Von gesetzt,
    nur Bis gesetzt, beide gesetzt, beide `nil`.
  - `EintrittVon`/`EintrittBis`: Randtage zählen ein (einschließend),
    Grenzfälle nur Von, nur Bis, beide `nil`.
  - Kombination mehrerer Filter gleichzeitig (z. B. Rückstand + Beitrag +
    Eintritt) grenzt konjunktiv ein (UND-Verknüpfung), wie es die
    bestehenden Filter bereits tun.
- `app/`-Handler und Templates (Zahnrad-Menüs, Fallback-Element,
  Reset-Button, Popover-Verhalten) werden wie bisher nicht unit-getestet;
  manueller Smoke-Test in `wails dev` genügt (CLAUDE.md).

## Out of Scope

- Filter für die Anschrift-Spalte (bewusst ausgeklammert).
- Neue Datenbankspalten oder Schemaänderungen — alle gefilterten Werte
  existieren bereits.
- Relative Datums-Presets (z. B. "letzte 30 Tage") beim Eintritt-Filter.
- Feste/konfigurierbare Grenzen beim Beitrag-Schieberegler — nur dynamisch
  aus vorhandenen Werten.
- Persistenz von Filterwerten über Navigation hinweg (bleibt wie heute:
  Reset bei Verlassen der Liste).
- Eine eigene Geschlecht-Spalte in der Tabelle (Filter bleibt am
  Zahnrad der Name-Spalte, ohne dass Geschlecht angezeigt wird).
- Verschieben der Filterung von Go/Speicher nach SQL (ADR-0004 bleibt
  unangetastet).

## Further Notes

Vor der Umsetzung ist eine ADR zu schreiben, die ADR-0015 ablöst (siehe
Implementation Decisions). Die Tickets sollten so geschnitten werden, dass
die Erweiterung von `Suchfilter`/`Search` (inkl. Tests) unabhängig von der
Template-/Frontend-Umstellung lauffähig ist, da Ersteres der einzige
Test-Seam ist und Letzteres nur manuell geprüft wird.
