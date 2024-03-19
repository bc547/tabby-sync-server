package admin

import (
	"context"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"tabby-syncd/internal/configstore"
	"tabby-syncd/web"
)

type ApiSyncTokensResponse struct {
	Msg        string   `json:"msg"`
	SyncTokens []string `json:"synctokens"`
}

type ApiSyncTokenResponse struct {
	Msg       string `json:"msg"`
	SyncToken string `json:"synctoken"`
}

func Init(e *echo.Echo, cstore *configstore.ConfigStore, logger *slog.Logger) {

	e.Pre(middleware.Rewrite(map[string]string{
		"/admin":        "/admin/",
		"/admin/page1*": "/admin/",
		"/admin/page2*": "/admin/",
	}))

	admin := e.Group("/admin")

	admin.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:        true,
		LogRemoteIP:      true,
		LogUserAgent:     true,
		LogURI:           true,
		LogError:         true,
		LogLatency:       true,
		LogResponseSize:  true,
		LogContentLength: true,
		HandleError:      true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			BytesIn, _ := strconv.ParseInt(v.ContentLength, 10, 64)
			LogAttrs := []slog.Attr{
				slog.String("remote_ip", v.RemoteIP),
				slog.Int("status", v.Status),
				slog.String("method", c.Request().Method),
				slog.String("uri", v.URI),
				slog.String("useragent", v.UserAgent),
				slog.String("latency", v.Latency.String()),
				slog.Group("bytes",
					slog.Int64("in", BytesIn),
					slog.Int64("out", v.ResponseSize),
				),
			}
			if v.Error != nil {
				LogAttrs = append(LogAttrs, slog.String("error", v.Error.Error()))
			}
			logger.LogAttrs(context.Background(), slog.LevelDebug, "Admin API request", LogAttrs...)
			return nil
		},
	}))

	// Serve static admin files (live if WEB_DEV, embedded FS otherwise)
	if os.Getenv("WEB_DEV") != "" {
		admin.GET("/*", echo.WrapHandler(http.FileServerFS(os.DirFS("web"))))
	} else {
		admin.GET("/*", echo.WrapHandler(http.FileServerFS(web.EmbedFS)))
	}

	// Admin API handlers
	adminAPI := admin.Group("/api/1")

	// Check authorization header and bearer key
	adminAPI.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == os.Getenv("ADMIN_KEY"), nil
	}))

	// Return list of synctokens
	adminAPI.GET("/synctokens", func(c echo.Context) error {
		SyncTokens, err := cstore.GetSyncTokens()
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, &ApiSyncTokensResponse{"List of all synctokens", SyncTokens})
	})

	// Create new SyncToken
	adminAPI.POST("/synctokens", func(c echo.Context) error {
		SyncToken := uuid.New().String()

		err := cstore.CreateSyncToken(SyncToken)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, &ApiSyncTokenResponse{"synctoken created", SyncToken})
	})

	// Delete existing synctoken
	adminAPI.DELETE("/synctokens/:id", func(c echo.Context) error {
		SyncToken := c.Param("id")

		err := cstore.DeleteSyncToken(SyncToken)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, &ApiSyncTokenResponse{"synctoken deleted", SyncToken})
	})

}
