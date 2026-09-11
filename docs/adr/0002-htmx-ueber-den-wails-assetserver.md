# ADR-0002: htmx spricht über den Wails-Assetserver, nicht über Wails-Bindings

**Status:** Accepted
**Datum:** 2026-09-10

## Kontext

[ADR-0001](0001-go-wails-fuer-desktop-gui.md) legt Wails v2 mit einem
htmx-Frontend fest. Die Spec skizziert den Frontend-Backend-Kontrakt als
Wails-Bindings, die HTML-Fragmente zurückgeben — z. B.
`ListMembers(filter) → HTML-Fragment`.

Beim ersten Vertikal-Cut (Ticket 02) zeigte sich, dass beides nicht
zusammenpasst: Wails-Bindings sind **JavaScript-Funktionen**, die ein Promise
liefern. htmx dagegen löst Interaktionen über HTTP aus und tauscht die Antwort
selbst in den DOM ein. Um Bindings mit htmx zu verheiraten, müsste man für jeden
Aufruf von Hand Promise auflösen, Ziel-Element suchen und Fragment einsetzen —
also genau die Arbeit nachbauen, für die htmx überhaupt eingesetzt wird. Der
Gewinn aus ADR-0001 ("kein SPA-Framework, wenig Frontend-Aufwand") wäre dahin.

Wails bietet dafür einen dokumentierten Ausweg: `assetserver.Options.Handler`.
Der Assetserver reicht **alle Nicht-GET-Requests** sowie **GET-Requests, die
keine statische Datei treffen**, an einen selbst gestellten `http.Handler`
durch.

## Entscheidung

Das Frontend spricht per htmx über normale HTTP-Requests gegen `/api/...` mit
dem Go-Backend. Diese Routen bedient ein `http.ServeMux` aus dem Package `app/`,
das über `assetserver.Options.Handler` im Wails-Assetserver hängt.

**Wails-Bindings werden nicht verwendet.** `options.App.Bind` bleibt leer.

Die Rolle von `app/` ändert sich dadurch nicht: es bleibt die dünne
Adapter-Schicht, die Eingaben parst, an `MemberService` bzw. `ExcelImporter`
delegiert und ein HTML-Fragment aus `templates/` rendert. Maßgeblich ist nicht
die Zahl der Service-Aufrufe, sondern dass **keine Fachregel** in `app/` liegt:
Pflichtfelder, Rückstand und Lebenszyklus gehören in `service/`, hier wird
nur angezeigt, was von dort kommt. Nur der Transport ist HTTP statt Wails-IPC.

## Betrachtete Alternativen

### Wails-Bindings plus JS-Shim

Bindings behalten und in `main.js` eine Hilfsfunktion schreiben, die das Promise
auflöst und das Fragment einsetzt. **Contra:** htmx wird zum toten Gewicht — die
Swap-Logik läge wieder handgeschrieben im Frontend. Widerspricht ADR-0001.

### Eigener HTTP-Server auf einem lokalen Port

Ein `net/http`-Server auf `localhost:<port>` neben Wails. **Contra:** öffnet
einen Port für andere Prozesse auf dem Rechner, braucht Portwahl und
CORS-Betrachtung — für eine reine Desktop-App unnötiges Risiko. Der Assetserver
ist prozessintern.

## Konsequenzen

- **Positiv:** htmx funktioniert wie vorgesehen, ohne eigenes JavaScript.
  `frontend/src/main.js` besteht aus zwei Import-Zeilen. Die Handler sind
  gewöhnliche `http.Handler` und damit ohne Wails-Toolchain per `httptest`
  ansteuerbar, falls sie später doch Tests brauchen.
- **Negativ:** Der in der Spec skizzierte Binding-Kontrakt gilt nicht mehr;
  spätere Tickets müssen Routen statt Binding-Methoden benennen.
- **Fallstrick (behoben):** Vites Dev-Server beantwortet standardmäßig jeden
  unbekannten GET-Pfad mit `index.html` (SPA-Fallback). Der Assetserver reicht
  aber nur weiter, wenn das Frontend 404 meldet — unter `wails dev` liefen
  GET-Aufrufe auf `/api/...` deshalb still ins Leere und gaben die Startseite
  zurück. `frontend/vite.config.js` setzt darum `appType: 'mpa'`. Unter
  `wails build` tritt das Problem nicht auf, weil dort die eingebettete
  `embed.FS` bedient wird.
- **Reversibilität:** hoch. Ein Wechsel zurück auf Bindings beträfe nur `app/`
  und die `hx-*`-Attribute in `templates/`; `service/` und `importer/` bleiben
  unberührt.
