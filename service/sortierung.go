package service

import (
	"slices"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// Sortierspalte ist eine Spalte, nach der sich die Mitgliederliste sortieren
// lässt — eine je Kopfzeile der Liste (Ticket 27: Spalten ein- und ausblenden,
// nach Spalte sortieren). Der Nullwert ist der Standard, den es schon vor
// diesem Ticket gab.
type Sortierspalte int

const (
	// SortierspalteName ist der Nullwert: Nachname, dann Vorname — dieselbe
	// Sortierung, die nachNamenSortieren schon vor diesem Ticket lieferte.
	SortierspalteName Sortierspalte = iota
	SortierspalteMitgliedID
	SortierspalteStatus
	SortierspalteAnschrift
	SortierspalteTraining
	SortierspalteBeitrag
	SortierspalteRueckstand
	SortierspalteEintritt
)

// Sortierrichtung dreht eine Sortierung um. Der Nullwert ist aufsteigend: der
// erste Klick auf eine noch nicht aktive Spaltenüberschrift sortiert immer so
// herum, ein erneuter Klick auf dieselbe Spalte dreht auf absteigend.
type Sortierrichtung int

const (
	SortierrichtungAufsteigend Sortierrichtung = iota
	SortierrichtungAbsteigend
)

// Sortierung ist Spalte und Richtung zusammen. Der Nullwert ist "Name,
// aufsteigend" — dieselbe Standardansicht, die Suchfilter{} schon vor diesem
// Ticket lieferte, und damit auch das, worauf eine ausgeblendete aktive
// Sortierspalte zurückfällt (siehe app.spaltenkopf).
type Sortierung struct {
	Spalte   Sortierspalte
	Richtung Sortierrichtung
}

// eintraegeSortieren ordnet die Liste nach der gewählten Spalte und Richtung.
// Sortiert wird ausschließlich hier im Speicher und nie in SQL (ADR-0004) —
// wie Suche und Filter es schon vorher taten.
//
// Leere Werte landen immer am Ende, unabhängig von der Richtung: ein
// Mitglied ohne Anschrift oder ohne Trainingstermin soll die erste Seite
// nicht füllen, nur weil "leer" beim Sortieren zufällig vorn landet — weder
// aufsteigend noch absteigend. Bei Gleichstand, und ebenso wenn beide Werte
// leer sind, entscheidet derselbe Name-Vergleich wie beim Nullwert: die
// Reihenfolge bleibt damit vorhersagbar, unabhängig von der gewählten Spalte.
func eintraegeSortieren(liste []Listeneintrag, sortierung Sortierung) {
	// Der Standardfall bleibt exakt das, was nachNamenSortieren schon vor
	// diesem Ticket tat — dafür bürgt derselbe Code und keine zweite,
	// gleichlautende Implementierung. Er wird damit zum Sonderfall dieser
	// allgemeineren Sortierung, nicht ersetzt.
	if sortierung == (Sortierung{}) {
		nachNamenSortieren(liste)
		return
	}

	// Ein Collator ist nicht nebenläufigkeitssicher, deshalb einer pro Aufruf
	// (wie in nachNamenSortieren).
	sortierer := collate.New(language.German)

	slices.SortStableFunc(liste, func(a, b Listeneintrag) int {
		ea, eb := spaltenwertLeer(a, sortierung.Spalte), spaltenwertLeer(b, sortierung.Spalte)
		if ea != eb {
			if ea {
				return 1
			}
			return -1
		}

		if !ea {
			if c := spaltenVergleich(a, b, sortierung.Spalte, sortierer); c != 0 {
				if sortierung.Richtung == SortierrichtungAbsteigend {
					return -c
				}
				return c
			}
		}

		// Gleichstand oder beide leer: wie beim Nullwert nach Namen auflösen.
		if c := sortierer.CompareString(a.Nachname, b.Nachname); c != 0 {
			return c
		}
		if c := sortierer.CompareString(a.Vorname, b.Vorname); c != 0 {
			return c
		}

		return int(a.MitgliedID - b.MitgliedID)
	})
}

// spaltenwertLeer sagt, ob eine Zeile in der gewählten Spalte nichts zu zeigen
// hat. Nur Anschrift und Training kennen das: die übrigen Spalten sind bei
// jedem Mitglied gesetzt (Name, Nummer, Status, Eintritt) oder ihr Nullwert
// ist ein gültiger Wert und keine fehlende Angabe — 0 € Beitrag zahlt zum
// Beispiel, wer nichts zahlt, und ist nicht "kein Beitrag erfasst" (siehe
// Mitgliedschaft.BeitragCents).
func spaltenwertLeer(e Listeneintrag, spalte Sortierspalte) bool {
	switch spalte {
	case SortierspalteAnschrift:
		return e.Anschrift.Leer()
	case SortierspalteTraining:
		return len(e.Trainingstermine) == 0
	default:
		return false
	}
}

// spaltenVergleich vergleicht zwei nicht-leere Werte einer Spalte aufsteigend
// — negativ, wenn a vor b gehört, positiv sonst. Die Richtung dreht der
// Aufrufer um; hier steht deshalb nur die aufsteigende Regel je Spaltentyp.
func spaltenVergleich(a, b Listeneintrag, spalte Sortierspalte, sortierer *collate.Collator) int {
	switch spalte {
	case SortierspalteMitgliedID:
		return int(a.MitgliedID - b.MitgliedID)
	case SortierspalteStatus:
		// Der Lebenszyklus, nicht das Alphabet: Neu, Aktiv, In
		// Kündigungsfrist, Ausgetreten — genau die Reihenfolge, in der die
		// Status-Konstanten stehen (siehe status.go).
		return int(a.Status()) - int(b.Status())
	case SortierspalteAnschrift:
		// Ort vor Straße: wer nach Anschrift sortiert, will meist wissen, wer
		// sich in derselben Stadt trifft — nicht, wessen Straßenname
		// alphabetisch zuerst kommt.
		if c := sortierer.CompareString(a.Anschrift.Ort, b.Anschrift.Ort); c != 0 {
			return c
		}
		return sortierer.CompareString(a.Anschrift.Adresse, b.Anschrift.Adresse)
	case SortierspalteTraining:
		return int(a.Trainingsfrequenz() - b.Trainingsfrequenz())
	case SortierspalteBeitrag:
		return int(a.BeitragCents - b.BeitragCents)
	case SortierspalteRueckstand:
		return vergleichBool(a.Rueckstand.Offen, b.Rueckstand.Offen)
	case SortierspalteEintritt:
		return a.Eintritt.Compare(b.Eintritt)
	default: // SortierspalteName absteigend — aufsteigend fängt eintraegeSortieren oben schon ab.
		if c := sortierer.CompareString(a.Nachname, b.Nachname); c != 0 {
			return c
		}
		return sortierer.CompareString(a.Vorname, b.Vorname)
	}
}

// vergleichBool ordnet "in Ordnung" vor "im Rückstand" — dieselbe Reihenfolge
// wie in rueckstandsoptionen (app.go): der Normalfall zuerst, die
// Ausnahmeliste, die der Verein tatsächlich abarbeitet, danach.
func vergleichBool(a, b bool) int {
	switch {
	case a == b:
		return 0
	case !a:
		return -1
	default:
		return 1
	}
}
