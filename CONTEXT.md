# Boxclub Mitgliederverwaltung – Kontext

Glossar der Begriffe, wie sie in diesem Projekt verwendet werden. Kein Implementierungsdetail.

## Mitglied

Eine natürliche Person, die dem Verein bekannt ist. Ein Mitglied bleibt derselbe Datensatz, auch wenn es zwischenzeitlich austritt und wieder eintritt.

**Nicht zu verwechseln mit:** Mitgliedschaft (das ist die zeitliche Zuordnung).

## Mitglieds-ID

Die laufende Nummer, unter der ein Mitglied im Verein geführt wird. Eindeutig, wird beim Anlegen automatisch vergeben und **nie neu verwendet** — auch nicht bei Austritt und Wiedereintritt. Damit ist sie die Identität eines Mitglieds über die Zeit.

_Vermeiden:_ Mandatsreferenz. Die so benannte Spalte der alten Excel-Tabelle enthält trotz ihres Namens keine SEPA-Mandatsreferenz, sondern genau diese laufende Nummer.

## Mitgliedschaft

Der zeitliche Zeitraum, in dem ein Mitglied aktiv im Verein ist. Hat ein Eintritts- und optional ein Austrittsdatum. Ein Mitglied kann mehrere Mitgliedschaften über die Zeit haben (Aus- und Wiedereintritt).

**Nicht zu verwechseln mit:** dem Tarif. Wie oft jemand pro Woche trainieren darf, ist die *Trainingsfrequenz* — nie die "Mitgliedschaft". _Vermeiden:_ Mitgliedschaft im Sinne von Tarif, Paket oder Stufe.

## Beitrag

Der monatliche Betrag, den ein bestimmtes Mitglied zahlt. Er ist **individuell pro Mitglied** und nicht aus der Trainingsfrequenz ableitbar — im Verein sind Beträge von 0 € bis 144 € im Umlauf, bei gleicher Frequenz unterschiedlich hoch.

**Nicht zu verwechseln mit:** der Anmeldegebühr (einmalig bei Eintritt).

**Abgelöst:** *Beitragsklasse* — eine gemeinsame Preisstufe, der Mitglieder zugeordnet werden. Das Konzept existiert in diesem Verein nicht; siehe [ADR-0005](docs/adr/0005-beitrag-individuell-statt-beitragsklasse.md). _Vermeiden:_ Beitragsklasse, Tarif, Preisstufe.

## Anschrift

Die Postanschrift eines Mitglieds. Sie besteht aus **drei getrennten Angaben** — *Adresse* (Straße samt Hausnummer), *Postleitzahl* und *Ort* —, weil die Excel-Tabelle des Vereins sie so führt und nur getrennt Postleitzahl und Ort als eigene Angaben auswertbar bleiben. Alle drei sind freiwillig; ein Mitglied ohne Anschrift ist ein gültiges Mitglied.

**Nicht zu verwechseln mit:** einem einzelnen Adress-Textblock. _Vermeiden:_ „Adresse" als Sammelbegriff für die ganze Anschrift — Adresse ist hier nur die Straße mit Hausnummer.

## Trainingstermin

Ein wiederkehrender wöchentlicher Termin des Vereins: Wochentag, Beginn, dazu wahlweise ein Ende und eine Bezeichnung — "Samstag 10:30 – 12:00, Anfänger". Er existiert **für sich**, unabhängig davon, ob jemand dafür angemeldet ist. Der Verein pflegt seine Termine als Stundenplan, aus dem ausgewählt wird.

Er ist **kein einzelnes Datum**. Der Termin am 15.09. um 10:30 ist keine eigene Sache, sondern dieser Termin in dieser Woche. Festgehalten wird der Plan, nie die einzelne Stunde und nie, wer da war; siehe [ADR-0008](docs/adr/0008-trainingstermine-als-wochenplan.md).

Eine **Mitgliedschaft** ist für null bis drei Termine angemeldet — nicht die Person, denn welche Zeiten gelten, ist Teil der Vereinbarung eines Zeitraums. Bei einem Wiedereintritt wird deshalb neu ausgewählt: die neue Mitgliedschaft beginnt ohne Termine, die der alten bleiben als Historie stehen. Die Anmeldung selbst hat keinen eigenen Namen; man sagt, eine Mitgliedschaft ist für einen Termin angemeldet.

