FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY todo-api /app/todo-api

# Create directory for persistent SQLite data
RUN mkdir -p /app/data && chown -R appuser:appgroup /app

USER appuser

ENV PORT=8080
ENV APP_ENV=production
ENV DATABASE_PATH=/app/data/todo.db

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/todo-api"]
