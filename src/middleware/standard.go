package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"

	"github.com/labstack/echo/v4"
)

// Recover is a middleware that recovers from panics anywhere in the chain,
// logs the stack trace, and returns an HTTP 500 error.
func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					stack := make([]byte, 2048)
					length := runtime.Stack(stack, false)
					log.Printf("[PANIC RECOVER] %v\n%s\n", err, stack[:length])
					c.Error(echo.NewHTTPError(http.StatusInternalServerError, err.Error()))
				}
			}()
			return next(c)
		}
	}
}

// CORSConfig defines the config for CORS middleware.
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
}

// CORS returns a CORS middleware configured with the provided options.
func CORS(config CORSConfig) echo.MiddlewareFunc {
	allowMethodsStr := strings.Join(config.AllowMethods, ", ")
	allowHeadersStr := strings.Join(config.AllowHeaders, ", ")

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			origin := req.Header.Get(echo.HeaderOrigin)

			// Check if origin is allowed
			allowed := false
			for _, o := range config.AllowOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed && origin != "" {
				res.Header().Set(echo.HeaderAccessControlAllowOrigin, origin)
				if config.AllowCredentials {
					res.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
				}
			} else if allowed && len(config.AllowOrigins) > 0 && config.AllowOrigins[0] == "*" {
				res.Header().Set(echo.HeaderAccessControlAllowOrigin, "*")
			}

			if req.Method == http.MethodOptions {
				res.Header().Set(echo.HeaderAccessControlAllowMethods, allowMethodsStr)
				res.Header().Set(echo.HeaderAccessControlAllowHeaders, allowHeadersStr)
				if allowed && origin != "" {
					res.Header().Set(echo.HeaderAccessControlAllowOrigin, origin)
					if config.AllowCredentials {
						res.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
					}
				}
				return c.NoContent(http.StatusNoContent)
			}

			return next(c)
		}
	}
}

// Logger returns a simple Logger middleware.
func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			log.Printf("API: method=%s, uri=%s, status=%d\n", req.Method, req.RequestURI, res.Status)
			return nil
		}
	}
}
