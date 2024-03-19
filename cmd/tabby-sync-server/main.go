package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log/slog"
	"net/http"
	"os"
	"tabby-syncd/internal/buildinfo"
	"tabby-syncd/internal/configstore"
	"tabby-syncd/internal/web"
	admin_api "tabby-syncd/internal/web/admin"
	tabby_api "tabby-syncd/internal/web/tabby"
)

func main() {

	// Setup logging
	LogLevel := slog.LevelInfo
	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		LogLevel = slog.LevelDebug
	case "WARN":
		LogLevel = slog.LevelWarn
	case "ERROR":
		LogLevel = slog.LevelError
	default:
		LogLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: LogLevel,
	}))

	logger.Info("Tabby-sync-server starting...",
		slog.String("version", buildinfo.Version),
		slog.String("repository", buildinfo.RepoUrl),
		slog.String("commit_hash", buildinfo.ShaCommit),
		slog.String("build_time", buildinfo.BuildTime),
	)

	// Setup database
	dbFile := os.Getenv("DB_FILE")
	if dbFile == "" {
		dbFile = "configstore.db"
	}
	cstore, err := configstore.Open(dbFile)
	if err != nil {
		logger.Error("Error opening database", slog.String("reason", err.Error()))
		os.Exit(1)
	}

	// Init HTTP server
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.BodyLimit("2M")) // safety feature
	e.Use(middleware.Gzip())
	e.Use(middleware.Recover())

	echo.NotFoundHandler = func(c echo.Context) error {
		// redirect to main page on "404 - Not found" error
		// in case one uses the extra clientinfo paramater in the url in tabby
		return c.Redirect(http.StatusMovedPermanently, "/")
	}

	web.Init(e, logger)

	// Init Tabby webapi
	tabby_api.Init(e, cstore, logger)

	// Enable admin?
	if os.Getenv("ADMIN_KEY") != "" {
		logger.Info("Admin API enabled")
		admin_api.Init(e, cstore, logger)
	} else {
		logger.Info("Admin API disabled", slog.String("reason", "No ADMIN_KEY found"))
	}

	httpAddress := os.Getenv("HTTP_ADDRESS")
	if httpAddress == "" {
		httpAddress = ":8080"
	}
	logger.Info("API server enabled", slog.String("address", httpAddress))
	e.Logger.Fatal(e.Start(httpAddress))
}
