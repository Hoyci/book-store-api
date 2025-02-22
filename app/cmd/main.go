package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hoyci/book-store-api/cmd/api"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/db"
	"github.com/hoyci/book-store-api/service/auth"
	"github.com/hoyci/book-store-api/service/book"
	"github.com/hoyci/book-store-api/service/healthcheck"
	"github.com/hoyci/book-store-api/service/user"
	"github.com/hoyci/book-store-api/utils"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// @title Book Store API
// @version 1.0
// @description API para gestão de livros
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	shutdown := initTracer()
	defer shutdown()

	db := db.NewPGStorage()
	path := fmt.Sprintf("0.0.0.0:%s", config.Envs.Port)

	apiServer := api.NewApiServer(path, db)

	healthCheckHandler := healthcheck.NewHealthCheckHandler(config.Envs)

	bookStore := book.NewBookStore(db)
	bookHandler := book.NewBookHandler(bookStore)

	userStore := user.NewUserStore(db)
	userHandler := user.NewUserHandler(userStore)

	authStore := auth.NewAuthStore(db)
	uuidGen := &utils.UUIDGeneratorUtil{}
	authHandler := auth.NewAuthHandler(userStore, authStore, uuidGen)

	apiServer.SetupRouter(healthCheckHandler, bookHandler, userHandler, authHandler)

	log.Println("Listening on:", path)
	http.ListenAndServe(path, apiServer.Router)
}

func initTracer() func() {
	ctx := context.Background()
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("otel-collector.observability.svc.cluster.local:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create OTLP trace exporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("book-store-api"),
			semconv.DeploymentEnvironment(config.Envs.Environment),
		)),
	)
	otel.SetTracerProvider(tp)

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("Error shutting down tracer provider: %v", err)
		}
	}
}
