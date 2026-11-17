package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"

	"sitegen/internal/config"
	"sitegen/internal/pages"
)

func New(cfg config.Config, store pages.Store) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), logger())
	if len(cfg.Domains()) > 0 {
		r.Use(func(c *gin.Context) {
			c.Header("Strict-Transport-Security", "max-age=31536000")
			c.Next()
		})
	}

	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.Static("/static", cfg.StaticDir)

	for _, p := range store.Pages() {
		html := p.HTML
		r.GET(p.Path, func(c *gin.Context) {
			c.Data(http.StatusOK, "text/html; charset=utf-8", html)
		})
	}

	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "404 page not found")
	})

	return r
}

func logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start),
		)
	}
}

func RunTLS(ctx context.Context, h http.Handler, cfg config.Config) error {
	if err := os.MkdirAll(cfg.TLSCacheDir, 0o700); err != nil {
		return fmt.Errorf("create tls cache dir: %w", err)
	}

	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.Domains()...),
		Cache:      autocert.DirCache(cfg.TLSCacheDir),
		Email:      cfg.TLSEmail,
	}

	acme := m.HTTPHandler(nil)
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/", acme)

	httpsSrv := &http.Server{
		Addr:              cfg.HTTPSAddr,
		Handler:           h,
		TLSConfig:         m.TLSConfig(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	httpSrv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 2)
	go func() {
		errCh <- httpsSrv.ListenAndServeTLS("", "")
	}()
	go func() {
		errCh <- httpSrv.ListenAndServe()
	}()

	var runErr error
	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			runErr = err
		}
	case <-ctx.Done():
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, s := range []*http.Server{httpsSrv, httpSrv} {
		if err := s.Shutdown(shutCtx); err != nil && !errors.Is(err, http.ErrServerClosed) && runErr == nil {
			runErr = err
		}
	}
	return runErr
}

func Run(ctx context.Context, h http.Handler, addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
