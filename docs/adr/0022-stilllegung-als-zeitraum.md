# ADR-0022: Stilllegung als Zeitraum statt Kennzeichen

**Status:** Accepted — präzisiert [ADR-0006](0006-rueckstand-statt-bezahlt-bis.md) und [ADR-0009](0009-rechnungen-und-monatssoll-ohne-zahlungsmodell.md), löst das Kennzeichen `mitgliedschaft.ruhend` ab.
**Datum:** 2026-10-20

## Kontext

*Ruhend* war ein Boolean an der Mitgliedschaft, per Knopf umgeschaltet. Es kannte kein Ende — man vergaß, wieder aktiv zu setzen, und der Beitrag blieb ausgesetzt —, keinen Grund und keine Historie, und es wirkte nicht auf die Laufzeit: wer ein halbes Jahr stillgelegt war, hatte trotzdem die volle Mindestlaufzeit abgesessen.

## Entscheidung

Eine Tabelle `stilllegung(id, mitgliedschaft_id, von, bis, grund NULL)`; `mitgliedschaft.ruhend` entfällt. Der Begriff heißt jetzt durchgängig *Stilllegung* / *stillgelegt* (Oberfläche, Code, Schema, Glossar); *Ruhend* wird vermieden.

- **Abgelesen statt gespeichert.** Stillgelegt ist, wer heute in einer Stilllegung liegt. Damit entfällt die letzte gespeicherte Ausnahme neben dem abgelesenen Status.
- **`von` und `bis` sind Pflicht.** Kein „bis auf Weiteres": ein offenes Ende wäre dasselbe Vergessen wie beim Boolean, und die Verlängerung der Laufzeit braucht eine Dauer.
- **Keine Überlappung; keine neue, solange eine läuft.** „Wieder aktiv" setzt `bis` auf gestern (beide Ränder zählen, mit heute bliebe das Mitglied bis Mitternacht stillgelegt); eine heute oder später beginnende wird gelöscht. Das macht die Zählung der Stilllegungsmonate eindeutig.
- **Ein Monat zählt voll, sobald ein Tag darin stillgelegt ist.** Monatssoll und MoneyMoney-Export lassen ihn aus; es gibt keine anteilige Berechnung, passend zu „individuell vereinbarter Beitrag" (ADR-0005).
- **Mindestlaufzeit** (`mitgliedschaft.mindestlaufzeit_monate`, optional) ist neu. Ihr Ende ist abgeleitet (Eintritt + Monate + Stilllegungsmonate) und warnt nur; `austritt` bleibt, was eingetragen wird.
- **Kündigung in einer Stilllegung gesperrt**, mit Hinweis am Knopf.
- **Excel-Import legt nichts an.** Die Datei hat keine Daten; ein erfundenes `bis` würde den Beitrag still aussetzen.
- **Keine Migration.** Der Schnitt macht aus einem Boolean Zeiträume (ADR-0019 deckt nur additive Spalten); die Entwicklungsdatenbank wird gelöscht.

## Erwogen und verworfen

- **Boolean mit `ruhend_von`/`ruhend_bis`.** Einfacher, hält aber nur einen Zeitraum je Mitgliedschaft und damit keine Historie.
- **Manueller Schalter plus Protokoll.** Beides könnte sich widersprechen.

## Konsequenzen

Das ist **keine Zahlungshistorie**: ADR-0006 und ADR-0009 schließen weiter Zahlungen, offene Posten und einen Beitragsverlauf aus. Die Stilllegungen sind Stammdaten der Mitgliedschaft, nicht Buchungen. Auch die Anwesenheit bleibt ausgeschlossen; eine Stilllegung sagt nur, dass jemand nicht trainiert, nie wann er da war.

Die Teilnehmerliste liest *heute stillgelegt*, nicht den Stilllegungsmonat: wer ab dem 28. stillgelegt ist, trainiert bis dahin noch.
