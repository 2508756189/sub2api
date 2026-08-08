package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/ensembleproposer"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type ensembleProposerRepository struct {
	client *dbent.Client
}

// NewEnsembleProposerRepository wires the ent client into the ensemble proposer
// repository contract consumed by the service layer.
func NewEnsembleProposerRepository(client *dbent.Client) service.EnsembleProposerRepository {
	return &ensembleProposerRepository{client: client}
}

func (r *ensembleProposerRepository) ListByGroup(ctx context.Context, groupID int64, includeDisabled bool) ([]service.EnsembleProposer, error) {
	q := clientFromContext(ctx, r.client).EnsembleProposer.Query().
		Where(ensembleproposer.GroupIDEQ(groupID)).
		Order(
			dbent.Asc(ensembleproposer.FieldPriority),
			dbent.Asc(ensembleproposer.FieldID),
		)
	if !includeDisabled {
		q = q.Where(ensembleproposer.EnabledEQ(true))
	}
	rows, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]service.EnsembleProposer, 0, len(rows))
	for _, row := range rows {
		out = append(out, *ensembleProposerEntityToService(row))
	}
	return out, nil
}

func (r *ensembleProposerRepository) Create(ctx context.Context, proposer *service.EnsembleProposer) error {
	if proposer == nil {
		return service.ErrEnsembleProposerNotFound
	}
	created, err := clientFromContext(ctx, r.client).EnsembleProposer.Create().
		SetGroupID(proposer.GroupID).
		SetRole(proposer.Role).
		SetModel(proposer.Model).
		SetPlatform(proposer.Platform).
		SetPriority(proposer.Priority).
		SetEnabled(proposer.Enabled).
		Save(ctx)
	if err != nil {
		return translatePersistenceError(err, nil, service.ErrEnsembleProposerExists)
	}
	*proposer = *ensembleProposerEntityToService(created)
	return nil
}

func (r *ensembleProposerRepository) Update(ctx context.Context, proposer *service.EnsembleProposer) error {
	if proposer == nil {
		return service.ErrEnsembleProposerNotFound
	}
	updated, err := clientFromContext(ctx, r.client).EnsembleProposer.UpdateOneID(proposer.ID).
		SetRole(proposer.Role).
		SetModel(proposer.Model).
		SetPlatform(proposer.Platform).
		SetPriority(proposer.Priority).
		SetEnabled(proposer.Enabled).
		Save(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrEnsembleProposerNotFound, service.ErrEnsembleProposerExists)
	}
	*proposer = *ensembleProposerEntityToService(updated)
	return nil
}

func (r *ensembleProposerRepository) Delete(ctx context.Context, id int64) error {
	err := clientFromContext(ctx, r.client).EnsembleProposer.DeleteOneID(id).Exec(ctx)
	return translatePersistenceError(err, service.ErrEnsembleProposerNotFound, nil)
}

func ensembleProposerEntityToService(row *dbent.EnsembleProposer) *service.EnsembleProposer {
	if row == nil {
		return nil
	}
	return &service.EnsembleProposer{
		ID:        row.ID,
		GroupID:   row.GroupID,
		Role:      row.Role,
		Model:     row.Model,
		Platform:  row.Platform,
		Priority:  row.Priority,
		Enabled:   row.Enabled,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
