package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func LoggerPostBody(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
		}

		log.Debug().Msg(string(body))

		c.Request.Body = io.NopCloser(bytes.NewReader(body))
	}

	c.Next()
}
