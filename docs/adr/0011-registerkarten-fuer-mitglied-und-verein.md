# ADR-0011: Registerkarten für Mitglied und Verein, nicht überall

**Status:** Accepted
**Datum:** 2026-09-18

## Kontext

Die Mitglied-Bearbeiten-Seite reihte Stammdaten, Verträge und Rechnung als
drei eigenständige Formulare untereinander — zusammen mit dem Verein-Formular
(drei Gruppenboxen in einem Formular) war das der Anlass für eine vierte
Prototyp-Runde dieses UI-Refactorings (siehe
[Spec](../../.scratch/ui-refactoring/spec.md), Branch
`prototype/registerkarten`). Drei strukturell unterschiedliche Varianten
standen zur Wahl (Reiter oben mit Unterstrich, Pillen-Segmented-Control,
seitliche Reiter); Verdict siehe
`.scratch/prototyp-registerkarten-verdict.md` auf dem Prototyp-Branch.

Zwei naheliegende Kandidaten wurden dabei bewusst **nicht** auf Registerkarten
umgestellt, obwohl beide ebenfalls mehrere Gruppenboxen zeigen:

- **Trainingstermin-Formular:** nur eine Gruppenbox mit vier Feldern — "viel
  Inhalt" trifft hier nicht zu, Reiter wären Overhead ohne Nutzen.
- **Eigenständige Rechnung** (Empfänger ohne Mitglied, `templates/rechnung.html`):
  ein einziges `<form>` mit drei Gruppenboxen (Empfänger / Rechnungsdaten /
  Positionen), und Pflichtfelder (`required`) verteilen sich über mehrere
  davon. Versteckt eine Registerkarte eine Box per CSS, kann der Browser ein
  unsichtbares Pflichtfeld beim Abschicken nicht fokussieren und bricht die
  native Formularvalidierung **stumm** ab — ohne sichtbare Fehlermeldung für
  den Nutzer. Lösbar nur mit zusätzlichem JavaScript (bei einem ungültigen
  Feld automatisch die richtige Registerkarte aktivieren, bevor der Browser
  validiert).

## Entscheidung

Registerkarten gelten für die Mitglied-Bearbeiten-Seite (Stammdaten /
Verträge / Rechnung, drei eigenständige Formulare) und das Verein-Formular
(Anschrift & Kontakt / Bankverbindung / Rechnungstext, ein Formular in drei
Gruppenboxen). Beide nutzen dasselbe Muster: seitliche Reiter mit
Trennlinie und getöntem Hintergrund, umgesetzt in Tailwind-Klassen
(`aria-selected:`-Varianten) plus einem gemeinsamen Klick-Handler
(`frontend/src/registerkarten.js`).

Die eigenständige Rechnung und das Trainingstermin-Formular bleiben bei
Gruppenboxen ohne Registerkarten.

## Betrachtete Alternativen

### Registerkarten auch für die eigenständige Rechnung

**Pro:** Konsistenz — dieselbe Seite wäre am Mitglied (eingebettet) anders
gegliedert als eigenständig aufgerufen nie der Fall, weil beide Wege
dieselbe `rechnung-bereich`-Vorlage nutzen und dort ohnehin nie Reiter
stehen.

**Contra:** Der Mehraufwand (JavaScript-gestützte Aktivierung der richtigen
Registerkarte bei einem Validierungsfehler) steht in keinem Verhältnis zu
einem Formular mit nur drei kompakten Boxen, das heute schon funktioniert
und dessen Pflichtfelder alle in Sichtweite bleiben, solange es ungeteilt ist.

### Akkordeon statt Registerkarten

**Pro:** Erlaubt mehrere Bereiche gleichzeitig offen.

**Contra:** Die betroffenen Bereiche sind fachlich unabhängig (eigene
Formulare, kein gemeinsamer Zustand) — es gibt keinen Anwendungsfall für
"zwei gleichzeitig offen". Registerkarten passen außerdem zum Vorbild wiso
mein Verein, an dem sich das gesamte UI-Refactoring orientiert.

## Konsequenzen

- **Positiv:** Die beiden am stärksten überladenen Seiten der App sind
  entzerrt, ohne dass Formularverhalten (welches Feld Pflicht ist, was beim
  Abschicken passiert) sich ändert.
- **Negativ:** Das Muster gilt nicht einheitlich für jede Gruppenboxen-Stelle
  der App — ein späterer Leser könnte sich fragen, warum die Rechnung fehlt;
  dieses ADR beantwortet das.
- **Reversibilität:** hoch für die Optik (reine Tailwind-Klassen), mittel für
  den Vorbehalt bei der Rechnung — fällt später doch eine Registerkarten-Lösung
  dafür, ist der hier beschriebene Validierungs-Konflikt der Ausgangspunkt,
  nicht ein Neuentwurf von vorn.
