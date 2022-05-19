package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

var ErrInvalidURL = errors.New("invalid url")

const (
	htmlLinkPrefix string = "https://github.com/"
	apiLinkPrefix  string = "https://api.github.com/repos/"
	contents       string = "contents"
)

type Content struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
}
type gitAPI struct {
	err error
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

	rootDir := filepath.Base(htmlLink)
	err = os.MkdirAll(rootDir, os.ModePerm)
	if err != nil {
		log.Fatalf("error creating dir: %v\n", err)
	}

	contents, err := getContent(apiLink)
	if err != nil {
		log.Fatal(err)
	}

	app := &gitAPI{}

	var wg sync.WaitGroup
	err = app.walk(&wg, contents, rootDir)
	if err != nil {
		log.Fatal(err)
	}
	wg.Wait()

}

func (g *gitAPI) walk(wg *sync.WaitGroup, contents []Content, rootDir string) error {
	if g.err != nil {
		return g.err
	}

	for _, content := range contents {
		if content.Type == "dir" {
			rootDir := getRootDir(content.Path, content.Name, rootDir)
			err := os.MkdirAll(rootDir, os.ModePerm)
			if err != nil {
				g.err = fmt.Errorf("error creating dir: %w", err)
				return g.err
			}

			newContents, err := getContent(content.URL)
			if err != nil {
				g.err = fmt.Errorf("error getContent: %w", err)
				return g.err
			}

			g.walk(wg, newContents, rootDir)
			continue
		}
		wg.Add(1)
		go g.downloadFiles(wg, content, rootDir)
	}

	return nil
}

func getContent(link string) ([]Content, error) {
	resp, err := http.Get(link)
	if err != nil && resp.StatusCode != http.StatusOK {
		return nil, err
	}
	defer resp.Body.Close()

	var c []Content
	err = json.NewDecoder(resp.Body).Decode(&c)
	if err != nil {
		return nil, fmt.Errorf("link: %q\nerr: %w", link, err)
	}

	return c, nil
}

func (g *gitAPI) downloadFiles(wg *sync.WaitGroup, content Content, rootDir string) error {
	defer wg.Done()
	resp, err := http.Get(content.DownloadURL)
	if err != nil {
		g.err = fmt.Errorf("error http get: %w", err)
		return g.err
	}
	defer resp.Body.Close()

	fPath := getRootDir(content.Path, content.Name, rootDir)
	f, err := os.Create(fPath)
	if err != nil {
		g.err = fmt.Errorf("error creating: %w", err)
		return g.err
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		g.err = fmt.Errorf("error copy: %w", err)
		return g.err
	}

	return nil
}

func getRootDir(cPath, cName, rootDir string) string {
	rootDir = filepath.Join(rootDir, cName)
	splitted := strings.Split(cPath, rootDir)
	splitted[0] = rootDir
	return filepath.Join(splitted...)
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

// getFilename returns utf8 filename to save
// since the link may contain escaped characters
func getFilename(link string) (string, error) {
	unescaped, err := url.QueryUnescape(link)
	if err != nil {
		return "", err
	}
	return path.Base(unescaped), nil
}
