package service

import (
	"context"
	"fmt"
	"strings"
)

// requireEnsembleGroup ensures the target group exists and is an ensemble-platform
// group. Ensemble members and config are only meaningful for such groups.
func (s *adminServiceImpl) requireEnsembleGroup(ctx context.Context, groupID int64) error {
	group, err := s.groupRepo.GetByIDLite(ctx, groupID)
	if err != nil {
		return err
	}
	if group.Platform != PlatformEnsemble {
		return fmt.Errorf("group %d is not an ensemble group", groupID)
	}
	return nil
}

// ListEnsembleProposers returns every member (proposers + aggregator) of an
// ensemble group, including disabled ones so the admin UI can render them.
func (s *adminServiceImpl) ListEnsembleProposers(ctx context.Context, groupID int64) ([]EnsembleProposer, error) {
	if err := s.requireEnsembleGroup(ctx, groupID); err != nil {
		return nil, err
	}
	if s.ensembleProposerRepo == nil {
		return nil, fmt.Errorf("ensemble proposer repository is not configured")
	}
	return s.ensembleProposerRepo.ListByGroup(ctx, groupID, true)
}

// CreateEnsembleProposer adds one member to an ensemble group.
func (s *adminServiceImpl) CreateEnsembleProposer(ctx context.Context, groupID int64, input EnsembleProposerInput) (*EnsembleProposer, error) {
	if err := s.requireEnsembleGroup(ctx, groupID); err != nil {
		return nil, err
	}
	if s.ensembleProposerRepo == nil {
		return nil, fmt.Errorf("ensemble proposer repository is not configured")
	}
	proposer, err := ensembleProposerFromInput(groupID, input)
	if err != nil {
		return nil, err
	}
	if err := s.validateEnsembleMemberLimit(ctx, groupID, 0, proposer.Role); err != nil {
		return nil, err
	}
	if err := s.validateEnsembleModel(ctx, groupID, proposer.Model, proposer.Platform); err != nil {
		return nil, err
	}
	if err := s.ensembleProposerRepo.Create(ctx, proposer); err != nil {
		return nil, err
	}
	return proposer, nil
}

// UpdateEnsembleProposer mutates one member of an ensemble group.
func (s *adminServiceImpl) UpdateEnsembleProposer(ctx context.Context, groupID, proposerID int64, input EnsembleProposerInput) (*EnsembleProposer, error) {
	if err := s.requireEnsembleGroup(ctx, groupID); err != nil {
		return nil, err
	}
	if s.ensembleProposerRepo == nil {
		return nil, fmt.Errorf("ensemble proposer repository is not configured")
	}
	if ok, err := s.ensembleProposerBelongsToGroup(ctx, groupID, proposerID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrEnsembleProposerNotFound
	}
	proposer, err := ensembleProposerFromInput(groupID, input)
	if err != nil {
		return nil, err
	}
	if err := s.validateEnsembleMemberLimit(ctx, groupID, proposerID, proposer.Role); err != nil {
		return nil, err
	}
	if err := s.validateEnsembleModel(ctx, groupID, proposer.Model, proposer.Platform); err != nil {
		return nil, err
	}
	if err := s.clampEnsembleMinimumForChange(ctx, groupID, proposerID, proposer); err != nil {
		return nil, err
	}
	proposer.ID = proposerID
	if err := s.ensembleProposerRepo.Update(ctx, proposer); err != nil {
		return nil, err
	}
	return proposer, nil
}

// DeleteEnsembleProposer removes one member from an ensemble group.
func (s *adminServiceImpl) DeleteEnsembleProposer(ctx context.Context, groupID, proposerID int64) error {
	if err := s.requireEnsembleGroup(ctx, groupID); err != nil {
		return err
	}
	if s.ensembleProposerRepo == nil {
		return fmt.Errorf("ensemble proposer repository is not configured")
	}
	if ok, err := s.ensembleProposerBelongsToGroup(ctx, groupID, proposerID); err != nil {
		return err
	} else if !ok {
		return ErrEnsembleProposerNotFound
	}
	if err := s.clampEnsembleMinimumForChange(ctx, groupID, proposerID, nil); err != nil {
		return err
	}
	return s.ensembleProposerRepo.Delete(ctx, proposerID)
}

// GetEnsembleConfig returns the group-level ensemble options.
func (s *adminServiceImpl) GetEnsembleConfig(ctx context.Context, groupID int64) (EnsembleConfig, error) {
	if err := s.requireEnsembleGroup(ctx, groupID); err != nil {
		return EnsembleConfig{}, err
	}
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return EnsembleConfig{}, err
	}
	return normalizeEnsembleConfig(group.EnsembleConfig), nil
}

// UpdateEnsembleConfig persists the group-level ensemble options.
func (s *adminServiceImpl) UpdateEnsembleConfig(ctx context.Context, groupID int64, config EnsembleConfig) (EnsembleConfig, error) {
	if err := s.requireEnsembleGroup(ctx, groupID); err != nil {
		return EnsembleConfig{}, err
	}
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return EnsembleConfig{}, err
	}
	group.EnsembleConfig = normalizeEnsembleConfig(config)
	if s.ensembleProposerRepo == nil {
		return EnsembleConfig{}, ErrEnsembleRuntimeUnavailable
	}
	members, err := s.ensembleProposerRepo.ListByGroup(ctx, groupID, false)
	if err != nil {
		return EnsembleConfig{}, err
	}
	enabledProposers := 0
	for _, member := range members {
		if member.Enabled && member.Role == EnsembleRoleProposer {
			enabledProposers++
		}
	}
	// A newly created group may be configured before its members are added. The
	// runtime will reject calls until a proposer exists, but the admin workflow
	// must be able to save the config first.
	if enabledProposers > 0 && group.EnsembleConfig.MinProposers > enabledProposers {
		return EnsembleConfig{}, fmt.Errorf("min_proposers %d exceeds enabled proposer count %d", group.EnsembleConfig.MinProposers, enabledProposers)
	}
	if err := s.groupRepo.Update(ctx, group); err != nil {
		return EnsembleConfig{}, err
	}
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByGroupID(ctx, groupID)
	}
	return group.EnsembleConfig, nil
}

