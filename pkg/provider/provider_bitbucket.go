package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mouuff/go-rocket-update/internal/fileio"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// BitBucket provider finds an archive file in the repository's downloads to provide updated binaries
// As BitBucket does not have a releases feature, we use the downloads feature instead.
type BitBucket struct {
	RepositoryURL string // The URL of the BitBucket repository, e.g. bitbucket.org/BenTomsett/go-rocket-update-example
	ArchiveName   string // The name of the archive file uploaded to BitBucket Downloads, e.g. project-v0.0.1.tar.gz

	tmpDir             string   // Temporary directory to download the updated version to
	decompressProvider Provider // Provider used to decompress the downloaded archive
	archivePath        string   // Path to the downloaded archive (should be in tmpDir)
}

// bitbucketRepositoryInfo is used to get the owner and name of the repository
// from these fields we are able to get other links (such as the downloads link)
type bitbucketRepositoryInfo struct {
	RepositoryOwner string
	RepositoryName  string
}

type bitbucketTag struct {
	Name string `json:"name"`
}

type bitbucketTagResponse struct {
	Values []bitbucketTag `json:"values"`
}

// getRepositoryInfo parses the BitBucket repository URL
func (c *BitBucket) repositoryInfo() (*bitbucketRepositoryInfo, error) {
	re := regexp.MustCompile(`bitbucket\.org/(.*?)/(.*?)$`)
	submatches := re.FindAllStringSubmatch(c.RepositoryURL, 1)
	if len(submatches) < 1 {
		return nil, errors.New("Invalid BitBucket URL: " + c.RepositoryURL)
	}
	return &bitbucketRepositoryInfo{
		RepositoryOwner: submatches[0][1],
		RepositoryName:  submatches[0][2],
	}, nil
}

// getTagsURL gets the API URL for the repository which returns the list of tags
func (c *BitBucket) getTagsUrl() (string, error) {
	info, err := c.repositoryInfo()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/refs/tags?sort=target.date",
		info.RepositoryOwner,
		info.RepositoryName,
	), nil
}

func (c *BitBucket) getTags() (tags []bitbucketTag, err error) {
	url, err := c.getTagsUrl()
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var tagResponse bitbucketTagResponse
	err = json.NewDecoder(resp.Body).Decode(&tagResponse)
	if err != nil {
		return nil, err
	}
	return tagResponse.Values, nil
}

func (c *BitBucket) getArchiveUrl() (string, error) {
	info, err := c.repositoryInfo()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://bitbucket.org/%s/%s/downloads/%s",
		info.RepositoryOwner,
		info.RepositoryName,
		c.ArchiveName,
	), nil
}

func (c *BitBucket) GetLatestVersion() (string, error) {
	tags, err := c.getTags()
	if err != nil {
		return "", err
	}
	if len(tags) == 0 {
		return "", errors.New("no tags found in this BitBucket repository")
	}
	return tags[0].Name, nil
}

func (c *BitBucket) Open() (err error) {
	archiveUrl, err := c.getArchiveUrl()
	if err != nil {
		return err
	}

	resp, err := http.Get(archiveUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	c.tmpDir, err = fileio.TempDir()
	if err != nil {
		return err
	}

	c.archivePath = filepath.Join(c.tmpDir, c.ArchiveName)
	archiveFile, err := os.Create(c.archivePath)
	if err != nil {
		return err
	}

	_, err = io.Copy(archiveFile, resp.Body)
	archiveFile.Close()
	if err != nil {
		return err
	}

	c.decompressProvider, err = Decompress(c.archivePath)
	if err != nil {
		return err
	}

	return c.decompressProvider.Open()
}

func (c *BitBucket) Close() error {
	if c.decompressProvider != nil {
		c.decompressProvider.Close()
		c.decompressProvider = nil
	}

	if len(c.tmpDir) > 0 {
		os.RemoveAll(c.tmpDir)
		c.tmpDir = ""
		c.archivePath = ""
	}

	return nil
}

func (c *BitBucket) Walk(walkFn WalkFunc) error {
	if c.decompressProvider == nil {
		return ErrNotOpenned
	}
	return c.decompressProvider.Walk(walkFn)
}

func (c *BitBucket) Retrieve(src string, dest string) error {
	return c.decompressProvider.Retrieve(src, dest)
}
