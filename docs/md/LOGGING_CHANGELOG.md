# Logging Changes

## Summary
- Added configurable log level (LOG_LEVEL)
- Introduced centralized Zap logger (pkg/logging)
- Added request-scoped logger middleware with request_id
- Upgraded access logger to level-by-status and include request_id
- Replaced fmt.Printf in user handlers with structured Zap logs
- Updated ENV docs and added Logging Guide

## Files touched
- pkg/config/config.go (+LogLevel)
- pkg/logging/logger.go (new)
- pkg/middleware/request_logger.go (new)
- pkg/middleware/logger.go (level selection and request_id)
- pkg/middleware/performance.go (fields reuse & request_id)
- pkg/api/router.go (mount request logger)
- pkg/api/user.go (replace printf with zap)
- cmd/server/main.go (use logging.NewLogger)
- docs/md/ENV_CONFIG.md (LOG_LEVEL)
- docs/md/LOGGING_GUIDE.md (new)

