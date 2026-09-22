package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ToniBlankenburg/boxclub/i18n"
	"github.com/ToniBlankenburg/boxclub/service"
)

func TestSpracheAendern(t *testing.T) {
	svc, err := service.Open(filepath.Join(t.TempDir(), "boxclub.db"))
	if err != nil {
		t.Fatalf("service.Open: %v", err)
	}
	t.Cleanup(func() { svc.Close() })

	einstellungenPfad := filepath.Join(t.TempDir(), "einstellungen.json")
	a, err := New(svc, i18n.Deutsch, einstellungenPfad)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	h := a.Handler()

	w := httptest.NewRecorder()
	h.ServeHTTP(w, sprachAnfrage(i18n.Englisch))

	if w.Code != 200 {
		t.Fatalf("Status = %d, Body: %s", w.Code, w.Body)
	}
	if a.Sprache() != i18n.Englisch {
		t.Errorf("a.Sprache() = %q, want %q", a.Sprache(), i18n.Englisch)
	}
	// "Members" statt "Mitglieder" in der mitgeschickten Navigation zeigt, dass
	// die Antwort selbst schon in der neuen Sprache gerendert wurde — kein
	// zweiter Aufruf nötig, um die Umschaltung zu sehen.
	if !strings.Contains(w.Body.String(), "Members") {
		t.Errorf("Antwort enthält nicht die englische Navigation:\n%s", w.Body.String())
	}

	// Die Einstellungsdatei überlebt einen Neustart: ein frisches Laden
	// derselben Datei liefert dieselbe Sprache.
	geladen, err := i18n.EinstellungenLaden(einstellungenPfad)
	if err != nil {
		t.Fatalf("EinstellungenLaden: %v", err)
	}
	if geladen.Sprache != i18n.Englisch {
		t.Errorf("gespeicherte Sprache = %q, want %q", geladen.Sprache, i18n.Englisch)
	}
}

func TestSpracheAendern_ungueltigeSpracheBleibtUnveraendert(t *testing.T) {
	svc, err := service.Open(filepath.Join(t.TempDir(), "boxclub.db"))
	if err != nil {
		t.Fatalf("service.Open: %v", err)
	}
	t.Cleanup(func() { svc.Close() })

	a, err := New(svc, i18n.Deutsch, filepath.Join(t.TempDir(), "einstellungen.json"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	h := a.Handler()

	w := httptest.NewRecorder()
	h.ServeHTTP(w, sprachAnfrage(i18n.Sprache("fr")))

	if w.Code == 200 {
		t.Fatalf("Status = 200, erwartet einen Fehlerstatus für eine unbekannte Sprache")
	}
	if a.Sprache() != i18n.Deutsch {
		t.Errorf("a.Sprache() = %q, eine unbekannte Sprache darf die aktuelle nicht ersetzen", a.Sprache())
	}
}

func sprachAnfrage(sprache i18n.Sprache) *http.Request {
	form := url.Values{"sprache": {string(sprache)}}
	req := httptest.NewRequest("POST", "/api/einstellungen/sprache", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}
