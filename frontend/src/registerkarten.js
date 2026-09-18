// Registerkarten (Tabs, Ticket 08): schaltet zwischen den Bereichen einer
// Seite um, die mehrere eigenständige Formulare stapelt (Mitglied:
// Stammdaten/Verträge/Rechnung) oder ein Formular in mehrere Gruppenboxen
// aufteilt (Verein). Optik steckt vollständig in den Tailwind-Klassen der
// Templates (aria-selected:-Varianten) — hier steht nur das Umschalten.
//
// Delegiert auf document, damit ein htmx-Austausch von #inhalt keine erneute
// Bindung braucht — genau wie die Spaltenwahl der Mitgliederliste oben.
document.addEventListener('click', (ereignis) => {
    const knopf = ereignis.target.closest('[data-rk-tab]');
    if (!knopf) return;

    const gruppe = knopf.closest('[data-registerkarten]');
    if (!gruppe) return;

    const ziel = knopf.dataset.rkTab;
    gruppe.querySelectorAll('[data-rk-tab]').forEach((b) => {
        b.setAttribute('aria-selected', String(b === knopf));
    });
    gruppe.querySelectorAll(':scope > [data-rk-panel]').forEach((panel) => {
        panel.hidden = panel.dataset.rkPanel !== ziel;
    });
});
