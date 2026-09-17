# ADR-0010: Navigation bleibt eine horizontale Top-Leiste, keine Sidebar

**Status:** Accepted
**Datum:** 2026-09-17

## Kontext

Im Zuge des UI-Refactorings (drei Prototyp-Runden, siehe Branches
`prototype/mitgliederliste-layout`, `prototype/stilrichtungen`,
`prototype/vier-restbereiche`) stand zur Debatte, ob die Bereichsnavigation
von der bisherigen flachen Top-Bar auf eine Sidebar links umgestellt wird —
das Muster, das die meisten modernen Business-Apps und native macOS-Anwendungen
(Mail, Finder, Notizen) verwenden, passend zum Produktions-Ziel macOS
([ADR-0001](0001-go-wails-fuer-desktop-gui.md)).

Dagegen steht: Die Mitgliederliste ist die meistgenutzte Ansicht und
spaltenreich — Ticket 27 hat gerade erst Ein-/Ausblenden einzelner Spalten
gebaut, damit auf begrenzter Breite alle relevanten Daten Platz finden. Eine
Sidebar nimmt genau die horizontale Breite weg, die dafür gebraucht wird. Die
App hat außerdem nur 6 Bereiche, alle gleichrangig und bewusst ungruppiert —
der Hauptvorteil einer Sidebar (Gruppierung vieler Einträge) greift hier
nicht.

## Entscheidung

Die Navigation bleibt eine horizontale Leiste oberhalb des Inhalts
(`templates/navigation.html`), umgesetzt als kompakte Pillen mit Icon je
Bereich. Der Hauptbereich bekommt dafür volle Breite statt der bisherigen
zentrierten `max-w-5xl`-Spalte (`frontend/index.html`).

## Betrachtete Alternativen

### Sidebar links

**Pro:** macOS-native Konvention, Raum für spätere Gruppierung, wenn weitere
Bereiche dazukommen.

**Contra:** Kostet genau die Breite, die die Mitgliederliste braucht. Bei nur
6 ungruppierten Einträgen liefert die Sidebar keinen Übersichts-Gewinn, der
diese Kosten aufwiegt.

## Konsequenzen

- **Positiv:** Datenlastige Ansichten (Mitgliederliste, später auch andere
  tabellarische Bereiche) behalten die volle Breite. Die bestehende
  Navigationsstruktur musste nicht umgebaut werden, nur neu gestaltet.
- **Negativ:** Weicht von der macOS-Sidebar-Konvention ab, die Nutzer anderer
  nativer Mac-Apps erwarten könnten.
- **Reversibilität:** mittel. Betrifft `templates/navigation.html` und die
  Breiten-Annahme in `frontend/index.html` sowie in jedem Bereichs-Template;
  wird die Bereichszahl deutlich größer als 6, ist eine Sidebar mit
  Gruppierung der naheliegende Revisionsgrund.
