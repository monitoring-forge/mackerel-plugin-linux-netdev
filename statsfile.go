package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/prometheus/procfs"
)

func fileExists(dir, filename string) bool {
	_, err := os.Stat(filepath.Join(dir, filename))
	return err == nil
}

func statFile() string {
	curUser, _ := user.Current()
	uid := "0"
	if curUser != nil {
		uid = curUser.Uid
	}
	path := fmt.Sprintf("mackerel-plugin-linux-netdev-%s", uid)
	return path
}

type stats struct {
	Interfaces map[string]procfs.NetDevLine `json:"interfaces"`
	Time       int64                        `json:"time"`
}

func writeStats(dir, filename string, st map[string]procfs.NetDevLine) error {
	newFile, err := os.CreateTemp(dir, "mackerel-plugin-linux-netdev-")
	if err != nil {
		return err
	}
	defer func() {
		errRemove := os.Remove(newFile.Name())
		if errRemove != nil && !os.IsNotExist(errRemove) {
			log.Printf("Failed to remove temporary file: %s, error: %v", newFile.Name(), errRemove)
		}
	}()

	je := json.NewEncoder(newFile)
	err = je.Encode(stats{
		Interfaces: st,
		Time:       time.Now().Unix(),
	})
	if err != nil {
		_ = newFile.Close()
		return err
	}

	err = newFile.Close()
	if err != nil {
		return err
	}

	return os.Rename(newFile.Name(), filepath.Join(dir, filename))
}

func readStats(dir, filename string) (float64, map[string]procfs.NetDevLine, error) {
	st := stats{}
	file, err := openRD(filepath.Join(dir, filename))
	if err != nil {
		return 0, nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&st)
	if err != nil {
		return 0, nil, err
	}

	if st.Time == 0 {
		return 0, nil, fmt.Errorf("failed to get previous time")
	}
	n := time.Now().Unix()
	timeDiff := float64(n - st.Time)
	if timeDiff > tooOldDuration {
		return 0, nil, fmt.Errorf("too long duration")
	}

	return timeDiff, st.Interfaces, nil
}
