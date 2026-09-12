Status: ready-for-agent

# 24: Dokumente ablegen und den Vertrag am Zeitraum hinterlegen

**What to build:** PDFs am Mitglied — als **Blob in der Datenbank**, nicht als Dateien daneben. Dieses Ticket baut die Ablage und ihren ersten Anwendungsfall: den eingescannten **Vertrag**, der an der **Mitgliedschaft** hängt.

**Blocked by:** —
**Siehe:** [ADR-0007](../../../docs/adr/0007-dokumente-als-blob-in-sqlite.md), `CONTEXT.md` → Dokument, Vertrag

## Acceptance Criteria

- [ ] Tabelle `dokument` mit Art, Dateiname, Inhalt (Blob), Zeitpunkt und **zwei** Fremdschlüsseln — `mitglied_id` und `mitgliedschaft_id` —, von denen immer **genau einer** gesetzt ist. Ein Vertrag hängt an der Mitgliedschaft, eine Rechnung (Ticket 25) am Mitglied
- [ ] Hochladen eines PDFs beim Mitglied, zugeordnet zur **Mitgliedschaft**, die man gerade ansieht
- [ ] Nur PDF wird angenommen; alles andere wird mit verständlicher Meldung abgewiesen
- [ ] Eine Obergrenze für die Dateigröße (Vorschlag: 10 MB) mit eigener Meldung — ein versehentlich hochgeladenes 400-MB-Video darf die Datenbank nicht aufblähen
- [ ] **Ersetzen** und **Entfernen** eines Dokuments, ohne Rückfrage-Zeremonie, ohne Papierkorb, ohne Versionen
- [ ] **Export**: das PDF lässt sich an einen frei gewählten Ort speichern. Ohne diesen Weg kommt niemand mehr an seinen Vertrag
- [ ] Bei mehreren Mitgliedschaften steht jeder Vertrag **bei seinem Zeitraum**, nicht in einer gemeinsamen Liste
- [ ] `inhalt` erscheint in **keiner** Abfrage, die mehr als eine Zeile liefert — insbesondere nicht in `List`, `Search` oder `eintraegeLesen`. Ein Test hält das fest
- [ ] Service-Tests: ablegen, lesen, ersetzen, entfernen, Größengrenze, Zuordnung zur richtigen Mitgliedschaft
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

**Warum Blob und nicht Dateien:** ADR-0007. Kurzfassung: der Vereinsadmin sichert, indem er eine Datei kopiert, und was danebenliegt, vergisst er.

**Der Blob ist die einzige Stelle im Schema, die teuer werden kann.** Ein `SELECT *` über die Mitgliederliste, das die Dokumente mitzieht, macht die Liste unbenutzbar, und zwar erst dann, wenn Daten drin sind — also nicht in der Entwicklung. Deshalb das ausdrückliche Kriterium oben.

**Mitglieder werden nicht gelöscht** (es gibt keinen Löschpfad), also gibt es auch keine verwaisten Dokumente. Sollte je einer gebaut werden, müssen die Dokumente mit.