Ein Termin, den es nicht mehr gibt, wird **archiviert**: er verschwindet aus der Auswahl, bestehende Anmeldungen bleiben lesbar. Gelöscht wird er nicht — das nähme jeder daran angemeldeten Mitgliedschaft still ihre Trainingsfrequenz.

**Nicht zu verwechseln mit:** der Trainingsfrequenz — das ist die Anzahl. _Vermeiden:_ Trainingseinheit; das klingt nach der einzelnen Stunde am einzelnen Tag.

## Trainingsslot

**Abgelöst:** der wöchentliche Termin als Freitext an der Mitgliedschaft ("Samstag 10:30 Uhr"). An seine Stelle tritt der Verweis auf einen *Trainingstermin* aus dem Stundenplan; siehe [ADR-0008](docs/adr/0008-trainingstermine-als-wochenplan.md). _Vermeiden:_ Slot, auch im Sinne von „Platz" — eine Platzzahl je Termin führt der Verein nicht.

## Trainingsfrequenz

Wie oft pro Woche ein Mitglied trainieren darf: 1×, 2× oder 3×. Sie **ergibt sich aus der Anzahl der Trainingstermine**, für die seine Mitgliedschaft angemeldet ist, und ist keine davon unabhängige Angabe. Null Termine sind **keine Frequenz** und nicht "1×"; drei sind das Höchste, was der Verein vergibt — ein vierter ist ein Tippfehler, kein neuer Tarif.

Ein **archivierter** Termin zählt weiter mit, solange die Anmeldung steht. Sonst änderte das Aufräumen des Stundenplans stillschweigend die Vereinbarungen der Mitglieder. _Vermeiden:_ Mitgliedschaft, Tarif, Paket.

## Anmeldedatum

Der Tag, an dem ein Mitglied seine Anmeldung abgegeben hat. Liegt in der Regel vor dem Beginn der Mitgliedschaft.

**Nicht zu verwechseln mit:** dem Eintritt — das ist der Tag, ab dem die Mitgliedschaft läuft und der Beitrag fällig wird (üblicherweise ein Monatserster).

## Anmeldegebühr

Der einmalige Betrag, der beim Eintritt fällig war. Er ist ein **historischer Wert**: er hält fest, was tatsächlich gezahlt wurde, und wird nie neu berechnet. 0 € heißt „keine erhoben" — zwischen „keine" und „null Euro" unterscheidet der Verein nicht, denn geflossen ist in beiden Fällen nichts.

Sie hängt an der **Mitgliedschaft**, nicht an der Person: jeder Zeitraum hatte seine eigene Anmeldung. Bei einem Wiedereintritt wird sie deshalb nicht fortgeschrieben.

Ob sie schon **eingezogen** ist, hält die Mitgliedschaft seit [ADR-0013](docs/adr/0013-moneymoney-csv-export-fuer-lastschrifteinzug.md) zusätzlich fest — die einzige Ausnahme vom sonst rein historischen Charakter dieser Angabe. Gesetzt wird das ausschließlich vom MoneyMoney-Export, nachdem eine Zeile mit dieser Gebühr tatsächlich gespeichert wurde, nie von Hand.

**Nicht zu verwechseln mit:** dem *Beitrag* (monatlich, laufend).

## Kündigungsdatum

Der Tag, an dem ein Mitglied seine Kündigung erklärt hat. Das Datum, zu dem die Mitgliedschaft dann tatsächlich endet, ist der *Austritt*; er wird von Hand eingetragen. Das Formular schlägt den regulären Termin nach der Satzungsfrist vor (drei Monate zum Monatsende), gespeichert wird aber immer nur der eingetragene Tag — abweichende Fristen (Aufhebungsvertrag, Kulanz) bleiben erfassbar, und ein leeres Feld heißt weiterhin „Kündigung liegt vor, Termin noch offen".

## Kündigungsfrist

Der Zeitraum zwischen Kündigungsdatum und Austritt. Ein Mitglied in der Kündigungsfrist ist **noch aktiv** — es trainiert weiter und zahlt weiter.

Die **reguläre** Frist der Satzung sind drei Monate zum Monatsende: eine am 12.09. erklärte Kündigung wirkt zum 31.12. Sie ist der Vorschlag des Formulars (`service.RegulaererAustritt`) und keine Schranke — die tatsächliche Frist steht im Einzelfall im Austrittsdatum, nicht in dieser Regel.

## Status

