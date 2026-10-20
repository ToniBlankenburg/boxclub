# 01: Umbenennung ruhend → stillgelegt

**What to build:** Der Begriff *ruhend* heißt überall *stillgelegt*, ohne Verhaltensänderung: Schema (`mitgliedschaft.ruhend` → `stillgelegt`), Service (`SetRuhend`, Filter, Status), Handler, Templates, i18n-Schlüssel und -Texte (de/en), Tests und Kommentare. Excel-Eingabewerte `Stillgelegt` und `Inakiv` bleiben. Alte ADRs bleiben unverändert (Geschichte).

**Blocked by:** None (can start immediately)

**Status:** done

- [ ] Kein `ruhend`/`Ruhend` mehr in Code, Templates und i18n, außer in alten ADRs und im Glossar unter „Vermeiden"
- [ ] Oberfläche zeigt „stillgelegt"/„Stilllegung" in Deutsch, englische Texte sinngemäß
- [ ] Entwicklungsdatenbank löschen, `go test ./...` grün
- [ ] `wails dev` und `wails build` laufen auf Linux
