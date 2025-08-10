package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/flarexio/core/endpoint"
	"github.com/flarexio/talkix"
)

func ListSessionsHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := c.Get("user")
		if !ok {
			err := errors.New("user not found in context")
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, talkix.UserKey, u)

		resp, err := endpoint(ctx, nil)
		if err != nil {
			c.String(http.StatusExpectationFailed, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, &resp)
	}
}

func SessionHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := c.Get("user")
		if !ok {
			err := errors.New("user not found in context")
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		sessionID := c.Param("session")
		if sessionID == "" {
			err := errors.New("session is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, talkix.UserKey, u)

		resp, err := endpoint(ctx, sessionID)
		if err != nil {
			c.String(http.StatusExpectationFailed, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, &resp)
	}
}

func CreateSessionHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := c.Get("user")
		if !ok {
			err := errors.New("user not found in context")
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, talkix.UserKey, u)

		resp, err := endpoint(ctx, nil)
		if err != nil {
			c.String(http.StatusExpectationFailed, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.JSON(http.StatusCreated, &resp)
	}
}

func SwitchSessionHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := c.Get("user")
		if !ok {
			err := errors.New("user not found in context")
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		sessionID := c.Param("session")
		if sessionID == "" {
			err := errors.New("session is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, talkix.UserKey, u)

		_, err := endpoint(ctx, sessionID)
		if err != nil {
			c.String(http.StatusExpectationFailed, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.String(http.StatusOK, "Session switched successfully")
	}
}

func DeleteSessionHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := c.Get("user")
		if !ok {
			err := errors.New("user not found in context")
			c.String(http.StatusInternalServerError, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		sessionID := c.Param("session")
		if sessionID == "" {
			err := errors.New("session is required")
			c.String(http.StatusBadRequest, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, talkix.UserKey, u)

		_, err := endpoint(ctx, sessionID)
		if err != nil {
			c.String(http.StatusExpectationFailed, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.String(http.StatusOK, "Session deleted successfully")
	}
}
