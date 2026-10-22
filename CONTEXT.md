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

Ein Termin kann **nur für Frauen** sein. Dann werden nur Mitglieder angemeldet, deren Geschlecht „Frau“ lautet (ohne Rücksicht auf Groß-/Kleinschreibung); jede andere Angabe, auch eine leere, wird beim Anmelden abgewiesen. Das Geschlecht bleibt Freitext — geprüft wird gegen den einen Wert „Frau“, nicht gegen eine Werteliste. Wer schon angemeldet ist, bleibt es, auch wenn sich die Angabe später ändert oder der Termin erst danach auf „nur Frauen“ gestellt wird: dieselbe Haltung wie beim archivierten Termin, die Vereinbarung wird nicht still geändert.

Ein Termin kann **nur für Kinder** sein. Dann werden nur Mitglieder angemeldet, die zum Zeitpunkt des Anmeldens ein *Kind* sind — der abgeleitete Status, keine eigene Altersgrenze; ein Mitglied ohne Geburtsdatum und ohne Kennzeichen wird abgewiesen. Wer schon angemeldet ist, bleibt es auch nach dem 18. Geburtstag; das Mitgliedsformular zeigt dann am angekreuzten Termin einen Hinweis, ändert aber nichts (wie beim *Kontoinhaber*). Kinder dürfen weiter in jeden Termin ohne Einschränkung. Ein Termin ist **entweder** nur für Frauen **oder** nur für Kinder **oder** offen für alle, nie zwei davon: er hat eine *Zielgruppe*. „Frau“ bleibt der Wert im Feld Geschlecht, keine Altersaussage — eine Minderjährige mit „Frau“ darf in einen Frauentermin.

Ein Termin, den es nicht mehr gibt, wird **archiviert**: er verschwindet aus der Auswahl, bestehende Anmeldungen bleiben lesbar. Gelöscht wird er nicht — das nähme jeder daran angemeldeten Mitgliedschaft still ihre Trainingsfrequenz.

**Nicht zu verwechseln mit:** der Trainingsfrequenz — das ist die Anzahl. _Vermeiden:_ Trainingseinheit; das klingt nach der einzelnen Stunde am einzelnen Tag.

## Teilnehmerliste

Die Mitglieder, die für einen Trainingstermin angemeldet sind — die Antwort auf „wer trainiert samstags um 10:30?". Sie wird **abgelesen, nicht geführt**: es gibt keine Liste, die jemand pflegt, sondern nur die Anmeldungen der Mitgliedschaften, nach Termin gelesen.

Aufgenommen ist, wer **Neu**, **Aktiv** oder **In Kündigungsfrist** ist und dessen Mitgliedschaft **heute nicht stillgelegt** ist. Maßgeblich ist die Mitgliedschaft, die auch die Mitgliederliste zeigt; ein früherer Zeitraum mit denselben Terminen zählt nicht. Ein Ausgetretener steht nicht darin, seine Anmeldung bleibt nur Historie. Ein *Neuer* steht schon darin, obwohl er noch nicht trainiert — bewusst, weil er sich für den Termin verbindlich angemeldet hat. Die Liste zeigt nur die Namen, alphabetisch nach Nachname und Vorname, dazu die Anzahl je Termin; der Anzahl steht keine Obergrenze gegenüber (siehe [ADR-0008](docs/adr/0008-trainingstermine-als-wochenplan.md)).

**Nicht zu verwechseln mit:** einer Anwesenheitsliste. Die Teilnehmerliste sagt, wer *angemeldet* ist, nie, wer an einem bestimmten Tag *da war* — das wäre Anwesenheit und ist aus v1 ausgeschlossen. _Vermeiden:_ Anwesenheitsliste, „wer da ist", Belegung; die beiden ersten behaupten Anwesenheit, das dritte eine Platzzahl.

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

Der Lebenszyklus-Zustand eines Mitglieds. Er wird **nicht gepflegt, sondern abgelesen**: *Neu* (Eintritt liegt in der Zukunft), *Aktiv* (Eintritt erreicht, Austritt nicht erreicht und keine Kündigung erfasst), *In Kündigungsfrist* (eine Kündigung ist erfasst, der Austritt aber noch nicht erreicht), *Ausgetreten* (Austritt erreicht). Auch *stillgelegt* ist kein Zustand dieser Reihe, sondern ein Merkmal daneben, das ebenfalls abgelesen wird — aus den *Stilllegungen* der Mitgliedschaft.

