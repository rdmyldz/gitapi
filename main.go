package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var ErrInvalidURL = errors.New("invalid url")

const (
	htmlLinkPrefix string = "https://github.com/"
	apiLinkPrefix  string = "https://api.github.com/repos/"
	contents       string = "contents"
)

type Content struct {
	Name        string `json:"name"`
	Size        int    `json:"size"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
}

type gitapi struct {
	dirname  string
	Contents []Content
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("you must specify a URL which must end with a directory to download files inside of it")
	}

	htmlLink := os.Args[1]
	apiLink, err := getApiLink(htmlLink)
	if err != nil {
		log.Fatal(err)
	}

	g := gitapi{
		dirname: getDirName(htmlLink),
	}
	err = createDirIfNotExist(g.dirname)
	if err != nil {
		log.Fatal(err)
	}

	b, err := downloadContent(apiLink)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(b, &g.Contents)
	if err != nil {
		log.Fatal(err)
	}

	for _, content := range g.Contents {
		fmt.Printf("name: %q, size: %d, url: %q, type: %q\n", content.Name, content.Size, content.DownloadURL, content.Type)
		if content.Type != "file" {
			continue
		}
		g.downloadFiles(content.DownloadURL)
	}

}

func downloadContent(link string) ([]byte, error) {
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (a *gitapi) downloadFiles(link string) error {
	filename, err := getFilename(link)
	if err != nil {
		log.Fatal(err)
	}

	filePath := filepath.Join(a.dirname, filename)

	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func getApiLink(link string) (string, error) {
	if !strings.HasPrefix(link, htmlLinkPrefix) || !strings.Contains(link, "/tree/") {
		return "", ErrInvalidURL
	}

	linkPath := strings.Split(link, htmlLinkPrefix)[1]
	partialPath := strings.SplitN(linkPath, "/tree/", 2)
	repo := partialPath[0]
	dirPath := partialPath[1]
	s := strings.SplitN(dirPath, "/", 2)
	branchName := s[0]
	dirname := s[1]

	return apiLinkPrefix + path.Join(repo, contents, dirname) + "?ref=" + branchName, nil
}

func getDirName(link string) string {
	return filepath.Base(link)
}

// getFilename returns utf8 filename to save
// since the link may contain escaped characters
func getFilename(link string) (string, error) {
	unescaped, err := url.QueryUnescape(link)
	if err != nil {
		return "", err
	}
	return path.Base(unescaped), nil
}

func createDirIfNotExist(p string) error {
	_, err := os.Stat(p)
	if errors.Is(err, fs.ErrNotExist) {
		err = os.Mkdir(p, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
