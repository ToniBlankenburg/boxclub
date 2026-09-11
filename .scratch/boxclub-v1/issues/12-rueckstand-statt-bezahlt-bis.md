Status: ready-for-human

# 12: Rückstand statt `bezahlt_bis`

**What to build:** Der Vereinsadmin kann ein Mitglied als **im Rückstand** markieren, wenn dessen Lastschrift zurückgekommen ist, eine Notiz zum Vorgang hinterlegen und den Rückstand wieder aufheben. Die Liste zeigt pro Zeile ein zweiwertiges Kennzeichen und lässt sich darauf filtern. Das Datum `bezahlt_bis` verschwindet ersatzlos.

**Blocked by:** None (kann sofort starten)

Begründung und Kostenseite: [ADR-0006](../../../docs/adr/0006-rueckstand-statt-bezahlt-bis.md). Dieses Ticket **ersetzt Ticket 05**.

## Acceptance Criteria

- [x] `bezahlt_bis` ist aus Schema, Service, Templates und Tests entfernt
- [x] Am Mitglied gibt es ein Rückstands-Kennzeichen und eine freie Notiz dazu
- [x] Der Rückstand lässt sich aus der Mitgliederliste heraus setzen und wieder aufheben
- [x] Beim Setzen kann eine Notiz erfasst werden ("Rücklastschrift Oktober, angeschrieben am 05.10."), die später änderbar ist
- [x] Die Listenzeile zeigt ein **zweiwertiges** Kennzeichen: grün für *in Ordnung*, rot für *im Rückstand*. Es gibt **keinen** dritten oder neutralen Zustand mehr
- [x] Filter über der Liste: alle / nur im Rückstand / nur in Ordnung
- [x] Ein Rückstand bleibt bestehen, wenn das Mitglied austritt — geprüft am Service-Seam
- [x] Die Suche trifft zusätzlich die **Mitglieds-ID**, damit die Nummer direkt eingegeben werden kann (Story 5; das Suchprädikat wird in diesem Ticket ohnehin angefasst)
- [x] Kein Rest von `bezahlt_bis`, "nicht bezahlt" oder "nicht gesetzt" in Code, Templates oder Tests
- [ ] `go test ./...` ist grün; `wails dev` und `wails build` laufen unter Windows und Linux — **offen:** Windows und `wails dev` sind nicht verifiziert (siehe Kommentar)

## Notes

Es gibt keinen Migrationsmechanismus — das Schema entsteht über `CREATE TABLE IF NOT EXISTS`. Die Entwicklungs-Datenbank (`BOXCLUB_DB`) muss vor dem ersten Start gelöscht werden. Genau deshalb passiert dieser Umbau jetzt und nicht nach dem ersten Mac-Release.

## Comments

### Umgesetzt (2026-09-11)

**Modell.** Ein eigener Typ `Rueckstand{Offen bool, Notiz string}` hängt am
`mitglied`, nicht an der `mitgliedschaft`. Damit erledigt sich "der Rückstand
bleibt beim Austritt bestehen" strukturell statt durch eine Regel — geprüft ist
es trotzdem, und zwar auch über einen Wiedereintritt hinweg.

Der Typ kam erst im Review dazu: vorher reisten `ImRueckstand` und
`RueckstandNotiz` als Paar durch `Mitglied`, `Listeneintrag`, den Setter und die
Templates, obwohl das Glossar genau **einen** Begriff dafür kennt. Nebenbei
verschwindet der nackte Boolean im Aufruf — aus `SetRueckstand(id, true, "…")`
wird `SetRueckstand(id, Rueckstand{Offen: true, Notiz: "…"})`. Die Beschriftung
("im Rückstand" / "in Ordnung") sitzt als `Bezeichnung()` am Typ, damit Liste und
spätere Ansichten nicht auseinanderlaufen.

