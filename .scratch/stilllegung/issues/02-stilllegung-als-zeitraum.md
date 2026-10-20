# 02: Stilllegung als Zeitraum und Mindestlaufzeit

**What to build:** Tabelle `stilllegung(id, mitgliedschaft_id, von, bis, grund NULL)` löst das Boolean ab; `mitgliedschaft.mindestlaufzeit_monate NULL` kommt dazu. Erfassen, Historie, Bearbeiten und Löschen laufen in „Mitglied bearbeiten"; der Knopf in der Mitgliederliste entfällt, das Abzeichen bleibt. „Wieder aktiv" setzt `bis` auf heute (eine noch nicht begonnene wird gelöscht). Stillgelegt ist, wer heute in einer Stilllegung liegt.

**Blocked by:** 01

**Status:** ready-for-agent

- [x] `von` und `bis` Pflicht, `bis ≥ von`; keine Überlappung und keine neue, solange eine läuft (Fehler als Schlüssel, ADR-0018); `bis` hinter `austritt` erlaubt, ohne Prüfung
- [x] Monatssoll und MoneyMoney-Export lassen einen Monat aus, sobald ein Tag darin stillgelegt ist; Teilnehmerliste liest „heute stillgelegt"
- [x] Kündigung erfassen ist in einer Stilllegung gesperrt, der Knopf erklärt es beim Darüberfahren; eine schon erfasste Kündigung verbietet keine Stilllegung
- [x] `mindestlaufzeit_monate` im Formular neben dem Beitrag; Ende abgelesen (Eintritt + Monate + Stilllegungsmonate), angezeigt im Formular und im Kündigungsformular; Warnung ohne Sperre, Vorschlag bleibt
- [x] Wiedereintritt: neue Mitgliedschaft ohne Stilllegungen und ohne Mindestlaufzeit, alte bleiben
- [x] Excel-Import legt nichts an; `Stillgelegt`/`Inakiv` erscheinen im Fehlerbericht
- [x] Tests nach dem Muster von `service/member_service_test.go` (echte SQLite, nur Service-API); `go test ./...` grün
- [x] `wails dev` und `wails build` laufen auf Linux
