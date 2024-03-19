package tabby

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"tabby-syncd/internal/configstore"
	"time"
)

func Init(e *echo.Echo, cstore *configstore.ConfigStore, logger *slog.Logger) {
	// Allow passing extra clientinfo from Tabby
	// e.g. instead of https://tabby-sync.example.net as sync host, you can add extra text in the path
	// -> https://tabby-sync.example.net/mylaptop
	// this extra info will get logged as clientinfo="mylaptop"
	e.Pre(middleware.Rewrite(map[string]string{
		"/*/api/1/user":      "/api/1/user?clientinfo=$1",
		"/*/api/1/configs":   "/api/1/configs?clientinfo=$1",
		"/*/api/1/configs/*": "/api/1/configs/$2?clientinfo=$1",
	}))

	tabbyAPI := e.Group("/api/1")

	logger.Info("Tabby config sync API enabled")

	tabbyAPI.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
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
			SyncTokenFragment := c.Get("synctokenfragment").(string)

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
				slog.String("synctoken.fragment", SyncTokenFragment),
			}
			ClientInfo := c.Get("clientinfo").(string)
			if ClientInfo != "" {
				LogAttrs = append(LogAttrs, slog.String("clientinfo", ClientInfo))
			}
			if v.Error != nil {
				LogAttrs = append(LogAttrs, slog.String("error", v.Error.Error()))
			}
			logger.LogAttrs(context.Background(), slog.LevelDebug, "Tabby API request", LogAttrs...)
			return nil
		},
	}))

	// Check authorization header and bearer key
	// Extract clientinfo (and limit max size)
	tabbyAPI.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		c.Set("synctoken", key) // store synctoken in context for use in handlers
		if len(key) > 2 {
			SyncTokenFragment := key[strings.LastIndex(key, "-")+1:] // last word of synctoken (for logging purposes)
			c.Set("synctokenfragment", SyncTokenFragment)            // store synctokenfragment use logs
		}
		c.Set("clientinfo", fmt.Sprintf("%.32s", c.QueryParam("clientinfo"))) // limit clientinfo length to 32 characters
		return cstore.IsValidSyncToken(key)
	}))

	// Return 200 if user exists (only used by Tabby to check if Authorization is ok)
	tabbyAPI.GET("/user", func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	})

	// Return array of all configs (belonging to the synctoken)
	tabbyAPI.GET("/configs", func(c echo.Context) error {
		SyncToken := c.Get("synctoken").(string)

		cfgs, err := cstore.LoadConfigs(SyncToken)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, cfgs)
	})

	// Create new config object (belonging to the user)
	tabbyAPI.POST("/configs", func(c echo.Context) error {
		SyncToken := c.Get("synctoken").(string)

		// create default empty config
		cfg := &configstore.Config{
			Id:         uuid.New().String(),
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
		}

		// Overwrite defaults with whatever tabby sent us (typically the name field)
		if err := c.Bind(cfg); err != nil {
			return err
		}

		// store in database
		err := cstore.SaveConfig(SyncToken, cfg)
		if err != nil {
			return err
		}

		LogAttrs := []slog.Attr{
			slog.String("config.name", cfg.Name),
			slog.String("config.id", cfg.Id),
			slog.String("synctoken.fragment", c.Get("synctokenfragment").(string)),
		}
		ClientInfo := c.Get("clientinfo").(string)
		if ClientInfo != "" {
			LogAttrs = append(LogAttrs, slog.String("clientinfo", ClientInfo))
		}
		logger.LogAttrs(context.Background(), slog.LevelInfo, "Config created", LogAttrs...)

		return c.JSON(http.StatusOK, cfg)
	})

	// Return a specific config (belonging to a user)
	tabbyAPI.GET("/configs/:id", func(c echo.Context) error {
		SyncToken := c.Get("synctoken").(string)
		cfgid := c.Param("id")

		cfg, err := cstore.LoadConfig(SyncToken, cfgid)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, cfg)
	})

	// Return a specific config (belonging to a user)
	tabbyAPI.PATCH("/configs/:id", func(c echo.Context) error {
		SyncToken := c.Get("synctoken").(string)
		cfgid := c.Param("id")

		// Load config
		cfg, err := cstore.LoadConfig(SyncToken, cfgid)
		if err != nil {
			return err
		}

		h1 := sha256.Sum256([]byte(cfg.Content))

		// Overwrite config with whatever tabby sent us (typically the content field)
		if err = c.Bind(cfg); err != nil {
			return err
		}

		h2 := sha256.Sum256([]byte(cfg.Content))

		// store in database if the config is different
		if h1 != h2 {
			cfg.ModifiedAt = time.Now()
			err = cstore.SaveConfig(SyncToken, cfg)
			if err != nil {
				return err
			}
			LogAttrs := []slog.Attr{
				slog.String("config.name", cfg.Name),
				slog.String("config.id", cfgid),
				slog.String("synctoken.fragment", c.Get("synctokenfragment").(string)),
			}
			ClientInfo := c.Get("clientinfo").(string)
			if ClientInfo != "" {
				LogAttrs = append(LogAttrs, slog.String("clientinfo", ClientInfo))
			}
			logger.LogAttrs(context.Background(), slog.LevelInfo, "Config modified", LogAttrs...)
		}

		return c.JSON(http.StatusOK, cfg)
	})

	// Return a specific config (belonging to a user)
	tabbyAPI.DELETE("/configs/:id", func(c echo.Context) error {
		SyncToken := c.Get("synctoken").(string)
		cfgid := c.Param("id")

		// Load config
		cfg, err := cstore.LoadConfig(SyncToken, cfgid)
		if err != nil {
			return err
		}

		err = cstore.DeleteConfig(SyncToken, cfgid)
		if err != nil {
			return err
		}

		LogAttrs := []slog.Attr{
			slog.String("config.name", cfg.Name),
			slog.String("config.id", cfgid),
			slog.String("synctoken.fragment", c.Get("synctokenfragment").(string)),
		}
		ClientInfo := c.Get("clientinfo").(string)
		if ClientInfo != "" {
			LogAttrs = append(LogAttrs, slog.String("clientinfo", ClientInfo))
		}
		logger.LogAttrs(context.Background(), slog.LevelInfo, "Config removed", LogAttrs...)

		return c.NoContent(http.StatusOK)
	})

}
