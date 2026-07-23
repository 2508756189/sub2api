package securityaudit

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const localSecretScannerID = "credential_exposure"

var secretFindingPatterns = []struct {
	kind    string
	pattern *regexp.Regexp
}{
	{"anthropic_api_key", regexp.MustCompile(`\bsk-ant-[A-Za-z0-9_-]{16,}\b`)},
	{"openai_compatible_key", regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{16,}\b`)},
	{"github_token", regexp.MustCompile(`\b(?:ghp_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,})\b`)},
	{"bearer_token", regexp.MustCompile(`(?i)\bauthorization\s*:\s*bearer\s+[A-Za-z0-9._~+\-/]+=*`)},
	{"assigned_secret", regexp.MustCompile(`(?i)\b(?:api[_-]?key|access[_-]?token|refresh[_-]?token|secret|password)\s*[:=]\s*["']?[A-Za-z0-9._~+\-/=]{12,}`)},
}

type SecretFinding struct {
	Kind string
}

type secretFindingRange struct {
	start    int
	end      int
	kind     string
	priority int
}

func isValidSecretGuardMode(mode SecretGuardMode) bool {
	return mode == "" || mode == SecretGuardModeOff || mode == SecretGuardModeBlock
}

// DetectSecretFindings returns only stable credential categories. It never
// returns, stores, logs, or exposes matched source text.
func DetectSecretFindings(value string) []SecretFinding {
	matches := make([]secretFindingRange, 0)
	for priority, definition := range secretFindingPatterns {
		for _, indexes := range definition.pattern.FindAllStringIndex(value, -1) {
			matches = append(matches, secretFindingRange{start: indexes[0], end: indexes[1], kind: definition.kind, priority: priority})
		}
	}
	if len(matches) == 0 {
		return nil
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].priority == matches[j].priority {
			return matches[i].start < matches[j].start
		}
		return matches[i].priority < matches[j].priority
	})
	accepted := make([]secretFindingRange, 0, len(matches))
	for _, match := range matches {
		overlaps := false
		for _, existing := range accepted {
			if match.start < existing.end && existing.start < match.end {
				overlaps = true
				break
			}
		}
		if overlaps {
			continue
		}
		accepted = append(accepted, match)
	}
	sort.SliceStable(accepted, func(i, j int) bool { return accepted[i].start < accepted[j].start })
	findings := make([]SecretFinding, 0, len(accepted))
	for _, match := range accepted {
		findings = append(findings, SecretFinding{Kind: match.kind})
	}
	return findings
}

func NewSecretGuardResult(findings []SecretFinding) *NormalizedResult {
	kinds := secretFindingKinds(findings)
	evidence := fmt.Sprintf("%d credential candidate(s) detected locally; values withheld", len(findings))
	if len(kinds) > 0 {
		evidence += ": " + strings.Join(kinds, ", ")
	}
	return &NormalizedResult{
		Decision: EventCritical, RiskLevel: RiskCritical, Action: ActionBlock,
		Categories: []string{localSecretScannerID}, MatchedScanners: []string{localSecretScannerID},
		ScannerScores:   map[string]float64{localSecretScannerID: 1},
		ScannerEvidence: map[string]string{localSecretScannerID: evidence},
		ScannerBackend:  "tokenport-local-secret-guard", ScannerVersion: "v1",
		GuardEndpointID: "local-secret-guard", PolicyID: "local_secret_guard", PolicyVersion: 1,
		ChunkTotal: 1, LatencyMS: 0,
	}
}

// SecretGuardSnapshot removes all recoverable prompt material before the
// existing audit repository receives the event. Metadata remains available for
// accountability, while values and prompt hashes are deliberately withheld.
func SecretGuardSnapshot(snapshot PromptSnapshot, findings []SecretFinding) PromptSnapshot {
	kinds := secretFindingKinds(findings)
	snapshot.ScanText = ""
	snapshot.FullPrompt = ""
	snapshot.PromptHash = ""
	snapshot.RedactedPreview = fmt.Sprintf("Local secret guard blocked %d credential candidate(s)", len(findings))
	if len(kinds) > 0 {
		snapshot.RedactedPreview += ": " + strings.Join(kinds, ", ")
	}
	return snapshot
}

func secretFindingKinds(findings []SecretFinding) []string {
	seen := make(map[string]struct{}, len(findings))
	for _, finding := range findings {
		if finding.Kind != "" {
			seen[finding.Kind] = struct{}{}
		}
	}
	kinds := make([]string, 0, len(seen))
	for kind := range seen {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return kinds
}
