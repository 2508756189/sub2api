package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
)

// Ensemble fan-out is protocol-agnostic in substance and protocol-specific only
// at its two edges: the request shape it parses and the response shape it writes.
//
// A group whose platform is `ensemble` is a group, not an upstream, so every
// client protocol the platform serves has to reach it. Before this, only
// /v1/chat/completions did; /v1/messages and /v1/responses fell through to normal
// account resolution, found no account carrying platform `ensemble`, and failed
// with 503 "No available accounts" — a message that describes nothing the
// operator can act on.
//
// The fix translates at the edge rather than duplicating the orchestrator. An
// Anthropic or Responses request is converted to Chat Completions shape on the
// way in, runs through the single existing fan-out, and is converted back on the
// way out. That keeps one implementation of the parts that are genuinely hard —
// the member race, the min-proposers rule, aggregation and its fallback, cost
// accounting — instead of three that drift apart. The converters are the same
// apicompat bridges the gateway already uses for cross-protocol upstream calls,
// so the shapes are the ones the platform is already tested against.
type ensembleWireProtocol int

const (
	// ensembleWireChat is the zero value, so anything that does not explicitly
	// declare a protocol keeps the original Chat Completions behaviour.
	ensembleWireChat ensembleWireProtocol = iota
	ensembleWireAnthropic
	ensembleWireResponses
)

const (
	ensembleWireProtocolKey   = "ensemble_wire_protocol"
	ensembleResponsesToolsKey = "ensemble_responses_tools"
)

// ensembleResponsesToolCtx carries the tool metadata that a Responses request
// needs on the way back out. Custom tools, the tool-search declaration, and
// namespaced names are all decided by the inbound request, and the response
// converter cannot re-derive them from a Chat Completions payload.
type ensembleResponsesToolCtx struct {
	customTools    map[string]bool
	toolSearch     bool
	namespaceTools map[string]apicompat.NamespacedToolName
}

func ensembleProtocolFromContext(c *gin.Context) ensembleWireProtocol {
	if v, ok := c.Get(ensembleWireProtocolKey); ok {
		if p, ok := v.(ensembleWireProtocol); ok {
			return p
		}
	}
	return ensembleWireChat
}

func ensembleResponsesToolsFromContext(c *gin.Context) ensembleResponsesToolCtx {
	if v, ok := c.Get(ensembleResponsesToolsKey); ok {
		if t, ok := v.(ensembleResponsesToolCtx); ok {
			return t
		}
	}
	return ensembleResponsesToolCtx{}
}

// replaceEnsembleRequestBody swaps in the translated body so the fan-out reads
// Chat Completions shape. ContentLength has to move with it: middleware and the
// sub-call dispatcher both read it, and a stale length truncates the body.
func replaceEnsembleRequestBody(c *gin.Context, body []byte) {
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	c.Request.ContentLength = int64(len(body))
}

