// Spaltenreihenfolge der Mitgliederliste: wie Spaltenwahl und -breite eine
// reine Ansichtsvorliebe an diesem Rechner, in localStorage und nie in Go.
// Verschiebbar sind alle Spalten mit data-col (Name und die sechs
// ausblendbaren); Auswahlkästchen und Mitglieds-ID tragen kein data-col, stehen
// immer vorn und bleiben fest.
//
// Die Zellen werden im DOM umsortiert statt per CSS "order": eine Tabelle kennt
// kein flex-order. Angewendet wird nach jedem htmx-Austausch, weil Kopf- und
// Datenzeilen dabei neu in Standardreihenfolge gerendert werden.
//
// Gezogen wird mit Zeigerereignissen statt HTML5-Drag&Drop, damit die Spalten
// schon während des Ziehens einrasten: Beim Ziehen haben alle verschiebbaren
// Spalten dieselbe feste Breite, die gezogene folgt als Kopf-Schatten dem
// Zeiger, und sobald dessen Mitte eine andere Spalte erreicht, tauschen beide
// ihren Platz. Beim Loslassen kehren alle zu ihrer eigenen Breite zurück.
const standardReihenfolge = ['name', 'status', 'anschrift', 'training', 'beitrag', 'rueckstand', 'eintritt'];
const speicherSchluessel = 'boxclub.mitgliederliste.spaltenreihenfolge';

function reihenfolgeLesen() {
    try {
        const gespeichert = JSON.parse(localStorage.getItem(speicherSchluessel));
        if (Array.isArray(gespeichert)) {
            // Unbekannte Schlüssel fallen weg, neue stehen am Ende.
            const bekannt = gespeichert.filter((s) => standardReihenfolge.includes(s));
            return bekannt.concat(standardReihenfolge.filter((s) => !bekannt.includes(s)));
        }
    } catch {
        // Kein Zugriff oder kein gültiges JSON — dann gilt der Standard.
    }

    return standardReihenfolge.slice();
}

function reihenfolgeSchreiben(reihenfolge) {
    try {
        localStorage.setItem(speicherSchluessel, JSON.stringify(reihenfolge));
    } catch {
        // Die Reihenfolge gilt trotzdem für diese Sitzung.
    }
}

function reihenfolgeAnwenden(reihenfolge = reihenfolgeLesen()) {
    document.querySelectorAll('#mitglieder-ergebnis table tr').forEach((zeile) => {
        const zellen = Array.from(zeile.children);
        const beweglich = zellen.filter((z) => z.hasAttribute('data-col'));
        if (beweglich.length < 2) {
            return;
        }

        const sortiert = reihenfolge
            .map((schluessel) => beweglich.find((z) => z.getAttribute('data-col') === schluessel))
            .filter(Boolean);
        const schonRichtig = sortiert.every((z, i) => z === beweglich[i]);
        if (!schonRichtig) {
            zeile.append(...zellen.filter((z) => !beweglich.includes(z)), ...sortiert);
        }
    });
}

document.body.addEventListener('htmx:afterSwap', () => reihenfolgeAnwenden());

// Einheitsbreite aller verschiebbaren Spalten, solange eine gezogen wird.
const ziehBreite = '10rem';
// Ab dieser Zeigerbewegung (px) gilt ein Klick auf den Kopf als Ziehen —
// darunter bleibt es ein Sortierklick.
const ziehSchwelle = 5;

let zug = null;

function spaltenZellen(schluessel) {
    return document.querySelectorAll(`#mitglieder-ergebnis [data-col="${schluessel}"]`);
}

function sichtbareKoepfe() {
    return Array.from(document.querySelectorAll('#mitglieder-ergebnis th[data-col]')).filter((k) => !k.hidden);
}

