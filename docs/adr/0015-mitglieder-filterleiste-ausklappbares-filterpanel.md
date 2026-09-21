# ADR-0015: Mitglieder-Filterleiste mit ausklappbarem Filterpanel

**Status:** Accepted
**Datum:** 2026-09-21

## Kontext

[ADR-0014](0014-werkzeugleisten-zweizeilig-mit-icon-aktionen.md) löste die
unübersichtliche Mitglieder-Werkzeugleiste durch zwei feste Zeilen: oben
Suche und Aktions-Icons, unten die vier Filter samt Checkbox. Nach dem
Einbau blieb der Eindruck: die vier Filter laufen dauerhaft mit, auch wenn
fast nie gefiltert wird — die zweite Zeile besteht praktisch immer, nur
selten mit Inhalt, der gerade zählt.

Drei strukturell unterschiedliche Varianten wurden dazu auf dem
Prototyp-Branch `prototype/mitglieder-werkzeugleiste` durchgespielt (siehe
dort für den vollständigen Variantensatz als primäre Quelle):

- **a** — die zweizeilige Variante aus ADR-0014.
- **b** — eine Zeile mit Suche und Aktionen; die Filter liegen hinter einem
  "Filter"-Button mit Zähler-Badge und werden nur auf Klick sichtbar.
- **c** — die Filter stehen dauerhaft in einer linken Seitenleiste, jeder mit
  eigener Beschriftung darüber.

## Entscheidung

Variante b gewinnt. Die Mitglieder-Filterleiste (`templates/mitglieder_liste.html`,
Define `mitglieder-filterleiste`) besteht aus:

- Einer Zeile mit Suchfeld (wächst) links und drei Aktions-Icons rechts:
  Filter, Zurücksetzen, Spalten.
- Dem Filter-Button, der ein Panel darunter auf- und zuklappt (`hx-on:click`
  auf ein reines Kästchen-Umschalten, kein `<details>` — dessen Inhaltsbox
  hinge sonst im selben Flex-Kind wie die Suche, siehe die ausführliche
  Begründung im Template-Kommentar). Eine Zähler-Badge neben "Filter" zeigt
  `AktiveFilterAnzahl()`: wie viele der fünf Filter (vier Selects plus
  Ehemalige-Checkbox) vom Standard abweichen. Ist mindestens einer aktiv,
  rendert das Panel serverseitig bereits aufgeklappt — sonst verschwände ein
  aktiver Filter scheinbar spurlos hinter einem geschlossenen Panel, etwa
  nach einer Rückkehr aus dem Bearbeitungsformular.
- Dem Spalten-Menü, verfeinert gegenüber der ersten Prototyp-Runde: Icon+Text
  statt reines Icon (an "Spalten" allein schwerer zu erraten als am
  Trichter-Icon bei "Filter" daneben), ein Zähler für versteckte Spalten und
  eine "Alle anzeigen"-Kurzaktion im Panel. Zähler und Kurzaktion laufen
  vollständig client-seitig über die bestehende localStorage-Spaltenwahl
  (`frontend/src/main.js`, Ticket 27) — Go bekommt sie nie zu sehen.

Die Trainingstermine-Werkzeugleiste ist von dieser Entscheidung nicht
berührt: sie hat weder Suche noch Aktionen und bleibt einzeilig, wie schon
in ADR-0014 festgehalten. Diese ADR schärft die Empfehlung für künftige
Werkzeugleisten *mit* Suche und Aktionen: ausklappbares Panel statt fester
zweiter Zeile.

## Betrachtete Alternativen

### Variante a (zweizeilig, ADR-0014)

**Pro:** Bereits gebaut und einsatzbereit; Filter jederzeit sichtbar, kein
zusätzlicher Klick zum Filtern.

**Contra:** Beansprucht dauerhaft eine zweite Zeile, auch wenn nicht
gefiltert wird — genau der Eindruck, der zur zweiten Prototyp-Runde führte.

### Variante c (Filter-Seitenleiste)

**Pro:** Filter ständig sichtbar wie bei a, aber ohne die Ergebnis-Tabelle
nach unten zu verdrängen; mehr Platz pro Filter durch eigene Beschriftung.

**Contra:** Kostet permanent horizontale Breite — genau die Ressource, die
die datenlastige Mitgliederliste am meisten braucht (siehe bereits
[ADR-0010](0010-navigation-bleibt-horizontale-top-leiste.md) zur selben
Abwägung bei der Navigation). Zweispaltiges Layout ist zudem die
invasivste der drei Änderungen an `mitglieder-liste`.

## Konsequenzen

- **Positiv:** Die Standardansicht (kein aktiver Filter) zeigt nur eine
  ruhige Zeile statt zwei; wer filtert, sieht sofort per Badge, wie viel
  gerade eingegrenzt ist, an Filter und an Spalten gleichermaßen.
- **Negativ:** Ein zusätzlicher Klick, um die Filter überhaupt zu sehen,
  wenn man von vornherein weiß, dass man filtern will. Das Auf-/Zuklappen
  läuft über einen kleinen Inline-Klick-Handler statt eines deklarativen
  `<details>`-Elements wie beim Spaltenmenü — zwei verschiedene
  Umschaltmechanismen in derselben Leiste.
- **Reversibilität:** mittel. Betrifft `templates/mitglieder_liste.html`,
  `app/app.go` (`AktiveFilterAnzahl`) und `frontend/src/main.js`
  (Spalten-Badge, "Alle anzeigen"); die verworfenen Varianten a und c
  bleiben als primäre Quelle auf `prototype/mitglieder-werkzeugleiste`
  erhalten, falls sich die Abwägung später umkehrt.
