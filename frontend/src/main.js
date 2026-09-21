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

    document.querySelectorAll('[data-spalten-kasten]').forEach((kasten) => {
        kasten.checked = sichtbar.includes(kasten.value);
    });

    // Der Zähler neben "Spalten" (ADR-0015) zeigt, wie viele der sechs
    // ausblendbaren Spalten gerade versteckt sind — Go kennt diesen Wert nie,
    // er lebt ausschließlich hier.
    const versteckt = spaltenAusblendbar.length - sichtbar.length;
    document.querySelectorAll('[data-spalten-badge]').forEach((badge) => {
        badge.textContent = versteckt;
        badge.hidden = versteckt === 0;
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
    const ehemalige = document.getElementById('filter-ehemalige');
    if (ehemalige && ehemalige.checked) parameter.set('ehemalige', '1');
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

// Ein Kästchen im Spaltenmenü ändert sich: neu speichern und sofort anwenden.
document.body.addEventListener('change', (ereignis) => {
    if (!ereignis.target.matches('[data-spalten-kasten]')) {
        return;
    }

    const sichtbar = spaltenAusblendbar.filter((spalte) => {
        const kasten = document.querySelector(`[data-spalten-kasten][value="${spalte}"]`);
        return !kasten || kasten.checked;
    });
    sichtbareSpaltenSchreiben(sichtbar);
    spaltenAnwenden();
});

// "Alle anzeigen" im Spaltenmenü setzt alle sechs Kästchen auf einmal zurück
// statt jedes einzeln (ADR-0015).
document.body.addEventListener('click', (ereignis) => {
    if (!ereignis.target.closest('[data-spalten-alle-anzeigen]')) {
        return;
    }

    sichtbareSpaltenSchreiben(spaltenAusblendbar.slice());
    spaltenAnwenden();
});

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
