package admin

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// EnsembleProposerRequest is the create/update payload for one ensemble member.
type EnsembleProposerRequest struct {
	Role     string `json:"role" binding:"omitempty,oneof=proposer aggregator"`
	Model    string `json:"model" binding:"required"`
	Platform string `json:"platform"`
	Priority int    `json:"priority"`
	Enabled  *bool  `json:"enabled"`
}

// EnsembleConfigRequest is the group-level ensemble options payload.
type EnsembleConfigRequest struct {
	SourceGroupIDs    []int64 `json:"source_group_ids"`
	AggregatorEnabled bool    `json:"aggregator_enabled"`
	MinProposers      int     `json:"min_proposers"`
	TimeoutSeconds    int     `json:"timeout_seconds"`
	MaxTokens         int     `json:"max_tokens"`
	ExposeMetadata    bool    `json:"expose_metadata"`
	// StreamTrace is a pointer so an older admin client that omits the field
	// keeps the default-on behaviour instead of silently disabling the trace.
	StreamTrace *bool `json:"stream_trace"`
}

func ensembleProposerRequestToInput(req EnsembleProposerRequest, defaultEnabled bool) service.EnsembleProposerInput {
	enabled := defaultEnabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	return service.EnsembleProposerInput{
		Role:     req.Role,
		Model:    req.Model,
		Platform: req.Platform,
		Priority: req.Priority,
		Enabled:  enabled,
	}
}

// ListEnsembleProposers lists every member of an ensemble group.
// GET /api/v1/admin/groups/:id/ensemble-proposers
func (h *GroupHandler) ListEnsembleProposers(c *gin.Context) {
	groupID, ok := parsePositiveIDParam(c, "id")
	if !ok {
		return
	}
	proposers, err := h.adminService.ListEnsembleProposers(c.Request.Context(), groupID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, proposers)
}

// CreateEnsembleProposer adds one member to an ensemble group.
// POST /api/v1/admin/groups/:id/ensemble-proposers
func (h *GroupHandler) CreateEnsembleProposer(c *gin.Context) {
	groupID, ok := parsePositiveIDParam(c, "id")
	if !ok {
		return
	}
	var req EnsembleProposerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}
	proposer, err := h.adminService.CreateEnsembleProposer(c.Request.Context(), groupID, ensembleProposerRequestToInput(req, true))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, proposer)
}

// UpdateEnsembleProposer replaces one member of an ensemble group.
// PUT /api/v1/admin/groups/:id/ensemble-proposers/:proposer_id
func (h *GroupHandler) UpdateEnsembleProposer(c *gin.Context) {
	groupID, ok := parsePositiveIDParam(c, "id")
	if !ok {
		return
	}
	proposerID, ok := parsePositiveIDParam(c, "proposer_id")
	if !ok {
		return
	}
	var req EnsembleProposerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}
	existing, err := h.adminService.ListEnsembleProposers(c.Request.Context(), groupID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	defaultEnabled := true
	for _, member := range existing {
		if member.ID != proposerID {
			continue
		}
		if req.Role == "" {
			req.Role = member.Role
		}
		defaultEnabled = member.Enabled
		if strings.TrimSpace(req.Platform) == "" {
			req.Platform = member.Platform
		}
		break
	}
	proposer, err := h.adminService.UpdateEnsembleProposer(c.Request.Context(), groupID, proposerID, ensembleProposerRequestToInput(req, defaultEnabled))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, proposer)
}

// DeleteEnsembleProposer removes one member from an ensemble group.
// DELETE /api/v1/admin/groups/:id/ensemble-proposers/:proposer_id
func (h *GroupHandler) DeleteEnsembleProposer(c *gin.Context) {
	groupID, ok := parsePositiveIDParam(c, "id")
	if !ok {
		return
	}
	proposerID, ok := parsePositiveIDParam(c, "proposer_id")
	if !ok {
		return
	}
	if err := h.adminService.DeleteEnsembleProposer(c.Request.Context(), groupID, proposerID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Ensemble proposer deleted"})
}

// GetEnsembleConfig returns the group-level ensemble options.
// GET /api/v1/admin/groups/:id/ensemble-config
func (h *GroupHandler) GetEnsembleConfig(c *gin.Context) {
	groupID, ok := parsePositiveIDParam(c, "id")
	if !ok {
		return
	}
	cfg, err := h.adminService.GetEnsembleConfig(c.Request.Context(), groupID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// UpdateEnsembleConfig persists the group-level ensemble options.
// PUT /api/v1/admin/groups/:id/ensemble-config
func (h *GroupHandler) UpdateEnsembleConfig(c *gin.Context) {
	groupID, ok := parsePositiveIDParam(c, "id")
	if !ok {
		return
	}
	var req EnsembleConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}
	cfg, err := h.adminService.UpdateEnsembleConfig(c.Request.Context(), groupID, service.EnsembleConfig{
		SourceGroupIDs:    req.SourceGroupIDs,
		AggregatorEnabled: req.AggregatorEnabled,
		MinProposers:      req.MinProposers,
		TimeoutSeconds:    req.TimeoutSeconds,
		MaxTokens:         req.MaxTokens,
		ExposeMetadata:    req.ExposeMetadata,
		StreamTrace:       req.StreamTrace,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}
