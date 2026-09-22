package i18n

import "testing"

func TestText(t *testing.T) {
	tests := []struct {
		name     string
		sprache  Sprache
		schluess string
		args     []any
		want     string
	}{
		{"deutsch bekannter schluessel", Deutsch, "nav.mitglieder", nil, "Mitglieder"},
		{"englisch bekannter schluessel", Englisch, "nav.mitglieder", nil, "Members"},
		{"unbekannte sprache faellt auf deutsch zurueck", Sprache("fr"), "nav.mitglieder", nil, "Mitglieder"},
		{"fehlender schluessel in englisch faellt auf deutsch zurueck", Englisch, "nur.deutsch.vorhanden", nil, "nur deutsch vorhanden"},
		{"fehlender schluessel ueberall liefert den schluessel selbst", Deutsch, "voellig.unbekannt", nil, "voellig.unbekannt"},
		{"platzhalter werden eingesetzt", Deutsch, "test.platzhalter", []any{"Anna", 3}, "Anna hat 3 Rückstände"},
	}

	// nur.deutsch.vorhanden und test.platzhalter existieren nur für diesen
	// Test — sie zeigen den Rückfall bzw. die Sprintf-Interpolation, ohne
	// dass ein echter Katalogschlüssel dafür zweckentfremdet werden muss.
	deutsch["nur.deutsch.vorhanden"] = "nur deutsch vorhanden"
	deutsch["test.platzhalter"] = "%s hat %d Rückstände"
	t.Cleanup(func() {
		delete(deutsch, "nur.deutsch.vorhanden")
		delete(deutsch, "test.platzhalter")
	})

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Text(tc.sprache, tc.schluess, tc.args...)
			if got != tc.want {
				t.Errorf("Text(%q, %q, %v) = %q, want %q", tc.sprache, tc.schluess, tc.args, got, tc.want)
			}
		})
	}
}

func TestSpracheGueltig(t *testing.T) {
	if !Deutsch.Gueltig() {
		t.Error("Deutsch soll gueltig sein")
	}
	if !Englisch.Gueltig() {
		t.Error("Englisch soll gueltig sein")
	}
	if Sprache("fr").Gueltig() {
		t.Error("fr soll nicht gueltig sein")
	}
	if Sprache("").Gueltig() {
		t.Error("leere Sprache soll nicht gueltig sein")
	}
}