Beide Ränder zählen einschließend: am Eintrittstag ist das Mitglied aktiv, am Austrittstag bereits ausgetreten. *Aktiv* heißt dabei „Austritt nicht erreicht" und nicht „kein Austrittsdatum gesetzt" — wer in der Kündigungsfrist steht, hat eines und bleibt trotzdem in der Standardansicht. Erfasst ist eine Kündigung, sobald **eines** der beiden Daten steht: ein Kündigungsdatum ohne Termin heißt „erklärt, Termin offen", ein Austritt ohne Kündigungsdatum ist der Normalfall der Altbestände.

_Vermeiden:_ „Mitglied" und „Aktiv" als zwei verschiedene Zustände — das ist derselbe. Ebenso „Inaktiv" für ein stillgelegtes Mitglied.

## Stilllegung

Ein **Zeitraum**, in dem ein Mitglied vorübergehend nicht trainiert (Verletzung, Auslandsaufenthalt), ohne zu kündigen. Er hat ein `von` und ein `bis` — beide Pflicht, kein „bis auf Weiteres" — und optional einen Grund als Freitext. Die Mitgliedschaft läuft weiter, der Beitrag wird für die Zeit **nicht eingezogen**.

Sie hängt an der **Mitgliedschaft**: eine Mitgliedschaft kann mehrere Stilllegungen haben, die zusammen ihre Historie bilden. Bei einem Wiedereintritt bleiben sie bei der alten Mitgliedschaft, die neue beginnt ohne. Ein Mitglied ist **stillgelegt**, wenn der heutige Tag in einer Stilllegung liegt — das wird abgelesen, nicht gepflegt; es gibt keinen Schalter. Beide Ränder zählen einschließend.

Zwei Stilllegungen einer Mitgliedschaft **überlappen nie**. Solange jemand stillgelegt ist, kommt keine neue hinzu; erst nach „wieder aktiv" (das setzt `bis` auf gestern — beide Ränder zählen, mit „heute“ wäre das Mitglied noch bis Mitternacht stillgelegt —, eine heute oder später beginnende Stilllegung wird dabei gelöscht). Eine Stilllegung darf über den Austritt hinausreichen; gespeichert wird, was eingetragen ist, nach dem Austritt wirkt sie nicht mehr.

**Wirkung auf den Beitrag:** ein Kalendermonat gilt als *Stilllegungsmonat*, sobald **irgendein Tag** darin in einer Stilllegung liegt — der Zeitraum 20.03. bis 05.04. macht März **und** April zu Stilllegungsmonaten. Für sie wird kein Beitrag eingezogen; Monatssoll und Lastschriftexport folgen derselben Regel. Es gibt keine anteilige Berechnung.

**Wirkung auf die Laufzeit:** jeder Stilllegungsmonat verlängert die *Mindestlaufzeit*. Eine **Kündigung** lässt sich in einer Stilllegung nicht erfassen; die Schaltfläche erklärt das beim Darüberfahren. Das Gegenteil gilt nicht: wer schon gekündigt hat, kann stillgelegt werden.

Unabhängig vom *Rückstand*: für einen Stillgelegten geht keine Lastschrift raus, es kann also kein neuer Rückstand entstehen, ein bestehender bleibt (ADR-0006). Eine Stilllegung ist **keine Anwesenheitsangabe** und kein Zahlungsvorgang.

_Vermeiden:_ Ruhend, Ruhezeit, Inaktiv, Pausiert. Die Excel-Werte `Stillgelegt` und `Inakiv` beschreiben denselben Sachverhalt, tragen aber keinen Zeitraum; der Import legt daraus keine Stilllegung an, sondern meldet sie im Fehlerbericht (siehe [ADR-0022](docs/adr/0022-stilllegung-als-zeitraum.md)).

## Mindestlaufzeit

Die Zahl der Monate ab Eintritt, für die sich ein Mitglied mit einer Mitgliedschaft gebunden hat. Optional; leer heißt: keine Bindung. Sie hängt an der **Mitgliedschaft** — wie Beitrag und Anmeldegebühr gilt sie für genau diesen Vertrag, ein Wiedereintritt beginnt ohne und bekommt eine neue.

Das **Ende der Mindestlaufzeit** wird abgelesen: Eintritt plus Mindestlaufzeit plus die Zahl der *Stilllegungsmonate* der Mitgliedschaft. Es wird nie gespeichert. Es bindet nichts: liegt ein erfasster Austritt davor, **warnt** das Kündigungsformular, speichert aber. Der Vorschlag dort (drei Monate zum Monatsende, siehe *Kündigungsfrist*) bleibt unverändert.

