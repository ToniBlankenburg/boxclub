Status: ready-for-agent

# 06: Suche & Filter

**What to build:** Über der Mitgliederliste liegen ein Suchfeld und Filter-Steuerungen. Suche trifft auf Name, E-Mail oder Telefon. Filter grenzen nach Zahlungsstatus, Beitragsklasse und Aktivität ein. Standardansicht zeigt nur aktive Mitglieder.

**Blocked by:** 03 (Mitgliederliste), 05 (Zahlungsstatus — für den Status-Filter)

## Acceptance Criteria

- [ ] Suchfeld findet Mitglieder mit Teiltreffer (case-insensitive) in Vorname, Nachname, E-Mail oder Telefonnummer
- [ ] Filter **Zahlungsstatus**: alle / bezahlt / nicht bezahlt
- [ ] Filter **Beitragsklasse**: alle / einzelne Klasse (dynamisch aus der DB gelistet)
- [ ] Filter **Aktivität**: nur aktive (Standard beim Öffnen) / auch ehemalige
- [ ] Suche und Filter kombinieren sich serverseitig in `MemberService.Search(query, filter)` — kein Client-Side-Filtern
- [ ] Reset-Aktion setzt Suchfeld und alle Filter auf Standardwerte zurück
- [ ] Ergebnisliste aktualisiert sich per htmx-Fragment bei jeder Eingabe oder Filter-Änderung
- [ ] Test am Seam für jede einzelne Filterdimension isoliert
- [ ] Test am Seam für kombinierte Filter (z. B. "aktive Erwachsen 2×/Woche mit unbezahltem Status")
- [ ] Test am Seam: leerer Query mit Filtern angewendet ≙ Filter-only
- [ ] Test am Seam: Query mit Sonderzeichen (Umlaute, Bindestriche) findet die passenden Mitglieder
