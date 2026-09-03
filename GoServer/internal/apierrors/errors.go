package apierrors

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const problemTypeBase = "https://api.ourdomain.com/errors/"

type InvalidParam struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type ProblemDetails struct {
	Type          string         `json:"type"`
	Title         string         `json:"title"`
	Status        int            `json:"status"`
	Detail        string         `json:"detail"`
	Instance      string         `json:"instance"`
	Code          string         `json:"code"`
	Timestamp     string         `json:"timestamp"`
	InvalidParams []InvalidParam `json:"invalid_params,omitempty"`
}

func NewAPIError(status int, code, title, detail, instance string) ProblemDetails {
	return ProblemDetails{
		Type:      problemTypeBase + codeToSlug(code),
		Title:     title,
		Status:    status,
		Detail:    detail,
		Instance:  instance,
		Code:      code,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func NewValidationError(detail, instance string, params []InvalidParam) ProblemDetails {
	problem := NewAPIError(http.StatusBadRequest, "VALIDATION_FAILED", "Validation Failed", detail, instance)
	problem.InvalidParams = params
	return problem
}

func Respond(c *gin.Context, problem ProblemDetails) {
	c.Header("Content-Type", "application/problem+json")
	c.JSON(problem.Status, problem)
}

func Error(c *gin.Context, status int, code, title, detail string) {
	Respond(c, NewAPIError(status, code, title, detail, c.Request.URL.Path))
}

func Validation(c *gin.Context, detail string, params []InvalidParam) {
	Respond(c, NewValidationError(detail, c.Request.URL.Path, params))
}

func codeToSlug(code string) string {
	return strings.ToLower(strings.ReplaceAll(code, "_", "-"))
}