Die Spalten heißen `rueckstand` und `rueckstand_notiz` — so, wie die
[Spec](../spec.md#schema) sie führt. Die erste Fassung hatte `im_rueckstand`;
das Review hat den Widerspruch gefunden.

`bezahlt_bis` ist ersatzlos weg: Spalte, `Zahlungsstatus`-Typ mit seinen drei
Werten, `zahlungsstatusAm`, `SetBezahltBis`, der `heute()`-Stichtag in
`eintraegeLesen` und die Template-Funktion `isodatum`, die danach niemand mehr
aufrief.

**Ein Setter für beides.** `SetRueckstand(id, imRueckstand, notiz)` schreibt
Kennzeichen und Notiz zusammen, weil der Nutzer im Formular beides zusammen vor
sich hat — Setzen, Aufheben und das Nachtragen der Notiz sind für ihn derselbe
Vorgang.

**Die Notiz überlebt das Aufheben.** Das war die eine offene Frage im Ticket.
ADR-0006 sagt, ein im April nachgezahlter März-Rückstand hinterlasse "außer der
Notiz keine Spur" — die Notiz *ist* also die Spur und darf beim Aufheben nicht
verschwinden. Wer sie loswerden will, leert sie ausdrücklich; umschließender
Leerraum wird dabei im Service abgeschnitten, nicht im Handler — normalisiert
wird, wo gespeichert wird.

Sie hing kurzzeitig als `title`-Tooltip am Kennzeichen. Das ist wieder raus: kein
AC verlangt es, und ein Tooltip ist per Tastatur nicht erreichbar, also keine
belastbare Anzeige. Die Notiz steht im Formular, wo sie auch bearbeitet wird.

**Mitglieds-ID trifft genau, nicht als Teilzeichenkette.** Alle übrigen
Suchfelder sind Teiltreffer; die ID ist die Ausnahme. Bei 200 Mitgliedern brächte
"7" sonst die 7, die 17, die 27 und die 70er zurück, und die Nummer wäre als
Sprungmarke gerade nicht mehr zu gebrauchen. Die ID hebelt dabei keinen Filter
aus: eine ausgetretene Person findet ihre Nummer nur unter "Auch Ehemalige".

**Liste hat jetzt vier Spalten statt fünf** — "Status" und "Bezahlt bis" fallen
zu einer Spalte "Rückstand" zusammen. Die `colspan` der Zeilenformulare zieht
entsprechend nach.

**Nachgezogene Dokumente.** `CLAUDE.md` (Schema und der Absatz zum
Zahlungsstatus), `spec.md` (die Suche hält jetzt fest, dass die Mitglieds-ID
**genau** trifft — das stand vorher nur im ADR), das Beispiel in
[ADR-0002](../../../docs/adr/0002-htmx-ueber-den-wails-assetserver.md), das noch
"Zahlungsstatus" als Fachregel nannte, und ein Nachtrag an
[ADR-0004](../../../docs/adr/0004-suche-und-filter-im-speicher.md): zwei seiner
drei Filter gibt es nicht mehr, und das Argument "der Zahlungsstatus ist ohnehin
abgeleitet und hat gar keine Spalte" trägt für den Rückstandsfilter **nicht** —
`im_rueckstand` ist eine echte Spalte. Die Entscheidung bleibt, aber nur noch
wegen der Umlaut-Faltung im Suchbegriff.

### Offen: Verifikation auf Windows

`go test ./...` ist grün und `wails build` läuft unter Linux durch. Nicht
verifiziert sind **Windows** (keine Maschine zur Hand) und **`wails dev`**, das
ein WebView-Fenster öffnet und sich hier nicht headless prüfen lässt.

Ersatzweise wurde die Adapter-Schicht mit einem Wegwerf-Test über
`httptest` end-to-end durchgespielt — Liste rendern, Rückstand setzen, aufheben,
beide Filterstufen, Suche über die Mitglieds-ID. Alles grün. Der Test ist danach
gelöscht worden, weil `CLAUDE.md` für `app/` und die Templates ausdrücklich
keinen Unit-Test vorsieht. Falls diese Abdeckung dauerhaft gewünscht ist, wäre
das eine bewusste Änderung der Konvention und ein eigenes Ticket wert.

### Entwicklungs-Datenbank löschen

Das Schema ändert sich, und es gibt keinen Migrationsmechanismus. Vor dem ersten
Start muss die Datei unter `BOXCLUB_DB` (bzw. im User-Config-Verzeichnis)
gelöscht werden — sonst startet die App gegen eine Tabelle ohne
`rueckstand`, und die Fehlermeldung zeigt nicht auf die Ursache.
