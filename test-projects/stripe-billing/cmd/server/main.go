package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/kabirnarang39/stripe-billing/internal/billing"
	"github.com/kabirnarang39/stripe-billing/internal/cache"
	"github.com/kabirnarang39/stripe-billing/internal/config"
	"github.com/kabirnarang39/stripe-billing/internal/httpx"
	"github.com/kabirnarang39/stripe-billing/internal/limits"
	"github.com/kabirnarang39/stripe-billing/internal/store"
	"github.com/kabirnarang39/stripe-billing/internal/stripeclient"
	"github.com/kabirnarang39/stripe-billing/internal/usage"
	"github.com/kabirnarang39/stripe-billing/internal/webhook"
)

func main() {
	// 1. Load config — fails fast on sk_live_ keys.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Connect to Postgres.
	pg, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer pg.Close()

	// 3. Connect to Redis.
	redisClient, err := cache.New(cfg.RedisURL)
	if err != nil {
		log.Fatalf("cache: %v", err)
	}

	// 4. Initialise Stripe client.
	stripeClient, err := stripeclient.New(cfg.StripeSecretKey)
	if err != nil {
		log.Fatalf("stripeclient: %v", err)
	}

	// 5. Initialise limits cache.
	limitsCache := limits.NewCache(redisClient, pg, cfg.APIQuotaPro)

	// 6. Initialise billing service.
	billingSvc := billing.NewService(pg, stripeClient, redisClient, cfg)

	// 7. Initialise webhook handler.
	webhookHandler := webhook.NewHandler(pg, redisClient, cfg.StripeWebhookSecret)

	// 8. Initialise usage flusher and start goroutine.
	flusher := usage.NewFlusher(pg, redisClient)
	go flusher.Start(ctx)

	// 9. Wire the chi router.
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(httpx.HTTPSOnlyMiddleware(cfg.DevMode))

	// Webhook route — NO JWT middleware; Stripe signs requests.
	r.Post("/webhooks/stripe", webhookHandler.ServeHTTP)

	// Billing API routes — JWT + workspace scope + rate limit.
	r.Route("/api/billing", func(r chi.Router) {
		r.Use(httpx.JWTMiddleware(cfg.JWTSecret))
		r.Use(httpx.WorkspaceScopeMiddleware(pg))
		r.Use(httpx.RateLimitMiddleware(redisClient))
		r.Use(usage.Middleware(redisClient, limitsCache, flusher))
		billing.RegisterRoutes(r, billingSvc, redisClient, pg, cfg)
	})

	// 10. Graceful shutdown.
	srv := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("server: listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("server: shutting down (30s timeout)")
	cancel() // signal background goroutines to stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server: forced shutdown: %v", err)
	}
	log.Println("server: stopped")
}
