package provider_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	provider "github.com/mouuff/easy-update/provider"
)

func testProvider(p provider.Provider) error {
	tmpDir, err := ioutil.TempDir("", "testProvider")
	if err != nil {
		return err
	}
	log.Print(tmpDir)
	//defer os.RemoveAll(tmpDir)

	err = p.Walk(func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", filePath, err)
			return err
		}
		destPath := path.Join(tmpDir, filePath)
		if info.IsDir() {
			os.MkdirAll(destPath, os.ModePerm)
		} else {
			os.MkdirAll(filepath.Dir(destPath), os.ModePerm)
			err = p.Retrieve(filePath, destPath)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func TestProviderLocal(t *testing.T) {
	p := provider.NewProviderLocal(path.Join("testdata", "Allum1"))
	if err := p.Open(); err != nil {
		return err
	}
	defer p.Close()

	err := testProvider(p)
	if err != nil {
		t.Error(err)
	}
}