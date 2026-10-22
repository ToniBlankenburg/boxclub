package service_test

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// sicherungMitDaten füllt eine Datenbank mit einem Mitglied, einem Vertrag und
// Vereinsdaten — den drei Arten von Inhalt, die eine Sicherung vollständig
// mitnehmen muss — und liefert Mitglieds- und Mitgliedschaftsnummer.
func sicherungMitDaten(t *testing.T, svc *service.MemberService) (mitgliedID, mitgliedschaftID int64) {
	t.Helper()

	mitgliedID = moneyMoneyMitglied(t, svc, "Anna", "Berger", "DE02120300000000202051")
	mitgliedschaftID = laufendeMitgliedschaftID(t, svc, mitgliedID)

	if err := svc.VertragAblegen(mitgliedschaftID, vertrag("vertrag-berger.pdf", "unterschrieben")); err != nil {
		t.Fatalf("VertragAblegen: %v", err)
	}
	if err := svc.SetVereinsdaten(service.Vereinsdaten{Name: "Boxclub Test"}); err != nil {
		t.Fatalf("SetVereinsdaten: %v", err)
	}

	return mitgliedID, mitgliedschaftID
}

func TestSicherungErstellen_EnthaeltAlleDaten(t *testing.T) {
	svc := neuerService(t)
	mitgliedID, mitgliedschaftID := sicherungMitDaten(t, svc)

	ziel := filepath.Join(t.TempDir(), "sicherung.db")
	if err := svc.SicherungErstellen(ziel); err != nil {
		t.Fatalf("SicherungErstellen: %v", err)
	}

	kopie, err := service.Open(ziel)
	if err != nil {
		t.Fatalf("Open(Sicherung): %v", err)
	}
	defer kopie.Close()

	m, err := kopie.Get(mitgliedID)
	if err != nil {
		t.Fatalf("Get in der Sicherung: %v", err)
	}
	if m.Nachname != "Berger" {
		t.Errorf("Nachname = %q, erwartet %q", m.Nachname, "Berger")
	}

	vertragInhalt, err := kopie.VertragInhalt(mitgliedschaftID)
	if err != nil {
		t.Fatalf("VertragInhalt in der Sicherung: %v", err)
	}
	if string(vertragInhalt.Inhalt) != string(pdf("unterschrieben")) {
		t.Errorf("Vertrag in der Sicherung weicht ab: %q", vertragInhalt.Inhalt)
	}

	verein, err := kopie.GetVereinsdaten()
	if err != nil {
		t.Fatalf("GetVereinsdaten in der Sicherung: %v", err)
	}
	if verein.Name != "Boxclub Test" {
		t.Errorf("Vereinsname = %q, erwartet %q", verein.Name, "Boxclub Test")
	}
}

func TestSicherungErstellen_UeberschreibtVorhandeneDatei(t *testing.T) {
	svc := neuerService(t)
	sicherungMitDaten(t, svc)

	ziel := filepath.Join(t.TempDir(), "sicherung.db")
	if err := os.WriteFile(ziel, []byte("alter inhalt"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := svc.SicherungErstellen(ziel); err != nil {
		t.Fatalf("SicherungErstellen: %v", err)
	}

	kopie, err := service.Open(ziel)
	if err != nil {
		t.Fatalf("Open(Sicherung): %v", err)
	}
	kopie.Close()
}

// oeffnenImTest legt eine Datenbank an einem bekannten Pfad an, den die
// Wiederherstellung braucht, um die Datei danach neu zu öffnen.
func oeffnenImTest(t *testing.T) (*service.MemberService, string) {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "boxclub.db")
	svc, err := service.Open(pfad)
	if err != nil {
		t.Fatalf("service.Open: %v", err)
	}
	t.Cleanup(func() { svc.Close() })

	return svc, pfad
}

func TestWiederherstellen_ErsetztDenBestandDurchDieSicherung(t *testing.T) {
	svc, pfad := oeffnenImTest(t)
	annaID := moneyMoneyMitglied(t, svc, "Anna", "Berger", "DE02120300000000202051")

	sicherung := filepath.Join(t.TempDir(), "sicherung.db")
	if err := svc.SicherungErstellen(sicherung); err != nil {
		t.Fatalf("SicherungErstellen: %v", err)
	}

	brunoID := moneyMoneyMitglied(t, svc, "Bruno", "Cordes", "DE02120300000000202052")

	if err := svc.Wiederherstellen(sicherung); err != nil {
		t.Fatalf("Wiederherstellen: %v", err)
	}

	neu, err := service.Open(pfad)
	if err != nil {
		t.Fatalf("Open nach Wiederherstellen: %v", err)
	}
	defer neu.Close()

	if _, err := neu.Get(annaID); err != nil {
		t.Errorf("Anna fehlt nach Wiederherstellen: %v", err)
	}
	if _, err := neu.Get(brunoID); err == nil {
		t.Errorf("Bruno steht noch da, obwohl er erst nach der Sicherung angelegt wurde")
	}
}

func TestWiederherstellen_LegtKopieDesBisherigenStandsAn(t *testing.T) {
	svc, pfad := oeffnenImTest(t)
	moneyMoneyMitglied(t, svc, "Anna", "Berger", "DE02120300000000202051")

	sicherung := filepath.Join(t.TempDir(), "sicherung.db")
	if err := svc.SicherungErstellen(sicherung); err != nil {
		t.Fatalf("SicherungErstellen: %v", err)
	}

	brunoID := moneyMoneyMitglied(t, svc, "Bruno", "Cordes", "DE02120300000000202052")

	if err := svc.Wiederherstellen(sicherung); err != nil {
		t.Fatalf("Wiederherstellen: %v", err)
	}

	vorher, err := service.Open(pfad + service.EndungVorWiederherstellung)
	if err != nil {
		t.Fatalf("Open(Kopie): %v", err)
	}
	defer vorher.Close()

	if _, err := vorher.Get(brunoID); err != nil {
		t.Errorf("die Kopie enthält den bisherigen Stand nicht, Bruno fehlt: %v", err)
	}
}

func TestWiederherstellen_LehntKeineSicherungAb(t *testing.T) {
	andere := filepath.Join(t.TempDir(), "andere.db")
	fremd, err := sql.Open("sqlite", andere)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fremd.Exec(`CREATE TABLE etwas (id INTEGER)`); err != nil {
		t.Fatal(err)
	}
	fremd.Close()

	kaputt := filepath.Join(t.TempDir(), "kaputt.db")
	if err := os.WriteFile(kaputt, []byte("das ist keine datenbank"), 0o600); err != nil {
		t.Fatal(err)
	}

	faelle := map[string]string{
		"fremde SQLite-Datei": andere,
		"Textdatei":           kaputt,
		"fehlende Datei":      filepath.Join(t.TempDir(), "gibtsnicht.db"),
	}

	for name, quelle := range faelle {
		t.Run(name, func(t *testing.T) {
			svc, _ := oeffnenImTest(t)
			id := moneyMoneyMitglied(t, svc, "Anna", "Berger", "DE02120300000000202051")

			err := svc.Wiederherstellen(quelle)
			if !errors.Is(err, service.ErrKeineSicherung) {
				t.Fatalf("Wiederherstellen = %v, erwartet ErrKeineSicherung", err)
			}

			// Eine abgelehnte Datei darf den laufenden Betrieb nicht antasten.
			if _, err := svc.Get(id); err != nil {
				t.Errorf("Service nach Ablehnung nicht mehr nutzbar: %v", err)
			}
		})
	}
}
