package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/window"
)

func TestRun(t *testing.T) {
	file := filepath.Join(t.TempDir(), "api.gen.ts")
	if err := run(t.Context(), file); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		header,
		"  publish_window: { input: PublishInput; output: Window };\n",
		"  list_rooms: { input: ListRoomsInput; output: RoomList };\n",
		"  list_domains: { input: ListDomainsInput; output: DomainList };\n",
		"export interface ServerOperations {",
		"  room?: string;\n",
		"  wrong_handle_or_password: null;\n",
		"  password_too_short: Length;\n",
		"  session_ended: null;\n",
		"  too_many_attempts_from_address: null;\n",
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("the types lack %q", want)
		}
	}
}

// sample holds what the generator meets in the operations.
type sample struct {
	Name     string            `json:"name" jsonschema:"a name"`
	Count    int64             `json:"count"`
	Tags     []string          `json:"tags,omitempty"`
	At       time.Time         `json:"at"`
	Seen     *time.Time        `json:"seen,omitempty"`
	Labels   map[string]bool   `json:"labels"`
	Window   window.Window     `json:"window"`
	Lists    [2]room.List      `json:"lists"`
	Inline   struct{ N int }   `json:"inline"`
	Secret   string            `json:"-"`
	Pointers *window.List      `json:"pointers"`
	Nested   map[string][]byte `json:"nested"`
}

// roomName fills a refusal of the sample.
type roomName struct {
	Name string `json:"name"`
}

func TestGenerate(t *testing.T) {
	ops := []operation.Info{{Name: "try", Description: "Try a sample */ out.", Input: reflect.TypeFor[sample](), Output: reflect.TypeFor[struct{}]()}}
	refusals := []operation.RefusalInfo{
		{Code: "plain", Message: "a plain */ refusal"},
		{Code: "filled", Message: "the room {name} is full", Values: reflect.TypeFor[roomName]()},
	}
	got := generate(ops, nil, refusals)
	for _, want := range []string{
		"export interface sample {\n  /** a name */\n  name: string;\n  count: number;\n  tags?: string[];\n  at: string;\n  seen?: string;\n" +
			"  labels: Record<string, boolean>;\n  window: Window;\n  lists: RoomList[];\n  inline: { N: number; };\n  pointers: WindowList;\n" +
			"  nested: Record<string, number[]>;\n}\n",
		"export interface RoomList {",
		"export interface WindowList {",
		"  /** Try a sample * / out. */\n  try: { input: sample; output: Record<string, never> };\n",
		"export interface roomName {\n  name: string;\n}\n",
		"export interface Refusals {\n  /** a plain * / refusal */\n  plain: null;\n  /** the room {name} is full */\n  filled: roomName;\n}\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the types lack %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "secret") || strings.Contains(got, "Secret") {
		t.Errorf("the types hold a field that JSON skips:\n%s", got)
	}
}
