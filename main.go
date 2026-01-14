package main

import (
	"HyLauncher/app"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	application := app.NewApp()

	err := wails.Run(&options.App{
		Title:         "HyLauncher",
		Width:         1280,
		Height:        720,
		DisableResize: true,
		Frameless:     true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 255},
		OnStartup:        application.Startup,
		CSSDragProperty:  "--wails-draggable",
		Bind: []interface{}{
			application,
		},
		Windows: &windows.Options{
			IsZoomControlEnabled: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
