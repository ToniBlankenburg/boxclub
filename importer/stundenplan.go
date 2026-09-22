package importer

import (
	"strings"

	"github.com/ToniBlankenburg/boxclub/i18n"
	"github.com/ToniBlankenburg/boxclub/service"
)

// Der Stundenplan ist die eine Sache, die der Importer über die Datenbank
// wissen muss: gegen ihn hält er die Freitexte der Spalten „Training - 1/2/3".
// Er kommt von außen herein und wird nicht hier geholt — der Importer spricht
// den Service nicht an (spec.md → Seams).
//
// Angelegt wird dabei nichts. Jede Schreibvariante der Excel würde sonst ein
// Eintrag im Stundenplan, und nach dem ersten Lauf stünden zwanzig Termine da,
// von denen sechs echt sind (ADR-0008 → Alternativen). Was nicht trifft, geht
// in den Fehlerbericht und wird von Hand nachgetragen.

// Stundenplan ist der Terminkatalog in der Form, in der sich ein Freitext darin
// nachschlagen lässt: je Schreibweise die Termine, auf die sie passt.
//
// Die archivierten stehen mit darin, obwohl sie nicht vergeben werden. Nur so
// kann der Bericht „den gibt es nicht mehr" von „den gibt es gar nicht"
// unterscheiden — zwei Auskünfte, aus denen zwei verschiedene Handgriffe
// folgen.
type Stundenplan struct {
	nachSchreibweise map[string][]service.Trainingstermin

	// offene ist die Zahl der nicht archivierten Termine — alles, was der
	// Import überhaupt vergeben kann.
	offene int
}

// StundenplanAus baut den Katalog aus den Terminen des Vereins. Übergeben
// werden sie einschließlich der archivierten.
func StundenplanAus(termine []service.Trainingstermin) Stundenplan {
	plan := Stundenplan{nachSchreibweise: make(map[string][]service.Trainingstermin, len(termine)*2)}

	for _, termin := range termine {
		if !termin.Archiviert {
			plan.offene++
		}

		// Je Termin zählt jede Schreibweise einmal: fallen zwei zusammen —
		// ohne Ende und ohne Bezeichnung ist die Anzeige die kurze Form —,
		// stünde derselbe Termin sonst zweimal unter demselben Schlüssel und
		// machte sich selbst mehrdeutig.
		vergeben := map[string]bool{}
		for _, schreibweise := range schreibweisen(termin) {
			schluessel := vereinheitlicht(schreibweise)
			if vergeben[schluessel] {
				continue
			}
			vergeben[schluessel] = true
			plan.nachSchreibweise[schluessel] = append(plan.nachSchreibweise[schluessel], termin)
		}
	}

	return plan
}

// Leer sagt, ob kein Termin zu vergeben ist. Die Import-Ansicht fragt danach:
// wer 200 Mitglieder gegen einen leeren Stundenplan importiert, bekommt 400
// Meldungen und keine Ahnung, warum.
func (p Stundenplan) Leer() bool {
	return p.offene == 0
}

// zeitzusatz ist das Wort, das die Excel hinter die Uhrzeit setzt: dort steht
// „Samstag 10:30 Uhr", während die App den Termin „Samstag 10:30" schreibt.
// Es ist keine Angabe, sondern Schreibweise derselben Zeit — deshalb bietet
// jeder Termin beide Formen an, statt dass der Verein 200 Zellen ändert.
const zeitzusatz = "Uhr"

// schreibweisen sind die Texte, unter denen ein Termin zu finden ist. Alle drei
// werden exakt verglichen (vereinheitlicht); geraten wird an keiner. Geschrieben
// werden sie vom Termin selbst (KurzeAnzeige, Anzeige) — hier steht nur, welche
// davon die Excel benutzt.
//
// Die kurze Form steht dabei, weil die Excel weder Ende noch Bezeichnung führt:
// ohne sie träfe ein Termin mit Endzeit nie, und der Abgleich liefe für den
// gepflegten Stundenplan gerade dann ins Leere, wenn er vollständig gepflegt
// ist. Die volle Anzeige steht dabei, weil sie zwei gleichzeitige Gruppen
// unterscheidet — die kurze Form kann das nicht und meldet sie als mehrdeutig.
func schreibweisen(termin service.Trainingstermin) []string {
	kurz := termin.KurzeAnzeige()

	return []string{kurz, kurz + " " + zeitzusatz, termin.Anzeige()}
}

// vereinheitlicht macht zwei Texte vergleichbar, ohne sie zu deuten:
// Groß- und Kleinschreibung und Leerraum sind egal, alles andere muss stimmen.
// „Sa 10:30" trifft „Samstag 10:30" damit nicht — das ist gewollt und landet im
// Bericht (ADR-0008).
func vereinheitlicht(text string) string {
	return strings.ToLower(strings.Join(strings.Fields(text), " "))
}

// Die Auskünfte über einen Freitext, der keinen Termin ergibt, stehen im
// Katalog unter import.hinweis.termin_* (i18n/de.go) — jede sagt, was zu tun
// ist: der Bericht ist das Werkzeug, mit dem der Verein seinen Stundenplan
// und seine Tabelle zusammenbringt, und dafür muss aus ihm der nächste
// Handgriff hervorgehen. Die Sprache kommt vom Aufrufer (Ticket 05):
// importer/ kennt selbst keine.

// zuordnen schlägt einen Freitext im Stundenplan nach: entweder die ID des
// Termins und eine leere Meldung, oder 0 und die Auskunft, warum es keinen gibt.
//
// Ein archivierter Termin wird nicht vergeben: ein Import darf keine Zeiten
// austeilen, die es nicht mehr gibt. Bestehende Anmeldungen bleiben davon
// unberührt — die stehen in der Datenbank und nicht in dieser Tabelle.
func (p Stundenplan) zuordnen(spalte, text string, sprache i18n.Sprache) (int64, string) {
	treffer := p.nachSchreibweise[vereinheitlicht(text)]

	offen := make([]service.Trainingstermin, 0, len(treffer))
	for _, termin := range treffer {
		if !termin.Archiviert {
			offen = append(offen, termin)
		}
	}

	switch {
	case len(offen) == 1:
		return offen[0].ID, ""
	case len(offen) > 1:
		return 0, i18n.Text(sprache, "import.hinweis.termin_mehrdeutig", spalte, text)
	case len(treffer) > 0:
		return 0, i18n.Text(sprache, "import.hinweis.termin_archiviert", spalte, text)
	default:
		return 0, i18n.Text(sprache, "import.hinweis.termin_unbekannt", spalte, text)
	}
}
