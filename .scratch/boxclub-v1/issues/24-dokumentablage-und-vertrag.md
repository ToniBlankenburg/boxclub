Status: ready-for-human

# 24: Dokumente ablegen und den Vertrag am Zeitraum hinterlegen

**What to build:** PDFs am Mitglied — als **Blob in der Datenbank**, nicht als Dateien daneben. Dieses Ticket baut die Ablage und ihren ersten Anwendungsfall: den eingescannten **Vertrag**, der an der **Mitgliedschaft** hängt.

**Blocked by:** —
**Siehe:** [ADR-0007](../../../docs/adr/0007-dokumente-als-blob-in-sqlite.md), `CONTEXT.md` → Dokument, Vertrag

## Acceptance Criteria

- [x] Tabelle `dokument` mit Art, Dateiname, Inhalt (Blob), Zeitpunkt und **zwei** Fremdschlüsseln — `mitglied_id` und `mitgliedschaft_id` —, von denen immer **genau einer** gesetzt ist. Ein Vertrag hängt an der Mitgliedschaft, eine Rechnung (Ticket 25) am Mitglied
- [x] Hochladen eines PDFs beim Mitglied, zugeordnet zur **Mitgliedschaft**, die man gerade ansieht
- [x] Nur PDF wird angenommen; alles andere wird mit verständlicher Meldung abgewiesen
- [x] Eine Obergrenze für die Dateigröße (Vorschlag: 10 MB) mit eigener Meldung — ein versehentlich hochgeladenes 400-MB-Video darf die Datenbank nicht aufblähen
- [x] **Ersetzen** und **Entfernen** eines Dokuments, ohne Rückfrage-Zeremonie, ohne Papierkorb, ohne Versionen
- [x] **Export**: das PDF lässt sich an einen frei gewählten Ort speichern. Ohne diesen Weg kommt niemand mehr an seinen Vertrag
- [x] Bei mehreren Mitgliedschaften steht jeder Vertrag **bei seinem Zeitraum**, nicht in einer gemeinsamen Liste
- [x] `inhalt` erscheint in **keiner** Abfrage, die mehr als eine Zeile liefert — insbesondere nicht in `List`, `Search` oder `eintraegeLesen`. Ein Test hält das fest
- [x] Service-Tests: ablegen, lesen, ersetzen, entfernen, Größengrenze, Zuordnung zur richtigen Mitgliedschaft
- [x] `go test ./...` grün; `wails build` unter Linux geprüft — Windows und `wails dev`
      stehen aus (kein Windows-Rechner, kein Display in dieser Umgebung),
      `GOOS=windows go build ./...` als Ersatz grün

## Notes

**Warum Blob und nicht Dateien:** ADR-0007. Kurzfassung: der Vereinsadmin sichert, indem er eine Datei kopiert, und was danebenliegt, vergisst er.

**Der Blob ist die einzige Stelle im Schema, die teuer werden kann.** Ein `SELECT *` über die Mitgliederliste, das die Dokumente mitzieht, macht die Liste unbenutzbar, und zwar erst dann, wenn Daten drin sind — also nicht in der Entwicklung. Deshalb das ausdrückliche Kriterium oben.

**Mitglieder werden nicht gelöscht** (es gibt keinen Löschpfad), also gibt es auch keine verwaisten Dokumente. Sollte je einer gebaut werden, müssen die Dokumente mit.

## Comments

**2026-09-13 — umgesetzt**

Alle Akzeptanzkriterien erfüllt. `go test ./...` grün, `wails build` unter Linux
erfolgreich, `GOOS=windows go build ./...` ebenfalls; ein Windows-`wails build`
und `wails dev` stehen aus.

`service/dokument.go` liegt wie die Vereinsdaten am `MemberService`: es ist der
eine Service dieser App (spec.md → Seams).

Fünf Entscheidungen, die über den Ticket-Text hinausgehen:

- **Ablegen *ist* Ersetzen.** Ein Zeitraum hat einen Vertrag, kein Bündel — ein
  eindeutiger Teilindex auf `mitgliedschaft_id` hält das fest, und
  `VertragAblegen` löscht und setzt in einer Transaktion neu. Ein eigener
  Aufruf fürs Ersetzen wäre ein zweites Wort für denselben Vorgang.
- **PDF wird an der Dateikennung erkannt, nicht an der Endung.** Eine umbenannte
  Word-Datei hieße sonst `.pdf` und läge unlesbar in der Datenbank.
- **Der Export geht über den Datei-Dialog von Wails, nicht über einen
  Download.** Wails v2.15 behandelt auf keiner Plattform Downloads (nichts in
  `internal/frontend/desktop/` fasst das an) — ein `Content-Disposition` liefe
  im WebView ins Leere, und das Kriterium „ohne diesen Weg kommt niemand mehr an
  seinen Vertrag" wäre auf dem Produktionsziel macOS genau nicht erfüllt.
  ADR-0002 bleibt gewahrt: der Dialog ist keine Wails-Bindung, sondern ein
  Funktionstyp `app.Speicherziel`, den `main.go` in `OnStartup` einsetzt und ein
  Test durch ein Verzeichnis ersetzt.
- **`Mitgliedschaft.Vertrag` trägt kein Blob.** `Dokument` hat kein Inhaltsfeld;
  den Blob liefert nur `VertragInhalt` über genau eine ID. Damit kann eine Liste
  von Zeiträumen ihn nicht mitschleppen, auch nicht aus Versehen.
- **`ArtRechnung` steht schon da, obwohl Ticket 25 sie erst baut.** CONTEXT.md
  legt die Werteliste auf genau zwei fest; ein Typ mit nur einem der beiden
  Werte lädt dazu ein, eine dritte Zeichenkette zu erfinden.

**Nach dem Review** (`/code-review`, Achsen Standards und Spec) noch geändert:

- **Der Test zur Blob-Regel hielt nicht, was er versprach.** Der Spec-Review hat
  recht, und die Gegenprobe bestätigt es: er las Literal für Literal, also kam
  `SELECT ` + `spaltenKonstante` + ` FROM dokument` durch — ausgerechnet das
  Muster, das derselbe Diff mit `mitgliedschaftsspalten` einführt —, und
  `SELECT *`, den ADR-0007 ausdrücklich nennt, ebenso. Er setzt jetzt die
  Anweisung über die Konstanten der Datei wieder zusammen, und ein zweiter Test
  verbietet `SELECT *` im ganzen Package. Beide Attrappen fallen nun durch.
- **Ein Test sicherte nichts zu.** Der Abbruch-Test legte ein Verzeichnis an,
  nannte es dem Dialog aber nie und prüfte dann, dass es leer ist. Er prüft
  jetzt, was tatsächlich schiefgehen kann: dass der Dialog gefragt wurde, keine
  Erfolgsmeldung kommt und der Block unverändert dasteht.
- **Die Meldung log.** „Diesen Vertrag gibt es nicht mehr" stand auch dann da,
  wenn der *Zeitraum* fehlte. Dafür gibt es jetzt `service.ErrKeinVertrag` — ein
  `ErrNichtGefunden`, das zusätzlich sagt, was fehlt.
- **Die Grenze stand in drei Schreibweisen da** („10,0 MB", „10 MB", „höchstens
  10 MB"). Sie kommt jetzt aus `service.Dokumentgrenze()`; `megabyte` lässt das
  Nachkomma weg, wo es nichts sagt.
- **`AND art = ?` an allen vier Abfragen.** Heute überflüssig — an einer
  Mitgliedschaft hängt nur der Vertrag —, aber eine Funktion namens „Vertrag"
  soll den Vertrag löschen und liefern und nicht, was immer dort hängt.

**Nicht geändert, bewusst:** Das CHECK „genau einer der beiden Fremdschlüssel"
hat keinen Test, weil es keinen Schreibweg auf `mitglied_id` gibt, bis Ticket 25
die Rechnung baut — und CLAUDE.md verbietet, dafür an SQL vorbei zu testen
(„Test via the service/importer API only"). Der Test gehört zu Ticket 25.
