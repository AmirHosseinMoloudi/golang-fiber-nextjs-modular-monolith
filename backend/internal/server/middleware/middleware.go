package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-todo/modular-monolith/internal/core/errors"
	"github.com/golang-todo/modular-monolith/internal/core/observability"
	identityApp "github.com/golang-todo/modular-monolith/internal/modules/identity/application"
	"github.com/golang-todo/modular-monolith/internal/modules/identity/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func ObservabilityMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()

		// Extract OpenTelemetry context
		carrier := propagation.HeaderCarrier{}
		c.Request().Header.VisitAll(func(key, val []byte) {
			carrier.Set(string(key), string(val))
		})
		ctx := otel.GetTextMapPropagator().Extract(c.Context(), carrier)

		ctx, span := observability.Tracer.Start(ctx, fmt.Sprintf("%s %s", method, path),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		traceID := span.SpanContext().TraceID().String()
		c.Locals("trace_id", traceID)
		c.Locals("otel_ctx", ctx)
		c.Set("X-Trace-Id", traceID)

		// Next handler
		err := c.Next()

		duration := time.Since(start).Seconds()
		statusCode := c.Response().StatusCode()
		statusStr := fmt.Sprintf("%d", statusCode)

		// Prometheus Metrics
		observability.HttpRequestsTotal.WithLabelValues(path, method, statusStr).Inc()
		observability.HttpRequestDuration.WithLabelValues(path, method).Observe(duration)

		// Structured Zerolog logging
		logEvent := observability.Logger.Info()
		if statusCode >= 500 {
			logEvent = observability.Logger.Error()
		} else if statusCode >= 400 {
			logEvent = observability.Logger.Warn()
		}

		logEvent.
			Str("trace_id", traceID).
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Float64("duration_sec", duration).
			Str("ip", c.IP()).
			Msg("HTTP request processed")

		return err
	}
}

func AuthMiddleware(tokenService identityApp.TokenService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return errors.NewUnauthorizedError("authorization header missing")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return errors.NewUnauthorizedError("invalid authorization header scheme")
		}

		claims, err := tokenService.ValidateToken(parts[1])
		if err != nil {
			return err
		}

		c.Locals("claims", claims)
		c.Locals("user_id", claims.UserID)
		c.Locals("tenant_id", claims.TenantID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", string(claims.Role))
		return c.Next()
	}
}

func RequireRoles(roles ...domain.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("claims").(*identityApp.JWTClaims)
		if !ok || claims == nil {
			return errors.NewUnauthorizedError("unauthenticated")
		}

		for _, r := range roles {
			if claims.Role == r {
				return c.Next()
			}
		}
		return errors.NewForbiddenError("insufficient permissions for this resource")
	}
}

func GlobalErrorHandler(c *fiber.Ctx, err error) error {
	code := http.StatusInternalServerError
	resp := fiber.Map{
		"error": fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		},
	}

	if appErr, ok := err.(*errors.AppError); ok {
		code = appErr.HTTPStatus
		errMap := fiber.Map{
			"code":    appErr.Code,
			"message": appErr.Message,
			"type":    appErr.Type,
		}
		if len(appErr.Details) > 0 {
			errMap["details"] = appErr.Details
		}
		resp = fiber.Map{"error": errMap}
	} else if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		resp = fiber.Map{
			"error": fiber.Map{
				"code":    "HTTP_ERROR",
				"message": e.Message,
			},
		}
	}

	traceID, _ := c.Locals("trace_id").(string)
	if traceID != "" {
		if m, ok := resp["error"].(fiber.Map); ok {
			m["trace_id"] = traceID
		}
	}

	return c.Status(code).JSON(resp)
}
