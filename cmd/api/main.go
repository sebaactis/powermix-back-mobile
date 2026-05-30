package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sebaactis/powermix-back-mobile/internal/clients/coffeeji"
	"github.com/sebaactis/powermix-back-mobile/internal/platform/logger"
	"github.com/sebaactis/powermix-back-mobile/internal/clients/mailer"
	"github.com/sebaactis/powermix-back-mobile/internal/clients/mercadopago"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/prode"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/proof"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/token"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/user"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/voucher"
	"github.com/sebaactis/powermix-back-mobile/internal/jobs"
	"github.com/sebaactis/powermix-back-mobile/internal/middlewares"
	"github.com/sebaactis/powermix-back-mobile/internal/platform/config"
	"github.com/sebaactis/powermix-back-mobile/internal/platform/database"
	"github.com/sebaactis/powermix-back-mobile/internal/routes"
	"github.com/sebaactis/powermix-back-mobile/internal/security/auth"
	jwtx "github.com/sebaactis/powermix-back-mobile/internal/security/jwt"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
)

func main() {
	// Configuramos el logger JSON para Render.com con inyección automática de request_id
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(logger.NewContextHandler(baseHandler)))

	if err := godotenv.Load(); err != nil {
		slog.Info("No se encontró .env (ok en prod)", "error", err)
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuración inválida", "error", err)
		os.Exit(1)
	}

	db, err := database.Open(cfg)
	if err != nil {
		slog.Error("Error al conectar la base de datos", "error", err)
		os.Exit(1)
	}

	if err := database.Migrate(db); err != nil {
		slog.Error("Error al migrar las entidades en la base de datos", "error", err)
		os.Exit(1)
	}

	// Utilidades
	jwt, err := jwtx.NewJWT()
	if err != nil {
		slog.Error("error inicializando JWT", "error", err)
		os.Exit(1)
	}
	validator := validations.NewValidator()
	rateLimiter := middlewares.NewRateLimiter(10, 2*time.Minute)

	// Mailer
	mailerClient := mailer.NewResendMailer(cfg.ResendKey, "safeimportsarg@gmail.com", "Powermix")

	// MercadoPago
	mpClient := mercadopago.NewClient(cfg.MercagoPagoToken)

	// Coffeeji
	coffejiClient := coffeeji.NewClient(cfg.CoffejiKey, cfg.CoffejiSecret)

	// Token DI
	tokenRepository := token.NewRepository(db)
	tokenService := token.NewService(tokenRepository, validator, cfg.HashToken)
	tokenHandler := token.NewHTTPHandler(tokenService)

	// Users DI
	userRepository := user.NewRepository(db)
	userService := user.NewService(userRepository, tokenService, validator, mailerClient)
	userHandler := user.NewHTTPHandler(userService, jwt)

	// Voucher DI
	voucherRepository := voucher.NewRepository(db)
	voucherService := voucher.NewService(voucherRepository, userRepository, mailerClient, coffejiClient)
	voucherHandler := voucher.NewHTTPHandler(voucherService)

	// Proof DI
	proofRepository := proof.NewRepository(db)
	proofService := proof.NewService(proofRepository, userService, voucherService, validator, mpClient, coffejiClient)
	proofHandler := proof.NewHTTPHandler(proofService)

	// Auth DI
	authHandler := auth.NewHTTPHandler(userService, tokenService, jwt, validator, mailerClient)

	// Prode DI
	prodeRepository := prode.NewRepository(db)
	prodeService := prode.NewService(prodeRepository, voucherRepository, userRepository, mailerClient, cfg.ProdeAdminEmails)
	prodeHandler := prode.NewHTTPHandler(prodeService)

	if cfg.IsProdeEnabled() {
		slog.Info("PRODE habilitado", "routes", "/api/v1/prode/*")
		if cfg.IsMaintenanceEnabled() {
			slog.Info("PRODE mantenimiento habilitado", "header", "X-Prode-Admin-Key")
		}
		if len(cfg.ProdeAdminEmails) > 0 {
			slog.Info("PRODE notificaciones admin", "emails", cfg.ProdeAdminEmails)
		}
	} else {
		slog.Info("PRODE deshabilitado")
	}

	// Middlewares
	authMiddleware := middlewares.NewAuthMiddleware(jwt)

	r := routes.Router(routes.Deps{
		UserHandler:    userHandler,
		TokenHandler:   tokenHandler,
		ProofHandler:   proofHandler,
		VoucherHandler: voucherHandler,
		AuthHandler:    authHandler,
		ProdeHandler:   prodeHandler,
		Config:         cfg,
		AuthMiddleware: authMiddleware,
		RateLimiter:    rateLimiter,
		Validator:      validator,
	})

	voucherCron := jobs.NewVoucherCron(voucherService, "@every 20m", 100, 30*time.Second)
	if err := voucherCron.Start(); err != nil {
		slog.Error("cannot start voucher cron", "error", err)
		os.Exit(1)
	}
	defer voucherCron.Stop()

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("API escuchando", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen error", "error", err)
			os.Exit(1)
		}

	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Error al apagar el servidor", "error", err)
	}

	slog.Info("Apagado limpio")
}
