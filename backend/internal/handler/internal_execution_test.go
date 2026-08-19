package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newInternalExecutionContext(body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/internal/v1/model-executions/chat/completions", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	groupID := int64(7)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		GroupID: &groupID,
		Group:   &service.Group{ID: groupID, Platform: service.PlatformComposite},
	})
	return c, recorder
}

func TestInternalExecutionHandlerRejectsMissingServiceToken(t *testing.T) {
	called := false
	h := NewInternalExecutionHandler("dsh-secret", func(*gin.Context) { called = true })
	c, recorder := newInternalExecutionContext(`{"model":"gpt-5","messages":[]}`)

	h.Handle(c)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.False(t, called)
}

func TestInternalExecutionHandlerRejectsMissingAPIKeyContext(t *testing.T) {
	h := NewInternalExecutionHandler("dsh-secret", func(*gin.Context) { t.Fatal("dispatcher must not run") })
	c, recorder := newInternalExecutionContext(`{"model":"gpt-5","messages":[]}`)
	c.Set(string(middleware2.ContextKeyAPIKey), nil)
	c.Request.Header.Set("X-TokenPort-Internal-Token", "dsh-secret")

	h.Handle(c)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestInternalExecutionHandlerSetsDirectTargetForCompositeGroup(t *testing.T) {
	var dispatched bool
	h := NewInternalExecutionHandler("dsh-secret", func(c *gin.Context) {
		dispatched = true
		platform, ok := service.ResolvedTargetPlatformFromContext(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, service.PlatformOpenAI, platform)
		c.Status(http.StatusNoContent)
		c.Writer.WriteHeaderNow()
	})
	c, recorder := newInternalExecutionContext(`{"model":"gpt-5","messages":[]}`)
	c.Request.Header.Set("X-TokenPort-Internal-Token", "dsh-secret")
	c.Request.Header.Set("X-TokenPort-Target-Platform", service.PlatformOpenAI)

	h.Handle(c)

	require.True(t, dispatched)
	require.Equal(t, http.StatusNoContent, recorder.Code)
}
