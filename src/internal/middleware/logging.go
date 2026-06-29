package middleware

import "backend/src/internal/logger"

// RequestLogger é um alias para o middleware de logging do pacote logger.
var RequestLogger = logger.RequestLoggerMiddleware
