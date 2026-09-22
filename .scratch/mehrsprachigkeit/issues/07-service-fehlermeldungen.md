Status: ready-for-agent

# 07: Fehlermeldungen und Validierungstexte in service/ übersetzt

**What to build:** `service.ValidierungsFehler`, Sentinel-Fehler
(`ErrNichtAktiv` u. ä.) und alle `fmt.Errorf`-Texte, die bis in die
Oberfläche durchgereicht werden (`app.fehlerAntwort`), bekommen
Übersetzungen. `service/` kennt heute keine Sprache — das ist der größte
Bruch mit dem bestehenden Seam (ADR-0002: `service/` ist reine Fachlogik,
keine Präsentation) und muss sorgfältig geschnitten werden: der Service
sollte weiterhin Sprache-unabhängige Fehlerwerte liefern (Sentinel-Fehler,
Fehlercodes), die `app/` erst am Rand in Text übersetzt — nicht der Service
selbst, der sonst eine Präsentationsabhängigkeit bekäme.

**Blocked by:** 01 (i18n-Infrastruktur)

## Acceptance Criteria

- [ ] Klären: bekommt jeder Fehlerfall einen stabilen Schlüssel/Code, den
      `app/` übersetzt, oder bleibt der Fehlertext selbst deutsch und wird
      nur an bekannten Stellen ersetzt? (Architekturentscheidung, ggf. ADR)
- [ ] `service/` bekommt **keine** Abhängigkeit auf `i18n` (ADR-0002 bleibt
      gültig) — Übersetzung passiert in `app/`
- [ ] `go test ./...` grün, `wails build` unter Linux grün
