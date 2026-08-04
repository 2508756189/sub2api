package service

import (
	"context"
	"strings"
)

// EnsembleCostEstimate is the cost shown in ensemble_metadata for one member
// call. It is an estimate from the same pricing resolver used by normal billing;
// the normal gateway remains responsible for recording the actual charge.
type EnsembleCostEstimate struct {
	Cost   float64
	Source string
}

// EnsembleCostEstimator exposes the existing billing calculation to the
// Ensemble handler without coupling the handler to billing implementation
// details.
type EnsembleCostEstimator interface {
	EstimateEnsembleCost(ctx context.Context, groupID int64, model, platform string, tokens UsageTokens) (*EnsembleCostEstimate, error)
}

// BillingEnsembleCostEstimator calculates member costs using the pricing
// configured for the Ensemble group's concrete target platform.
type BillingEnsembleCostEstimator struct {
	billingService *BillingService
	resolver       *ModelPricingResolver
}

func NewBillingEnsembleCostEstimator(billingService *BillingService, resolver *ModelPricingResolver) *BillingEnsembleCostEstimator {
	return &BillingEnsembleCostEstimator{
		billingService: billingService,
		resolver:       resolver,
	}
}

func (s *BillingEnsembleCostEstimator) EstimateEnsembleCost(
	ctx context.Context,
	groupID int64,
	model string,
	platform string,
	tokens UsageTokens,
) (*EnsembleCostEstimate, error) {
	if s == nil || s.billingService == nil || strings.TrimSpace(model) == "" {
		return nil, nil
	}

	platform = strings.ToLower(strings.TrimSpace(platform))
	ctx = WithResolvedTargetPlatform(ctx, platform)

	if s.resolver == nil {
		breakdown, err := s.billingService.CalculateCost(model, tokens, 1)
		if err != nil || breakdown == nil {
			return nil, err
		}
		return &EnsembleCostEstimate{Cost: breakdown.ActualCost, Source: PricingSourceFallback}, nil
	}

	resolvedGroupID := groupID
	resolved := s.resolver.Resolve(ctx, PricingInput{
		Model:   model,
		GroupID: &resolvedGroupID,
	})
	breakdown, err := s.billingService.CalculateCostUnified(CostInput{
		Ctx:            ctx,
		Model:          model,
		GroupID:        &resolvedGroupID,
		Tokens:         tokens,
		RequestCount:   1,
		RateMultiplier: 1,
		Resolver:       s.resolver,
		Resolved:       resolved,
	})
	if err != nil || breakdown == nil {
		return nil, err
	}

	source := resolved.Source
	if source == "" {
		source = PricingSourceFallback
	}
	return &EnsembleCostEstimate{Cost: breakdown.ActualCost, Source: source}, nil
}
