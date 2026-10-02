// Zwei-Griff-Regler für den Beitragsfilter der Mitgliederliste: zwei
// übereinandergelegte <input type="range"> (siehe mitglieder_liste.html),
// die dieses Skript aufeinander abstimmt — von darf nie über bis hinaus,
// die Füllung folgt den Griffen, die Anzeige nennt beide Werte. Der Wert
// selbst reist weiter über die beiden Eingabefelder und "change" zum Server.
//
// Delegiert auf document und läuft nach jedem htmx-Austausch neu, weil die
// Regler in #mitglieder-ergebnis liegen und dabei durch frische Knoten
// ersetzt werden.
function beitragsbereichAktualisieren(bereich, geaendert) {
    const von = bereich.querySelector('#filter-beitrag-von');
    const bis = bereich.querySelector('#filter-beitrag-bis');
    if (!von || !bis) return;

    // Der gerade bewegte Griff schiebt den anderen nicht mit, er hält an ihm an.
    if (parseFloat(von.value) > parseFloat(bis.value)) {
        if (geaendert === bis) bis.value = von.value;
        else von.value = bis.value;
    }

    const min = parseFloat(von.min);
    const spanne = parseFloat(von.max) - min;
    const anteil = (eingabe) => (spanne > 0 ? (parseFloat(eingabe.value) - min) / spanne : 0);
    const anteilVon = anteil(von);
    const anteilBis = anteil(bis);
    bereich.style.setProperty('--von', anteilVon);
    bereich.style.setProperty('--bis', anteilBis);

    // Liegen beide Griffe am selben Ende, muss der obere der sein, der noch
    // von dort wegkann — sonst bliebe der andere unerreichbar darunter.
    von.style.zIndex = anteilVon > 0.5 ? '2' : '1';
    bis.style.zIndex = anteilVon > 0.5 ? '1' : '2';

    const euro = (eingabe) => eingabe.value.replace('.', ',');
    bereich.querySelector('output').textContent = euro(von) + ' – ' + euro(bis) + ' €';
}

document.addEventListener('input', (ereignis) => {
    const bereich = ereignis.target.closest && ereignis.target.closest('[data-beitragsbereich]');
    if (bereich) beitragsbereichAktualisieren(bereich, ereignis.target);
});

document.body.addEventListener('htmx:afterSwap', () => {
    document.querySelectorAll('[data-beitragsbereich]').forEach((b) => beitragsbereichAktualisieren(b, null));
});
