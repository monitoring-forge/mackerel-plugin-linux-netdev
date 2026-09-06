package main

import "testing"

func TestValidate(t *testing.T) {
	opt := &Opt{
		IgnoreInterfaces: "eth.*",
	}
	if err := opt.Validate(nil); err != nil {
		t.Errorf("Validate() returned an error: %v", err)
	}
	if opt.ignoreInterfacesRegexp == nil {
		t.Errorf("ignoreInterfacesRegexp is nil after Validate()")
	}
}

func TestValidate_InvalidRegexp(t *testing.T) {
	opt := &Opt{
		IgnoreInterfaces: "eth[",
	}
	if err := opt.Validate(nil); err == nil {
		t.Errorf("Validate() did not return an error for invalid regexp")
	}
}
