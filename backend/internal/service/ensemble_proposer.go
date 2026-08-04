package service

import (
	"context"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	EnsembleRoleProposer      = domain.EnsembleRoleProposer
	EnsembleRoleAggregator    = domain.EnsembleRoleAggregator
	MaxEnsembleProposers      = 6
	MaxEnsembleTimeoutSeconds = 600
)

var (
	ErrEnsembleProposerNotFound = infraerrors.NotFound("ENSEMBLE_PROPOSER_NOT_FOUND", "ensemble proposer not found")
	ErrEnsembleProposerExists   = infraerrors.Conflict("ENSEMBLE_PROPOSER_EXISTS", "ensemble proposer already exists")

	// ErrEnsembleRuntimeUnavailable means the ensemble runtime dependencies were
	// not wired; the request cannot be fanned out.
	ErrEnsembleRuntimeUnavailable = infraerrors.InternalServer("ENSEMBLE_RUNTIME_UNAVAILABLE", "ensemble runtime is not available")
	// ErrEnsembleNoProposers means the group has no enabled proposer members.
	ErrEnsembleNoProposers = infraerrors.BadRequest("ENSEMBLE_NO_PROPOSERS", "ensemble group has no enabled proposer models")
	// ErrEnsembleInsufficientProposers means fewer proposers succeeded than min_proposers requires.
	ErrEnsembleInsufficientProposers = infraerrors.ServiceUnavailable("ENSEMBLE_INSUFFICIENT_PROPOSERS", "not enough ensemble proposers succeeded")
	// ErrEnsembleModelUnavailable means no schedulable account bound to the group
	// can serve the configured model.
	ErrEnsembleModelUnavailable = infraerrors.BadRequest("ENSEMBLE_MODEL_UNAVAILABLE", "ensemble model is not available from a bound account")
)

// EnsembleProposer is one member of an ensemble group. role="proposer" members
// are called in parallel; a single role="aggregator" member (optional) synthesizes
// the final answer. Models are served by the group's own bound accounts.
type EnsembleProposer struct {
	ID        int64     `json:"id"`
	GroupID   int64     `json:"group_id"`
	Role      string    `json:"role"`
	Model     string    `json:"model"`
	Platform  string    `json:"platform"`
	Priority  int       `json:"priority"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EnsembleProposerInput is the mutable payload for create/update.
type EnsembleProposerInput struct {
	Role     string
	Model    string
	Platform string
	Priority int
	Enabled  bool
}

type EnsembleProposerRepository interface {
	ListByGroup(ctx context.Context, groupID int64, includeDisabled bool) ([]EnsembleProposer, error)
	Create(ctx context.Context, proposer *EnsembleProposer) error
	Update(ctx context.Context, proposer *EnsembleProposer) error
	Delete(ctx context.Context, id int64) error
	DeleteByGroup(ctx context.Context, groupID int64) error
}

func normalizeEnsembleRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case EnsembleRoleAggregator:
		return EnsembleRoleAggregator
	default:
		return EnsembleRoleProposer
	}
}

func normalizeEnsembleProposerInput(input EnsembleProposerInput) EnsembleProposerInput {
	input.Role = normalizeEnsembleRole(input.Role)
	input.Model = strings.TrimSpace(input.Model)
	input.Platform = strings.ToLower(strings.TrimSpace(input.Platform))
	if input.Priority == 0 {
		input.Priority = 100
	}
	return input
}
