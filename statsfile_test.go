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
	"github.com/stretchr/testify/require"
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

	if err := writeStats(dir, filename, metrics, time.Now().Add(-5*time.Second)); err != nil {
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

func writeJSONToFile(dir, filename string, data any) error {
	file, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

func TestReadStatsZeroTime(t *testing.T) {
	dir := t.TempDir()
	filename := "stats.json"

	st := stats{
		Interfaces: map[string]procfs.NetDevLine{},
		Time:       0,
	}
	err := writeJSONToFile(dir, filename, st)
	require.NoError(t, err, "failed to write stats with zero time")

	_, _, err = readStats(dir, filename)
	require.Error(t, err, "expected readStats to fail with zero time")
	require.Contains(t, err.Error(), "failed to get previous time", "unexpected error message")
}

func TestReadStatsTooLongDuration(t *testing.T) {
	dir := t.TempDir()
	filename := "stats.json"

	st := stats{
		Interfaces: map[string]procfs.NetDevLine{},
		Time:       time.Now().Unix() - int64(tooOldDuration) - 10,
	}

	err := writeJSONToFile(dir, filename, st)
	require.NoError(t, err, "failed to write stats with zero time")

	_, _, err = readStats(dir, filename)
	require.Error(t, err, "expected readStats to fail for too old stats")
	require.Contains(t, err.Error(), "too long duration", "unexpected error message")
}
