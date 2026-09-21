// Spaltenbreite der Mitgliederliste: wie die Spaltenwahl in main.js eine reine
// Ansichtsvorliebe an diesem einen Rechner, die localStorage trägt und nie
// durch Go läuft — der Server kennt nur die Sortierung, nicht wie breit
// jemand eine Spalte gezogen hat. Ziehbar sind nur die sechs Spalten mit
// data-col (siehe die lange Begründung am Tabellenkopf in
// mitglieder_liste.html); Nr. und Name bleiben fest und tragen deshalb keinen
// Ziehgriff.
//
// Angewendet wird nach jedem htmx-Austausch, aus demselben Grund wie bei
// spaltenAnwenden in main.js: Kopf- und Datenzeilen werden bei jeder Suche,
// jedem Filter und jedem Sortierklick neu gerendert und verlieren dabei jede
// zuvor gesetzte Breite.
const speicherSchluessel = 'boxclub.mitgliederliste.spaltenbreiten';

// breitenLesen liefert die gespeicherten Breiten je Spalte — oder ein leeres
// Objekt, wenn noch nichts gespeichert ist oder der Speicher nicht lesbar ist
// (privates Fenster, blockierter Seitenzugriff). Eine Spalte ohne gespeicherte
// Breite behält dann einfach die Breite aus ihrer Tailwind-Klasse im Template.
function breitenLesen() {
    try {
        const gespeichert = JSON.parse(localStorage.getItem(speicherSchluessel));
        if (gespeichert && typeof gespeichert === 'object' && !Array.isArray(gespeichert)) {
            return gespeichert;
        }
    } catch {
        // Siehe oben.
    }

    return {};
}

// breitenSchreiben speichert die Breiten. Schlägt das fehl, bleibt die
// gezogene Breite für diese Sitzung trotzdem wirksam — nur der nächste
// Neustart vergisst sie dann wieder (siehe main.js, sichtbareSpaltenSchreiben).
function breitenSchreiben(breiten) {
    try {
        localStorage.setItem(speicherSchluessel, JSON.stringify(breiten));
    } catch {
        // Siehe breitenLesen.
    }
}

// breitenAnwenden setzt die gespeicherten Breiten auf die passenden Kopfzellen
// im ganzen Dokument — wie spaltenAnwenden in main.js bewusst dokumentweit
// gescannt statt nur im ausgetauschten Teilbaum, aus demselben Grund (siehe
// dort): ein outerHTML-Tausch liefert in htmx.detail.target den schon
// ersetzten alten Knoten.
function breitenAnwenden() {
    const breiten = breitenLesen();
    document.querySelectorAll('th[data-col]').forEach((kopf) => {
        const breite = breiten[kopf.getAttribute('data-col')];
        if (breite) {
            kopf.style.width = breite + 'px';
        }
    });
}

document.body.addEventListener('htmx:afterSwap', breitenAnwenden);

// ziehend hält den Zustand eines laufenden Ziehvorgangs — außerhalb jeder
// Funktion, weil pointerdown, pointermove und pointerup drei getrennte
// Ereignisse sind, die sich denselben Zustand teilen müssen.
let ziehend = null;

document.body.addEventListener('pointerdown', (ereignis) => {
    const griff = ereignis.target.closest('[data-resize-griff]');
    if (!griff) {
        return;
    }

    const kopf = griff.closest('th[data-col]');
    if (!kopf) {
        return;
    }

    ziehend = {
        spalte: griff.getAttribute('data-resize-spalte'),
        kopf,
        startX: ereignis.clientX,
        startBreite: kopf.getBoundingClientRect().width,
    };
    // Pointer Capture hält Bewegung und Loslassen am Griff, auch wenn der
    // Zeiger dabei die schmale Griff-Spanne verlässt — sonst bräche das Ziehen
    // bei schneller Mausbewegung ab, sobald der Zeiger die zwei Pixel breite
    // Spanne verlässt.
    griff.setPointerCapture(ereignis.pointerId);
    ereignis.preventDefault();
});

// mindestbreite verhindert, dass eine Spalte auf einen Bruchteil ihres
// Inhalts schrumpft — Kennzeichen und Schaltflächen darin bräuchten sonst
// mehr Platz, als die Spalte noch hätte.
const mindestbreite = 60;

document.body.addEventListener('pointermove', (ereignis) => {
    if (!ziehend) {
        return;
    }

    const breite = Math.max(mindestbreite, ziehend.startBreite + (ereignis.clientX - ziehend.startX));
    ziehend.kopf.style.width = breite + 'px';
});

function ziehenBeenden() {
    if (!ziehend) {
        return;
    }

    const breiten = breitenLesen();
    breiten[ziehend.spalte] = Math.round(ziehend.kopf.getBoundingClientRect().width);
    breitenSchreiben(breiten);
    ziehend = null;
}

document.body.addEventListener('pointerup', ziehenBeenden);

// pointercancel statt pointerup, wenn das Fenster den Zeiger mitten im Ziehen
// verliert (z.B. Fensterwechsel) — ohne diesen Fall bliebe "ziehend" gesetzt
// und jede spätere, unbeteiligte Zeigerbewegung würde die Spalte weiterziehen.
document.body.addEventListener('pointercancel', ziehenBeenden);
