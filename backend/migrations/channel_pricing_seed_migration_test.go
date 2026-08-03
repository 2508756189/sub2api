package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelPricingSeedMigrationsGuardMissingChannels(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		markers []string
	}{
		{
			name: "TokenRhythm",
			file: "194_seed_tokenrhythm_model_pricing.sql",
			markers: []string{
				"JOIN channels AS c ON c.id = s.channel_id",
				"SELECT\n    c.id",
			},
		},
		{
			name: "Grok Composer",
			file: "195_seed_grok_composer_pricing.sql",
			markers: []string{
				"FROM channels AS c",
				"WHERE c.id = 8",
				"SELECT\n    c.id",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := FS.ReadFile(tt.file)
			require.NoError(t, err)
			sql := string(content)
			for _, marker := range tt.markers {
				require.Contains(t, sql, marker)
			}
		})
	}
}
