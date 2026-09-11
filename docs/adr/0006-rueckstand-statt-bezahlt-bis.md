# ADR-0006: Kein `bezahlt_bis` — der Verein zieht ein und verwaltet nur Rückstände

**Status:** Accepted
**Datum:** 2026-09-11

## Kontext

Die Spec baut auf der Annahme auf, dass Mitglieder ihre Beiträge **überweisen**
und der Vereinsadmin hinterherschaut: *"der Nutzer muss regelmäßig prüfen, welche
Mitglieder ihren Beitrag bezahlt haben"*. Daraus entstanden `mitglied.bezahlt_bis`,
die zweistufige Ampel (Story 5), der Zahlungsstatus-Filter (Story 3), das manuelle
Setzen des Datums (Story 8) und der neutrale Zustand *nicht gesetzt* aus Ticket 05.

Die Bestandsaufnahme der bestehenden Excel-Tabelle
([excel-vorlage.md](../../.scratch/boxclub-v1/excel-vorlage.md)) widerlegt die
Annahme zweifach:

- Jede Zeile führt eine **IBAN**. Der Verein zieht die Beiträge per
  SEPA-Lastschrift ein; es gibt keine Überweiser. (Die Excel-Spalte
  `Mandatsreferenz` ist trotz ihres Namens keine SEPA-Mandatsreferenz, sondern
  die laufende Mitglieds-Nummer — siehe [CONTEXT.md](../../CONTEXT.md).)
- Die Tabelle enthält **keine einzige Spalte zum Zahlungsstand**. Kein
  Bezahlt-bis-Datum, kein Kennzeichen, nichts. Die Werteliste `Bank`
  (*Eingezogen, Bezahlt, Offen, …*) im Blatt "Quelle" gehört zu keiner Spalte —
  sie ist verwaist, ebenso wie vier Textwerte am Ende der `Beitrag`-Liste.

Das ist keine Lücke in der Tabelle, sondern die logische Folge des Einzugsverfahrens:
**wer das Geld selbst holt, führt keine Bezahlt-Liste.** Das macht die Bank. Ein
`bezahlt_bis` hätte in diesem Verein nie eine Quelle gehabt, und die Ampel wäre
dauerhaft grün gewesen — bei 200 Mitgliedern ein Signal ohne Information.

## Entscheidung

`bezahlt_bis` entfällt. An seine Stelle tritt am Mitglied ein **Rückstands-Kennzeichen**:
zweiwertig (*in Ordnung* / *im Rückstand*) plus ein freies Notizfeld für den Vorgang
("Rücklastschrift Oktober, angeschrieben am 05.10.").

Ampel und Filter aus Ticket 05 und 06 bleiben als Bedienkonzept erhalten, **wechseln
aber ihre Bedeutung**: Rot heißt nicht mehr "hat nicht bezahlt", sondern
**"Lastschrift ist geplatzt"**. Aus einer Liste, die im Normalfall alle 200 Mitglieder
bewertet, wird eine Ausnahmeliste mit einer Handvoll Einträgen — also die Liste, die
der Admin tatsächlich abarbeitet.

**Mitentschieden:** Die App speichert die **IBAN** als reines Textfeld, ohne
Validierung und ohne SEPA-Export. Datensparsamkeit wäre das Gegenargument (die
Mandate liegen auch im Bankportal), aber der Zweck des Projekts ist, die Excel
*abzulösen*. Was die App nicht halten kann, hält die Excel weiter — und dann laufen
zwei Systeme parallel, genau der Zustand, der abgeschafft werden soll. Story 15
(Daten verlassen die Platte nicht) plus FileVault ist die Antwort auf die
Datenschutz-Seite.

## Konsequenzen

**Gut:**

- Das Modell beschreibt, was der Verein wirklich tut. Ein importiertes Mitglied
  startet zutreffend als *in Ordnung*, weil bei Lastschrift die Abwesenheit einer
  Rückgabe tatsächlich bedeutet, dass gezahlt wurde.
- Der neutrale Zustand *nicht gesetzt* aus Ticket 05 wird **überflüssig** und
  entfällt. Er existierte nur, weil ein leeres Datum nicht als "nicht bezahlt"
  gelten durfte; ein Kennzeichen mit zwei Werten hat dieses Problem nicht.
- Es gibt erstmals einen Ort für die Notiz zum Vorgang. Heute steht die nirgends.
- v2 dockt additiv an: eine Rücklastschrift taucht im MoneyMoney-Import auf und
  kann das Kennzeichen künftig automatisch setzen.

**Schlecht:**

- **Ticket 05 wird zurückgebaut**, der Zahlungsstatus-Filter aus **Ticket 06**
  umgehängt. Die Glossar-Einträge *bezahlt_bis* und *Statusanzeige* sind ersetzt,
  die **Stories 3, 5 und 8 umformuliert**.
- **Keine Zahlungshistorie in v1.** Ein Mitglied, dessen Lastschrift im März
  platzt und das im April nachzahlt, hinterlässt außer der Notiz keine Spur.
  Bewusst: Historie ist v2 mit echten Kontodaten, nicht handgepflegt.
- Das Kennzeichen hat **keine Importquelle**. Nach dem Excel-Import steht jedes
  Mitglied auf *in Ordnung*, auch wenn dort gerade etwas offen ist. Das muss der
  Admin einmalig nacharbeiten — es gibt keine Daten, aus denen man es ableiten könnte.
- Ein **ruhendes** Mitglied kann nicht in Rückstand geraten, weil für es keine
  Lastschrift losgeschickt wird. Die beiden Felder sind unabhängig, hängen aber
  fachlich zusammen.

## Alternativen

- **`bezahlt_bis` behalten und umdeuten** ("bis hierhin ist nichts offen") — ein
  Datum, das niemand pflegt, weil es im Regelfall niemanden interessiert. Es wäre
  nach drei Monaten flächendeckend veraltet und die Ampel flächendeckend rot.
- **Rücklastschriften als einzelne Vorfälle** in einer eigenen Tabelle — echte
  Historie, aber handgepflegt für ein Ereignis, das ein paar Mal im Monat vorkommt.
  Das ist die v2-Lösung, und dort kommt sie aus den Kontodaten statt aus Handarbeit.
- **Gar kein Zahlungsfeld in v1** — wäre konsequent (die Bank weiß es ohnehin),
  aber dann hat die App zum Thema Geld nichts zu sagen und die Notiz zum Vorgang
  keinen Ort.
