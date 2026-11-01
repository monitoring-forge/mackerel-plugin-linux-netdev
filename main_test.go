package main

import (
	"testing"

	"github.com/monitoring-forge/flagrun"
)

func TestValidate(t *testing.T) {
	re := flagrun.Re{}
	err := re.UnmarshalFlag("eth.*")
	if err != nil {
		t.Fatalf("Failed to unmarshal flag: %v", err)
	}
	opt := &Opt{
		IgnoreInterfaces: re,
	}
	if err := opt.Validator(nil); err != nil {
		t.Errorf("Validator() returned an error: %v", err)
	}
	if opt.workDir == "" {
		t.Errorf("workDir is not set after Validator()")
	}
}
