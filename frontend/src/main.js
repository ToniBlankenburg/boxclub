// htmx wird von Vite mitgebündelt, damit die App vollständig offline läuft.
// Eigene Logik gibt es hier bis auf zwei Ausnahmen nicht: alle übrigen
// Interaktionen laufen über hx-*-Attribute gegen die Go-Handler in app/.
//
// Die erste Ausnahme ist die Spaltenwahl der Mitgliederliste (Ticket 27):
// welche Spalten sichtbar sind, ist eine reine Ansichtsvorliebe an diesem
// einen Rechner und überlebt den Neustart über localStorage — "kein Go, keine
// Datenbank" (CLAUDE.md). Dafür braucht es zwangsläufig Code im Browser: Go
// bekommt localStorage nie zu Gesicht. Die Sortierung selbst läuft dagegen
// wie gehabt über hx-*-Attribute gegen den Server (ADR-0004) — hier unten
// steht nur der Rückfall, wenn eine ausgeblendete Spalte gerade die aktive
// Sortierspalte ist.
//
// Die zweite ist das "Alle auswählen"-Kästchen der Serienmail-Spalte: es
// setzt nur andere Kästchen im DOM und hat mit dem Server nichts zu tun,
// anders als die Auswahl selbst, die als Formularwert über hx-include mitfährt.
//
// Die dritte ist der mailto-Link der Serienmail selbst (siehe unten): ein
// gewöhnlicher <a href="mailto:..."> lässt das WebView versuchen, sich selbst
// dorthin zu navigieren — es kann das Schema nicht darstellen, und die ganze
// Ansicht wird schwarz. Wails' Laufzeit-Funktion BrowserOpenURL reicht die
// Adresse stattdessen an das Betriebssystem weiter, wie ein Klick in einem
// gewöhnlichen Browser das täte.
import 'htmx.org';
import './style.css';
import './registerkarten.js';
import './tastenkuerzel.js';
import './spaltenbreite.js';

// spaltenAusblendbar sind die Schlüssel der Spalten, die sich ausblenden
// lassen — dieselben, die app.spaltenAusblendbar in Go kennt (Status,
// Anschrift, Training, Beitrag, Rückstand, Eintritt). Nr. und Name bleiben
// immer sichtbar und stehen deshalb nicht hier.
const spaltenAusblendbar = ['status', 'anschrift', 'training', 'beitrag', 'rueckstand', 'eintritt'];

// speicherSchluessel ist der localStorage-Schlüssel der Spaltenwahl.
const speicherSchluessel = 'boxclub.mitgliederliste.spalten';

// sichtbareSpaltenLesen liefert die gespeicherte Auswahl der ausblendbaren
// Spalten — oder alle, wenn noch nichts gespeichert ist (der Standard: nichts
// ausgeblendet) oder der Speicher nicht lesbar ist (privates Fenster,
// blockierter Seitenzugriff).
function sichtbareSpaltenLesen() {
    try {
        const gespeichert = JSON.parse(localStorage.getItem(speicherSchluessel));
        if (Array.isArray(gespeichert)) {
            return spaltenAusblendbar.filter((spalte) => gespeichert.includes(spalte));
        }
    } catch {
        // Kein Zugriff oder kein gültiges JSON — dann gilt der Standard unten.
    }

    return spaltenAusblendbar.slice();
}

// sichtbareSpaltenSchreiben speichert die Auswahl. Schlägt das fehl, bleibt
// sie für diese Sitzung trotzdem wirksam — nur der nächste Neustart vergisst
// sie dann wieder.
function sichtbareSpaltenSchreiben(sichtbar) {
    try {
        localStorage.setItem(speicherSchluessel, JSON.stringify(sichtbar));
    } catch {
        // Siehe sichtbareSpaltenLesen.
    }
}

