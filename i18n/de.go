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

	// dashboard.* ist der Überblick für Trainer und Admin (Ticket 03,
	// templates/dashboard.html). monatssoll ist eigens benannt, weil das Wort
	// zweimal vorkommt (Kennzahl und Fließtext) und beide Stellen dasselbe
	// Wort tragen sollen.
	"dashboard.titel":            "Dashboard",
	"dashboard.beschreibung_vor": "Mitgliederzahlen und das ",
	"dashboard.monatssoll":       "Monatssoll",
	"dashboard.beschreibung_nach": " des laufenden Monats. Das Monatssoll ist ein Soll und kein Ist: es " +
		"zeigt, was eingezogen werden soll, nicht, was tatsächlich ankam — das weiß nur die Bank. Es gibt hier " +
		"deshalb auch keinen Verlauf über mehrere Monate.",
	"dashboard.stat_davon_ruhend":       "davon ruhend",
	"dashboard.stat_neu_ausgetreten":    "Neu / Ausgetreten (Monat)",
	"dashboard.zustand_ueberschrift":    "Mitglieder nach Zustand (%d gesamt)",
	"dashboard.status_aktiv":            "Aktiv",
	"dashboard.status_kuendigungsfrist": "In Kündigungsfrist",
	"dashboard.status_neu":              "Neu",
	"dashboard.status_ruhend":           "Ruhend",
	"dashboard.status_ausgetreten":      "Ausgetreten",
	"dashboard.moneymoney_knopf":        "MoneyMoney-Export …",
	"dashboard.moneymoney_hinweis":      "CSV für den Lastschrifteinzug dieses Monats, eine Zeile je Mitgliedschaft.",
	"dashboard.moneymoney_warnung": "Achtung: In den Vereinsdaten ist ein eigener Verwendungszweck " +
		"hinterlegt — der Export verwendet ihn statt \"Vereinsname Beitrag MM/JJJJ\" und aktualisiert ihn " +
		"nicht selbst. Vor dem Export prüfen, ob er noch zu diesem Monat passt.",

	// trainingstermine.* ist der Stundenplan (Ticket 04,
	// templates/trainingstermine.html, app/trainingstermin.go). Die
	// Schreibweise eines einzelnen Termins selbst (Trainingstermin.Anzeige,
	// z. B. "Samstag 10:30 – 12:00 · Anfänger") entsteht im Service und bleibt
	// unübersetzt wie die übrigen Bezeichnung()-Ausgaben (Ticket 07) — hier
	// stehen nur die Texte drumherum. Feldbeschriftungen des Terminformulars
	// stehen unter feld.*, damit spätere Formulare sie mitbenutzen können.
	"trainingstermine.titel":                     "Trainingstermine",
	"trainingstermine.neu_knopf":                 "Neuer Trainingstermin",
	"trainingstermine.archivierte_anzeigen":      "Archivierte anzeigen",
	"trainingstermine.eintrag_singular":          "Termin",
	"trainingstermine.eintrag_plural":            "Termine",
	"trainingstermine.archiviert_abzeichen":      "archiviert",
	"trainingstermine.archiviert_titel":          "Archiviert — nicht mehr im Stundenplan und in keiner Auswahl. Bestehende Anmeldungen bleiben bestehen.",
	"trainingstermine.angemeldet_anzahl":         "%d angemeldet",
	"trainingstermine.niemand_angemeldet":        "niemand angemeldet",
	"trainingstermine.serienmail_aria":           "Serienmail an die Teilnehmer von %s",
	"trainingstermine.serienmail_knopf":          "Serienmail",
	"trainingstermine.bearbeiten_aria":           "%s bearbeiten",
	"trainingstermine.bearbeiten_knopf":          "Bearbeiten",
	"trainingstermine.zurueckholen_aria":         "%s wieder in den Stundenplan aufnehmen",
	"trainingstermine.archivieren_aria":          "%s archivieren",
	"trainingstermine.zurueckholen_knopf":        "Zurückholen",
	"trainingstermine.archivieren_knopf":         "Archivieren",
	"trainingstermine.leer_alle":                 "Kein Trainingstermin erfasst.",
	"trainingstermine.leer_aktiv":                "Kein Trainingstermin im Stundenplan.",
	"trainingstermine.leer_hinweis_import":       "Vor dem ersten Excel-Import sollte der Stundenplan stehen: der Import legt keine Termine an.",
	"trainingstermine.formular_titel_bearbeiten": "Trainingstermin bearbeiten",
	"trainingstermine.formular_titel_neu":        "Neuen Trainingstermin anlegen",
	"trainingstermine.formular_archiviert_hinweis": "Dieser Termin ist archiviert: er steht nicht im Stundenplan " +
		"und in keiner Auswahl. Das Speichern ändert daran nichts.",
	"trainingstermine.abschnitt_titel":   "Trainingstermin",
	"trainingstermine.ende_hinweis":      "Das Ende ist freiwillig. Ein Termin ist wiederkehrend und kein Datum: „Samstag 10:30“ gilt bis auf Weiteres.",
	"trainingstermine.wochentag_waehlen": "Bitte wählen",

	// Die vier Rückmeldungen nach Anlegen/Speichern/Archivieren/Zurückholen
	// kommen paarweise: die "_mit_name"-Variante nennt den Termin
	// (Trainingstermin.Anzeige), die "_ohne_name"-Variante greift, wenn er
	// zwischen Aktion und erneutem Lesen verschwunden ist (terminMeldung).
	"trainingstermine.angelegt_mit_name":     "»%s« wurde angelegt.",
	"trainingstermine.angelegt_ohne_name":    "Der Trainingstermin wurde angelegt.",
	"trainingstermine.gespeichert_mit_name":  "»%s« wurde gespeichert.",
	"trainingstermine.gespeichert_ohne_name": "Der Trainingstermin wurde gespeichert.",
	"trainingstermine.archiviert_mit_name":   "»%s« wurde archiviert und steht nicht mehr zur Auswahl.",
	"trainingstermine.archiviert_ohne_name":  "Der Trainingstermin wurde archiviert und steht nicht mehr zur Auswahl.",
	"trainingstermine.reaktiviert_mit_name":  "»%s« steht wieder im Stundenplan.",
	"trainingstermine.reaktiviert_ohne_name": "Der Trainingstermin steht wieder im Stundenplan.",
	"trainingstermine.nicht_gefunden":        "Diesen Trainingstermin gibt es nicht mehr.",

	"feld.wochentag":   "Wochentag",
	"feld.beginn":      "Beginn",
	"feld.ende":        "Ende",
	"feld.bezeichnung": "Bezeichnung",

	// wochentag.* sind die Wochentagsnamen der Auswahlliste im Terminformular
	// (Ticket 04) — anders als Wochentag.Bezeichnung() im Service (die die
	// Kurzschreibweise eines Termins bildet und deshalb unübersetzt bleiben
	// muss, siehe oben) ist diese Auswahlliste reine Formularbeschriftung
	// ohne Bezug zum Excel-Abgleich. Nummeriert nach service.Wochentag
	// (Montag = 1 … Sonntag = 7), damit die Zuordnung ohne zweite
	// Wochentagsliste im App-Paket auskommt.
	"wochentag.1": "Montag",
	"wochentag.2": "Dienstag",
	"wochentag.3": "Mittwoch",
	"wochentag.4": "Donnerstag",
	"wochentag.5": "Freitag",
	"wochentag.6": "Samstag",
	"wochentag.7": "Sonntag",

	// import.* ist der Excel-Import (Ticket 05, templates/import.html,
	// app/import.go). „Verwaltung“ und „Training - 1/2/3“ bleiben in beiden
	// Sprachen unverändert stehen: das sind die tatsächlichen Namen aus der
	// Vorlage des Vereins (importer.blatt, importer.spalteTraining1 usw.) und
	// keine Beschriftung, die zur Anzeigesprache gehört.
	"import.titel":            "Excel-Import",
	"import.beschreibung_vor": "Liest das Blatt „Verwaltung“ einer ",
	"import.beschreibung_nach": "-Datei. Der Import lässt sich beliebig oft wiederholen: bestehende " +
		"Mitglieder werden aktualisiert, nicht doppelt angelegt. Was sich nicht zuordnen lässt, steht " +
		"danach zeilenweise im Bericht.",
	"import.datei_aria":        "Excel-Datei auswählen",
	"import.importieren_knopf": "Importieren",
	"import.wird_gelesen":      "Wird gelesen …",

	"import.bericht_titel":        "Import-Bericht",
	"import.nochmal_knopf":        "Nochmal importieren",
	"import.zu_mitgliedern_knopf": "Zu den Mitgliedern",
	"import.stat_uebernommen":     "Übernommen",
	"import.stat_neu":             "Neu angelegt",
	"import.stat_aktualisiert":    "Aktualisiert",
	"import.stat_gescheitert":     "Gescheitert",
	"import.datei_praefix":        "Datei:",

	"import.gescheiterte_zeilen_titel": "Gescheiterte Zeilen",
	"import.gescheiterte_zeilen_hinweis": "Diese Zeilen wurden nicht übernommen. In Excel korrigieren und " +
		"erneut importieren — bereits übernommene Zeilen werden dabei nur aktualisiert.",
	"import.spalte_zeile": "Zeile",
	"import.spalte_grund": "Grund",

	"import.nachzutragen_titel": "Übernommen, aber nachzutragen",
	"import.nachzutragen_hinweis": "Diese Zeilen sind übernommen. Ihre Trainingstermine ließen sich im " +
		"Stundenplan nicht finden — erst dort pflegen und erneut importieren oder die Termine am Mitglied " +
		"von Hand zuordnen.",
	"import.spalte_hinweis": "Hinweis",

	"import.stundenplan_leer_vor": "Der Stundenplan ist leer. Der Import ordnet die Spalten " +
		"„Training - 1/2/3“ nur ",
	"import.stundenplan_leer_fett": "vorhandenen",
	"import.stundenplan_leer_nach": " Trainingsterminen zu und legt selbst keine an — bis dahin kommt " +
		"jedes Mitglied ohne Trainingszeit herein.",
	"import.stundenplan_leer_knopf": "Stundenplan pflegen",

	// import.fehler.* sind Gründe, an denen eine Zeile oder die ganze Datei
	// scheitert (importer.ExcelImporter) — import.hinweis.* sind Meldungen zu
	// Zeilen, die trotzdem durchgehen (siehe Ergebnis.Hinweise). Beide
	// brauchen die Sprache vom Aufrufer, weil importer/ selbst keine kennt
	// (Ticket 05).
	"import.fehler.datei_nicht_lesbar": "die Datei ließ sich nicht als .xlsx öffnen",
	"import.fehler.blatt_fehlt":        "das Blatt „%s“ fehlt in der Datei",
	"import.fehler.blatt_nicht_lesbar": "das Blatt „%s“ ließ sich nicht lesen",
	"import.fehler.keine_kopfzeile":    "das Blatt „%s“ hat keine Kopfzeile",
	"import.fehler.spalte_fehlt": "im Blatt „%s“ fehlt die Spalte „%s“ — die Datei passt nicht zur " +
		"erwarteten Tabelle",
	"import.fehler.id_doppelt":                "Mitglieds-ID %d doppelt — sie steht schon in Zeile %d",
	"import.fehler.eintrittsdatum_fehlt":      "in der Spalte „%s“ fehlt das Eintrittsdatum",
	"import.fehler.angabe_fehlt":              "in der Spalte „%s“ fehlt die Angabe",
	"import.fehler.mitglieds_id_fehlt":        "in der Spalte „%s“ fehlt die Mitglieds-ID",
	"import.fehler.mitglieds_id_ungueltig":    "die Mitglieds-ID „%s“ ist keine positive Zahl",
	"import.fehler.datum_unlesbar":            "in der Spalte „%s“ ist „%s“ kein lesbares Datum",
	"import.fehler.nur_jahreszahl":            "in der Spalte „%s“ ist „%s“ nur eine Jahreszahl und kein vollständiges Datum",
	"import.fehler.datum_ausserhalb":          "in der Spalte „%s“ ergibt „%s“ kein Datum, das in dieser Tabelle stehen kann",
	"import.fehler.spalte_mit_grund":          "in der Spalte „%s“: %s",
	"import.fehler.bewertung_unbekannt":       "in der Spalte „%s“ ist „%s“ kein bekannter Wert",
	"import.fehler.frequenz_unlesbar":         "in der Spalte „%s“ ist „%s“ keine lesbare Frequenz",
	"import.fehler.status_fehlt":              "in der Spalte „%s“ fehlt der Status",
	"import.fehler.status_unerwartetes_datum": "Status „%s“, aber in der Spalte „%s“ steht ein Datum",
	"import.fehler.status_datum_fehlt":        "Status „%s“, aber in der Spalte „%s“ fehlt das Datum",
	"import.fehler.status_unbekannt":          "unbekannter Status „%s“",

	"import.hinweis.frequenz_widerspruch": "Frequenz %d× widerspricht %d zugeordneten Trainingsterminen",
	"import.hinweis.termin_unbekannt": "in der Spalte „%s“ steht „%s“ nicht im Stundenplan — erst den " +
		"Termin dort anlegen, dann erneut importieren",
	"import.hinweis.termin_archiviert": "in der Spalte „%s“ ist „%s“ ein archivierter Termin und wird " +
		"nicht mehr vergeben — erst den Stundenplan pflegen, dann erneut importieren",
	"import.hinweis.termin_mehrdeutig": "in der Spalte „%s“ passt „%s“ auf mehrere Trainingstermine — in " +
		"der Excel die vollständige Schreibweise des gemeinten eintragen",

	"import.fehler.upload_zu_gross": "Die Datei ließ sich nicht entgegennehmen. Ist sie größer als %d MB?",
	"import.fehler.keine_datei":     "Bitte eine Datei auswählen.",
	"import.fehler.falsche_endung": "%q ist keine %s-Datei. Das alte .xls-Format liest der Import nicht " +
		"— in Excel einmal als .xlsx speichern.",

	// feld.*-Ergänzungen für Rechnung und Verein (Ticket 06) — dieselbe
	// Namensraum-Logik wie bei feld.wochentag & Co. (Ticket 04): Formulare,
	// die dieselbe Angabe brauchen, teilen sich den Schlüssel.
	"feld.empfaenger":        "Empfänger",
	"feld.rechnungsnummer":   "Rechnungsnummer",
	"feld.steuersatz":        "Steuersatz (%)",
	"feld.rechnungsdatum":    "Rechnungsdatum",
	"feld.zahlungsziel":      "Zahlungsziel",
	"feld.menge":             "Menge",
	"feld.einzelpreis_netto": "Einzelpreis (netto)",
	"feld.name_verein":       "Name des Vereins",
	"feld.bic":               "BIC",
	"feld.kreditinstitut":    "Kreditinstitut",
	// feld.position_bezeichnung ist eigens von feld.bezeichnung getrennt: das
	// dort für den Trainingstermin passende "Label" (englischer Katalog)
	// passt nicht auf eine Rechnungsposition — dort steht eine Beschreibung
	// der Leistung, kein Etikett.
	"feld.position_bezeichnung": "Bezeichnung",

	// rechnung.* ist das Rechnungsformular selbst (Ticket 06,
	// templates/rechnung.html, app/rechnung.go). rechnung.pdf.* sind die
	// Textbausteine auf dem erzeugten PDF (service.RechnungBeschriftungen,
	// service/rechnung_pdf.go) — service/ übersetzt sie nicht selbst
	// (ADR-0002, kein i18n-Import dort); app/rechnung.go löst sie über
	// rechnungBeschriftungen auf und reicht den fertigen Text durch. Damit
	// trägt eine erzeugte Rechnung die zum Erstellzeitpunkt aktive Sprache.
	"rechnung.titel": "Rechnung",
	"rechnung.beschreibung": "Für eine Leistung neben dem Beitrag — typisch ein Einzeltraining. Der " +
		"Empfänger ist frei überschreibbar: ein Einzeltraining nimmt auch, wer nie eintritt, nur bleibt " +
		"das PDF dann nirgends abgelegt.",
	"rechnung.abschnitt_empfaenger":     "Empfänger",
	"rechnung.abschnitt_rechnungsdaten": "Rechnungsdaten",
	"rechnung.abschnitt_positionen":     "Positionen",
	"rechnung.positionen_hinweis": "Menge und Einzelpreis (netto), etwa \"2\" und \"30,00\". Leere " +
		"Zeilen zählen nicht mit.",
	"rechnung.erstellen_knopf": "Rechnung erstellen",
	"rechnung.wird_erstellt":   "Wird erstellt …",

	"rechnung.pdf.telefon_praefix":         "Telefon: ",
	"rechnung.pdf.email_praefix":           "E-Mail: ",
	"rechnung.pdf.rechnungsnummer_praefix": "Rechnungsnummer: ",
	"rechnung.pdf.rechnungsdatum_praefix":  "Rechnungsdatum: ",
	"rechnung.pdf.zahlungsziel_praefix":    "Zahlungsziel: ",
	"rechnung.pdf.titel_praefix":           "Rechnung ",
	"rechnung.pdf.spalte_summe":            "Summe (netto)",
	"rechnung.pdf.netto_praefix":           "Netto: ",
	"rechnung.pdf.steuer_vorlage":          "zzgl. %s %% USt: %s",
	"rechnung.pdf.gesamtbetrag_praefix":    "Gesamtbetrag: ",
	"rechnung.pdf.iban_praefix":            "IBAN: ",
	"rechnung.pdf.bic_praefix":             "BIC: ",

	// rechnung.fehler_*/dialog_*/kein_speicherort_*/gespeichert_nach/
	// dateiname_praefix sind Text, den app/rechnung.go selbst zusammenbaut
	// (Formularprüfung, der Datei-Dialog beim Anbieten) — anders als die
	// Meldungen aus service.ValidierungsFehler (Ticket 07) entstehen sie
	// nicht im Service.
	"rechnung.fehler_rechnungsdatum":   "Rechnungsdatum ist kein gültiges Datum.",
	"rechnung.fehler_zahlungsziel":     "Zahlungsziel ist kein gültiges Datum.",
	"rechnung.fehler_position_menge":   "Position %d: die Menge ist keine gültige Zahl.",
	"rechnung.fehler_position_praefix": "Position %d: %s",
	"rechnung.dateiname_praefix":       "Rechnung ",
	"rechnung.dialog_nicht_verfuegbar": "Der Datei-Dialog steht nicht zur Verfügung — die Rechnung wurde " +
		"erstellt, aber nicht angeboten.",
	"rechnung.kein_speicherort_ohne_mitglied": "Ohne Speicherort ist das PDF weg — es gab niemanden, an " +
		"dem es hätte abgelegt werden können.",
	"rechnung.kein_speicherort_mit_mitglied": "Die Rechnung wurde nicht gespeichert. Sie ist am Mitglied " +
		"abgelegt, lässt sich von hier aus aber nicht noch einmal exportieren.",
	"rechnung.speichern_fehlgeschlagen": "Die Rechnung ließ sich nicht speichern: %v",
	"rechnung.gespeichert_nach":         "Die Rechnung wurde nach %s gespeichert.",

	// serienmail.* ist die Rückmeldung nach dem Vorbereiten einer Serienmail
	// (Ticket 06, templates/serienmail.html) — die Knöpfe, die dorthin
	// führen, sind schon unter mitglieder.serienmail_knopf/
	// trainingstermine.serienmail_knopf übersetzt (Ticket 02/04).
	"serienmail.vorbereitet":   "%d Empfänger für die Serienmail vorbereitet.",
	"serienmail.oeffnen_knopf": "Mail-Programm öffnen",
	"serienmail.keine_adresse": "Keine E-Mail-Adresse unter den Ausgewählten.",

	// verein.* ist die Vereinsseite (Ticket 06, templates/verein.html,
	// app/verein.go) — der Sprachumschalter selbst (sprache.*) ist schon aus
	// Ticket 01.
	"verein.titel": "Verein",
	"verein.beschreibung": "Name, Anschrift, Kontakt und Bankverbindung des Vereins. Sie stehen auf " +
		"jeder Rechnung, die die App schreibt. Alle Angaben sind freiwillig — was hier fehlt, fehlt dort.",
	"verein.bereiche_aria":     "Verein-Bereiche",
	"verein.tab_anschrift":     "Anschrift & Kontakt",
	"verein.tab_bank":          "Bankverbindung",
	"verein.tab_rechnungstext": "Rechnungstext",
	"verein.logo_feld":         "Vereinslogo",
	"verein.logo_entfernen":    "Entfernen",
	"verein.logo_alt":          "Aktuelles Vereinslogo",
	"verein.logo_datei_aria":   "Vereinslogo auswählen",
	"verein.logo_hinweis": "PNG, JPEG oder SVG, höchstens %s. Erscheint in der App und im Briefkopf " +
		"jeder erzeugten Rechnung.",
	"verein.moneymoney_feld": "MoneyMoney-Export: eigener Verwendungszweck",
	"verein.moneymoney_hinweis_vor": "Ersetzt im MoneyMoney-Export den automatisch erzeugten Text " +
		"\"Vereinsname Beitrag MM/JJJJ\". Leer lassen für den Automatismus. Ein gesetzter Text führt sich ",
	"verein.moneymoney_hinweis_fett": "nicht von selbst",
	"verein.moneymoney_hinweis_nach": " nach Monat und Jahr fort — nach dem Export hier wieder leeren, " +
		"sonst steht er auch im nächsten Monatslauf noch so da.",
	"verein.fusszeile_feld": "Fußzeile der Rechnung",
	// Das nicht-brechende Leerzeichen um "19" steht als  -Escape, nicht
	// als HTML-Entity: html/template escapt den Text aus i18n.Text beim
	// Einsetzen, und ein escaptes "&nbsp;" würde als sichtbares "&amp;nbsp;"
	// erscheinen statt als Leerzeichen (dieselbe Falle wie beim
	// Import-Spaltennamen, Ticket 05).
	"verein.fusszeile_hinweis": "Mehrzeilig. Hier steht, was unter der Rechnung stehen muss — etwa der " +
		"Hinweis auf § 19 UStG. Die App prüft daran nichts.",
	"verein.gespeichert": "Die Vereinsdaten wurden gespeichert.",
	"verein.formular_zu_gross": "Die Formulardaten ließen sich nicht entgegennehmen — das Logo ist " +
		"größer als %s.",
}
