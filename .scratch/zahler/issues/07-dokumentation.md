# 07: Dokumentation

**What to build:** `CONTEXT.md` bekommt die Begriffe *Zahler* (technisch, in der UI nicht sichtbar), *Kind* (aus dem Geburtsdatum abgeleitet; Handkennzeichen nur ohne Geburtsdatum) und *Kontoinhaber* (das UI-Wort), mit „nicht zu verwechseln mit“ gegenüber dem Mitglied. Eine neue ADR „Zahler als technische Entität“ hält fest: Entität statt Felder an der Person; Zusammenführen über die normalisierte IBAN; IBAN-Normalisierung weicht an dieser einen Stelle von ADR-0006 ab; `mitglied.iban` entfällt; Verweise auf ADR-0006, ADR-0013, ADR-0019. Der Schemaabschnitt in `CLAUDE.md` wird nachgezogen.

**Blocked by:** 06 (Hinweis bei Volljährigkeit)

**Status:** ready-for-agent

- [ ] `CONTEXT.md`: Einträge Zahler, Kind, Kontoinhaber mit „Vermeiden“ und Verweisen
- [ ] Neue ADR in `docs/adr/` im Format der bestehenden (Status, Datum, Kontext, Entscheidung, Konsequenzen, Alternativen)
- [ ] ADR-0006 bekommt im Statuskopf den Hinweis auf die IBAN-Normalisierung
- [ ] `CLAUDE.md`: Schema und Beschreibung von IBAN/Rückstand stimmen mit dem Stand überein