_Vermeiden:_ Vertragslaufzeit, Bindefrist.

## Lastschrift

Der Einzug des Beitrags vom Konto des Mitglieds. **Der Verein holt das Geld** — Mitglieder überweisen nicht. Daraus folgt: im Normalfall ist jeder Beitrag bezahlt, und nur die Ausnahme ist es wert, erfasst zu werden.

## Rücklastschrift

Eine fehlgeschlagene Lastschrift, die von der Bank zurückgegeben wird (etwa wegen fehlender Deckung). Der einzige Anlass, an dem der Verein sich mit dem Zahlungsstand eines einzelnen Mitglieds befassen muss.

## Rückstand

Das Kennzeichen, dass bei einem Mitglied nach einer Rücklastschrift Geld offen ist. Zweiwertig — *in Ordnung* oder *im Rückstand* — und von Hand gepflegt, begleitet von einer freien Notiz für den Vorgang.

**Abgelöst:** *bezahlt_bis* — ein Datum, bis zu dem der Beitrag als bezahlt gilt. Das Konzept setzt voraus, dass Mitglieder von sich aus zahlen und der Verein hinterherschaut; siehe [ADR-0006](docs/adr/0006-rueckstand-statt-bezahlt-bis.md). _Vermeiden:_ bezahlt_bis, Zahlungsstatus, „nicht bezahlt".

## Zahler

Wer die Lastschrift zahlt, getrennt von der Frage, wer trainiert. Eine rein technische Entität mit Name und IBAN; jede Mitgliedschaft zeigt auf genau einen. Bei Erwachsenen ist der Zahler die Person selbst und unsichtbar. Dieselbe normalisierte IBAN heißt derselbe Zahler, daher teilen sich Geschwister mit einem Konto einen; ein Zahler ohne Mitgliedschaft wird automatisch gelöscht. Siehe [ADR-0021](docs/adr/0021-zahler-als-technische-entitaet.md).

**Nicht zu verwechseln mit:** dem Mitglied. Das Mitglied trainiert und hat Anschrift, Rückstand und Dokumente; der Zahler hat nur Name und IBAN. _Vermeiden:_ „Zahler" in der Oberfläche — dort heißt es *Kontoinhaber*. Ebenso Elternteil, Erziehungsberechtigter oder Rechnungsempfänger: der Zahler ist nur, wer zahlt, nicht, in welchem Verhältnis er zum Mitglied steht.

## Kind

Ein Mitglied, das heute jünger als 18 ist. Der Status wird bei jedem Lesen aus dem Geburtsdatum **abgeleitet** und nie gespeichert. Nur ohne Geburtsdatum zählt ein von Hand gesetztes Kennzeichen an der Person; neben einem Geburtsdatum wird es ignoriert. Nur bei einem Kind lässt sich ein abweichender *Kontoinhaber* eintragen.

**Nicht zu verwechseln mit:** der Altersangabe oder einem Mitgliedstarif. Es gibt keine Kindermitgliedschaft mit eigenem Beitrag (ADR-0005). _Vermeiden:_ Jugendlicher, Minderjähriger als Feldname, „Kind" als gespeicherte Spalte oder Filter.

## Kontoinhaber

Das Wort der Oberfläche für den *Zahler*, wenn er nicht das Mitglied selbst ist: das Feld „Abweichender Kontoinhaber" im Formular, daneben die eine IBAN-Zeile. Leer heißt: das Mitglied zahlt selbst. Wird ein Mitglied mit Kontoinhaber volljährig, ändert sich nichts an den Abbuchungen; das Formular zeigt nur einen Hinweis, bis der Admin das Feld leert.

**Nicht zu verwechseln mit:** dem Mitglied und dem Rechnungsempfänger. Der Kontoinhaber ist im MoneyMoney-Export der „Zahlungspflichtige" und füllt den Empfängernamen einer Rechnung vor (überschreibbar); die Anschrift kommt weiter vom Mitglied. _Vermeiden:_ Zahler (in der UI), Elternteil.

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

Name, Anschrift, Steuernummer, Bankverbindung, Fußzeile, *Vereinslogo* und ein eigener Verwendungszweck-Text für den *MoneyMoney-Export* — alles, was auf einer Rechnung über dem Inhalt steht oder den monatlichen Lastschriftlauf betrifft. Die einzigen Daten der App, die kein Mitglied betreffen; sie werden einmal von Hand gepflegt und nicht importiert.

