package provider_test

import (
	"fmt"
	"runtime"
	"testing"

	provider "github.com/mouuff/go-rocket-update/pkg/provider"
)

func TestProviderGitlab(t *testing.T) {
	p := &provider.Gitlab{
		ProjectID: 24021648,
		ZipName:   fmt.Sprintf("binaries_%s.zip", runtime.GOOS),
	}

	if err := p.Open(); err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	err := ProviderTestWalkAndRetrieve(p)
	if err != nil {
		t.Fatal(err)
	}
}
