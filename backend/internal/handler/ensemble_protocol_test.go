package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// ensembleBodyCapture records the exact body each member sub-call received, which
// is the only way to prove the inbound protocol was translated rather than passed
// through: a member that receives an Anthropic body silently fails or, worse,
// succeeds against an upstream that tolerates it and diverges from the Chat
// Completions path.
type ensembleBodyCapture struct {
	mu     sync.Mutex
	bodies []string
}

func (b *ensembleBodyCapture) dispatch(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	b.mu.Lock()
	b.bodies = append(b.bodies, string(body))
	b.mu.Unlock()
	c.JSON(http.StatusOK, chatCompletion("answer", 3, 4))
}

func (b *ensembleBodyCapture) first() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.bodies) == 0 {
		return ""
	}
	return b.bodies[0]
}

// newEnsembleProtocolRequest drives one of the protocol entry points with an
// Ensemble-platform key, mirroring newEnsembleHandlerRequest but letting the
// caller choose the entry point and request path.
func newEnsembleProtocolRequest(
	t *testing.T,
	path string,
	body string,
	members []service.EnsembleProposer,
	cfg service.EnsembleConfig,
	dispatch gin.HandlerFunc,
	entry func(*EnsembleHandler, *gin.Context),
) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		ID:      99,
		GroupID: ensembleInt64Ptr(7),
		Group: &service.Group{
			ID:             7,
			Platform:       service.PlatformEnsemble,
			EnsembleConfig: cfg,
		},
	})
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 7, Concurrency: 2})
	h := NewEnsembleHandler(service.NewEnsembleRuntimeService(&ensembleHandlerRepoStub{members: members}))
	h.SetSubCallDispatcher(dispatch)
	entry(h, c)
	return recorder
}

