package web

import (
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"os"
	"tabby-syncd/internal/buildinfo"
	"tabby-syncd/web"
)

func Init(e *echo.Echo, logger *slog.Logger) {

	if os.Getenv("WEB_DEV") != "" {
		dir, _ := os.Getwd()
		logger.Info("Web development mode enabled", slog.String("webroot", dir+"\\web"))
		e.GET("/", echo.WrapHandler(http.FileServerFS(os.DirFS("web"))))
		e.GET("/favicon.png", echo.WrapHandler(http.FileServerFS(os.DirFS("web"))))
	} else {
		e.GET("/", echo.WrapHandler(http.FileServerFS(web.EmbedFS)))
		e.GET("/favicon.png", echo.WrapHandler(http.FileServerFS(web.EmbedFS)))
	}

	// Health check for docker
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK,
			struct {
				Version   string
				RepoUrl   string
				ShaCommit string
				BuildTime string
			}{
				Version:   buildinfo.Version,
				RepoUrl:   buildinfo.RepoUrl,
				ShaCommit: buildinfo.ShaCommit,
				BuildTime: buildinfo.BuildTime,
			})
	})
}
