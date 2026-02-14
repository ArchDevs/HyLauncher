package main

import (
	"HyLauncher/internal/app"
	"HyLauncher/internal/env"
	"HyLauncher/pkg/logger"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := logger.Init(env.GetDefaultAppDir()+"/logs", logger.INFO, true); err != nil {
		println("Failed to init logger:", err.Error())
	}
	defer logger.Close()

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
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  true,
				HideTitleBar:               false,
				FullSizeContent:            true,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
