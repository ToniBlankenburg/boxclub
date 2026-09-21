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

// mindestbreite verhindert, dass eine Spalte auf einen Bruchteil ihres
// Inhalts schrumpft — Kennzeichen und Schaltflächen darin bräuchten sonst
// mehr Platz, als die Spalte noch hätte. Sie gilt auch für Name (siehe
// maximaleBreite): die Tabelle steht unter table-fixed auf w-full, und die
// sechs gezogenen Spalten behalten dabei immer genau ihre gesetzte Breite —
// bei zu wenig Platz geht der ganze Fehlbetrag sonst auf die einzige Spalte
// ohne eigene Breite, Name, die dabei bis auf 0 kollabieren kann (in Chromium
// beobachtet; CSS min-width auf der th greift dort nicht, weil table-fixed
// den Fehlbetrag über die Spaltenbreiten-Formel verteilt statt über die
// normale Box-Größenberechnung).
const mindestbreite = 60;

// ziehend hält den Zustand eines laufenden Ziehvorgangs — außerhalb jeder
// Funktion, weil pointerdown, pointermove und pointerup drei getrennte
// Ereignisse sind, die sich denselben Zustand teilen müssen.
let ziehend = null;

// andereSpalten liefert die Summe der Breiten aller Kopfzellen einer Zeile
// außer der übergebenen und außer Name (deren Platz maximaleBreite reserviert)
// — Nr. zählt mit, weil sie wie die sechs gezogenen Spalten der Tabelle ihre
// Breite entzieht, nur eben fest statt gezogen.
function andereSpalten(kopf) {
    const zeile = kopf.closest('tr');
    let summe = 0;
    for (const zelle of zeile.children) {
        if (zelle === kopf || zelle === zeile.children[1]) {
            continue;
        }
        summe += zelle.getBoundingClientRect().width;
    }
    return summe;
}

// maximaleBreite deckelt, wie weit sich eine Spalte ziehen lässt: die Tabelle
// selbst wächst nicht über ihren Container hinaus (w-full), also muss für
// Name mindestens mindestbreite übrig bleiben — sonst kollabiert sie (siehe
// mindestbreite oben). Während des Ziehens ändert sich nur die eine Spalte,
// die Tabellenbreite und die übrigen Spalten bleiben also fest, weshalb sich
// das einmal beim Griff-Runterdrücken ausrechnen lässt.
function maximaleBreite(kopf) {
    const tabelle = kopf.closest('table');
    return tabelle.getBoundingClientRect().width - andereSpalten(kopf) - mindestbreite;
}

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
        maxBreite: maximaleBreite(kopf),
    };
    // Pointer Capture hält Bewegung und Loslassen am Griff, auch wenn der
    // Zeiger dabei die schmale Griff-Spanne verlässt — sonst bräche das Ziehen
    // bei schneller Mausbewegung ab, sobald der Zeiger die zwei Pixel breite
    // Spanne verlässt.
    griff.setPointerCapture(ereignis.pointerId);
    ereignis.preventDefault();
});

document.body.addEventListener('pointermove', (ereignis) => {
    if (!ziehend) {
        return;
    }

    const gewuenscht = ziehend.startBreite + (ereignis.clientX - ziehend.startX);
    const breite = Math.min(ziehend.maxBreite, Math.max(mindestbreite, gewuenscht));
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