// Messages runs an Anthropic /v1/messages request through the ensemble fan-out.
func (h *EnsembleHandler) Messages(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.errorResponseAs(c, ensembleWireAnthropic, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}
	var req apicompat.AnthropicRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.errorResponseAs(c, ensembleWireAnthropic, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
		return
	}
	converted, err := apicompat.AnthropicToChatCompletionsRequest(&req)
	if err != nil {
		h.errorResponseAs(c, ensembleWireAnthropic, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}
	translated, err := json.Marshal(converted)
	if err != nil {
		h.errorResponseAs(c, ensembleWireAnthropic, http.StatusInternalServerError, "api_error", "Failed to translate request")
		return
	}

	c.Set(ensembleWireProtocolKey, ensembleWireAnthropic)
	replaceEnsembleRequestBody(c, translated)
	h.ChatCompletions(c)
}

// Responses runs an OpenAI /v1/responses request through the ensemble fan-out.
//
// Compact is not routed here: it is context maintenance rather than a question
// to answer, so it keeps its own single-aggregator path.
func (h *EnsembleHandler) Responses(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.errorResponseAs(c, ensembleWireResponses, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}
	var req apicompat.ResponsesRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.errorResponseAs(c, ensembleWireResponses, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
		return
	}

	// Tool metadata is derived from the inbound request, exactly as the normal
	// Responses path does, because a Chat Completions payload no longer says which
	// tools were custom or namespaced.
	tools, err := apicompat.EffectiveResponsesTools(&req)
	if err != nil {
		h.errorResponseAs(c, ensembleWireResponses, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}
	toolCtx := ensembleResponsesToolCtx{
		customTools:    apicompat.CustomToolNames(tools),
		toolSearch:     apicompat.HasToolSearchTool(tools),
		namespaceTools: apicompat.NamespaceToolNames(tools),
	}

	converted, err := apicompat.ResponsesToChatCompletionsRequest(&req)
	if err != nil {
		h.errorResponseAs(c, ensembleWireResponses, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}
	translated, err := json.Marshal(converted)
	if err != nil {
		h.errorResponseAs(c, ensembleWireResponses, http.StatusInternalServerError, "api_error", "Failed to translate request")
		return
	}

	c.Set(ensembleWireProtocolKey, ensembleWireResponses)
	c.Set(ensembleResponsesToolsKey, toolCtx)
	replaceEnsembleRequestBody(c, translated)
	h.ChatCompletions(c)
}

// writeEnsemblePayload writes a completed non-streaming fan-out result in the
// protocol the client asked for.
//
// ensemble_metadata does not survive the conversion, and cannot: it is a Chat
// Completions extension with no field to occupy in an Anthropic message or a
// Responses object, and inventing one would put a non-conforming key in a
// response strict clients validate. It stays available where it is defined —
// Chat Completions and the admin diagnostic stream.
func (h *EnsembleHandler) writeEnsemblePayload(c *gin.Context, payload map[string]any) {
	protocol := ensembleProtocolFromContext(c)
	if protocol == ensembleWireChat {
		c.JSON(http.StatusOK, payload)
		return
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		h.errorResponseAs(c, protocol, http.StatusInternalServerError, "api_error", "Failed to encode response")
		return
	}
	var ccResp apicompat.ChatCompletionsResponse
	if err := json.Unmarshal(encoded, &ccResp); err != nil {
		h.errorResponseAs(c, protocol, http.StatusInternalServerError, "api_error", "Failed to encode response")
		return
	}

	switch protocol {
	case ensembleWireAnthropic:
		c.JSON(http.StatusOK, apicompat.ChatCompletionsResponseToAnthropic(&ccResp, ccResp.Model))
	case ensembleWireResponses:
		toolCtx := ensembleResponsesToolsFromContext(c)
		c.JSON(http.StatusOK, apicompat.ChatCompletionsResponseToResponses(
			&ccResp, ccResp.Model, toolCtx.customTools, toolCtx.toolSearch, toolCtx.namespaceTools))
	default:
		c.JSON(http.StatusOK, payload)
	}
}

// errorResponseAs writes a pre-stream error in one protocol's envelope.
//
// The three envelopes disagree in ways a client notices: Anthropic expects a
// top-level "type":"error" alongside the error object, and Responses keys the
// discriminator as "code" rather than "type". A client that cannot parse the
// envelope reports its own generic failure instead of our message, which is how
// a clear "this group has no enabled proposers" turns into an unexplained error
// at the caller.
func (h *EnsembleHandler) errorResponseAs(c *gin.Context, protocol ensembleWireProtocol, status int, errType, message string) {
	switch protocol {
	case ensembleWireAnthropic:
		c.JSON(status, gin.H{
			"type":  "error",
			"error": gin.H{"type": errType, "message": message},
		})
	case ensembleWireResponses:
		c.JSON(status, gin.H{
			"error": gin.H{"code": errType, "message": message},
		})
	default:
		c.JSON(status, gin.H{
			"error": gin.H{"type": errType, "message": message},
		})
	}
}

// ensembleStreamTerminator writes the terminal frames for a protocol.
//
// Anthropic and Responses differ here in a way that is easy to get wrong:
// Responses ends with a "data: [DONE]" sentinel after its finalize events, while
// an Anthropic stream has no sentinel at all and ends on message_stop. Emitting
// [DONE] to an Anthropic client appends a frame it cannot parse; omitting the
// finalize events leaves the message open and the turn unresumable.
func (w *ensembleStreamWriter) writeTerminalLocked() {
	switch w.protocol {
	case ensembleWireAnthropic:
		for _, evt := range apicompat.FinalizeChatCompletionsAnthropicStream(w.anthropicState) {
			if sse, err := apicompat.ResponsesAnthropicEventToSSE(evt); err == nil {
				w.writeRawLocked(sse)
			}
		}
	case ensembleWireResponses:
		for _, evt := range apicompat.FinalizeChatCompletionsResponsesStream(w.responsesState) {
			if sse, err := apicompat.ResponsesEventToSSE(evt); err == nil {
				w.writeRawLocked(sse)
			}
		}
		w.writeRawLocked("data: [DONE]\n\n")
	default:
		w.writeRawLocked("data: [DONE]\n\n")
	}
}

// transcodeChunkLocked converts one Chat Completions chunk into the client's
// protocol and returns the SSE frames to write.
//
// The chunk is routed through the same apicompat state machines the gateway uses
// for cross-protocol upstream streams, so block indices, tool-call announcement
// and usage roll-up behave exactly as they do on a direct call. State mutation
// and the write share the caller's lock: the fan-out emits progress from several
// goroutines, and an unsynchronised state machine would interleave block indices
// between two members and produce a stream no client can reassemble.
func (w *ensembleStreamWriter) transcodeChunkLocked(encoded []byte) []string {
	var chunk apicompat.ChatCompletionsChunk
	if err := json.Unmarshal(encoded, &chunk); err != nil {
		return nil
	}
	var frames []string
	switch w.protocol {
	case ensembleWireAnthropic:
		for _, evt := range apicompat.ChatCompletionsChunkToAnthropicEvents(&chunk, w.anthropicState) {
			if sse, err := apicompat.ResponsesAnthropicEventToSSE(evt); err == nil {
				frames = append(frames, sse)
			}
		}
	case ensembleWireResponses:
		for _, evt := range apicompat.ChatCompletionsChunkToResponsesEvents(&chunk, w.responsesState) {
			if sse, err := apicompat.ResponsesEventToSSE(evt); err == nil {
				frames = append(frames, sse)
			}
		}
	}
	return frames
}

// writeTerminal emits the stream's terminal frames, taking the writer's lock so
// a protocol whose terminator is several frames cannot be split by a concurrent
// write.
func (w *ensembleStreamWriter) writeTerminal() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.writeTerminalLocked()
}
