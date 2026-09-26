package main

import (
	"context"
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/ToniBlankenburg/boxclub/app"
	"github.com/ToniBlankenburg/boxclub/i18n"
	"github.com/ToniBlankenburg/boxclub/service"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	pfad, err := datenbankPfad()
	if err != nil {
		log.Fatalf("Datenbankpfad bestimmen: %v", err)
	}

	svc, err := service.Open(pfad)
	if err != nil {
		log.Fatalf("Datenbank %s öffnen: %v", pfad, err)
	}
	defer svc.Close()

	einstellungenPfad, err := einstellungenPfad()
	if err != nil {
		log.Fatalf("Einstellungspfad bestimmen: %v", err)
	}

	// Eine unlesbare Einstellungsdatei ist anders als ein Datenbankfehler
	// nicht fatal: sie enthält nur die Anzeigesprache, keine Mitgliederdaten
	// — die App startet dann eben deutsch (siehe i18n.EinstellungenLaden).
	einstellungen, err := i18n.EinstellungenLaden(einstellungenPfad)
	if err != nil {
		log.Printf("Einstellungen %s lesen, verwende Deutsch: %v", einstellungenPfad, err)
		einstellungen = i18n.StandardEinstellungen
	}

	anwendung, err := app.New(svc, einstellungen.Sprache, einstellungenPfad)
	if err != nil {
		log.Fatalf("Anwendung initialisieren: %v", err)
	}

	err = wails.Run(&options.App{
		Title: "Mitgliederverwaltung",
		// 1360 statt zuvor 1024: die Mitgliederliste hat acht Spalten fester
		// Breite (mitglieder_liste.html) und braucht das, um ohne horizontales
		// Scrollen ins Standardfenster zu passen — bei 1024px reichte selbst
		// die unberührte Breite nicht, Name lief auf 0px zusammen.
		Width:  1360,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
			// Der Assetserver reicht alle Nicht-GET-Requests und alle GET-Requests,
			// die keine statische Datei treffen, an diesen Handler durch — darüber
			// spricht htmx mit dem Go-Backend.
			Handler: anwendung.Handler(),
		},
		BackgroundColour: &options.RGBA{R: 250, G: 250, B: 250, A: 1},
		// Ohne dieses Feld bleibt Mac nil, und Wails deaktiviert dann den
		// grünen Zoom-Button am Fenster explizit (siehe darwin/WailsContext.m:
		// !zoomable && resizable → Button wird disabled) — natives Vollbild
		// wäre sonst aus, nicht nur nicht extra an.
		Mac: &mac.Options{},
		// Der Datei-Dialog braucht den Wails-Kontext, den es erst ab hier gibt.
		// Er ist die eine Stelle, an der die App das Betriebssystem braucht:
		// Dokumente liegen als Blob in der Datenbank (ADR-0007), und ohne einen
		// Weg heraus käme niemand mehr an seinen Vertrag. Ein Download über den
		// Assetserver ist kein solcher Weg — das WebView von Wails behandelt
		// keine.
		OnStartup: func(ctx context.Context) {
			anwendung.SpeicherzielSetzen(func(vorschlag, filterBeschriftung, filterMuster string) (string, error) {
				return runtime.SaveFileDialog(ctx, runtime.SaveDialogOptions{
					Title:                "Datei speichern",
					DefaultFilename:      vorschlag,
					CanCreateDirectories: true,
					Filters: []runtime.FileFilter{
						{DisplayName: filterBeschriftung, Pattern: filterMuster},
					},
				})
			})
		},
	})
	if err != nil {
		log.Fatalf("Wails starten: %v", err)
	}
}

// datenbankPfad liefert den Ort der SQLite-Datei. Standard ist ein eigenes
// Verzeichnis im Benutzer-Konfigurationsordner (macOS:
// ~/Library/Application Support/Boxclub); BOXCLUB_DB überschreibt ihn, was für
// Entwicklung und manuelle Tests praktisch ist.
func datenbankPfad() (string, error) {
	if pfad := os.Getenv("BOXCLUB_DB"); pfad != "" {
		return pfad, nil
	}

	basis, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	verzeichnis := filepath.Join(basis, "Boxclub")
	// 0700: die Datei enthält personenbezogene Daten und geht niemanden sonst an.
	if err := os.MkdirAll(verzeichnis, 0o700); err != nil {
		return "", err
	}

	return filepath.Join(verzeichnis, "boxclub.db"), nil
}

// einstellungenPfad liefert den Ort der Datei mit der Anzeigesprache
// (ADR-0017). Sie liegt bewusst neben, nicht in der Datenbank: sie hat nichts
// mit den Mitgliederdaten zu tun und soll ein Löschen der Dev-Datenbank
// überstehen. BOXCLUB_EINSTELLUNGEN überschreibt den Pfad, analog zu
// BOXCLUB_DB.
func einstellungenPfad() (string, error) {
	if pfad := os.Getenv("BOXCLUB_EINSTELLUNGEN"); pfad != "" {
		return pfad, nil
	}

	basis, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(basis, "Boxclub", "einstellungen.json"), nil
}