func ensembleSoloProposer() []service.EnsembleProposer {
	return []service.EnsembleProposer{
		{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
	}
}

// An Anthropic client validates the envelope it gets back, so a Chat Completions
// body reaching /v1/messages is unusable even when the fan-out itself succeeded.
func TestEnsembleMessagesReturnsAnthropicShape(t *testing.T) {
	capture := &ensembleBodyCapture{}
	recorder := newEnsembleProtocolRequest(t, "/v1/messages",
		`{"model":"ensemble","max_tokens":1024,"messages":[{"role":"user","content":"hi"}]}`,
		ensembleSoloProposer(),
		service.EnsembleConfig{MinProposers: 1},
		capture.dispatch,
		func(h *EnsembleHandler, c *gin.Context) { h.Messages(c) },
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	body := recorder.Body.String()
	require.Equal(t, "message", gjson.Get(body, "type").String())
	require.Equal(t, "assistant", gjson.Get(body, "role").String())
	require.Equal(t, "text", gjson.Get(body, "content.0.type").String())
	require.Equal(t, "answer", gjson.Get(body, "content.0.text").String())
	// Anthropic spells token counts differently; a Chat Completions usage block
	// here would leave the client reporting zero tokens.
	require.True(t, gjson.Get(body, "usage.input_tokens").Exists())
	require.False(t, gjson.Get(body, "choices").Exists())
}

// The fan-out has exactly one implementation, so the inbound body must be
// translated before it reaches a member rather than each member learning a second
// protocol.
func TestEnsembleMessagesTranslatesBodyForMembers(t *testing.T) {
	capture := &ensembleBodyCapture{}
	newEnsembleProtocolRequest(t, "/v1/messages",
		`{"model":"ensemble","max_tokens":1024,"system":"be brief","messages":[{"role":"user","content":"hi"}]}`,
		ensembleSoloProposer(),
		service.EnsembleConfig{MinProposers: 1},
		capture.dispatch,
		func(h *EnsembleHandler, c *gin.Context) { h.Messages(c) },
	)

	memberBody := capture.first()
	require.NotEmpty(t, memberBody)
	require.True(t, gjson.Get(memberBody, "messages").IsArray())
	// The member is called with its own model, not the group's public alias.
	require.Equal(t, "gpt-5", gjson.Get(memberBody, "model").String())
	// Anthropic's top-level system prompt becomes a system message; leaving it at
	// the top level would drop it entirely on the Chat Completions path.
	require.Contains(t, memberBody, "be brief")
	require.False(t, gjson.Get(memberBody, "system").Exists())
}

// Anthropic streams are framed as named SSE events and terminate on message_stop.
// A [DONE] sentinel is a Chat Completions convention: appended to an Anthropic
// stream it is a frame the client cannot parse.
func TestEnsembleMessagesStreamsAnthropicEvents(t *testing.T) {
	capture := &ensembleBodyCapture{}
	recorder := newEnsembleProtocolRequest(t, "/v1/messages",
		`{"model":"ensemble","max_tokens":1024,"stream":true,"messages":[{"role":"user","content":"hi"}]}`,
		ensembleSoloProposer(),
		service.EnsembleConfig{MinProposers: 1},
		capture.dispatch,
		func(h *EnsembleHandler, c *gin.Context) { h.Messages(c) },
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Header().Get("Content-Type"), "text/event-stream")
	body := recorder.Body.String()
	require.Contains(t, body, "event: content_block_delta")
	require.Contains(t, body, "text_delta")
	require.Contains(t, body, "answer")
	require.NotContains(t, body, "chat.completion.chunk")

	// Counted, not merely present. A duplicated terminal is the failure mode
	// here: the chunk carrying finish_reason and the finalize pass both have a
	// claim on ending the message, and a second message_stop puts a strict client
	// into an invalid state rather than producing a visible error.
	require.Equal(t, 1, strings.Count(body, "event: message_start"))
	require.Equal(t, 1, strings.Count(body, "event: message_delta"))
	require.Equal(t, 1, strings.Count(body, "event: message_stop"))
	// An Anthropic stream has no [DONE] sentinel; it ends on message_stop.
	require.NotContains(t, body, "[DONE]")

	// The execution trace rides delta.reasoning_content, which the bridge maps to
	// a native thinking block. That is the point of transcoding rather than
	// inventing a field: the trace arrives on the channel an Anthropic client
	// already renders. The fan-out boundary strips these back off on the next
	// turn, so a client echoing them cannot leak an unsigned thinking block to a
	// member.
	require.Contains(t, body, "thinking_delta")
}

func TestEnsembleResponsesReturnsResponsesShape(t *testing.T) {
	capture := &ensembleBodyCapture{}
	recorder := newEnsembleProtocolRequest(t, "/v1/responses",
		`{"model":"ensemble","input":"hi"}`,
		ensembleSoloProposer(),
		service.EnsembleConfig{MinProposers: 1},
		capture.dispatch,
		func(h *EnsembleHandler, c *gin.Context) { h.Responses(c) },
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	body := recorder.Body.String()
	require.Equal(t, "response", gjson.Get(body, "object").String())
	require.Equal(t, "completed", gjson.Get(body, "status").String())
	require.Contains(t, body, "answer")
	require.False(t, gjson.Get(body, "choices").Exists())

	memberBody := capture.first()
	require.True(t, gjson.Get(memberBody, "messages").IsArray())
	require.Equal(t, "gpt-5", gjson.Get(memberBody, "model").String())
}

// Unlike Anthropic, a Responses stream does end on [DONE], after its own
// terminal events.
func TestEnsembleResponsesStreamsResponsesEvents(t *testing.T) {
	capture := &ensembleBodyCapture{}
	recorder := newEnsembleProtocolRequest(t, "/v1/responses",
		`{"model":"ensemble","input":"hi","stream":true}`,
		ensembleSoloProposer(),
		service.EnsembleConfig{MinProposers: 1},
		capture.dispatch,
		func(h *EnsembleHandler, c *gin.Context) { h.Responses(c) },
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Header().Get("Content-Type"), "text/event-stream")
	body := recorder.Body.String()
	require.Contains(t, body, "answer")
	require.NotContains(t, body, "chat.completion.chunk")

	require.Equal(t, 1, strings.Count(body, "event: response.created"))
	require.Equal(t, 1, strings.Count(body, "event: response.completed"))
	// Unlike Anthropic, a Responses stream does end on the [DONE] sentinel, and
	// exactly once.
	require.Equal(t, 1, strings.Count(body, "data: [DONE]"))

	// The trace reaches a Responses client as its own reasoning summary, the
	// channel that protocol already has for model thinking.
	require.Contains(t, body, "response.reasoning_summary_text.delta")
}

// A misconfigured group is the case an operator most needs to read, so the error
// has to arrive in an envelope the client can parse. Anthropic carries a
// top-level "type":"error" beside the error object.
func TestEnsembleMessagesErrorUsesAnthropicEnvelope(t *testing.T) {
	recorder := newEnsembleProtocolRequest(t, "/v1/messages",
		`{"model":"ensemble","max_tokens":1024,"messages":[{"role":"user","content":"hi"}]}`,
		nil,
		service.EnsembleConfig{MinProposers: 1},
		func(c *gin.Context) { t.Fatal("no member should be dispatched") },
		func(h *EnsembleHandler, c *gin.Context) { h.Messages(c) },
	)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	body := recorder.Body.String()
	require.Equal(t, "error", gjson.Get(body, "type").String())
	require.Equal(t, "invalid_request_error", gjson.Get(body, "error.type").String())
	require.Contains(t, gjson.Get(body, "error.message").String(), "no enabled proposer")
}

// Responses keys the discriminator as "code"; emitting "type" there leaves a
// strict client with an error it reports as unknown.
func TestEnsembleResponsesErrorUsesResponsesEnvelope(t *testing.T) {
	recorder := newEnsembleProtocolRequest(t, "/v1/responses",
		`{"model":"ensemble","input":"hi"}`,
		nil,
		service.EnsembleConfig{MinProposers: 1},
		func(c *gin.Context) { t.Fatal("no member should be dispatched") },
		func(h *EnsembleHandler, c *gin.Context) { h.Responses(c) },
	)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	body := recorder.Body.String()
	require.Equal(t, "invalid_request_error", gjson.Get(body, "error.code").String())
	require.False(t, gjson.Get(body, "error.type").Exists())
	require.False(t, gjson.Get(body, "type").Exists())
}

// Making the writer protocol-aware must not move the Chat Completions envelope,
// which existing clients are already parsing.
func TestEnsembleChatCompletionsErrorKeepsChatEnvelope(t *testing.T) {
	recorder := newEnsembleHandlerRequest(t,
		nil,
		service.EnsembleConfig{MinProposers: 1},
		func(c *gin.Context) { t.Fatal("no member should be dispatched") },
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}]}`,
	)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	body := recorder.Body.String()
	require.Equal(t, "invalid_request_error", gjson.Get(body, "error.type").String())
	require.False(t, gjson.Get(body, "type").Exists())
}
