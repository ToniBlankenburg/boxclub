package i18n

// deutsch ist der deutsche Katalog — zugleich der Rückfall für jeden
// Schlüssel, der im aktiven Katalog fehlt (siehe Text). Schlüssel sind nach
// Bereich benannt ("nav." für die Kopfzeilen-Navigation, "sprache." für den
// Sprachumschalter), damit spätere Tickets eigene Namensräume ergänzen
// können, ohne bestehende Schlüssel zu berühren.
var deutsch = map[string]string{
	"nav.mitglieder":       "Mitglieder",
	"nav.dashboard":        "Dashboard",
	"nav.trainingstermine": "Trainingstermine",
	"nav.import":           "Excel-Import",
	"nav.rechnung":         "Rechnung",
	"nav.verein":           "Verein",

	"sprache.name.de": "Deutsch",
	"sprache.name.en": "English",
	"sprache.hinweis": "Sprache",
}
