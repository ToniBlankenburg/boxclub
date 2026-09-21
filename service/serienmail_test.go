package service_test

import (
	"slices"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

func TestSerienmailVorbereiten_OhneEmpfaengerBleibtLeer(t *testing.T) {
	for name, empfaenger := range map[string][]string{
		"nil":             nil,
		"leere Liste":     {},
		"nur Leerraum":    {"", "   "},
		"nur leere Werte": {"", ""},
	} {
		t.Run(name, func(t *testing.T) {
			sm := service.SerienmailVorbereiten(empfaenger)
			if sm.Link != "" || sm.Empfaenger != 0 {
				t.Fatalf("SerienmailVorbereiten(%v) = %+v, erwartet Nullwert", empfaenger, sm)
			}
		})
	}
}

func TestSerienmailVorbereiten_EinEmpfaenger(t *testing.T) {
	sm := service.SerienmailVorbereiten([]string{"anna@example.org"})

	if sm.Empfaenger != 1 {
		t.Errorf("Empfaenger = %d, erwartet 1", sm.Empfaenger)
	}
	if want := "mailto:?bcc=anna%40example.org"; sm.Link != want {
		t.Errorf("Link = %q, erwartet %q", sm.Link, want)
	}
}

func TestSerienmailVorbereiten_MehrereEmpfaengerImBcc(t *testing.T) {
	sm := service.SerienmailVorbereiten([]string{"anna@example.org", "ben@example.org"})

	if sm.Empfaenger != 2 {
		t.Errorf("Empfaenger = %d, erwartet 2", sm.Empfaenger)
	}
	if want := "mailto:?bcc=anna%40example.org%2Cben%40example.org"; sm.Link != want {
		t.Errorf("Link = %q, erwartet %q", sm.Link, want)
	}
}

// Doppelte und leere Adressen fallen heraus — wer für zwei Trainingstermine
// angemeldet ist, soll trotzdem nur einmal in der Serienmail stehen.
func TestSerienmailVorbereiten_DoppelteUndLeereFallenHeraus(t *testing.T) {
	sm := service.SerienmailVorbereiten([]string{"anna@example.org", "", "anna@example.org", "  "})

	if sm.Empfaenger != 1 {
		t.Errorf("Empfaenger = %d, erwartet 1", sm.Empfaenger)
	}
	if want := "mailto:?bcc=anna%40example.org"; sm.Link != want {
		t.Errorf("Link = %q, erwartet %q", sm.Link, want)
	}
}

// mitEmailAnlegen legt ein Mitglied mit einer E-Mail-Adresse an und liefert
// dessen ID.
func mitEmailAnlegen(t *testing.T, svc *service.MemberService, vorname, nachname, email string) int64 {
	t.Helper()

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      vorname,
		Nachname:     nachname,
		Email:        email,
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create(%s %s): %v", vorname, nachname, err)
	}

	return id
}

func TestEmailsZuIDs_LeereListeOhneAbfrage(t *testing.T) {
	svc := neuerService(t)

	emails, err := svc.EmailsZuIDs(nil)
	if err != nil || len(emails) != 0 {
		t.Fatalf("EmailsZuIDs(nil) = %v, %v, erwartet leer", emails, err)
	}
}

func TestEmailsZuIDs_LiestNurDieAngegebenen(t *testing.T) {
	svc := neuerService(t)

	anna := mitEmailAnlegen(t, svc, "Anna", "Berger", "anna@example.org")
	_ = mitEmailAnlegen(t, svc, "Ben", "Dietz", "ben@example.org")
	chris := mitEmailAnlegen(t, svc, "Chris", "Ernst", "chris@example.org")

	emails, err := svc.EmailsZuIDs([]int64{anna, chris})
	if err != nil {
		t.Fatalf("EmailsZuIDs: %v", err)
	}

	if erwartet := []string{"anna@example.org", "chris@example.org"}; !slices.Equal(zeichenkettenSortiert(emails), erwartet) {
		t.Errorf("EmailsZuIDs = %v, erwartet %v", emails, erwartet)
	}
}

// Ein Mitglied ohne E-Mail-Adresse liefert eine leere Zeichenkette statt
// keiner Zeile — SerienmailVorbereiten sortiert sie ohnehin selbst aus.
func TestEmailsZuIDs_MitgliedOhneEmailLiefertLeereZeichenkette(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Anna", "Berger")

	emails, err := svc.EmailsZuIDs([]int64{id})
	if err != nil {
		t.Fatalf("EmailsZuIDs: %v", err)
	}
	if erwartet := []string{""}; !slices.Equal(emails, erwartet) {
		t.Errorf("EmailsZuIDs = %v, erwartet %v", emails, erwartet)
	}
}

func zeichenkettenSortiert(werte []string) []string {
	kopie := slices.Clone(werte)
	slices.Sort(kopie)
	return kopie
}
