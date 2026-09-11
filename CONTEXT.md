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

## Trainingsslot

Ein konkreter wöchentlicher Trainingstermin, für den ein Mitglied angemeldet ist — bezeichnet durch Wochentag und Uhrzeit (z. B. "Samstag 10:30 Uhr"). Ein Mitglied hat null bis drei Slots.

## Trainingsfrequenz

Wie oft pro Woche ein Mitglied trainieren darf: 1×, 2× oder 3×. Sie **ergibt sich aus der Anzahl der Trainingsslots** des Mitglieds und ist keine davon unabhängige Angabe. _Vermeiden:_ Mitgliedschaft, Tarif, Paket.

## Anmeldedatum

Der Tag, an dem ein Mitglied seine Anmeldung abgegeben hat. Liegt in der Regel vor dem Beginn der Mitgliedschaft.

**Nicht zu verwechseln mit:** dem Eintritt — das ist der Tag, ab dem die Mitgliedschaft läuft und der Beitrag fällig wird (üblicherweise ein Monatserster).

## Kündigungsdatum

Der Tag, an dem ein Mitglied seine Kündigung erklärt hat. Das Datum, zu dem die Mitgliedschaft dann tatsächlich endet, ist der *Austritt*; er wird von Hand eingetragen und nicht aus der Kündigungsfrist berechnet.

## Kündigungsfrist

Der Zeitraum zwischen Kündigungsdatum und Austritt. Ein Mitglied in der Kündigungsfrist ist **noch aktiv** — es trainiert weiter und zahlt weiter.

## Status

Der Lebenszyklus-Zustand eines Mitglieds. Er wird **nicht gepflegt, sondern abgelesen**: *Neu* (Eintritt liegt in der Zukunft), *Aktiv* (Eintritt erreicht, kein Austritt), *In Kündigungsfrist* (Kündigung erklärt, Austritt liegt in der Zukunft), *Ausgetreten* (Austritt erreicht). Einzige Ausnahme ist *ruhend*, das sich aus keinem Datum ergibt.

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
