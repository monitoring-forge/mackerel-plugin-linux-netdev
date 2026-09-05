package main

import (
	"fmt"
	"os/user"
	"time"

	"github.com/monitoring-forge/saferio"
	"github.com/prometheus/procfs"
)

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

func writeStats(dir, filename string, st map[string]procfs.NetDevLine, t ...time.Time) error {
	s := stats{
		Interfaces: st,
		Time:       time.Now().Unix(),
	}
	if len(t) > 0 {
		s.Time = t[0].Unix()
	}
	return saferio.WriteJSON(dir, filename, s)
}

func readStats(dir, filename string) (float64, map[string]procfs.NetDevLine, error) {
	st := stats{}
	if err := saferio.ReadJSON(dir, filename, &st); err != nil {
		return 0, nil, err
	}

	if st.Time == 0 {
		return 0, nil, fmt.Errorf("failed to get previous time")
	}
	n := time.Now().Unix()
	timeDiff := float64(n - st.Time)
	if timeDiff <= 0 {
		return 0, nil, fmt.Errorf("invalid elapsed time")
	}
	if timeDiff > tooOldDuration {
		return 0, nil, fmt.Errorf("too long duration")
	}

	return timeDiff, st.Interfaces, nil
}
