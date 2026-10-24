package notice

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"cahut.cloud/internal/mail/mailtest"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
)

var cli = operation.Actor{Kind: operation.Admin, Via: "cli"}

func TestDiskAlmostFull(t *testing.T) {
	ctx := context.Background()
	srv, err := server.Open(ctx, server.Config{DataDir: t.TempDir(), ConsoleURL: func(d string) string { return "https://" + d + "/console/" }})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { srv.Close() })
	if _, err := operation.Do(ctx, srv.Registry, srv.AddDomain, cli, server.AddDomainInput{Domain: "cahut.example.com"}); err != nil {
		t.Fatal(err)
	}
	relay := mailtest.Start(t, mailtest.Options{NoStartTLS: true})
	in, _ := json.Marshal(map[string]any{"host": "127.0.0.1", "port": relay.Port, "security": "none", "from": "cahut@example.com"})
	if _, err := srv.Registry.Invoke(ctx, "set_mail_relay", cli, in); err != nil {
		t.Fatal(err)
	}
	if _, err := operation.Do(ctx, srv.Registry, server.CreateAdmin, cli, server.CreateAdminInput{Handle: "root", Name: "Root", Email: "root@example.org"}); err != nil {
		t.Fatal(err)
	}
	relay.Wait(t, 1) // the activation link

	roomy := func(string) (uint64, uint64, error) { return 100 << 30, 20 << 30, nil }
	full := func(string) (uint64, uint64, error) { return 100 << 30, 5 << 30, nil }
	for _, disk := range []Disk{roomy, full, full} { // the second full disk sends nothing: once a day
		if err := Server(ctx, srv, disk); err != nil {
			t.Fatal(err)
		}
	}
	msgs := relay.Wait(t, 2)
	if b := strings.ReplaceAll(string(msgs[1].Data), "=\r\n", ""); msgs[1].To[0] != "root@example.org" ||
		!strings.Contains(b, "is almost full") || !strings.Contains(b, "is 95% full, with 5.4 GB free") {
		t.Errorf("the notice of the disk:\n%s", b)
	}
	if n := len(relay.Messages()); n != 2 {
		t.Errorf("the relay received %d messages, want 2", n)
	}
}
