package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/urmzd/zoro/internal/app"
	"github.com/urmzd/zoro/internal/config"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:assets
var embeddedAssets embed.FS

func setupLogging() *os.File {
	logDir, err := os.UserConfigDir()
	if err != nil {
		logDir = os.TempDir()
	}
	logDir = filepath.Join(logDir, "zoro")
	os.MkdirAll(logDir, 0o755)

	logPath := filepath.Join(logDir, "zoro.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Printf("failed to open log file %s: %v", logPath, err)
		return nil
	}

	log.SetOutput(f)
	log.Printf("=== Zoro desktop starting (log: %s) ===", logPath)
	return f
}

func buildMenu(appCtx *context.Context) *menu.Menu {
	appMenu := menu.NewMenu()

	// App menu (macOS "Zoro" menu)
	appMenu.Append(menu.AppMenu())

	// View menu
	viewMenu := appMenu.AddSubmenu("View")
	viewMenu.AddText("New Chat", keys.CmdOrCtrl("n"), func(_ *menu.CallbackData) {
		wailsRuntime.WindowExecJS(*appCtx, `window.location.hash=''; window.location.pathname='/'`)
	})
	viewMenu.AddSeparator()
	viewMenu.AddText("Logs", keys.CmdOrCtrl("l"), func(_ *menu.CallbackData) {
		wailsRuntime.WindowExecJS(*appCtx, `window.location.pathname='/logs'`)
	})
	viewMenu.AddText("Help", keys.CmdOrCtrl("?"), func(_ *menu.CallbackData) {
		wailsRuntime.WindowExecJS(*appCtx, `window.location.pathname='/help'`)
	})
	viewMenu.AddSeparator()
	viewMenu.AddText("Knowledge Graph", keys.CmdOrCtrl("k"), func(_ *menu.CallbackData) {
		wailsRuntime.WindowExecJS(*appCtx, `window.location.pathname='/knowledge'`)
	})

	// Edit menu (standard copy/paste/undo)
	appMenu.Append(menu.EditMenu())

	return appMenu
}

func main() {
	logFile := setupLogging()
	if logFile != nil {
		defer logFile.Close()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Load()

	echoInstance, cleanup, err := app.Wire(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to wire app: %v", err)
	}
	defer cleanup()

	assets, err := fs.Sub(embeddedAssets, "assets")
	if err != nil {
		log.Fatalf("failed to load embedded assets: %v", err)
	}

	// Wails context is set in OnStartup; menu callbacks need a pointer to it
	var wailsCtx context.Context
	appMenu := buildMenu(&wailsCtx)

	log.Println("starting Wails app")
	err = wails.Run(&options.App{
		Title:     "Zoro",
		Width:     1280,
		Height:    800,
		MinWidth:  800,
		MinHeight: 600,
		Menu:      appMenu,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: echoInstance,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
			About: &mac.AboutInfo{
				Title:   "Zoro",
				Message: "Privacy-first research agent with a personal knowledge graph.\n\nv0.1.0",
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
		},
		OnStartup: func(ctx context.Context) {
			wailsCtx = ctx
		},
		OnShutdown: func(ctx context.Context) {
			log.Println("shutting down")
			cancel()
		},
	})
	if err != nil {
		log.Fatalf("wails: %v", err)
	}
}