// spaltenAnwenden blendet die ausgeblendeten Spalten im ganzen Dokument aus:
// jede Zelle mit passendem data-col, plus die zusammengefasste Zelle der drei
// Inline-Formulare (Kündigung, Wiedereintritt, Rückstand), deren colspan
// mitschrumpft, damit die Zeile mit der Tabelle ausgerichtet bleibt.
//
// Aufgerufen wird sie nach jedem htmx-Austausch (siehe unten), weil jeder von
// ihnen Kopf- oder Datenzeilen neu einsetzt, die ohne das hier wieder alle
// Spalten zeigen würden. Gescannt wird bewusst immer das ganze Dokument statt
// nur des ausgetauschten Teilbaums: htmx.detail.target ist bei einem
// outerHTML-Tausch — genau das, was die drei Inline-Formulare benutzen — der
// schon ersetzte, nicht mehr im Baum hängende alte Knoten, und ein Scan gegen
// ihn träfe die neue Zeile gar nicht. Bei 50–200 Mitgliedern kostet der
// vollständige Durchlauf nichts.
function spaltenAnwenden() {
    const sichtbar = sichtbareSpaltenLesen();

    document.querySelectorAll('[data-col]').forEach((zelle) => {
        const spalte = zelle.getAttribute('data-col');
        if (spaltenAusblendbar.includes(spalte)) {
            zelle.hidden = !sichtbar.includes(spalte);
        }
    });

    document.querySelectorAll('[data-colspan-optional]').forEach((zelle) => {
        zelle.colSpan = Math.max(1, sichtbar.length);
    });

    // Das Fallback-Element am Ende der Kopfzeile (ADR-0020) zeigt je einen
    // Knopf zum Wiedereinblenden — aber nur für Spalten, die gerade
    // tatsächlich versteckt sind; die übrigen bleiben hier verborgen.
    document.querySelectorAll('[data-spalten-zeigen]').forEach((knopf) => {
        knopf.hidden = sichtbar.includes(knopf.getAttribute('data-spalten-zeigen'));
    });

    // Der Zähler am Fallback-Element zeigt, wie viele der sechs ausblendbaren
    // Spalten gerade versteckt sind — Go kennt diesen Wert nie, er lebt
    // ausschließlich hier.
    const versteckt = spaltenAusblendbar.length - sichtbar.length;
    document.querySelectorAll('[data-spalten-fallback-badge]').forEach((badge) => {
        badge.textContent = versteckt;
        badge.hidden = versteckt === 0;
    });
    document.querySelectorAll('[data-spalten-fallback-leer]').forEach((hinweis) => {
        hinweis.hidden = versteckt !== 0;
    });

    sortierfallbackPruefen(sichtbar);
}

// sortierfallbackPruefen setzt die Sortierung auf Namen zurück, sobald die
// gerade aktive Sortierspalte ausgeblendet wird: eine ausgeblendete Spalte
// lässt sich nicht als Sortierspalte auswählen (Ticket 27). Das läuft hier
// und nicht in Go, weil die Sichtbarkeit nie durch Go läuft — der Server weiß
// gar nicht, welche Spalte gerade ausgeblendet ist.
function sortierfallbackPruefen(sichtbar) {
    const ergebnis = document.getElementById('mitglieder-ergebnis');
    const sortFeld = ergebnis && ergebnis.querySelector('input[name="sort"]');
    if (!sortFeld || !window.htmx) {
        return;
    }

    // Das leere Feld heißt "Name" (service.Sortierung-Nullwert) — und Name
    // ist nie ausgeblendet, dann ist hier ohnehin nichts zu tun.
    const aktiveSpalte = sortFeld.value;
    if (!spaltenAusblendbar.includes(aktiveSpalte) || sichtbar.includes(aktiveSpalte)) {
        return;
    }

    const parameter = new URLSearchParams();
    const suchfeld = document.getElementById('suchfeld');
    if (suchfeld && suchfeld.value) parameter.set('q', suchfeld.value);
    const rueckstand = document.getElementById('filter-rueckstand');
    if (rueckstand && rueckstand.value) parameter.set('rueckstand', rueckstand.value);
    const frequenz = document.getElementById('filter-frequenz');
    if (frequenz && frequenz.value) parameter.set('frequenz', frequenz.value);
    const geschlecht = document.getElementById('filter-geschlecht');
    if (geschlecht && geschlecht.value) parameter.set('geschlecht', geschlecht.value);
    const termin = document.getElementById('filter-termin');
    if (termin && termin.value) parameter.set('termin', termin.value);
    const status = document.getElementById('filter-status');
    if (status && status.value) parameter.set('status', status.value);
    const ruhend = document.getElementById('filter-ruhend');
    if (ruhend && ruhend.value) parameter.set('ruhend', ruhend.value);
    // Der Regler trägt immer einen Wert (nie leer wie ein Auswahlfeld) — an
    // seiner eigenen Grenze steht er unbewegt, genau wie Go ihn ohne Filter
    // rendert (listeDaten.BeitragVonWert/-BisWert), und grenzt dann nicht ein.
    const beitragVon = document.getElementById('filter-beitrag-von');
    if (beitragVon && beitragVon.value !== beitragVon.min) parameter.set('beitragVon', beitragVon.value);
    const beitragBis = document.getElementById('filter-beitrag-bis');
    if (beitragBis && beitragBis.value !== beitragBis.max) parameter.set('beitragBis', beitragBis.value);
    // sort und richtung bleiben weg: der fehlende Wert ist bereits der
    // Standard "Name, aufsteigend" (service.Sortierung-Nullwert).

    window.htmx.ajax('GET', '/api/mitglieder/ergebnis?' + parameter.toString(), {
        target: '#mitglieder-ergebnis',
        swap: 'innerHTML',
    });
}

