# ADR-0008: Trainingstermine sind ein gepflegter Wochenplan, keine datierten Einheiten

**Status:** Accepted
**Datum:** 2026-09-12

## Kontext

Bis heute steht der Trainingstermin einer Mitgliedschaft als **Freitext** in
`trainingsslot.bezeichnung` — "Samstag 10:30 Uhr", so wie es in der Excel-Tabelle
stand. Der Verein wünscht sich stattdessen eine **eigene Liste von Terminen**, die
er pflegt und aus der er auswählt.

Freitext hat drei Kosten, die alle schon anfallen: dieselbe Trainingszeit steht in
fünf Schreibweisen in der Datenbank, eine Zeitänderung muss bei jedem Mitglied
einzeln nachgetragen werden, und die Frage „wer trainiert samstags?" ist nicht
beantwortbar, ohne Texte zu vergleichen.

Heikel ist die Nachbarschaft zum **attendance tracking**, das `CLAUDE.md`
ausdrücklich aus v1 ausschließt. Ein „Katalog von Trainingsterminen" klingt danach,
ist es aber nicht zwingend — die Grenze muss festgeschrieben werden, sonst wandert
sie.

## Entscheidung

Es gibt eine eigene Tabelle **Trainingstermin**: Wochentag (feste Auswahl),
Beginn, wahlweise Ende und Bezeichnung. Eine **Mitgliedschaft** wird für null bis
drei dieser Termine angemeldet; der Freitext entfällt. Die *Trainingsfrequenz*
bleibt die Anzahl und wird weiterhin nicht gespeichert.

Drei Festlegungen dazu:

**Wiederkehrend, nicht datiert.** Ein Termin ist „Samstag 10:30" und gilt bis auf
Weiteres. Es gibt keine Instanz je Kalenderwoche und kein Feld, in dem stehen
könnte, wer am 15.09. da war.

**Archivieren statt löschen.** Ein Termin, den es nicht mehr gibt, verschwindet aus
der Auswahl; bestehende Anmeldungen bleiben bestehen und zählen weiter zur
Frequenz.

**Der Import legt keine Termine an.** Der Excel-Import ordnet einen Freitext einem
vorhandenen Termin per Textvergleich zu und meldet alles andere im Fehlerbericht
aus Ticket 09.

## Konsequenzen

**Gut:**

- Eine Zeitänderung ist **eine** Änderung. Heute sind es so viele wie Mitglieder.
- „Wer trainiert samstags?" wird eine Abfrage statt einer Textsuche.
- Der Stundenplan wird zum ersten Mal **sichtbar**. Bisher existierte er nur in
  200 Einzelangaben und im Kopf des Trainers.
- Die Grenze zum ausgeschlossenen attendance tracking ist damit **benannt**: sie
  liegt nicht beim Termin, sondern bei der Frage, wer da war. Ein Wochenplan hat
  keine Stelle, an der eine Anwesenheit stehen könnte — das ist der eigentliche
  Grund für „wiederkehrend statt datiert".

**Schlecht:**

- **Vor dem ersten Import muss der Stundenplan stehen.** Bei einer leeren
  Datenbank hat der Admin einen zusätzlichen Schritt, den es vorher nicht gab.
  Es sind wenige Einträge und einmalig.
- Der **Excel-Import wird umgebaut**, obwohl er gerade fertig geworden ist
  (Ticket 09). Er liest `Training - 1/2/3` heute direkt in Freitext-Slots.
- Die Zuordnung per Textvergleich ist **unscharf**: „Sa 10:30" trifft „Samstag
  10:30" nicht. Das landet im Fehlerbericht und muss von Hand nachgetragen werden.
  Eine Aufräumrunde beim ersten Import ist eingepreist.
- Ein archivierter Termin bleibt für immer in der Datenbank, und die Liste wächst
  langsam. Bei einem Verein mit einer Halle ist das in zehn Jahren keine
  nennenswerte Menge.

## Alternativen

- **Datierte Trainingseinheiten** (der Termin am 15.09., am 22.09., …). Präziser
  und die Grundlage für alles Weitere — aber der Verein pflegt einen Stundenplan
  und keinen Kalender, und die erste Frage an eine solche Liste ist immer „wer war
  da?". Das ist attendance tracking und für v1 ausgeschlossen. Wer es später will,
  legt Instanzen unter die Termine; der Wochenplan bleibt dabei stehen.
- **Freitext behalten und nur Vorschläge anbieten** (Autovervollständigung aus
  vorhandenen Werten). Billig und ohne Umbau, aber die Schreibvarianten bleiben,
  und die Zeitänderung bleibt Handarbeit an 14 Mitgliedern.
- **Import legt unbekannte Termine automatisch an.** Bequem, erzeugt aber aus
  jeder Schreibvariante einen Stundenplaneintrag. Nach dem ersten Import stünden
  20 Termine da, von denen sechs echt sind — und aufgeräumt wird das nie.
- **Kapazität je Termin.** Naheliegend, sobald es Termine gibt, aber der Verein
  zählt heute niemanden mit. Eine Obergrenze verlangt sofort Antworten auf
  Warteliste, Überbuchung und Ausnahme; das ist ein eigenes Feature.
