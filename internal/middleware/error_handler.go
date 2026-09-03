package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gopulse/internal/errors"
)

type ErrorResponse struct {
	Success   bool        `json:"success"`
	Error     AppErrorDTO `json:"error"`
	RequestID string      `json:"requestId"`
}

type AppErrorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			// Only handle the first error
			err := c.Errors[0].Err

			var appErr *errors.AppError
			var ok bool

			if appErr, ok = err.(*errors.AppError); !ok {
				appErr = errors.NewInternal(err)
			}

			statusCode := mapErrorCodeToHTTPStatus(appErr.Code)

			reqID, _ := c.Get("RequestID")

			c.JSON(statusCode, ErrorResponse{
				Success: false,
				Error: AppErrorDTO{
					Code:    string(appErr.Code),
					Message: appErr.Message,
				},
				RequestID: reqID.(string),
			})
		}
	}
}

func mapErrorCodeToHTTPStatus(code errors.ErrorCode) int {
	switch code {
	case errors.InvalidRequest, errors.ValidationError:
		return http.StatusBadRequest
	case errors.Unauthorized:
		return http.StatusUnauthorized
	case errors.Forbidden:
		return http.StatusForbidden
	case errors.NotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
