Status: needs-triage

# 08: frontend/index.html über das Backend ausliefern (optional)

**What to build:** `frontend/index.html` ist eine statische, von Vite
gebaute Datei mit drei deutschen Strings (Fenstertitel, `aria-label`,
Ladehinweis "Laden …"). Sie über den `App.Handler()` statt über den
Wails-Assetserver auszuliefern, würde sie als Go-Template behandelbar
machen. Das ist ein Eingriff in das Aufbau-Muster aus ADR-0002 (htmx über
den Assetserver, `frontend/dist` ist ein statischer Build) und lohnt sich
nur, wenn der kurze Blitz unübersetzten Texts beim Programmstart tatsächlich
stört.

**Blocked by:** 01 (i18n-Infrastruktur)

**Status needs-triage:** ist der Aufwand gerechtfertigt, oder bleibt die
Lücke wie in der Spec beschrieben bestehen? Rücksprache mit Nutzer vor dem
Start.
