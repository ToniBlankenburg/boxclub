Status: ready-for-human

# 17: Kündigung erfassen und Mitglied ruhend schalten

**What to build:** Der Vereinsadmin erfasst eine **Kündigung** mit zwei Daten: wann gekündigt wurde und wann der Austritt wirksam wird. Dazwischen liegt die **Kündigungsfrist**, in der das Mitglied weiter trainiert und weiter zahlt. Unabhängig davon kann er ein Mitglied **ruhend** schalten, wenn es vorübergehend nicht trainiert (Verletzung, Ausland) — es bleibt Mitglied, aber für diese Zeit wird kein Beitrag eingezogen.

**Blocked by:** 13 (Beitrag individuell) — beide Felder hängen an der Mitgliedschaft, die dort umstrukturiert wird

## Acceptance Criteria

- [x] An der Mitgliedschaft: Kündigungsdatum und ein Ruhend-Kennzeichen
- [x] Eine Kündigung lässt sich mit Kündigungsdatum **und** Austrittsdatum erfassen
- [x] Das **Austrittsdatum wird nicht aus dem Kündigungsdatum berechnet**. Fristen haben Sonderfälle (Kulanz, Aufhebungsvertrag, Quartalsende); ein errechnetes Datum, das man überschreiben muss, ist lästiger als ein leeres Feld — **für das Speichern weiter gültig, für die Formular-Vorbelegung später umgedreht, siehe Nachtrag unten**
- [x] Ein Austrittsdatum **darf in der Zukunft liegen** — das ist der Normalfall bei laufender Kündigungsfrist
- [x] Ein Kündigungsdatum ohne Austrittsdatum ist erlaubt (Kündigung liegt vor, Termin noch offen)
- [x] Ein Mitglied lässt sich ruhend schalten und wieder aktiv setzen
- [x] Ruhend und Rückstand sind **unabhängig**: für ein ruhendes Mitglied wird keine Lastschrift losgeschickt, es kann also kein *neuer* Rückstand entstehen — ein bestehender bleibt aber stehen
- [x] Tests am Service-Seam: Kündigung mit und ohne Austrittstermin, Austritt in der Zukunft, ruhend setzen und zurücknehmen, ruhend mit bestehendem Rückstand
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux — **unter Linux erledigt, Windows steht aus** (von hier aus nicht ausführbar; siehe Rauchtest unten)

## Notes

Dieses Ticket legt nur die **Felder und ihre Pflege** an. Dass daraus der Status *In Kündigungsfrist* abgelesen wird und sich die Bedeutung von "aktiv" ändert, ist Ticket 18 — sonst wäre die Abhängigkeit zirkulär.

Entwicklungs-Datenbank vor dem ersten Start löschen.

## Comments

**Umgesetzt.** Seam: `MemberService.SetKuendigung(id, service.Kuendigung{Datum, Austritt})`
und `MemberService.SetRuhend(id, ruhend)`. An der Mitgliedschaft stehen dazu
`kuendigungsdatum TEXT` und `ruhend INTEGER NOT NULL DEFAULT 0`; gelesen werden
beide über `Mitgliedschaft` und — für die Liste — über `Listeneintrag`.

**`MarkExit` ist ersetzt, nicht ergänzt.** Eine Kündigung zu erfassen und einen
Austritt einzutragen sind für den Nutzer ein Vorgang, und beide schreiben
dieselbe Spalte. Zwei Methoden nebeneinander hätten zwei Validierungen für
dasselbe Feld bedeutet. `service.Kuendigung` ist dabei ein reiner Eingabewert wie
`NeuesMitglied` — an der Mitgliedschaft stehen die beiden Daten einzeln, weil der
Austritt dort den Zeitraum begrenzt und nicht nur einen Vorgang festhält.

**Entscheidungen, die das Ticket offen ließ:**

1. **Beide Daten zugleich leer ist ein `ValidierungsFehler`.** Das schützt eine
   bereits erfasste Kündigung vor einem versehentlich leer abgeschickten
   Formular. Eine Kündigung *zurückzunehmen* ist damit nicht möglich — das wäre
   ein eigener Vorgang und steht in keinem Akzeptanzkriterium.
2. **`SetKuendigung` schreibt beide Spalten, auch die nicht gesetzte** (wie die
   Anmeldung im Patch). Ein erfasstes Austrittsdatum kann dabei nicht verloren
   gehen: sobald eines steht, läuft der Zeitraum nicht mehr und der nächste
   Aufruf endet in `ErrNichtAktiv`. Mit Ticket 18 ändert sich, was "läuft" heißt
   — dann wird aus dieser Sperre eine Korrekturmöglichkeit während der Frist.
3. **Ein Kündigungsdatum *vor* dem Eintritt bleibt erlaubt.** Geprüft wird nur
   Austritt ≥ Eintritt und Kündigungsdatum ≤ Austritt. Wer zurücktritt, bevor
   seine Mitgliedschaft beginnt, ist kein Tippfehler.
4. **Ruhend hängt an der laufenden Mitgliedschaft** (Spec → Datenmodell): auf
   einem Ehemaligen ist `SetRuhend` ein `ErrNichtAktiv`. Ein Wiedereintritt
   beginnt weder gekündigt noch ruhend — die Spalten stehen gar nicht erst in der
   `INSERT`-Anweisung.

