package handler

import (
	"bytes"
	"net/http/httptest"
	"sync"

	"github.com/tidwall/gjson"
)

// ensembleReasoningTee is the sub-call response writer that lets a member
// model's own reasoning reach the caller while the request is still running.
//
// Members are dispatched with stream=true and their SSE body is buffered so the
// ensemble can normalize a complete answer before aggregating. That buffering is
// why a member's delta.reasoning_content used to be dropped on the floor: by the
// time the body was parsed, only content and tool_calls were kept. A direct call
// to the same model streams its thinking, so discarding it made the ensemble a
// capability regression rather than a superset.
//
// The tee keeps the recorder semantics intact — callers still read Code and Body
// exactly as before — and additionally scans whole SSE lines as they are written,
// forwarding reasoning deltas to a sink. Partial writes are carried over between
// calls, because a provider may split one line across several Write calls.
// ensembleReasoningSink receives one member reasoning fragment as it arrives. A
// nil sink means the member stays fully buffered, which is what every proposer
// does.
type ensembleReasoningSink func(string)

type ensembleReasoningTee struct {
	*httptest.ResponseRecorder

	mu      sync.Mutex
	pending []byte
	sink    ensembleReasoningSink
}

// newEnsembleReasoningTee wraps the sub-call recorder. The recorder stays the
// caller's handle for reading Code and Body, so the normalizing path below is
// unchanged; the tee only adds the live scan on the way through.
func newEnsembleReasoningTee(recorder *httptest.ResponseRecorder, sink ensembleReasoningSink) *ensembleReasoningTee {
	return &ensembleReasoningTee{ResponseRecorder: recorder, sink: sink}
}

func (t *ensembleReasoningTee) Write(p []byte) (int, error) {
	n, err := t.ResponseRecorder.Write(p)
	if n > 0 {
		t.scan(p[:n])
	}
	return n, err
}

// WriteString must be overridden explicitly. gin's responseWriter routes strings
// through io.WriteString, which prefers an io.StringWriter — the embedded
// recorder provides one, so without this the promoted method would bypass Write
// and the scan would never see a byte.
func (t *ensembleReasoningTee) WriteString(s string) (int, error) {
	return t.Write([]byte(s))
}

// disable detaches the sink. The dispatcher runs in its own goroutine precisely
// so a provider that ignores cancellation cannot hold the request open; on that
// path the sub-call returns while writes are still arriving. Detaching before
// returning is what guarantees a late member write can never appear in the client
// stream after the response has been finished.
func (t *ensembleReasoningTee) disable() {
	t.mu.Lock()
	t.sink = nil
	t.pending = nil
	t.mu.Unlock()
}

func (t *ensembleReasoningTee) scan(chunk []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sink == nil {
		return
	}
	t.pending = append(t.pending, chunk...)
	for {
		index := bytes.IndexByte(t.pending, '\n')
		if index < 0 {
			break
		}
		line := t.pending[:index]
		t.pending = t.pending[index+1:]
		t.emit(line)
	}
	// A provider that never emits a newline must not grow this buffer without
	// bound; the cap matches the SSE line cap used when normalizing member bodies.
	if len(t.pending) > maxEnsembleAggregatorBodyBytes {
		t.pending = t.pending[:0]
	}
}

// emit forwards one line's reasoning delta. Anything that is not a parseable
// data: frame carrying reasoning_content is ignored: content, tool calls, usage
// and the [DONE] sentinel are all handled by the normalizing parser instead.
func (t *ensembleReasoningTee) emit(line []byte) {
	line = bytes.TrimSpace(line)
	if !bytes.HasPrefix(line, []byte("data:")) {
		return
	}
	data := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
	if len(data) == 0 || bytes.Equal(data, []byte("[DONE]")) || !gjson.ValidBytes(data) {
		return
	}
	reasoning := gjson.GetBytes(data, "choices.0.delta.reasoning_content")
	if reasoning.Type != gjson.String || reasoning.String() == "" {
		return
	}
	t.sink(reasoning.String())
}
