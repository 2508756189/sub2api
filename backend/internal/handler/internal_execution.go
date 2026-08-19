package handler

import (
	"crypto/subtle"
	"net/http"
	"strings"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	InternalExecutionTokenHeader          = "X-TokenPort-Internal-Token"
	InternalExecutionTargetPlatformHeader = "X-TokenPort-Target-Platform"
)

// InternalExecutionHandler exposes one direct model execution seam for trusted
// orchestrators such as DSH. It deliberately delegates the actual call to the
// existing platform handler, so public /v1 behavior and billing stay unchanged.
type InternalExecutionHandler struct {
	token    string
	dispatch gin.HandlerFunc
}

func NewInternalExecutionHandler(token string, dispatch gin.HandlerFunc) *InternalExecutionHandler {
	return &InternalExecutionHandler{token: strings.TrimSpace(token), dispatch: dispatch}
}

// Handle authenticates the DSH service and the user API key before allowing a
// direct call. Composite groups must name a concrete target platform so this
// endpoint cannot bypass their routing policy.
func (h *InternalExecutionHandler) Handle(c *gin.Context) {
	if h == nil || c == nil || h.dispatch == nil || !h.authorized(c) {
		internalExecutionError(c, http.StatusUnauthorized, "Invalid internal execution credentials")
		return
	}

	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil || apiKey.Group == nil {
		internalExecutionError(c, http.StatusUnauthorized, "Valid API key context is required")
		return
	}

	groupPlatform := strings.ToLower(strings.TrimSpace(apiKey.Group.Platform))
	targetPlatform := strings.ToLower(strings.TrimSpace(c.GetHeader(InternalExecutionTargetPlatformHeader)))
	if targetPlatform != "" && !isInternalExecutionConcretePlatform(targetPlatform) {
		internalExecutionError(c, http.StatusBadRequest, "target platform must be concrete")
		return
	}
	if targetPlatform != "" && groupPlatform != service.PlatformComposite && targetPlatform != groupPlatform {
		internalExecutionError(c, http.StatusBadRequest, "target platform is not allowed for this group")
		return
	}
	if targetPlatform != "" && groupPlatform == service.PlatformComposite {
		c.Request = c.Request.WithContext(service.WithResolvedTargetPlatform(c.Request.Context(), targetPlatform))
	}

	h.dispatch(c)
}

func (h *InternalExecutionHandler) authorized(c *gin.Context) bool {
	provided := strings.TrimSpace(c.GetHeader(InternalExecutionTokenHeader))
	if h.token == "" || provided == "" || len(provided) != len(h.token) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(h.token)) == 1
}

func isInternalExecutionConcretePlatform(platform string) bool {
	switch platform {
	case service.PlatformAnthropic, service.PlatformOpenAI, service.PlatformGemini,
		service.PlatformAntigravity, service.PlatformGrok, service.PlatformKimi,
		service.PlatformZhipu, service.PlatformDeepseek, service.PlatformKiro:
		return true
	default:
		return false
	}
}

func internalExecutionError(c *gin.Context, status int, message string) {
	if c == nil {
		return
	}
	errorType := "authentication_error"
	if status >= http.StatusBadRequest && status != http.StatusUnauthorized {
		errorType = "invalid_request_error"
	}
	c.JSON(status, gin.H{
		"error": gin.H{
			"type":    errorType,
			"message": message,
		},
	})
}
