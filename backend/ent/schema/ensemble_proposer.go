package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
	"github.com/Wei-Shaw/sub2api/internal/domain"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// EnsembleProposer holds one member (proposer or aggregator) of an ensemble group.
// Members reference models served by the ensemble group's own bound accounts —
// there is no cross-group routing (in-group aggregation only).
type EnsembleProposer struct {
	ent.Schema
}

func (EnsembleProposer) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ensemble_proposers"},
	}
}

func (EnsembleProposer) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (EnsembleProposer) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("group_id"),
		field.String("role").
			MaxLen(20).
			Default(domain.EnsembleRoleProposer).
			Comment("proposer or aggregator."),
		field.String("model").
			MaxLen(200).
			NotEmpty().
			Comment("Model identifier served by the group's own accounts."),
		field.String("platform").
			MaxLen(32).
			Default("").
			Comment("Concrete upstream platform used to serve this model."),
		field.Int("priority").
			Default(100).
			Comment("Lower values are called/displayed first among proposers."),
		field.Bool("enabled").
			Default(true),
		// Vision marks a member that accepts image inputs (image_url /
		// input_image parts). When the request carries images, members with
		// vision=false are skipped instead of being called and failing, so a
		// non-vision model never breaks an image request that a vision-capable
		// member could serve.
		field.Bool("vision").
			Default(true).
			Comment("whether the member model accepts image inputs."),
	}
}

func (EnsembleProposer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("group", Group.Type).
			Unique().
			Required().
			Field("group_id"),
	}
}

func (EnsembleProposer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_id"),
		index.Fields("group_id", "enabled"),
		index.Fields("group_id", "role"),
		index.Fields("deleted_at"),
		index.Fields("priority"),
	}
}
