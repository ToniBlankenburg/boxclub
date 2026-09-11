package service

import "slices"

// Das Geschlecht ist ein Freitextfeld am Mitglied und hat deshalb keinen
// eigenen Typ. Was hier steht, ist allein die Eintipphilfe dazu.

// geschlechtVorschlaege sind die Werte, die die Excel-Tabelle des Vereins in
// ihrem Dropdown führt.
var geschlechtVorschlaege = []string{"Frau", "Mann"}

// GeschlechtVorschlaege sind Eintipphilfen und keine Werteliste: gespeichert
// wird, was eingetippt wurde (Spec → keine Wertelisten-Prüfung). Der Verein
// führt heute zwei Werte; daraus eine Einschränkung zu machen hieße, für ein
// Mitglied, das in keinen von beiden passt, gar keinen Platz zu haben.
//
// Kopiert, damit ein Aufrufer die Liste nicht versehentlich umschreibt.
func GeschlechtVorschlaege() []string {
	return slices.Clone(geschlechtVorschlaege)
}
