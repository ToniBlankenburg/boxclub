# Spec: Stilllegung als Zeitraum

Ersetzt das Kennzeichen *ruhend* durch Zeiträume und führt die *Mindestlaufzeit* ein. Entscheidungen und Begründung: [ADR-0022](../../docs/adr/0022-stilllegung-als-zeitraum.md), Begriffe in `CONTEXT.md` → Stilllegung, Mindestlaufzeit.

## Zuschnitt

1. Umbenennung *ruhend* → *stillgelegt* (Ticket 01), ohne Verhaltensänderung.
2. Zeiträume, Beitragswirkung, Kündigungssperre, Mindestlaufzeit (Ticket 02).

Keine Migration: Entwicklungsdatenbank löschen.