Der Lebenszyklus-Zustand eines Mitglieds. Er wird **nicht gepflegt, sondern abgelesen**: *Neu* (Eintritt liegt in der Zukunft), *Aktiv* (Eintritt erreicht, Austritt nicht erreicht und keine Kündigung erfasst), *In Kündigungsfrist* (eine Kündigung ist erfasst, der Austritt aber noch nicht erreicht), *Ausgetreten* (Austritt erreicht). Einzige Ausnahme ist *ruhend*, das sich aus keinem Datum ergibt.

Beide Ränder zählen einschließend: am Eintrittstag ist das Mitglied aktiv, am Austrittstag bereits ausgetreten. *Aktiv* heißt dabei „Austritt nicht erreicht" und nicht „kein Austrittsdatum gesetzt" — wer in der Kündigungsfrist steht, hat eines und bleibt trotzdem in der Standardansicht. Erfasst ist eine Kündigung, sobald **eines** der beiden Daten steht: ein Kündigungsdatum ohne Termin heißt „erklärt, Termin offen", ein Austritt ohne Kündigungsdatum ist der Normalfall der Altbestände.

_Vermeiden:_ „Mitglied" und „Aktiv" als zwei verschiedene Zustände — das ist derselbe. Ebenso „Inaktiv" für ein ruhendes Mitglied.

## Ruhend

Ein Mitglied, das vorübergehend nicht trainiert (Verletzung, Auslandsaufenthalt), ohne zu kündigen. Die Mitgliedschaft läuft weiter, der Beitrag wird in dieser Zeit **nicht eingezogen**. _Vermeiden:_ Stillgelegt, Inaktiv, Pausiert.

## Lastschrift

Der Einzug des Beitrags vom Konto des Mitglieds. **Der Verein holt das Geld** — Mitglieder überweisen nicht. Daraus folgt: im Normalfall ist jeder Beitrag bezahlt, und nur die Ausnahme ist es wert, erfasst zu werden.

## Rücklastschrift

Eine fehlgeschlagene Lastschrift, die von der Bank zurückgegeben wird (etwa wegen fehlender Deckung). Der einzige Anlass, an dem der Verein sich mit dem Zahlungsstand eines einzelnen Mitglieds befassen muss.

## Rückstand

Das Kennzeichen, dass bei einem Mitglied nach einer Rücklastschrift Geld offen ist. Zweiwertig — *in Ordnung* oder *im Rückstand* — und von Hand gepflegt, begleitet von einer freien Notiz für den Vorgang.

**Abgelöst:** *bezahlt_bis* — ein Datum, bis zu dem der Beitrag als bezahlt gilt. Das Konzept setzt voraus, dass Mitglieder von sich aus zahlen und der Verein hinterherschaut; siehe [ADR-0006](docs/adr/0006-rueckstand-statt-bezahlt-bis.md). _Vermeiden:_ bezahlt_bis, Zahlungsstatus, „nicht bezahlt".

## Zahlung

Ein einzelnes Ereignis auf dem Vereinskonto, das einem Mitglied zugeordnet werden kann. **In v1 nicht modelliert** — v1 hält nur den *Rückstand* als Kennzeichen. In v2 kommen Zahlungen und Rücklastschriften aus dem MoneyMoney-Import.

## Google-Bewertung

Ob ein Mitglied den Verein bei Google bewertet hat. Ja oder nein. **Nicht zu verwechseln mit:** einer Bewertung *des* Mitglieds — der Verein bewertet seine Mitglieder nicht.

## Dokument

Eine am Mitglied abgelegte PDF-Datei. Es gibt genau **zwei Arten**: den *Vertrag*, den der Verein einscannt und hochlädt, und die *Rechnung*, die die App selbst erzeugt. Andere Anhänge kennt v1 nicht — die Ablage ist ein Ort für Vereinsunterlagen und kein allgemeines Dateifach.

Dokumente liegen **in der Datenbank**, nicht als Dateien daneben; siehe [ADR-0007](docs/adr/0007-dokumente-als-blob-in-sqlite.md). Sie lassen sich ersetzen und entfernen — ohne Papierkorb und ohne Versionen, denn das Papier liegt beim Verein ohnehin noch.

## Vertrag

Die unterschriebene Vereinbarung über eine Mitgliedschaft, als *Dokument* abgelegt.

Er hängt an der **Mitgliedschaft**, nicht an der Person: wer austritt und Jahre später wiederkommt, unterschreibt einen neuen, und jeder Vertrag bleibt bei seinem Zeitraum. Am Mitglied abgelegt stünden zwei Verträge nebeneinander, ohne dass man sähe, welcher gilt.

