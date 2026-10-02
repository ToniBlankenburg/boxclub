package app

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ToniBlankenburg/boxclub/i18n"
	"github.com/ToniBlankenburg/boxclub/service"
)

// Die dritte Ausnahme von "app/ wird nicht unit-getestet" (CLAUDE.md):
// geprüft wird wie bei TestFragmenteWerdenNichtZwischengespeichert keine
// Fachlogik, sondern eine Verdrahtung, die man beim Anlegen einer neuen Route
// genau so leicht wieder vergisst.
//
// listeMitFilterRendern (volle Seite) holte Geschlechtswerte und Termine
// früher selbst, mitgliederErgebnis (nur das ausgetauschte Fragment) ging
// direkt über listeDatenLesen und damit ohne die beiden Listen — ein Rest aus
// der Zeit vor ADR-0020, als die zugehörigen Auswahlfelder noch in der
// stehenbleibenden Filterleiste standen. Seit ADR-0020 stehen sie an den
// Zahnrädern der Kopfzeile und damit selbst im ausgetauschten Fragment: ohne
// die Listen kommt das Auswahlfeld nach der ersten Filteränderung nur noch
// mit dem Platzhalter zurück, ein gesetzter Geschlecht- oder Termin-Filter
// geht dabei unbemerkt verloren, und keine weitere Auswahl lässt sich mehr
// treffen.
func TestMitgliederErgebnis_TraegtGeschlechtsUndTerminoptionen(t *testing.T) {
	svc, err := service.Open(filepath.Join(t.TempDir(), "boxclub.db"))
	if err != nil {
		t.Fatalf("service.Open: %v", err)
	}
	t.Cleanup(func() { svc.Close() })

	if _, err := svc.Create(service.NeuesMitglied{
		Vorname: "Anna", Nachname: "Frau", Geschlecht: "Frau", BeitragCents: 6500,
		Eintritt: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := svc.CreateTrainingstermin(service.Trainingsterminangabe{
		Wochentag: service.Samstag, Beginn: "10:30", Ende: "12:00", Bezeichnung: "Anfänger",
	}); err != nil {
		t.Fatalf("CreateTrainingstermin: %v", err)
	}

	a, err := New(svc, i18n.Deutsch, filepath.Join(t.TempDir(), "einstellungen.json"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	h := a.Handler()

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/api/mitglieder/ergebnis", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/mitglieder/ergebnis = %d, erwartet 200", w.Code)
	}

	koerper := w.Body.String()
	if !strings.Contains(koerper, `value="Frau"`) {
		t.Errorf("Antwort enthält keine Geschlecht-Option \"Frau\" — Geschlechtswerte fehlen im Fragment")
	}
	if !strings.Contains(koerper, "Anfänger") {
		t.Errorf("Antwort enthält keine Termin-Option \"Anfänger\" — Termine fehlen im Fragment")
	}
}
