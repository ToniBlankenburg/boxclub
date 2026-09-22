# ADR-0013: CSV-Export für den Lastschrifteinzug über MoneyMoney

**Status:** Accepted — löst den SEPA-Export-Ausschluss aus [ADR-0006](0006-rueckstand-statt-bezahlt-bis.md) in diesem einen Punkt ab. Der automatisch erzeugte Verwendungszweck ist seit [ADR-0016](0016-eigener-verwendungszweck-in-vereinsdaten.md) optional überschreibbar.
**Datum:** 2026-09-18

## Kontext

ADR-0006 hatte festgelegt: *"Die App speichert die IBAN als reines Textfeld,
ohne Validierung und ohne SEPA-Export."* Begründet war das damit, dass es für
den Lastschrifteinzug selbst keine App-Unterstützung brauche — die Bank weiß,
was sie einzieht, die App nur, was gilt.

In der Praxis stößt der Verein den Einzug aber nicht direkt bei der Bank an,
sondern über MoneyMoney (eine Banking-App), die dafür eine CSV-Datei
importiert. Diese Datei wurde bisher von Hand aus der Excel-Tabelle
abgetippt — genau die Art doppelter Datenhaltung, die diese App ablösen soll
(derselbe Gedanke wie bei der IBAN selbst, ADR-0006). Die alte Excel-Tabelle
hat dafür ein eigenes Blatt "Money Money" mit den Spalten Mandatsreferenz,
Zahlungspflichtiger, IBAN, BIC, Betrag, Verwendungszweck, Unterschrieben am.

Das ist etwas anderes als das in CLAUDE.md ausgeschlossene "MoneyMoney CSV
import/auto-matching": jenes meint den *Import* von Kontoumsätzen in die App
zum Abgleich (Rücklastschriften automatisch erkennen — das bleibt v2, siehe
ADR-0006). Hier geht die Datei in die andere Richtung: aus der App heraus, um
den Einzug überhaupt erst anzustoßen.

## Entscheidung

Ein Export erzeugt eine CSV-Zeile je Mitgliedschaft, die diesen Monat
tatsächlich Beitrag einzieht — dieselbe Menge wie das Monatssoll des
Dashboards (aktiv und in Kündigungsfrist, ohne ruhende und ohne noch nicht
begonnene). Die Spalten entsprechen exakt dem Blatt "Money Money" der alten
Tabelle:

- **Mandatsreferenz** ist die Mitglieds-ID — wie in der alten Tabelle auch:
  keine echte SEPA-Mandatsreferenz, sondern die laufende Nummer
  (CONTEXT.md → Mitglieds-ID).
- **Zahlungspflichtiger** ist Vorname und Nachname.
- **IBAN** wird unverändert übernommen, reiner Text wie überall (ADR-0006).
- **BIC** bleibt leer — das Mitglied trägt keines, eine deutsche IBAN braucht
  es für die Lastschrift nicht mehr.
- **Betrag** ist der monatliche Beitrag, zuzüglich einer noch nicht
  eingezogenen Anmeldegebühr (siehe unten).
- **Verwendungszweck** ist Vereinsname und Beitragsmonat, ergänzt um
  "Anmeldegebühr + ", wenn die Zeile sie enthält.
- **Unterschrieben am** ist das Anmeldedatum, sonst der Eintritt
  (CONTEXT.md → Anmeldedatum).

**Die Anmeldegebühr braucht einen neuen Bearbeitungsstand.** Anders als der
Beitrag ist sie ein einmaliger Betrag, der nicht in jedem Monatslauf wieder
auftauchen darf. Die Mitgliedschaft bekommt dafür die Spalte
`anmeldegebuehr_eingezogen`: 0, bis eine Exportzeile sie tatsächlich enthalten
und die Datei tatsächlich gespeichert wurde, danach 1. Das ist die einzige
Ausnahme von "die App führt keine Zahlungshistorie" (ADR-0006) — nötig, weil
sich "schon eingezogen" hier aus nichts anderem ableiten lässt, anders als der
Rückstand, den die Bank über die Rücklastschrift von selbst meldet.

Geschrieben wird die Datei über denselben Speicherziel-Dialog wie Vertrag und
Rechnung (app/dokument.go, app/rechnung.go) — das WebView von Wails kennt
keine Downloads. Der Dialog nimmt jetzt eine Dateifilter-Beschriftung und ein
Muster entgegen, damit er für CSV nicht denselben PDF-Filter zeigt wie bisher.

## Konsequenzen

**Gut:**

- Das doppelte Abtippen in MoneyMoney entfällt — genau der Zustand, den auch
  die IBAN-Speicherung selbst schon abgeschafft hat.
- Die Anmeldegebühr geht nicht mehr verloren, wenn ein Export einmal ausfällt:
  `anmeldegebuehr_eingezogen` ist ein Zustand und kein Kalendermonat, ein
  übersprungener Lauf holt sie beim nächsten automatisch nach.

**Schlecht:**

- **Die App weiß weiterhin nicht, ob der Einzug tatsächlich klappt.** Markiert
  wird "die Zeile wurde geschrieben", nicht "das Geld kam an" — eine
  Rücklastschrift bleibt so unsichtbar wie bisher, dafür gibt es weiterhin nur
  den Rückstand (ADR-0006).
- Ein Export, der zwar gespeichert, aber nie tatsächlich in MoneyMoney
  importiert wird, markiert die Anmeldegebühr trotzdem als eingezogen — die
  Markierung hängt am Speichern der Datei, nicht an ihrer Verwendung. Das
  entspricht demselben Vertrauen, das die App schon beim Rückstand aufbringt:
  sie bildet ab, was der Admin ihr sagt, nicht was auf dem Konto passiert.
- Der Speicherziel-Dialog ist jetzt parametrisiert statt fest auf PDF verdrahtet
  — eine kleine Signaturänderung an drei Aufrufstellen (Vertrag, Rechnung,
  dieser Export).

## Alternativen

- **Eine echte SEPA-XML-Datei (pain.008) statt CSV.** Der correcte, bankseitig
  vorgesehene Weg für Sammellastschriften — aber ungleich aufwendiger zu
  erzeugen und zu prüfen, und MoneyMoney selbst erwartet für seinen eigenen
  CSV-Import ohnehin kein XML. Für 50–200 Mitglieder ist das eine Größenordnung
  Aufwand, die der Nutzen nicht rechtfertigt.
- **Gar kein Export, nur die IBAN-Liste zum Nachschlagen.** Verworfen, weil
  genau das die Excel bisher schon leistet und die Handarbeit — das Abtippen in
  MoneyMoney — die eigentliche Zeitverschwendung ist, die abgelöst werden soll.
- **Anmeldegebühr immer nur im Eintrittsmonat addieren, ohne neue Spalte.**
  Einfacher, aber verliert die Gebühr ersatzlos, wenn der Export in diesem
  einen Monat ausfällt oder doppelt läuft, verdoppelt sie. Die neue Spalte ist
  die robustere, wenn auch nicht ganz so einfache Lösung.
