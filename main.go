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
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/ToniBlankenburg/boxclub/app"
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

	anwendung, err := app.New(svc)
	if err != nil {
		log.Fatalf("Anwendung initialisieren: %v", err)
	}

	err = wails.Run(&options.App{
		Title:  "Boxclub Mitgliederverwaltung",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
			// Der Assetserver reicht alle Nicht-GET-Requests und alle GET-Requests,
			// die keine statische Datei treffen, an diesen Handler durch — darüber
			// spricht htmx mit dem Go-Backend.
			Handler: anwendung.Handler(),
		},
		BackgroundColour: &options.RGBA{R: 250, G: 250, B: 250, A: 1},
		// Der Datei-Dialog braucht den Wails-Kontext, den es erst ab hier gibt.
		// Er ist die eine Stelle, an der die App das Betriebssystem braucht:
		// Dokumente liegen als Blob in der Datenbank (ADR-0007), und ohne einen
		// Weg heraus käme niemand mehr an seinen Vertrag. Ein Download über den
		// Assetserver ist kein solcher Weg — das WebView von Wails behandelt
		// keine.
		OnStartup: func(ctx context.Context) {
			anwendung.SpeicherzielSetzen(func(vorschlag string) (string, error) {
				return runtime.SaveFileDialog(ctx, runtime.SaveDialogOptions{
					Title:                "Dokument speichern",
					DefaultFilename:      vorschlag,
					CanCreateDirectories: true,
					Filters: []runtime.FileFilter{
						{DisplayName: "PDF-Dateien (*.pdf)", Pattern: "*.pdf"},
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
