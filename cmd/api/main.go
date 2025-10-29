package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sebaactis/powermix-back-mobile/internal/clients/mercadopago"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/proof"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/token"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/user"
	"github.com/sebaactis/powermix-back-mobile/internal/middlewares"
	"github.com/sebaactis/powermix-back-mobile/internal/platform/config"
	"github.com/sebaactis/powermix-back-mobile/internal/platform/database"
	"github.com/sebaactis/powermix-back-mobile/internal/routes"
	"github.com/sebaactis/powermix-back-mobile/internal/security/auth"
	jwtx "github.com/sebaactis/powermix-back-mobile/internal/security/jwt"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error al cargar el archivo .env", err)
	}

	cfg := config.Load()

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal("Error al conectar la base de datos", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal("Error al migrar las entidades en la base de datos", err)
	}

	// Utils
	jwt := jwtx.NewJWT()
	validator := validations.NewValidator()
	rateLimiter := middlewares.NewRateLimiter(10, 2*time.Minute)

	// MP
	mpClient := mercadopago.NewClient(cfg.MercagoPagoToken)

	// Token DI
	tokenRepository := token.NewRepository(db)
	tokenService := token.NewService(tokenRepository, validator)
	tokenHandler := token.NewHTTPHandler(tokenService)

	// Users DI
	userRepository := user.NewRepository(db)
	userService := user.NewService(userRepository, tokenService, validator)
	userHandler := user.NewHTTPHandler(userService)

	// Proof DI
	proofRepository := proof.NewRepository(db)
	proofService := proof.NewService(proofRepository, validator, mpClient)
	proofHandler := proof.NewHTTPHandler(proofService)

	// Auth DI
	authHandler := auth.NewHTTPHandler(userService, tokenService, jwt, validator)

	// Middlewares
	authMiddleware := middlewares.NewAuthMiddleware(jwt, userService, tokenService)

	r := routes.Router(routes.Deps{
		UserHandler:    userHandler,
		TokenHandler:   tokenHandler,
		ProofHandler:   proofHandler,
		AuthHandler:    authHandler,
		AuthMiddleware: authMiddleware,
		RateLimiter:    rateLimiter,
		Validator:      validator,
	})

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("API escuchando en %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}

	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	log.Println("Apago limpio")
}
