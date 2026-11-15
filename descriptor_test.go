package serverprotection

import (
	"encoding/json"
	"testing"
)

func TestManifestCopyPreservesComponentIdentity(t *testing.T) {
	first := ManifestJSON()
	var value struct {
		ID             string `json:"id"`
		Delivery       string `json:"delivery"`
		DefaultEnabled bool   `json:"defaultEnabled"`
		Version        string `json:"version"`
		Database       struct {
			Tables           []string `json:"tables"`
			MigrationVersion string   `json:"migrationVersion"`
		} `json:"database"`
	}
	if err := json.Unmarshal(first, &value); err != nil {
		t.Fatal(err)
	}
	if value.ID != "server-protection" || value.Delivery != "in-process" || value.DefaultEnabled || value.Version != "2" || value.Database.MigrationVersion != "2" || len(value.Database.Tables) != 30 {
		t.Fatalf("canonical descriptor changed: %+v", value)
	}
	first[0] = 0
	if !json.Valid(ManifestJSON()) {
		t.Fatal("caller changed the embedded descriptor")
	}
}
