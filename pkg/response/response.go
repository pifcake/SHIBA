package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type envelope struct {
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, envelope{Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, envelope{Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, envelope{Error: msg})
}

func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, envelope{Error: msg})
}

func Forbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, envelope{Error: "forbidden"})
}

func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, envelope{Error: msg})
}

func Conflict(c *gin.Context, msg string) {
	c.JSON(http.StatusConflict, envelope{Error: msg})
}

func UnprocessableEntity(c *gin.Context, msg string) {
	c.JSON(http.StatusUnprocessableEntity, envelope{Error: msg})
}

func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, envelope{Error: "internal server error"})
}

func Message(c *gin.Context, status int, msg string) {
	c.JSON(status, envelope{Message: msg})
}
