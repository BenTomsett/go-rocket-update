package provider_test

import (
	"path/filepath"
	"testing"

	provider "github.com/mouuff/easy-update/provider"
)

func TestProviderZip(t *testing.T) {
	p := provider.NewProviderZip(filepath.Join("testdata", "Allum1.zip"))
	if err := p.Open(); err != nil {
		t.Error(err)
	}
	defer p.Close()
	/*
		err := Provider(p)
		if err != nil {
			t.Error(err)
		}
	*/
}