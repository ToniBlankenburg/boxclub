# ADR-0014: Werkzeugleisten zweizeilig mit Icon-Aktionen

**Status:** Superseded by [ADR-0015](0015-mitglieder-filterleiste-ausklappbares-filterpanel.md)
**Datum:** 2026-09-21

## Kontext

Die Mitglieder-Werkzeugleiste (`templates/mitglieder_liste.html:69-153`) reiht
acht gleich gestylte Elemente (Suche, vier Filter-Selects, eine Checkbox,
Zurücksetzen-Button, Spalten-Dropdown) als Flex-Siblings in einer einzigen
`flex flex-wrap`-Zeile. Bei begrenzter Fensterbreite wird der Zeilenumbruch
unvorhersehbar: Suche, Filter und Aktionen verschmelzen optisch, weil nichts
sie voneinander trennt.

Die Trainingstermine-Werkzeugleiste (`templates/trainingstermine.html:30-41`)
nutzt dieselben Klassen, hat aber nur ein einziges Control (Checkbox
"Archivierte anzeigen") und ist davon nicht betroffen.

## Entscheidung

Werkzeugleisten mit Suche und Aktionen bekommen zwei Zeilen:

- **Zeile 1:** Suchfeld links (wächst), Aktions-Icons rechts
  (`justify-between`). Zurücksetzen und Spalten werden von Text- zu
  Icon-Buttons mit `title`-Tooltip, im selben stroke-basierten SVG-Stil wie
  `navigation.html`, aber mit größerer Klickfläche (`size-4` statt
  `size-3.5`): dort steht das Icon neben Text in einer Pille, hier trägt es
  die Aktion allein.
- **Zeile 2:** Filter (Selects + Checkboxen), Reihenfolge unverändert,
  weiterhin `flex flex-wrap items-center gap-2`.

Eine Werkzeugleiste ohne Suche/Aktionen — aktuell nur Trainingstermine —
bleibt einzeilig; das zweizeilige Muster wird nicht erzwungen, wenn Zeile 1
nichts zu zeigen hätte.

## Betrachtete Alternativen

### Eine Zeile mit visuellen Gruppen (Abstand/Trenner statt Zeilenumbruch)

**Pro:** Kleinere Änderung an bestehender Struktur.

**Contra:** Löst das Kernproblem nicht — bei Umbruch fallen die
Gruppengrenzen weg, genau der Fall, der aktuell unübersichtlich wirkt.

### Textbuttons für Zurücksetzen/Spalten beibehalten

**Pro:** Kein Tooltip nötig, für jeden sofort lesbar.

**Contra:** Kostet mehr Breite als Icons, ohne dass die App sonst Textbuttons
für wiederkehrende Aktionen dieser Art bevorzugt — Icons sind bereits
etabliertes Mittel (`navigation.html`, `verein.html`).

## Konsequenzen

- **Positiv:** Suche/Aktionen und Filter sind räumlich getrennt, Umbruch
  bleibt innerhalb der jeweiligen Zeile vorhersehbar. Muster gilt für alle
  Werkzeugleisten, nicht nur Mitglieder — künftige Listenansichten übernehmen
  es direkt.
- **Negativ:** Icon-Buttons ohne sichtbares Label sind für Screenreader-Nutzer
  schlechter erschließbar; bisher aber kein Accessibility-Anspruch im
  Projekt.
- **Reversibilität:** mittel. Betrifft `templates/mitglieder_liste.html` und
  `templates/trainingstermine.html`; bei deutlich mehr Filtern oder Aktionen
  je Zeile ist eine erneute Überarbeitung der naheliegende Revisionsgrund.