func (s *adminServiceImpl) ensembleProposerBelongsToGroup(ctx context.Context, groupID, proposerID int64) (bool, error) {
	proposers, err := s.ensembleProposerRepo.ListByGroup(ctx, groupID, true)
	if err != nil {
		return false, err
	}
	for i := range proposers {
		if proposers[i].ID == proposerID {
			return true, nil
		}
	}
	return false, nil
}

// clampEnsembleMinimumForChange keeps a saved minimum satisfiable when an
// admin disables, re-roles, or deletes a proposer. The group is updated before
// the member mutation so a failed member write cannot leave an impossible
// runtime configuration.
func (s *adminServiceImpl) clampEnsembleMinimumForChange(ctx context.Context, groupID, targetID int64, replacement *EnsembleProposer) error {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	members, err := s.ensembleProposerRepo.ListByGroup(ctx, groupID, true)
	if err != nil {
		return err
	}
	enabledProposers := 0
	for _, member := range members {
		if member.ID == targetID {
			if replacement != nil && replacement.Enabled && replacement.Role == EnsembleRoleProposer {
				enabledProposers++
			}
			continue
		}
		if member.Enabled && member.Role == EnsembleRoleProposer {
			enabledProposers++
		}
	}
	if enabledProposers == 0 || group.EnsembleConfig.MinProposers <= enabledProposers {
		return nil
	}
	group.EnsembleConfig.MinProposers = enabledProposers
	if err := s.groupRepo.Update(ctx, group); err != nil {
		return err
	}
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByGroupID(ctx, groupID)
	}
	return nil
}

func (s *adminServiceImpl) validateEnsembleMemberLimit(ctx context.Context, groupID, currentID int64, role string) error {
	members, err := s.ensembleProposerRepo.ListByGroup(ctx, groupID, true)
	if err != nil {
		return err
	}
	if role == EnsembleRoleProposer {
		count := 0
		for _, member := range members {
			if member.ID != currentID && member.Role == EnsembleRoleProposer {
				count++
			}
		}
		if count >= MaxEnsembleProposers {
			return fmt.Errorf("ensemble proposer count cannot exceed %d", MaxEnsembleProposers)
		}
		return nil
	}
	for _, member := range members {
		if member.ID != currentID && member.Role == EnsembleRoleAggregator {
			return fmt.Errorf("an ensemble group can have only one aggregator")
		}
	}
	return nil
}

func ensembleProposerFromInput(groupID int64, input EnsembleProposerInput) (*EnsembleProposer, error) {
	input = normalizeEnsembleProposerInput(input)
	if input.Model == "" {
		return nil, fmt.Errorf("model is required")
	}
	return &EnsembleProposer{
		GroupID:  groupID,
		Role:     input.Role,
		Model:    input.Model,
		Platform: input.Platform,
		Priority: input.Priority,
		Enabled:  input.Enabled,
	}, nil
}

// normalizeEnsembleConfig applies safe defaults for zero values.
func normalizeEnsembleConfig(cfg EnsembleConfig) EnsembleConfig {
	if cfg.MinProposers < 1 {
		cfg.MinProposers = 1
	}
	if cfg.TimeoutSeconds < 0 {
		cfg.TimeoutSeconds = 0
	}
	if cfg.MaxTokens < 0 {
		cfg.MaxTokens = 0
	}
	return cfg
}

func (s *adminServiceImpl) validateEnsembleModel(ctx context.Context, groupID int64, model, platform string) error {
	if s.accountRepo == nil {
		return ErrEnsembleRuntimeUnavailable
	}
	accounts, err := s.accountRepo.ListSchedulableByGroupID(ctx, groupID)
	if err != nil {
		return err
	}
	model = strings.TrimSpace(model)
	platform = strings.ToLower(strings.TrimSpace(platform))
	if platform != "" && !isConcreteRequestPlatform(platform) {
		return fmt.Errorf("%w: unsupported platform %s", ErrEnsembleModelUnavailable, platform)
	}
	for _, account := range accounts {
		if !account.IsModelSupported(model) {
			continue
		}
		targetPlatform := platform
		if targetPlatform == "" {
			targetPlatform, _ = DetectModelPlatform(model)
		}
		if targetPlatform != "" && !ensembleAccountCanServePlatform(account, targetPlatform) {
			continue
		}
		return nil
	}
	return fmt.Errorf("%w: %s", ErrEnsembleModelUnavailable, model)
}

func ensembleAccountCanServePlatform(account Account, targetPlatform string) bool {
	if account.Platform == targetPlatform {
		return true
	}
	if account.Platform != PlatformAntigravity || !account.IsMixedSchedulingEnabled() {
		return false
	}
	return targetPlatform == PlatformAnthropic || targetPlatform == PlatformGemini
}
