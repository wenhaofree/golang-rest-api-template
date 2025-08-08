# Logging Guide

This document describes the improved logging architecture and how to configure and use it.

## Overview
- Structured logging powered by Uber Zap
- Configurable log level via environment variable `LOG_LEVEL`
- Request-scoped logger with `request_id` injected by middleware
- Level-based request logging (2xx/3xx info, 4xx warn, 5xx error)
- Optional MongoDB sink for request audit (async, best-effort)

## Configuration
- LOG_LEVEL: debug | info | warn | error (default: info)
- MONGO_ENABLED: enable/disable MongoDB logging (default: true)

Example `.env`:
```
LOG_LEVEL=debug
MONGO_ENABLED=true
```

## Components
- pkg/logging/logger.go: centralized logger initialization
- pkg/middleware/request_logger.go: injects `logger` and `request_id` into context
- pkg/middleware/logger.go: access log with level by status code; forwards to MongoDB when enabled

## Usage in handlers
- Get the request logger from Gin context and log with proper level:

```go
lAny, _ := c.Get("logger")
logger, _ := lAny.(*zap.Logger)
if logger != nil { logger.Info("business event", zap.String("user", email)) }
```

Avoid using fmt.Printf; prefer structured fields.

## Best practices
- debug: development-only details (timings, branches)
- info: successful business events
- warn: validation failures, throttling, slow requests
- error: unexpected errors with `zap.Error(err)`

## MongoDB retention
If using MongoDB for logs, create a TTL index on `timestamp` field (e.g., 14d) and indexes on `method`, `path`, `status`, `request_id` for efficient queries.

