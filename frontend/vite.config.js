import tailwindcss from '@tailwindcss/vite';

export default {
    plugins: [tailwindcss()],

    // Ohne diese Zeile beantwortet Vites Dev-Server jeden unbekannten GET-Pfad mit
    // index.html (SPA-Fallback). Der Wails-Assetserver reicht Requests aber nur dann
    // an den Go-Handler weiter, wenn das Frontend 404 meldet — in `wails dev` würden
    // GET-Aufrufe auf /api/... sonst still die Startseite zurückliefern.
    // `mpa` liefert index.html weiterhin unter "/", aber ohne Fallback-Rewrite.
    appType: 'mpa',
};
