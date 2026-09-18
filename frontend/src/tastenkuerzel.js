// Tastenkürzel (Komfort-Features): drei App-weite Kürzel, die reine
// Bedienerleichterung sind und keine neue Fachlogik brauchen.
//
// Strg+Z steht bewusst NICHT hier: der native Undo des Browsers funktioniert
// in jedem Textfeld bereits ohne Zutun, solange kein eigener Tastatur-Handler
// ihn abfängt. Genau das vermeidet diese Datei — sie reagiert nur auf Tasten,
// die der Browser in Formularfeldern sonst nicht sinnvoll nutzt.
//
// Delegiert auf document, damit ein htmx-Austausch von #inhalt keine erneute
// Bindung braucht — wie registerkarten.js und die Spaltenwahl in main.js.
document.addEventListener('keydown', (ereignis) => {
    const strgOderCmd = ereignis.ctrlKey || ereignis.metaKey;

    // Strg+S schickt das Formular ab, in dem gerade der Fokus steht, statt
    // den Speichern-Dialog des Browsers zu öffnen. Reagiert nur, wenn der
    // Fokus tatsächlich in einem Formular steht — sonst gibt es nichts
    // Eindeutiges zum Abschicken, und der Browser-Dialog bleibt wie gewohnt.
    if (strgOderCmd && !ereignis.altKey && ereignis.key.toLowerCase() === 's') {
        const formular = ereignis.target.closest('form');
        if (formular) {
            ereignis.preventDefault();
            formular.requestSubmit();
        }
        return;
    }

    // Strg+F springt ins Suchfeld der Mitgliederliste, wenn es gerade auf der
    // Seite steht. Sonst bleibt die native Browsersuche unangetastet — die
    // ist auf jeder anderen Ansicht (Rechnung, Trainingstermine, Verein)
    // weiterhin die richtige Wahl.
    if (strgOderCmd && !ereignis.altKey && ereignis.key.toLowerCase() === 'f') {
        const suchfeld = document.getElementById('suchfeld');
        if (suchfeld) {
            ereignis.preventDefault();
            suchfeld.focus();
            suchfeld.select();
        }
        return;
    }

    // Esc klickt den Abbrechen-Knopf des Formulars, in dem der Fokus steht —
    // dieselbe hx-get-Aktion, die der Knopf sonst per Maus auslöst (Formular
    // verlassen, Zeile zurück in die Ansicht). Ein eigenes data-shortcut-
    // Attribut markiert die Knöpfe, statt auf den Anzeigetext "Abbrechen" zu
    // matchen.
    if (ereignis.key === 'Escape') {
        const formular = ereignis.target.closest('form');
        const abbrechenKnopf = formular && formular.querySelector('[data-shortcut="abbrechen"]');
        if (abbrechenKnopf) {
            ereignis.preventDefault();
            abbrechenKnopf.click();
        }
    }
});
