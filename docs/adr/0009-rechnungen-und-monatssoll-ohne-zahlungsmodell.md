# ADR-0009: Rechnungen sind Dokumente, das Monatssoll ist eine Hochrechnung

**Status:** Accepted
**Datum:** 2026-09-12

## Kontext

[ADR-0006](0006-rueckstand-statt-bezahlt-bis.md) hat aus gutem Grund **kein
Zahlungsmodell** in v1 gelassen: der Verein zieht per Lastschrift ein, also weiß
die Bank, wer gezahlt hat, und die App führt nur die Ausnahme — den *Rückstand*.

Der Vereinsadmin wünscht sich nun zweierlei, das auf den ersten Blick genau
dagegen läuft:

- **Rechnungen erstellen.** Gemeint ist aber nicht die Beitragsrechnung, sondern
  ein Dokument über Leistungen daneben — ausdrücklich genannt: Einzeltrainings.
  Es geht um das Erzeugen eines PDFs, nicht um das Überwachen eines Status.
- **Einen Überblick** über Mitgliederzahl und die Summe der Beiträge pro Monat,
  als Dashboard für Trainer und Admin.

`CLAUDE.md` hatte beides zuvor als v2 vertagt, mit der Begründung, Rechnungen zu
erzeugen hieße, das Zahlungsmodell einzuführen, das ADR-0006 weggelassen hat.
Diese Begründung trägt in dem Zuschnitt oben nicht: ein erzeugtes PDF ohne Status
ist kein Forderungsmanagement, und eine Summe hinterlegter Beiträge ist eine
Ableitung aus Daten, die längst da sind.

## Entscheidung

Beides kommt in v1 — **ohne** Zahlungsmodell. ADR-0006 wird dadurch **nicht
abgelöst**, sondern in seiner Reichweite präzisiert: es bleibt dabei, dass die App
weder Zahlungen noch Zahlungsstände führt.

**Eine Rechnung ist ein Dokument.** Sie hat einen frei eintragbaren Empfänger (aus
dem Mitglied vorbelegt), Positionen mit Menge und Einzelpreis, ein Rechnungsdatum,
ein Zahlungsziel und einen Steuersatz. Daraus entsteht ein PDF, das am Mitglied
abgelegt wird ([ADR-0007](0007-dokumente-als-blob-in-sqlite.md)). Es gibt **keinen
Rechnungsstatus**, keinen Zahlungseingang, keine Mahnung und keine offenen Posten.

**Die Rechnungsnummer wird eingetippt, nicht vergeben.** Der Verein führt seine
Nummernfolge in seiner Buchhaltung. Eine fortlaufende Nummer aus der App wäre die
Zusage, sie lückenlos und eindeutig zu halten — die kann eine App nicht einhalten,
neben der noch andere Rechnungen geschrieben werden.

**Beträge sind brutto, ein Steuersatz je Rechnung**, Vorgabe 19 %, änderbar bis 0.
Die App rechnet den Steueranteil zur Anzeige heraus und kennt keine steuerlichen
Regeln — welcher Satz gilt, weiß der Verein.

**Das Dashboard zeigt das *Monatssoll* des laufenden Monats**, nicht einen Verlauf:
die Summe der Beiträge aktiver Mitgliedschaften einschließlich derer in der
Kündigungsfrist, ohne ruhende, ohne noch nicht begonnene. Ruhende werden daneben
getrennt ausgewiesen.

## Konsequenzen

**Gut:**

- Der Verein kann sein Einzeltraining abrechnen, ohne dass die App eine
  Nebenbuchhaltung wird.
- Das Dashboard kostet **kein neues Datenfeld**. Es liest nur, was Beitrag, Status
  und `ruhend` ohnehin hergeben.
- ADR-0006 bleibt intakt und ist jetzt schärfer: nicht „die App sagt nichts zum
  Geld", sondern „die App verfolgt keine Zahlungsstände".
- Die Vereinsdaten (Briefkopf, Bankverbindung, Fußzeile) bekommen einen Ort. Den
  brauchte die App bisher nie, und ohne ihn wäre jede Adressänderung ein Release.

**Schlecht:**

- **Rechnungsnummern können doppelt vergeben werden.** Nichts hindert den Admin,
  zweimal die 2026-014 zu tippen. Bewusst: die Alternative wäre ein Versprechen,
  das die App nicht halten kann.
- **Keine Antwort auf „ist die Rechnung bezahlt?"** Wer das wissen will, schaut
  aufs Konto. Das ist dieselbe Antwort wie bei den Beiträgen und dieselbe
  Grenze wie in ADR-0006.
- **Das Monatssoll ist ein Soll.** Es zeigt, was eingezogen werden soll, nicht was
  ankam. Die Zahl steht ohne Gegenprobe da, und wer sie für Einnahmen hält, irrt.
  Der Name *Monatssoll* im Glossar ist die Gegenmaßnahme — er sagt es bei jeder
  Nennung mit.
- **Kein Verlauf.** Die Datenbank führt je Mitgliedschaft einen Beitrag, den
  heutigen; eine Kurve über zwölf Monate wäre zwölfmal die heutige Summe, nur nach
  Ein- und Austritten gefiltert. Ein Diagramm mit dieser Eigenschaft täuscht mehr,
  als es zeigt, und fehlt deshalb.
- Der Steuersatz ist **eine Zahl ohne Prüfung**. Ob der Verein 19 %, 7 % oder gar
  nichts ausweisen darf, entscheidet sein steuerlicher Status — die App weiß davon
  nichts und soll es nicht wissen.

## Alternativen

- **Weiter auf v2 vertagen.** Wäre konsistent mit dem Stand von gestern, hält den
  Verein aber von etwas ab, das ihn nichts kostet: ein PDF, das er sonst in Word
  tippt.
- **Rechnung als vollwertiger Datensatz mit fortlaufender Nummer**, Status und
  Zahlungseingang. Das ist die ehrliche Buchhaltungslösung und irgendwann
  vielleicht richtig — aber es ist genau das Zahlungsmodell, das ADR-0006
  begründet weggelassen hat, und der Verein hat seine Buchhaltung bereits
  woanders.
- **Beitragsrechnungen im Stapel für alle Mitglieder.** Wurde erwogen und
  verworfen: bei Lastschrifteinzug hat niemand eine Rechnung nötig, und 200 PDFs
  zu erzeugen ist ein anderes, größeres Feature.
- **Verlaufsdiagramm mit Beitragshistorie.** Setzt voraus, dass jede
  Beitragsänderung datiert festgehalten wird. Das ist ein eigener Umbau am
  Mitgliedschaftsmodell und wartet, bis jemand die Historie wirklich braucht.
