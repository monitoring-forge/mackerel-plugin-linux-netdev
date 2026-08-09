package main

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/procfs"
)

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	filename := "stats.json"

	if fileExists(dir, filename) {
		t.Fatalf("file should not exist: %s", filename)
	}

	if err := os.WriteFile(filepath.Join(dir, filename), []byte("{}"), 0o600); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if !fileExists(dir, filename) {
		t.Fatalf("file should exist: %s", filename)
	}
}

func TestStatFile(t *testing.T) {
	name := statFile()
	if !strings.HasPrefix(name, "mackerel-plugin-linux-netdev-") {
		t.Fatalf("unexpected filename prefix: %s", name)
	}

	curUser, err := user.Current()
	if err == nil && curUser != nil {
		if got, want := name, "mackerel-plugin-linux-netdev-"+curUser.Uid; got != want {
			t.Fatalf("unexpected filename, got=%s want=%s", got, want)
		}
	}
}

func TestWriteStatsAndReadStats(t *testing.T) {
	dir := t.TempDir()
	filename := "stats.json"

	metrics := map[string]procfs.NetDevLine{
		"eth0": {
			Name:      "eth0",
			TxErrors:  3,
			RxErrors:  2,
			TxDropped: 1,
			RxDropped: 4,
			TxPackets: 100,
			RxPackets: 200,
		},
	}

	if err := writeStats(dir, filename, metrics); err != nil {
		t.Fatalf("writeStats failed: %v", err)
	}

	timeDiff, gotMetrics, err := readStats(dir, filename)
	if err != nil {
		t.Fatalf("readStats failed: %v", err)
	}

	if timeDiff < 0 || timeDiff > tooOldDuration {
		t.Fatalf("unexpected timeDiff: %v", timeDiff)
	}

	got, ok := gotMetrics["eth0"]
	if !ok {
		t.Fatalf("missing interface eth0")
	}

	if got.TxErrors != metrics["eth0"].TxErrors ||
		got.RxErrors != metrics["eth0"].RxErrors ||
		got.TxDropped != metrics["eth0"].TxDropped ||
		got.RxDropped != metrics["eth0"].RxDropped ||
		got.TxPackets != metrics["eth0"].TxPackets ||
		got.RxPackets != metrics["eth0"].RxPackets {
		t.Fatalf("decoded metrics do not match written metrics")
	}
}

func TestReadStatsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	filename := "stats.json"

	if err := os.WriteFile(filepath.Join(dir, filename), []byte("{invalid"), 0o600); err != nil {
		t.Fatalf("failed to create invalid json file: %v", err)
	}

	if _, _, err := readStats(dir, filename); err == nil {
		t.Fatalf("expected readStats to fail with invalid JSON")
	}
}

func TestReadStatsZeroTime(t *testing.T) {
	dir := t.TempDir()
	filename := "stats.json"

	st := stats{
		Interfaces: map[string]procfs.NetDevLine{},
		Time:       0,
	}

	f, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := json.NewEncoder(f).Encode(st); err != nil {
		_ = f.Close()
		t.Fatalf("failed to encode stats: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close file: %v", err)
	}

	if _, _, err := readStats(dir, filename); err == nil {
		t.Fatalf("expected readStats to fail with zero time")
	}
}

func TestReadStatsTooLongDuration(t *testing.T) {
	dir := t.TempDir()
	filename := "stats.json"

	oldTooOldDuration := tooOldDuration
	tooOldDuration = 1
	t.Cleanup(func() {
		tooOldDuration = oldTooOldDuration
	})

	st := stats{
		Interfaces: map[string]procfs.NetDevLine{},
		Time:       time.Now().Unix() - 10,
	}

	f, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := json.NewEncoder(f).Encode(st); err != nil {
		_ = f.Close()
		t.Fatalf("failed to encode stats: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close file: %v", err)
	}

	if _, _, err := readStats(dir, filename); err == nil {
		t.Fatalf("expected readStats to fail for too old stats")
	}
}
