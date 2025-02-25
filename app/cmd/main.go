package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
	db := db.NewPGStorage()
	path := fmt.Sprintf("0.0.0.0:%s", config.Envs.Port)

	apiServer := api.NewApiServer(path, db)

	initTracer()

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

func initTracer() {
	exporter, _ := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithEndpoint("0.0.0.0:4317"),
		otlptracegrpc.WithInsecure(),
	)

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("book-store-api"),
		semconv.DeploymentEnvironment(config.Envs.Environment),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
}
