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

	// allgemein.* sind Wörter, die in mehreren Formularen gleich lauten
	// (Ticket 02) — ein Schlüssel für alle, damit ein späteres Ticket ihn
	// wiederverwendet statt neu zu übersetzen.
	"allgemein.speichern": "Speichern",
	"allgemein.abbrechen": "Abbrechen",

	// feld.* sind Beschriftungen von Eingabefeldern, die das Teil-Template
	// "feld" (templates/mitglied_formular.html) über mehrere Formulare
	// hinweg teilt — mitglied_formular.html führt sie in Ticket 02 ein,
	// rechnung.html/trainingstermine.html/verein.html verweisen in späteren
	// Tickets auf dieselben Schlüssel, statt sie zu verdoppeln.
	"feld.vorname":               "Vorname",
	"feld.nachname":              "Nachname",
	"feld.geburtsdatum":          "Geburtsdatum",
	"feld.geschlecht":            "Geschlecht",
	"feld.adresse":               "Adresse (Straße und Hausnummer)",
	"feld.postleitzahl":          "Postleitzahl",
	"feld.ort":                   "Ort",
	"feld.email":                 "E-Mail",
	"feld.telefon":               "Telefon",
	"feld.iban":                  "IBAN",
	"feld.anmeldedatum":          "Anmeldedatum",
	"feld.eintrittsdatum":        "Eintrittsdatum",
	"feld.anmeldegebuehr":        "Anmeldegebühr in € (einmalig)",
	"feld.beitrag":               "Beitrag in € pro Monat",
	"feld.google_bewertung":      "Google-Bewertung",
	"feld.training":              "Training",
	"feld.nicht_aenderbar_titel": "Nach der Anlage nicht mehr änderbar",

	// spalte.* sind die Kopfzeilen der Mitgliederliste (Ticket 02) — vorher
	// in spalte.Beschriftung fest verdrahtet, jetzt über
	// listeDaten.spaltenkopfFuer je Anzeigesprache aufgelöst.
	"spalte.nr":         "Nr.",
	"spalte.name":       "Name",
	"spalte.status":     "Status",
	"spalte.anschrift":  "Anschrift",
	"spalte.training":   "Training",
	"spalte.beitrag":    "Beitrag",
	"spalte.rueckstand": "Rückstand",
	"spalte.eintritt":   "Eintritt",

	// filter.* sind die Einträge der vier Auswahllisten in der Filterleiste
	// der Mitgliederliste (Ticket 02).
	"filter.rueckstand.alle":  "Alle Mitglieder",
	"filter.rueckstand.offen": "Im Rückstand",
	"filter.rueckstand.ok":    "In Ordnung",
	"filter.frequenz.alle":    "Jede Frequenz",
	"filter.geschlecht.alle":  "Jedes Geschlecht",
	"filter.termin.alle":      "Jeder Termin",

	// mitglieder.* ist die Listenansicht samt Filterleiste und Zeilenaktionen
	// (Ticket 02, templates/mitglieder_liste.html).
	"mitglieder.titel":                         "Mitglieder",
	"mitglieder.serienmail_knopf":              "Serienmail",
	"mitglieder.neu_knopf":                     "Neues Mitglied",
	"mitglieder.suche_platzhalter":             "Name, E-Mail, Telefon oder Mitglieds-ID",
	"mitglieder.suche_aria":                    "Mitglieder durchsuchen",
	"mitglieder.filter_knopf":                  "Filter",
	"mitglieder.zuruecksetzen_titel":           "Zurücksetzen",
	"mitglieder.zuruecksetzen_aria":            "Suche und Filter zurücksetzen",
	"mitglieder.spalten_aria":                  "Sichtbare Spalten wählen",
	"mitglieder.spalten_knopf":                 "Spalten",
	"mitglieder.spalten_alle_anzeigen":         "Alle anzeigen",
	"mitglieder.filter_rueckstand_aria":        "Nach Rückstand filtern",
	"mitglieder.filter_frequenz_aria":          "Nach Trainingsfrequenz filtern",
	"mitglieder.filter_geschlecht_aria":        "Nach Geschlecht filtern",
	"mitglieder.filter_termin_aria":            "Nach Trainingstermin filtern",
	"mitglieder.auch_ehemalige":                "Auch Ehemalige",
	"mitglieder.eintrag_singular":              "Mitglied",
	"mitglieder.eintrag_plural":                "Mitglieder",
	"mitglieder.leer_gefiltert":                "Kein Mitglied passt zu Suche und Filter.",
	"mitglieder.leer":                          "Keine Mitglieder erfasst.",
	"mitglieder.alle_auswaehlen_aria":          "Alle sichtbaren Mitglieder auswählen",
	"mitglieder.mitglieds_id_titel":            "Mitglieds-ID",
	"mitglieder.sortieren_nach":                "Nach %s sortieren",
	"mitglieder.aktuell_aufsteigend":           " (aktuell aufsteigend)",
	"mitglieder.aktuell_absteigend":            " (aktuell absteigend)",
	"mitglieder.zeile_bearbeiten_titel":        "Bearbeiten",
	"mitglieder.zeile_bearbeiten_aria":         "%s %s bearbeiten",
	"mitglieder.serienmail_auswaehlen_aria":    "%s %s für die Serienmail auswählen",
	"mitglieder.google_bewertung_titel":        "Google-Bewertung: %s",
	"mitglieder.ruhend_titel":                  "Ruhend — kein Beitragseinzug, die Mitgliedschaft läuft weiter (Status: %s)",
	"mitglieder.ruhend_abzeichen":              "ruhend",
	"mitglieder.austritt_zum":                  "zum %s",
	"mitglieder.gekuendigt_seit":               "gekündigt %s",
	"mitglieder.archiviert_suffix":             "(archiviert)",
	"mitglieder.rueckstand_aendern_knopf":      "Ändern",
	"mitglieder.rueckstand_pflegen_aria":       "Rückstand von %s %s pflegen",
	"mitglieder.wiedereintritt_fuer_aria":      "Wiedereintritt für %s %s",
	"mitglieder.kuendigung_erfassen_fuer_aria": "Kündigung erfassen für %s %s",
	"mitglieder.wiedereintritt_knopf":          "Wiedereintritt",
	"mitglieder.kuendigung_knopf":              "Kündigung",
	"mitglieder.ruhend_schalten_aria":          "%s %s ruhend schalten",
	"mitglieder.aktiv_setzen_aria":             "%s %s wieder aktiv setzen",
	"mitglieder.ruhend_knopf":                  "Ruhend",
	"mitglieder.aktiv_setzen_knopf":            "Aktiv setzen",
	"mitglieder.gekuendigt_am_feld":            "Gekündigt am",
	"mitglieder.austritt_zum_feld":             "Austritt zum",
	"mitglieder.kuendigung_erfassen_knopf":     "Kündigung erfassen",
	"mitglieder.kuendigung_hinweis": "Vorgeschlagen ist die reguläre Frist von drei Monaten zum " +
		"Monatsende. Steht der Termin noch nicht fest, leere das Austrittsfeld " +
		"— bis dahin trainiert und zahlt das Mitglied weiter.",
	"mitglieder.wiedereintritt_zum_feld": "Wiedereintritt zum",
	"mitglieder.wieder_eintreten_knopf":  "Wieder eintreten",
	"mitglieder.wiedereintritt_hinweis":  "Es entsteht eine neue Mitgliedschaft; Stammdaten und Historie bleiben erhalten.",
	"mitglieder.im_rueckstand_kasten":    "Im Rückstand",
	"mitglieder.notiz_feld":              "Notiz",
	"mitglieder.notiz_platzhalter":       "Rücklastschrift Oktober, angeschrieben am 05.10.",
	"mitglieder.rueckstand_hinweis":      "Die Notiz bleibt auch dann stehen, wenn der Rückstand erledigt ist.",

	// mitglied.* ist das Mitglied-Formular selbst: Registerkarten,
	// Überschriften und Hinweistexte (Ticket 02,
	// templates/mitglied_formular.html). Feldbeschriftungen stehen unter
	// feld.*, weil das Teil-Template auch andere Formulare bedient.
	"mitglied.bereiche_aria":            "Mitglied-Bereiche",
	"mitglied.tab_stammdaten":           "Stammdaten",
	"mitglied.tab_vertraege":            "Verträge",
	"mitglied.tab_rechnung":             "Rechnung",
	"mitglied.bearbeiten_titel":         "Mitglied bearbeiten",
	"mitglied.anlegen_titel":            "Neues Mitglied anlegen",
	"mitglied.mitglieds_id_titel":       "Mitglieds-ID — einmal vergeben und nie neu verwendet",
	"mitglied.nr_praefix":               "Nr.",
	"mitglied.abschnitt_person":         "Person",
	"mitglied.abschnitt_kontakt":        "Kontakt & Anschrift",
	"mitglied.abschnitt_mitgliedschaft": "Mitgliedschaft",
	"mitglied.abschnitt_sonstiges":      "Sonstiges",
	"mitglied.training_hinweis": "Höchstens drei Termine. Die Trainingsfrequenz ergibt sich aus der " +
		"Anzahl der Kreuze; kein Kreuz heißt „keine Frequenz“ und nicht „1×“.",
	"mitglied.stundenplan_leer_vor": "Der Stundenplan ist noch leer — Trainingszeiten lassen sich erst " +
		"zuordnen, wenn im Bereich ",
	"mitglied.stundenplan_leer_nach": " welche angelegt sind.",

	// meldung.* sind die Rückmeldungen nach einer abgeschlossenen Aktion an
	// der Mitgliederliste (Ticket 02, app.meldung).
	"meldung.angelegt":            "%s %s wurde angelegt.",
	"meldung.gespeichert":         "%s %s wurde gespeichert.",
	"meldung.austritt_erfasst":    "Für %s %s ist der Austritt zum %s erfasst; die Zeile steht jetzt unter »%s«.",
	"meldung.wiedereingetreten":   "%s %s ist zum %s wieder eingetreten.",
	"meldung.mitglied_nicht_mehr": "Dieses Mitglied gibt es nicht mehr.",
	"meldung.bereits_ausgetreten": "Dieses Mitglied ist bereits ausgetreten.",
	"meldung.bereits_aktiv":       "Dieses Mitglied ist bereits aktiv.",
}
