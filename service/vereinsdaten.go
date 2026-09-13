package service

import (
	"fmt"
	"strings"
)

// Die Vereinsdaten sind die einzigen Angaben der App, die kein Mitglied
// betreffen (CONTEXT.md → Vereinsdaten): der Verein selbst, so wie er über einer
// Rechnung steht. Sie werden einmal von Hand gepflegt und nicht importiert.
//
// Sie stehen in der Datenbank und nicht im Code, weil die App sonst für genau
// einen Verein gebaut wäre und ein Tippfehler im Namen ein Entwicklungsauftrag.

// Vereinsdaten sind Name, Anschrift, Kontakt, Bankverbindung und Fußzeile des
// Vereins.
//
// Jede Angabe ist freiwillig, und keine wird geprüft. Was fehlt, fehlt später
// auf der Rechnung — das ist eine Sache des Vereins und keine, an der die App
// etwas aufhalten dürfte (ADR-0009). Der Nullwert sind leere Vereinsdaten und
// genau das, was eine frische Datenbank enthält.
type Vereinsdaten struct {
	Name string

	// Anschrift ist derselbe Typ wie am Mitglied: drei getrennte Angaben, aus
	// demselben Grund (CONTEXT.md → Anschrift). Die Briefanschrift eines
	// Vereins ist genauso gebaut wie die einer Person.
	Anschrift Anschrift

	Email   string
	Telefon string

	// IBAN steht wie am Mitglied als reiner Text da: keine Prüfziffer, kein
	// Format, keine SEPA-Datei (ADR-0006). BIC und Kreditinstitut stehen
	// daneben, weil sie auf der Rechnung danebenstehen.
	IBAN           string
	BIC            string
	Kreditinstitut string

	// Fusszeile ist mehrzeiliger Freitext: dort steht, was der Verein
	// steuerlich schreiben muss — etwa der Hinweis auf § 19 UStG. Die App kennt
	// dazu keine Regel und prüft nichts; sie weiß nicht, was ein Verein
	// schreiben muss, und geraten wäre hier schlimmer als leer.
	//
	// Sie ist auch der Ort, an dem später ein Logo landen könnte. Solange
	// niemand danach fragt, bleibt es bei Text (Ticket 23).
	Fusszeile string
}

// vereinsdatenID ist der Schlüssel der einen Zeile. Die Vereinsdaten sind keine
// Liste: es gibt einen Verein, und eine zweite Zeile wäre ein zweiter — die
// Tabelle nagelt das mit einem CHECK fest (siehe schema).
const vereinsdatenID = 1

// GetVereinsdaten liest die Angaben über den Verein.
//
// Dass die Zeile da ist, sichert Open zu (migrate). Deshalb steht hier kein
// Zweig für ihr Fehlen: er liefe in keinem Aufruf und in keinem Test, und eine
// Zusicherung, die vorsichtshalber auch ohne sie auskommt, ist keine. Fehlt sie
// doch, ist das ein Fehler und wird als solcher gemeldet — ein erneuter Start
// legt sie wieder an.
func (s *MemberService) GetVereinsdaten() (Vereinsdaten, error) {
	var daten Vereinsdaten

	err := s.db.QueryRow(
		`SELECT name, adresse, postleitzahl, ort, email, telefon,
		 	iban, bic, kreditinstitut, fusszeile
		 FROM vereinsdaten WHERE id = ?`, vereinsdatenID).Scan(
		&daten.Name, &daten.Anschrift.Adresse, &daten.Anschrift.Postleitzahl,
		&daten.Anschrift.Ort, &daten.Email, &daten.Telefon,
		&daten.IBAN, &daten.BIC, &daten.Kreditinstitut, &daten.Fusszeile)
	if err != nil {
		return Vereinsdaten{}, fmt.Errorf("vereinsdaten lesen: %w", err)
	}

	return daten, nil
}

// SetVereinsdaten ersetzt alle Angaben über den Verein. Das Formular schickt sie
// immer vollständig; ein geleertes Feld heißt deshalb „steht nicht mehr da" und
// nicht „unverändert".
//
// Geprüft wird nichts — es gibt keine Pflichtangabe und keine Regel, gegen die
// zu prüfen wäre. Abgeschnitten wird nur umschließender Leerraum, und zwar an
// jeder Angabe: diese Felder werden gedruckt (ADR-0009), und ein Leerzeichen
// hinter dem Vereinsnamen steht dann im Briefkopf des PDF. Bei der Fußzeile ist
// es die leere Zeile, die beim Tippen im Textfeld hängen bleibt.
//
// Geschrieben wird als Upsert — nicht als Notnagel für eine fehlende Zeile,
// sondern weil das die Schreibform für eine Zeile ist, die es genau einmal gibt:
// ein UPDATE ohne getroffene Zeile bliebe stillschweigend wirkungslos.
func (s *MemberService) SetVereinsdaten(daten Vereinsdaten) error {
	daten = daten.bereinigt()

	if _, err := s.db.Exec(
		`INSERT INTO vereinsdaten
		 	(id, name, adresse, postleitzahl, ort, email, telefon,
		 	 iban, bic, kreditinstitut, fusszeile)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		 	name = excluded.name,
		 	adresse = excluded.adresse,
		 	postleitzahl = excluded.postleitzahl,
		 	ort = excluded.ort,
		 	email = excluded.email,
		 	telefon = excluded.telefon,
		 	iban = excluded.iban,
		 	bic = excluded.bic,
		 	kreditinstitut = excluded.kreditinstitut,
		 	fusszeile = excluded.fusszeile`,
		vereinsdatenID, daten.Name,
		daten.Anschrift.Adresse, daten.Anschrift.Postleitzahl, daten.Anschrift.Ort,
		daten.Email, daten.Telefon,
		daten.IBAN, daten.BIC, daten.Kreditinstitut, daten.Fusszeile); err != nil {
		return fmt.Errorf("vereinsdaten speichern: %w", err)
	}

	return nil
}

// bereinigt schneidet von jeder Angabe den umschließenden Leerraum ab. Die
// Zeilenumbrüche *innerhalb* der Fußzeile bleiben stehen — sie sind dort die
// Angabe und kein Leerraum.
func (v Vereinsdaten) bereinigt() Vereinsdaten {
	return Vereinsdaten{
		Name: strings.TrimSpace(v.Name),
		Anschrift: Anschrift{
			Adresse:      strings.TrimSpace(v.Anschrift.Adresse),
			Postleitzahl: strings.TrimSpace(v.Anschrift.Postleitzahl),
			Ort:          strings.TrimSpace(v.Anschrift.Ort),
		},
		Email:          strings.TrimSpace(v.Email),
		Telefon:        strings.TrimSpace(v.Telefon),
		IBAN:           strings.TrimSpace(v.IBAN),
		BIC:            strings.TrimSpace(v.BIC),
		Kreditinstitut: strings.TrimSpace(v.Kreditinstitut),
		Fusszeile:      strings.TrimSpace(v.Fusszeile),
	}
}
