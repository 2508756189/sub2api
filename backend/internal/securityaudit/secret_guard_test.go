package securityaudit

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetectSecretFindingsReturnsCategoriesWithoutValues(t *testing.T) {
	canary := "sk-demo-abcdefghijklmnopqrstuvwx"
	findings := DetectSecretFindings("export API_KEY=" + canary + " and Authorization: Bearer abcdefghijklmnop")

	require.Len(t, findings, 2)
	require.Equal(t, []SecretFinding{{Kind: "openai_compatible_key"}, {Kind: "bearer_token"}}, findings)
	for _, finding := range findings {
		require.NotContains(t, finding.Kind, canary)
	}
}

func TestSecretGuardSnapshotAndResultNeverRetainMatchedValue(t *testing.T) {
	canary := "sk-demo-abcdefghijklmnopqrstuvwx"
	snapshot := PromptSnapshot{
		PromptHash: canary, RedactedPreview: canary, FullPrompt: canary, ScanText: canary,
	}
	findings := DetectSecretFindings(canary)
	protected := SecretGuardSnapshot(snapshot, findings)
	result := NewSecretGuardResult(findings)

	require.Empty(t, protected.PromptHash)
	require.Empty(t, protected.FullPrompt)
	require.Empty(t, protected.ScanText)
	require.NotContains(t, protected.RedactedPreview, canary)
	require.NotContains(t, result.ScannerEvidence[localSecretScannerID], canary)
	require.Equal(t, EventCritical, result.Decision)
	require.Equal(t, ActionBlock, result.Action)
	require.False(t, strings.Contains(result.ScannerEvidence[localSecretScannerID], canary))
}