function zugStarten() {
    const koepfe = sichtbareKoepfe();
    zug.breiten = new Map(koepfe.map((k) => [k, k.style.width]));
    const reck = zug.kopf.getBoundingClientRect();

    // Der Schatten ist eine Kopie der Beschriftung, die dem Zeiger folgt.
    const schatten = document.createElement('div');
    schatten.textContent = zug.kopf.textContent.replace(/\s+/g, ' ').trim();
    schatten.style.cssText = `position:fixed;z-index:50;pointer-events:none;top:${reck.top}px;`
        + 'padding:0.5rem 1rem;font-weight:500;font-size:0.875rem;background:#fff;'
        + 'border:2px solid var(--akzent);border-radius:0.375rem;box-shadow:0 6px 16px rgb(0 0 0/.2);'
        + `width:${ziehBreite};box-sizing:border-box;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;`;
    document.body.append(schatten);
    zug.schatten = schatten;

    koepfe.forEach((k) => { k.style.width = ziehBreite; });
    // Erst nach dem Schrumpfen messen: die gezogene Spalte sitzt jetzt woanders.
    zug.startMitte = zug.kopf.getBoundingClientRect().left + zug.kopf.getBoundingClientRect().width / 2;
    spaltenZellen(zug.schluessel).forEach((z) => { z.style.opacity = '0.35'; z.style.background = 'var(--akzent-ring-15, #eee)'; });
    document.body.style.userSelect = 'none';
    document.body.style.cursor = 'grabbing';
}

function zugBewegen(ereignis) {
    const dx = ereignis.clientX - zug.startX;
    if (!zug.laeuft) {
        if (Math.abs(dx) < ziehSchwelle) {
            return;
        }
        zug.laeuft = true;
        zugStarten();
    }

    const mitte = zug.startMitte + dx;
    const breite = zug.schatten.offsetWidth;
    zug.schatten.style.left = `${mitte - breite / 2}px`;

    const ziel = sichtbareKoepfe().find((k) => {
        const reck = k.getBoundingClientRect();
        return mitte >= reck.left && mitte < reck.right;
    });
    if (!ziel || ziel.getAttribute('data-col') === zug.schluessel) {
        return;
    }

    const alt = reihenfolgeLesen();
    const zielSchluessel = ziel.getAttribute('data-col');
    const nachRechts = alt.indexOf(zug.schluessel) < alt.indexOf(zielSchluessel);
    const neu = alt.filter((s) => s !== zug.schluessel);
    neu.splice(neu.indexOf(zielSchluessel) + (nachRechts ? 1 : 0), 0, zug.schluessel);
    zug.reihenfolge = neu;
    reihenfolgeSchreiben(neu);
    reihenfolgeAnwenden(neu);
}

function zugBeenden() {
    if (!zug) {
        return;
    }

    const beendet = zug;
    zug = null;
    document.removeEventListener('pointermove', zugBewegen);
    document.removeEventListener('pointerup', zugBeenden);
    document.removeEventListener('pointercancel', zugBeenden);
    if (!beendet.laeuft) {
        return;
    }

    beendet.schatten.remove();
    beendet.breiten.forEach((breite, kopf) => { kopf.style.width = breite; });
    spaltenZellen(beendet.schluessel).forEach((z) => { z.style.opacity = ''; z.style.background = ''; });
    document.body.style.userSelect = '';
    document.body.style.cursor = '';
    // Der Klick nach dem Loslassen soll nicht auch noch sortieren.
    const schlucken = (e) => e.stopPropagation();
    document.addEventListener('click', schlucken, { capture: true, once: true });
    setTimeout(() => document.removeEventListener('click', schlucken, true), 0);
}

document.body.addEventListener('pointerdown', (ereignis) => {
    if (ereignis.button !== 0 || ereignis.target.closest('[data-resize-griff], details, input, select')) {
        return;
    }
    const kopf = ereignis.target.closest('#mitglieder-ergebnis th[data-col]');
    if (!kopf) {
        return;
    }

    zug = { kopf, schluessel: kopf.getAttribute('data-col'), startX: ereignis.clientX, laeuft: false };
    document.addEventListener('pointermove', zugBewegen);
    document.addEventListener('pointerup', zugBeenden);
    document.addEventListener('pointercancel', zugBeenden);
});