## Vereinslogo

Das Bild des Vereins. Es erscheint im Verein-Bereich der App und im Briefkopf jeder erzeugten Rechnung, ist aber **freiwillig** wie jede andere Angabe der Vereinsdaten — fehlt es, bleibt der Briefkopf wie bisher reiner Text.

Welches Format hochgeladen wurde und welches gespeichert ist, gehört nicht zum Begriff — das ist Ablage, kein Modell; siehe [ADR-0012](docs/adr/0012-vereinslogo-eigene-spalte-und-rasterisierung.md).

**Nicht zu verwechseln mit:** der *Fußzeile* — Ticket 23 hatte dort einmal vorgesehen, dass später "auch ein Bild landen kann". Dabei ist es nicht geblieben: die Fußzeile bleibt reiner Text, das Logo ist eine eigene, unabhängig von ihr entfernbare Angabe.

**Nicht zu verwechseln mit:** *Dokument* — ein Dokument ist eine PDF-Datei an einem Mitglied oder einer Mitgliedschaft. Das Vereinslogo hängt an keinem der beiden, sondern an den Vereinsdaten selbst, und ist kein PDF.

## MoneyMoney-Export

Eine CSV-Datei, mit der der Verein den Lastschrifteinzug eines Monats über MoneyMoney anstößt — eine Zeile je Mitgliedschaft, die diesen Monat einzieht, dieselbe Menge wie das Monatssoll ([ADR-0013](docs/adr/0013-moneymoney-csv-export-fuer-lastschrifteinzug.md)). Die Datei geht aus der App heraus, nicht herein.

**Nicht zu verwechseln mit:** dem in CLAUDE.md ausgeschlossenen *MoneyMoney-Import* — jener meint Kontoumsätze, die in die App **hinein** sollen, um Rücklastschriften automatisch zu erkennen (v2). Dieser Export geht die andere Richtung und ersetzt nur das Abtippen in MoneyMoney, das der Verein ohnehin schon von Hand macht.

Der Verwendungszweck ist zum Teil überschreibbar: die *Vereinsdaten* tragen dafür ein eigenes, optionales Textfeld ([ADR-0016](docs/adr/0016-eigener-verwendungszweck-in-vereinsdaten.md)). Leer heißt automatisch wie oben; gesetzt ersetzt es nur den Vereinsname-und-Monat-Teil — der Anmeldegebühr-Hinweis bleibt davon unberührt und immer automatisch, weil er Faktenstand der einzelnen Zeile ist, kein Stiltext.

## Sicherung

Eine vollständige Kopie der Datenbank in einer Datei, die der Verein selbst irgendwohin legt (USB-Stick, Cloud-Ordner). Sie enthält alles, auch die *Dokumente* und die *Vereinsdaten*, und lässt sich in der App wieder *wiederherstellen*. Sie wird von Hand ausgelöst, nie automatisch.

*Wiederherstellen* ersetzt die gesamte aktuelle Datenbank durch die Sicherung — kein Zusammenführen. Vorher bleibt eine Kopie des bisherigen Stands neben der Datenbank liegen.

**Nicht zu verwechseln mit:** dem *Export der Mitgliederliste* — jener ist eine Excel-Datei nur mit den Mitgliedern und ihren Mitgliedschaften, ohne *Dokumente*, und dient dem Weitergeben und Auswerten. Aus ihm lässt sich nichts wiederherstellen. _Vermeiden:_ Backup, Restore (in der UI).

## Monatssoll

Die Summe der Beiträge, die im laufenden Monat eingezogen werden: alle aktiven Mitgliedschaften **einschließlich** derer in der Kündigungsfrist, **ohne** Mitgliedschaften, für die der laufende Monat ein *Stilllegungsmonat* ist, und ohne solche, deren Eintritt noch bevorsteht.

Es ist ein **Soll und kein Ist**. Was tatsächlich einging, weiß die Bank; v1 kennt keine Zahlungen. Deshalb gibt es auch keinen Verlauf über Monate: die Datenbank führt je Mitgliedschaft einen Beitrag, den heutigen, und eine Kurve daraus wäre eine Hochrechnung im Gewand einer Messung. Siehe [ADR-0009](docs/adr/0009-rechnungen-und-monatssoll-ohne-zahlungsmodell.md). _Vermeiden:_ Einnahmen, Umsatz, Beitragsaufkommen.
