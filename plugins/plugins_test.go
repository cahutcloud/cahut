package plugins

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The marketplace of the repository, the manifest of the plugin and its MCP
// server, as Claude Code reads them. claude plugin validate checks them in
// full; this checks what Cahut depends on.
func TestPlugin(t *testing.T) {
	var market struct {
		Name    string `json:"name"`
		Plugins []struct {
			Name   string `json:"name"`
			Source string `json:"source"`
		} `json:"plugins"`
	}
	read(t, "../.claude-plugin/marketplace.json", &market)
	if market.Name != "cahut" || len(market.Plugins) != 1 || market.Plugins[0].Name != "cahut" || market.Plugins[0].Source != "./plugins/cahut" {
		t.Fatalf("marketplace = %+v", market)
	}
	var manifest struct {
		Name       string                    `json:"name"`
		Version    string                    `json:"version"`
		License    string                    `json:"license"`
		UserConfig map[string]map[string]any `json:"userConfig"`
	}
	read(t, "cahut/.claude-plugin/plugin.json", &manifest)
	if manifest.Name != "cahut" || manifest.License != "Apache-2.0" || manifest.UserConfig["domain"]["required"] != true ||
		!regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(manifest.Version) {
		t.Errorf("manifest = %+v", manifest)
	}
	var mcp struct {
		MCPServers map[string]struct {
			Type string `json:"type"`
			URL  string `json:"url"`
		} `json:"mcpServers"`
	}
	read(t, "cahut/.mcp.json", &mcp)
	if s := mcp.MCPServers["cahut"]; s.Type != "http" || s.URL != "https://${user_config.domain}/mcp" {
		t.Errorf("MCP server of the plugin = %+v", mcp.MCPServers)
	}
	if _, err := os.Stat("cahut/LICENSE"); err != nil {
		t.Error("the plugin has no LICENSE")
	}
}

func read(t *testing.T, path string, v any) {
	t.Helper()
	b, err := os.ReadFile(filepath.FromSlash(path))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}

// The zip holds the folder of the skill, as the plugin has it.
func TestSkillZip(t *testing.T) {
	b, err := SkillZip()
	if err != nil {
		t.Fatal(err)
	}
	again, _ := SkillZip()
	if !bytes.Equal(b, again) {
		t.Error("SkillZip changes from a call to the next")
	}
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, f := range r.File {
		names[f.Name] = true
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		got, _ := io.ReadAll(rc)
		rc.Close()
		want, err := os.ReadFile(filepath.Join("cahut", "skills", filepath.FromSlash(f.Name)))
		if err != nil || !bytes.Equal(got, want) {
			t.Errorf("%s in the zip differs from the plugin's: %v", f.Name, err)
		}
	}
	if !names["cahut/SKILL.md"] || !names["cahut/references/formats.md"] {
		t.Errorf("the zip holds %v", names)
	}
	skill, _ := os.ReadFile("cahut/skills/cahut/SKILL.md")
	front := regexp.MustCompile(`(?s)^---\nname: cahut\ndescription: (.+?)\n---\n`).FindSubmatch(skill)
	if front == nil || len(front[1]) > 1024 || !strings.Contains(string(front[1]), "Cahut") {
		t.Errorf("the front matter of SKILL.md does not name the skill or describe it in 1024 characters: %.200s", skill)
	}
}
