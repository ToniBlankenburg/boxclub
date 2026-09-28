# ADR-0020: Spaltenfilter am Zahnrad statt Filterleiste

**Status:** Accepted
**Datum:** 2026-09-28

## Kontext

[ADR-0015](0015-mitglieder-filterleiste-ausklappbares-filterpanel.md) legte
die Mitglieder-Filterleiste als ausklappbares Panel hinter einem
Trichter-Button an, daneben ein eigenes "Spalten"-Menü zum Ausblenden von
Spalten. In der Praxis trennt das zwei Fragen, die zusammengehören: welche
Spalte gerade wie eingegrenzt ist, und ob sie überhaupt sichtbar ist. Wer
wissen will, was gerade filtert, muss das Panel öffnen; wer nach einer
bestimmten Spalte filtern will, muss erst erraten, in welchem der beiden
Menüs das zugehörige Feld steckt. Die Zuordnung "dieser Filter gehört zu
dieser Spalte" ist nirgends sichtbar. Vier Spalten (Status, Ruhend, Beitrag,
Eintritt) hatten zudem noch gar keine Filtermöglichkeit — Details dazu in den
Folgetickets 02–04 (`.scratch/spaltenfilter/`), diese ADR betrifft nur die
Struktur, nicht die neuen Filterdimensionen.

## Entscheidung

Das Filterpanel und das "Spalten"-Menü entfallen vollständig. Stattdessen
bekommt jede Spaltenüberschrift, die heute schon ausblendbar ist (Status,
Anschrift, Training, Beitrag, Rückstand, Eintritt), ein eigenes
Zahnrad-Icon. Ein Klick öffnet ein natives `<details>`-Menü direkt unter
dieser Spalte mit genau den Filtern, die zu ihr gehören, plus der Aktion
"Spalte ausblenden". Nur ein Menü ist gleichzeitig offen — durchgesetzt über
einen `toggle`-Listener in der Erfassungsphase (`main.js`, da das
`toggle`-Ereignis eines `<details>` nicht blubbert). Das Zahnrad zeigt einen
gefüllten statt umrissenen Zustand, wenn mindestens einer der Filter dieser
Spalte vom Nullwert abweicht — serverseitig berechnet (`listeDaten.AktivXxx`),
analog zum bisherigen `AktiveFilterAnzahl()`, nur pro Spalte statt einmal
gesamt.

**Die Name-Spalte bekommt ebenfalls ein Zahnrad**, obwohl sie — anders als
die sechs oben — nicht ausblendbar ist: ihr Menü enthält nur den
Geschlecht-Filter (Geschlecht hat keine eigene Spalte), keine
"Spalte ausblenden"-Aktion. Das ist eine bewusste Lesart der Spec
(`.scratch/spaltenfilter/spec.md`), die an zwei Stellen unterschiedlich
klingt: User Story 20 sagt, Nr. und Name blieben "ohne Zahnrad", User
Story 9 verlangt zugleich ein Zahnrad an der Name-Spalte für den
Geschlecht-Filter. Aufgelöst zugunsten von Story 9: "ohne Zahnrad" bezieht
sich auf die Ausblenden-Funktion, die Name tatsächlich nie bekommt, nicht auf
das Filtern. Nr. bleibt ohne Zahnrad — sie hat keinen ihr zugeordneten
Filter und keine eigene Spalte, die etwas anderes trüge.

Ausgeblendete Spalten verlieren ihr Zahnrad mit sich selbst. Ein neues,
permanent sichtbares "+"-Element am Ende der Kopfzeile (eigene `<th>`, wie
das alte "Spalten"-Menü rein client-seitig über `localStorage`, Go bekommt
es nie zu sehen) listet die aktuell ausgeblendeten Spalten zum
Wiedereinblenden auf. Ein eigenständiger Reset-Icon-Button in der
Werkzeugleiste ersetzt den alten Trichter-Button: er setzt alle
Spaltenfilter auf einmal zurück, lässt die Volltextsuche aber unverändert
stehen (`hx-include` zählt dort nur Suchfeld und Sortierung auf) — anders als
der bisherige Zurücksetzen-Knopf, der auch die Suche mitlöschte.

Weil die Filterfelder jetzt im `<thead>` von `mitglieder-ergebnis` stehen
statt in einer eigenen, vom htmx-Austausch ausgenommenen Filterleiste,
werden sie bei jeder Suche und jedem Filterwechsel mitgerendert. Das
bestehende `form="mitglieder-filter"`-Muster (Felder als Geschwister statt
Kinder eines leeren, unsichtbaren Formulars, siehe ADR-0015 für die
ausführliche Chromium-Begründung) bleibt unverändert nötig und funktioniert
unabhängig von der DOM-Position der Felder, weil `HTMLFormElement.elements`
extern verknüpfte Felder unabhängig von ihrer Verschachtelung einschließt.
Die vormalige Out-of-Band-Aktualisierung des Filter-Zählers
(`filter-badge-oob`) entfällt ersatzlos: die aktiven Zustände der Zahnräder
liegen jetzt selbst im ausgetauschten Bereich und werden bei jedem Rendern
frisch berechnet.

Die `Ehemalige`-Checkbox des alten Panels entfällt in diesem Zug ersatzlos;
ihr Ersatz, ein Status-Filter am Zahnrad der Status-Spalte, ist Gegenstand
von Ticket 02 und ändert `service.Suchfilter` gesondert.

## Betrachtete Alternativen

### Eine gemeinsame Filter-Seitenleiste statt verteilter Zahnräder

**Pro:** Alle Filter auf einen Blick, ohne mehrere Menüs zu öffnen.

**Contra:** Genau die in ADR-0015 verworfene Variante c — kostet permanent
horizontale Breite, die datenlastige Mitgliederliste braucht sie am meisten.
Löst außerdem das eigentliche Problem nicht: die Zuordnung Filter↔Spalte
bliebe weiterhin unsichtbar, nur an anderer Stelle unübersichtlich statt im
Panel.

### Ein Popover pro Spalte, aber weiterhin zentral über einem Trichter-Button ausgewählt

**Pro:** Weniger Icons in der Kopfzeile.

**Contra:** Bringt die ursprüngliche Unklarheit zurück, welches Menü zu
welcher Spalte gehört — der Auswahlschritt selbst wäre das neue Rätsel.

## Konsequenzen

- **Positiv:** Filter und Spalte sind an derselben Stelle sichtbar; kein
  Rätselraten mehr zwischen zwei Menüs. Aktive Filter sind am Zahnrad direkt
  erkennbar, ohne ein Panel zu öffnen.
- **Negativ:** Sieben Zahnräder plus ein Fallback-Element sind mehr
  Kopfzeilen-Chrome als der bisherige einzelne Trichter- und Spalten-Button.
  Ein Popover, das an einer weit rechts stehenden Spalte (z. B. Eintritt)
  öffnet, kann durch `overflow-x-auto` der Tabelle horizontal wegscrollen
  statt sofort sichtbar zu sein — er wird dadurch nicht abgeschnitten
  (absolut positionierte Nachfahren zählen zum scrollbaren Überlauf ihres
  Containers), nur eben per Scroll statt auf den ersten Blick erreichbar.
- **Reversibilität:** mittel, wie schon bei ADR-0015. Betrifft
  `templates/mitglieder_liste.html`, `app/app.go` (`suchEingabe`,
  `listeDaten`) und `frontend/src/main.js` (Spaltensichtbarkeit,
  Menü-Exklusivität) gleichzeitig.
</content>