**Oberfläche:** Aus der Schaltfläche *Austritt* wird *Kündigung* mit zwei
Datumsfeldern; das Austrittsfeld ist bewusst **leer** vorbelegt. Ohne
Austrittsdatum kommt die Zeile mit dem Vermerk „gekündigt «Datum»" zurück, mit
Austrittsdatum die ganze Liste. Daneben steht ein Umschalter *Ruhend* /
*Aktiv setzen*, nur bei aktiven Mitgliedern; das Kennzeichen erscheint als eigenes
Etikett neben dem Namen, nicht anstelle des Austritts-Etiketts.

**Noch nicht hier:** Dass aus beiden Feldern der Status *In Kündigungsfrist*
abgelesen wird und ein Mitglied mit künftigem Austritt in der Standardansicht
bleibt, ist Ticket 18. Bis dahin verschwindet es mit dem Eintragen des Austritts
sofort unter »Auch Ehemalige« — das ist die bekannte Zwischenstufe, kein Fehler.

**Rauchtest:** Liste → Kündigung ohne Termin → Termin nachtragen → Fehlerfall
(Kündigung nach Austritt) → Ruhend hin und zurück → Ruhend auf Ausgetretenem →
Wiedereintritt lief gegen `app.Handler()` per `httptest` durch. Der Harnisch ist
danach gelöscht (siehe [CLAUDE.md](../../../CLAUDE.md)).

**Was geprüft ist:** `go test ./...`, `go vet`, `gofmt` und `wails build` unter
Linux; dazu kompiliert `GOOS=windows` und `GOOS=darwin` sauber durch. **Was
offen ist:** ein echter `wails build` bzw. `wails dev` unter Windows — der
braucht eine Windows-Maschine und ist von hier aus nicht zu machen. Das eine
offene Häkchen oben bleibt deshalb offen, bis das jemand nachholt.

**Entwicklungs-Datenbank vor dem ersten Start löschen** — die zwei neuen Spalten
kommen ohne Migration.

**Aus dem Review nachgezogen:** Schema-Absatz in [CLAUDE.md](../../../CLAUDE.md)
um `kuendigungsdatum`/`ruhend` ergänzt; „pausieren" aus allen Kommentaren
entfernt (CONTEXT.md → Ruhend führt es unter _Vermeiden_); das vollständige
Ersetzen beider Spalten ist jetzt im Kommentar von `SetKuendigung` ausgesprochen
und mit `TestSetKuendigung_ErsetztBeideAngabenVollstaendig` festgenagelt — es ist
der Weg, eine irrtümlich erfasste Kündigungserklärung wieder wegzunehmen.

---

**Nachtrag: das Austrittsfeld ist mit der regulären Frist vorbelegt.** Auf Wunsch
des Vereinsadmins schlägt das Kündigungsformular jetzt den Termin vor, statt das
Feld leer zu lassen. Die Regel steht als `service.RegulaererAustritt` in
[`service/kuendigung.go`](../../../service/kuendigung.go): drei Monate ab
Kündigungstag, aufgerundet auf das **Monatsende** (`KuendigungsfristMonate`).
Eine am 12.09. erklärte Kündigung wirkt damit zum 31.12.

Gerechnet wird über das Monatsende und nicht taggenau — sonst wäre der
30. November plus drei Monate der 30. Februar, den es nicht gibt, und
`time.AddDate` schöbe ihn stillschweigend in den März. Die Randfälle
(Monatsletzter, Jahreswechsel, Februar, Schaltjahr) stehen in
`TestRegulaererAustritt_DreiMonateZumMonatsende`.

**Was sich dadurch nicht ändert** — und was die obige Abnahme weiterhin trägt:

- `SetKuendigung` leitet nach wie vor **nichts** ab. Gespeichert wird genau, was
  abgeschickt wird; geprüft wird der Austritt nur gegen den Eintritt und nicht
  gegen die Satzungsfrist. Eine kürzere Frist (Aufhebungsvertrag, Kulanz) bleibt
  erfassbar — festgenagelt in `TestRegulaererAustritt_BindetSetKuendigungNicht`.
- Wer das Feld **leert**, erfasst weiterhin eine Kündigung ohne Termin. Der
  Vorschlag nimmt das Rechnen ab, nicht die Entscheidung.
- Ein **bereits vereinbarter** Termin schlägt den Vorschlag. Das ist seit der
  Kündigungsfrist (Ticket 18) wichtig: dort führt die Schaltfläche der Zeile
  zurück in dieses Formular, und `SetKuendigung` schreibt beide Spalten zusammen
  — ein frisch gerechneter Wert überschriebe sonst den vereinbarten Austritt.

**Bewusst nicht gebaut:** der Austritt rechnet **nicht** nach, wenn im Formular
das Kündigungsdatum geändert wird. Vorbelegt wird einmal beim Öffnen, aus dem
Datum, das dann oben steht (bei einer bereits erfassten Kündigung also aus deren
Tag, nicht aus heute). Ein Nachrechnen bräuchte einen eigenen htmx-Endpunkt und
überschriebe dabei auch einen Termin, den jemand gerade von Hand eingetippt hat.

Nachgezogen ist die Doku an allen Stellen, die das Gegenteil zusagten:
[CLAUDE.md](../../../CLAUDE.md) (Kündigung-Absatz), [CONTEXT.md](../../../CONTEXT.md)
→ Kündigungsdatum und → Kündigungsfrist, der Kommentar an `Kuendigung.Austritt`,
`app.kuendigungsformular` und der Hinweistext unter dem Formular.

**Rauchtest** gegen `app.Handler()` per `httptest`: ohne Kündigung (heute →
31.12.), mit erfasster Kündigung ohne Termin (05.01.2026 → 30.04.2026, also aus
dem erfassten Tag und nicht aus heute) und mit vereinbartem Termin (bleibt
stehen). Der Harnisch ist danach gelöscht.