// Nach jedem htmx-Austausch neu anwenden — das deckt sowohl den ersten Aufbau
// der Seite (#inhalt lädt selbst per hx-trigger="load") als auch jede spätere
// Suche, jeden Filter, jeden Sortierklick und jede Zeile ab, die ein
// Inline-Formular ersetzt.
document.body.addEventListener('htmx:afterSwap', spaltenAnwenden);

// "Spalte ausblenden" im Zahnrad-Menü einer Spalte (ADR-0020): sie fliegt aus
// der sichtbaren Auswahl, ihr eigenes Zahnrad verschwindet damit gleich mit —
// wiederzufinden ist sie über das Fallback-Element am Ende der Kopfzeile.
document.body.addEventListener('click', (ereignis) => {
    const knopf = ereignis.target.closest('[data-spalten-verstecken]');
    if (!knopf) {
        return;
    }

    const spalte = knopf.getAttribute('data-spalten-verstecken');
    const sichtbar = sichtbareSpaltenLesen().filter((s) => s !== spalte);
    sichtbareSpaltenSchreiben(sichtbar);
    spaltenAnwenden();
});

// Ein Knopf im Fallback-Element blendet seine Spalte wieder ein.
document.body.addEventListener('click', (ereignis) => {
    const knopf = ereignis.target.closest('[data-spalten-zeigen]');
    if (!knopf) {
        return;
    }

    const spalte = knopf.getAttribute('data-spalten-zeigen');
    const sichtbar = sichtbareSpaltenLesen();
    if (!sichtbar.includes(spalte)) {
        // In der Reihenfolge von spaltenAusblendbar wieder einsetzen, statt
        // ans Ende zu hängen — dieselbe Reihenfolge, in der main.js die
        // Spalten überall sonst aufzählt.
        sichtbareSpaltenSchreiben(spaltenAusblendbar.filter((s) => sichtbar.includes(s) || s === spalte));
    }
    spaltenAnwenden();
});

// Nur ein Zahnrad-Menü gleichzeitig offen (spec.md, Story 3): sobald eines
// aufklappt, schließen alle anderen. Das "toggle"-Ereignis eines <details>
// blubbert nicht — abgefangen wird es deshalb in der Erfassungsphase, die
// jedes Ereignis unabhängig davon erreicht.
document.body.addEventListener('toggle', (ereignis) => {
    const details = ereignis.target;
    if (!(details instanceof HTMLDetailsElement) || !details.matches('[data-spaltenmenu]') || !details.open) {
        return;
    }

    document.querySelectorAll('details[data-spaltenmenu][open]').forEach((andere) => {
        if (andere !== details) andere.open = false;
    });
}, true);

// Das Kästchen in der Kopfzelle der Auswahlspalte (Serienmail) setzt alle
// sichtbaren Kästchen der Zeilen auf einmal — reine Bedienungshilfe, ohne
// eigenen Zustand: bei jedem htmx-Austausch kommt die Kopfzelle unangekreuzt
// zurück, wie die Zeilen selbst auch (siehe mitglieder_liste.html). Gesucht
// wird bei jedem Klick neu statt einmalig gebunden, weil htmx die Zeilen bei
// Suche und Filter ersetzt.
document.body.addEventListener('change', (ereignis) => {
    if (!ereignis.target.matches('[data-mitglied-alle-auswaehlen]')) {
        return;
    }

    document.querySelectorAll('[name="mitglied_id"]').forEach((kasten) => {
        kasten.checked = ereignis.target.checked;
    });
});

// Der "Mail-Programm öffnen"-Knopf der Serienmail (serienmail.html) ist ein
// mailto-Link — abgefangen und an BrowserOpenURL gereicht statt das WebView
// selbst navigieren zu lassen (siehe die Begründung oben). window.runtime
// steht in jedem Wails-Fenster zur Verfügung, unabhängig von den eigenen
// Go-Bindings, die diese App sonst nicht nutzt (ADR-0002).
document.body.addEventListener('click', (ereignis) => {
    const link = ereignis.target.closest('a[href^="mailto:"]');
    if (!link) {
        return;
    }

    ereignis.preventDefault();
    window.runtime.BrowserOpenURL(link.href);
});
