Status: ready-for-agent

# 09: Excel-Import

**What to build:** Der Vereinsadmin lädt seine bestehende `.xlsx`-Tabelle in die App und bekommt einen Bericht, welche Zeilen übernommen wurden und welche mit welchem Grund nicht. Der Import ist **wiederholbar**: ein zweiter Lauf aktualisiert bestehende Mitglieder statt sie zu duplizieren, sodass der Admin Fehler in Excel korrigieren und einfach nochmal importieren kann. Die **Mitglieds-IDs aus der Excel bleiben erhalten**, damit die Nummern, unter denen er seine Mitglieder kennt, dieselben bleiben und beide Systeme während der Umstellung vergleichbar sind.

**Blocked by:** 12, 13, 14, 15, 16, 17, 18 — jede der 21 Spalten braucht erst ihren Ort im Modell

Das vollständige Spalten-Mapping steht in der [Spec](../spec.md#excel-import), die Bestandsaufnahme der echten Tabelle in [excel-vorlage.md](../excel-vorlage.md).

## Acceptance Criteria

**Einlesen**

- [ ] Eine Import-Aktion öffnet einen Datei-Dialog für `.xlsx` (nicht `.xls`)
- [ ] Gelesen wird ausschließlich das Blatt **„Verwaltung"**; das Blatt „Quelle" (Dropdown-Wertelisten) wird ignoriert
- [ ] Kopfzeile ist Zeile 1, Daten ab Zeile 2. Die Spalten werden **über ihre Überschrift** zugeordnet, nicht über die Position — inklusive der Tippfehler-Überschrift `Eintrit`
- [ ] Fehlt eine erwartete Spaltenüberschrift, bricht der Import mit einer klaren Meldung ab, statt Zeilen halb zu übernehmen

**Datumswerte**

- [ ] Datumsfelder werden sowohl als **Excel-Seriennummer** (Epoche 1899-12-30) als auch als **getippter Text** gelesen; beide Formen ergeben dasselbe Datum
- [ ] Ein Wert, der keins von beidem ist, macht die Zeile zum Fehlerfall

**Lebenszyklus aus `Status` und `Gekündigt`**

- [ ] `Status` wird **nicht als Feld gespeichert**, steuert aber die Zuordnung des Datums aus `Gekündigt`:
  - *Gekündigt* → das Datum wird der **Austritt**
  - *Kündigungsfrist* → das Datum wird das **Kündigungsdatum**, der Austritt bleibt leer
  - *Stillgelegt* / *Inakiv* → die Mitgliedschaft wird **ruhend**
  - *Neu* / *Aktiv* / *Mitglied* → keine Wirkung, diese Zustände werden ohnehin abgelesen
- [ ] Ein unbekannter `Status`-Wert macht die Zeile zum Fehlerfall, statt stillschweigend als "aktiv" durchzulaufen

**Trainingsslots und Frequenz**

- [ ] `Training - 1/2/3` ergeben je eine Slot-Zeile; eine leere Zelle ergibt **keine** Zeile
- [ ] `1x 2x Woche` wird **nicht importiert** — die Frequenz ergibt sich aus der Anzahl der Slots
- [ ] Weicht der Wert in `1x 2x Woche` von der Anzahl gefüllter Training-Spalten ab, ist das ein **Eintrag im Fehlerbericht und keine stille Korrektur**

**Mitglieds-IDs**

- [ ] Der Wert aus `Mandatsreferenz` wird die **Mitglieds-ID** — trotz des Spaltennamens ist das keine SEPA-Mandatsreferenz
- [ ] Nach dem Import zählt die automatische Vergabe **oberhalb der höchsten importierten Nummer** weiter
- [ ] Eine doppelte oder leere Nummer ist ein Fehlerfall im Bericht

**Beträge und restliche Felder**

- [ ] `Beitrag` und `Anmeldegebühr` werden von Euro in Cent umgerechnet; **0 € ist gültig**
- [ ] `Digital` wird **wortwörtlich als Freitext** übernommen, ohne Interpretation
- [ ] `Bewertung` wird auf die zweiwertige Google-Bewertung abgebildet
- [ ] Der **Rückstand wird nicht importiert** — es gibt keine Quellspalte. Jedes importierte Mitglied startet als *in Ordnung*, was bei Lastschrift der zutreffende Normalfall ist

**Bericht**

- [ ] Nach dem Import zeigt ein Bericht: Anzahl übernommener Zeilen, Anzahl neu angelegter, Anzahl aktualisierter, Anzahl gescheiterter
- [ ] Jede gescheiterte Zeile erscheint **einzeln mit Zeilennummer und Grund**, z. B. "fehlender Nachname", "Datum nicht lesbar", "Mitglieds-ID 47 doppelt", "Frequenz 1× widerspricht 3 Trainingsterminen", "unbekannter Status `Halbtot`"
- [ ] **Kein stilles Ausweichen auf Standardwerte.** Was nicht zugeordnet werden kann, steht im Bericht

**Wiederholbarkeit**

- [ ] Ein zweiter Import derselben Datei erzeugt **keine Duplikate**
- [ ] Geänderte Werte werden übernommen; der Match läuft über die Mitglieds-ID
- [ ] Für Zeilen ohne Nummer greift der Rückfall auf Vorname + Nachname + Geburtsdatum
- [ ] Eine Zeile aktualisiert die **laufende** Mitgliedschaft; existiert keine, wird eine angelegt

**Seam-Disziplin und Tests**

- [ ] Der `ExcelImporter` liefert geparste Zeilen plus Fehlerbericht und **ruft den `MemberService` nicht selbst auf**; ein dünner Orchestrator im Wails-Layer setzt die Zeilen ein
- [ ] Die Test-Fixture ist eine **anonymisierte** `.xlsx` — die echte Tabelle des Vereins enthält Klarnamen, IBANs und Kontaktdaten und darf **nicht** ins Repository
- [ ] Tests am Importer-Seam: Seriendatum und Textdatum ergeben dasselbe Datum; fehlender Nachname, doppelte ID, widersprüchliche Frequenz und unbekannter Status landen im Bericht; alle drei Training-Spalten werden zu Slots; zweiter Import erzeugt keine Duplikate und übernimmt Änderungen
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

**Vollständig neu geschrieben.** Die erste Fassung dieses Tickets nannte "unbekannte Beitragsklasse" als Fehlerbeispiel und ging von einem offenen Spalten-Mapping aus. Beides ist überholt: Beitragsklassen gibt es nicht mehr ([ADR-0005](../../../docs/adr/0005-beitrag-individuell-statt-beitragsklasse.md)), und die Spaltenliste ist vom Nutzer als echt und vollständig bestätigt.

**Offener Detailpunkt:** In der Musterzeile steht bei `Bewertung` ein ❌. Welche Werte für "hat bewertet" vorkommen, ist unbekannt. Beim Umsetzen die tatsächlich vorkommenden Werte prüfen und unbekannte in den Fehlerbericht geben, statt sie auf "nein" zu raten.
