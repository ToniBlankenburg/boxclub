package app

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
)

// Die Fragmente sind Daten und kein Inhalt: sie dürfen nicht im Zwischenspeicher
// des WebView landen. Ohne diese Zusicherung liefert WebKit ein erneut
// geöffnetes Formular aus dem Cache — mit dem Stand von vor der letzten
// Änderung. Wer es dann speichert, schreibt den alten Stand zurück, und die
// Änderung ist weg, ohne dass irgendwo ein Fehler aufgetaucht wäre.
//
// Das ist die eine Ausnahme von "app/ wird nicht unit-getestet" (CLAUDE.md):
// geprüft wird keine Fachlogik, sondern eine Kopfzeile, die man beim Anlegen
// einer neuen Route genau so leicht wieder vergisst.
func TestFragmenteWerdenNichtZwischengespeichert(t *testing.T) {
	svc, err := service.Open(filepath.Join(t.TempDir(), "boxclub.db"))
	if err != nil {
		t.Fatalf("service.Open: %v", err)
	}
	t.Cleanup(func() { svc.Close() })

	if _, err := svc.Create(service.NeuesMitglied{
		Vorname: "Nina", Nachname: "Klein", BeitragCents: 6500,
		Eintritt: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	a, err := New(svc)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	h := a.Handler()

	// Jede GET-Route, die einen Stand aus der Datenbank zeigt. Eine neue Route
	// gehört hier dazu.
	for _, pfad := range []string{
		"/api/mitglieder",
		"/api/mitglieder/ergebnis",
		"/api/mitglied/formular",
		"/api/mitglied/1/formular",
		"/api/mitglied/1/zeile",
		"/api/mitglied/1/rueckstand",
		"/api/mitglied/1/kuendigung",
		"/api/mitglied/1/wiedereintritt",
		"/api/import",
	} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", pfad, nil))

		if w.Code != http.StatusOK {
			t.Errorf("GET %s = %d, erwartet 200", pfad, w.Code)
			continue
		}
		if steuerung := w.Header().Get("Cache-Control"); steuerung != "no-store" {
			t.Errorf("GET %s: Cache-Control = %q, erwartet \"no-store\" — sonst zeigt das WebView beim "+
				"erneuten Öffnen den Stand von vor der letzten Änderung", pfad, steuerung)
		}
	}
}