## Rechnung

Ein vom Verein geschriebenes Dokument über eine Leistung **außerhalb des Beitrags** — typisch ein Einzeltraining. Sie entsteht von Hand: Empfänger, Positionen mit Menge und Einzelpreis, Steuersatz, Nummer. Die App setzt daraus ein PDF und legt es am *Mitglied* ab.

Eine Rechnung ist hier ein **Dokument und kein offener Posten**. Die App führt weder einen Rechnungsstatus noch einen Zahlungseingang, und die fortlaufende Nummer vergibt der Verein in seiner Buchhaltung, nicht die App; siehe [ADR-0009](docs/adr/0009-rechnungen-und-monatssoll-ohne-zahlungsmodell.md). _Vermeiden:_ „offene Rechnung", „Rechnung bezahlt", Mahnung — davon gibt es in v1 nichts.

Der **Empfänger** ist frei eintragbar und wird aus dem Mitglied vorbelegt: ein Einzeltraining nimmt auch, wer nie eintritt. Für einen Empfänger ohne Mitglied fällt das PDF nur heraus, statt abgelegt zu werden.

**Nicht zu verwechseln mit:** dem *Beitrag*. Der wird per Lastschrift eingezogen und nie in Rechnung gestellt.

## Vereinsdaten

Name, Anschrift, Bankverbindung, Fußzeile und *Vereinslogo* des Vereins selbst — alles, was auf einer Rechnung über dem Inhalt steht. Die einzigen Daten der App, die kein Mitglied betreffen; sie werden einmal von Hand gepflegt und nicht importiert.

## Vereinslogo

Das Bild des Vereins. Es erscheint im Verein-Bereich der App und im Briefkopf jeder erzeugten Rechnung, ist aber **freiwillig** wie jede andere Angabe der Vereinsdaten — fehlt es, bleibt der Briefkopf wie bisher reiner Text.

Welches Format hochgeladen wurde und welches gespeichert ist, gehört nicht zum Begriff — das ist Ablage, kein Modell; siehe [ADR-0012](docs/adr/0012-vereinslogo-eigene-spalte-und-rasterisierung.md).

**Nicht zu verwechseln mit:** der *Fußzeile* — Ticket 23 hatte dort einmal vorgesehen, dass später "auch ein Bild landen kann". Dabei ist es nicht geblieben: die Fußzeile bleibt reiner Text, das Logo ist eine eigene, unabhängig von ihr entfernbare Angabe.

**Nicht zu verwechseln mit:** *Dokument* — ein Dokument ist eine PDF-Datei an einem Mitglied oder einer Mitgliedschaft. Das Vereinslogo hängt an keinem der beiden, sondern an den Vereinsdaten selbst, und ist kein PDF.

## MoneyMoney-Export

Eine CSV-Datei, mit der der Verein den Lastschrifteinzug eines Monats über MoneyMoney anstößt — eine Zeile je Mitgliedschaft, die diesen Monat einzieht, dieselbe Menge wie das Monatssoll ([ADR-0013](docs/adr/0013-moneymoney-csv-export-fuer-lastschrifteinzug.md)). Die Datei geht aus der App heraus, nicht herein.

**Nicht zu verwechseln mit:** dem in CLAUDE.md ausgeschlossenen *MoneyMoney-Import* — jener meint Kontoumsätze, die in die App **hinein** sollen, um Rücklastschriften automatisch zu erkennen (v2). Dieser Export geht die andere Richtung und ersetzt nur das Abtippen in MoneyMoney, das der Verein ohnehin schon von Hand macht.

## Monatssoll

Die Summe der Beiträge, die im laufenden Monat eingezogen werden: alle aktiven Mitgliedschaften **einschließlich** derer in der Kündigungsfrist, **ohne** ruhende und ohne solche, deren Eintritt noch bevorsteht.

Es ist ein **Soll und kein Ist**. Was tatsächlich einging, weiß die Bank; v1 kennt keine Zahlungen. Deshalb gibt es auch keinen Verlauf über Monate: die Datenbank führt je Mitgliedschaft einen Beitrag, den heutigen, und eine Kurve daraus wäre eine Hochrechnung im Gewand einer Messung. Siehe [ADR-0009](docs/adr/0009-rechnungen-und-monatssoll-ohne-zahlungsmodell.md). _Vermeiden:_ Einnahmen, Umsatz, Beitragsaufkommen.
