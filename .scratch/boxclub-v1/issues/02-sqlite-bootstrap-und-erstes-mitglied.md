Status: ready-for-agent

# 02: SQLite-Bootstrap & erstes Mitglied anlegen

**What to build:** Die App bindet SQLite an und legt beim ersten Start eine Datenbankdatei mit Schema und Seed-Daten (die zwei **Beitragsklassen**) an. Der Nutzer kann in einem Formular die Stammdaten eines **Mitglied**s eingeben, speichern und den gespeicherten Datensatz sofort sehen. Erster echter Vertikal-Cut durch alle Schichten.

**Blocked by:** 01 (Wails-Projekt-Skelett)

## Acceptance Criteria

- [ ] Beim Start ohne bestehende DB wird `boxclub.db` angelegt und mit Schema für `mitglied`, `mitgliedschaft`, `beitragsklasse` initialisiert
- [ ] Die zwei Beitragsklassen `Erwachsen 1×/Woche` (60 €/Monat) und `Erwachsen 2×/Woche` (80 €/Monat) sind geseedet, Preise in Cent gespeichert
- [ ] Formular erlaubt: Vorname, Nachname, Geburtsdatum, Adresse, E-Mail, Telefon, Eintrittsdatum, Beitragsklasse (Auswahl)
- [ ] `MemberService.Create` legt gleichzeitig einen `mitglied`- und einen `mitgliedschaft`-Datensatz (`austritt` NULL) an
- [ ] Nach dem Speichern wird der angelegte Datensatz im UI angezeigt (`MemberService.Get`)
- [ ] Beim Neustart der App bleibt das Mitglied erhalten
- [ ] Tests am `MemberService`-Seam mit echtem SQLite in `t.TempDir()`: Create → Get liefert alle Werte zurück; Fremdschlüssel auf `beitragsklasse` funktioniert; Doppelanlage mit demselben Vor-/Nachnamen + Geburtsdatum bleibt möglich (kein UNIQUE-Constraint darauf)
- [ ] `MemberService.Create` etabliert als Referenz-Testdatei das Setup-Muster für alle späteren Service-Tests

## Notes

Treiber: **`modernc.org/sqlite`** (pure-Go). Kein Repository-Interface (siehe spec.md → "Bewusst nicht modelliert"). Vokabular: [CONTEXT.md](../../../CONTEXT.md).
