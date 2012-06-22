package antipaste

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

var gistDesc = flag.String("gist-desc", "", "gist description")
var gistFilename = flag.String("gist-filename", "README", "gist filename")
var gistPrefix = regexp.MustCompile("^gist:")

type ghandler struct {
	Description string
	Filename string
}

func init() {
	gh := &ghandler{
		Description: *gistDesc,
		Filename: *gistFilename }
	protocolHandlers[gh.Prefix()] = gh
}

func (gh *ghandler) Prefix() string {
	return "gist"
}

func (gh *ghandler) ReadPaste(url string) (io.ReadCloser, error) {
	url = strings.Trim(url, "/")
	fields := strings.Split(url, "/")
	if len(fields) == 0 {
		return nil, errors.New(fmt.Sprintf("Invalid gist paste URL %v", url))
	}
	id := fields[len(fields)-1]
	id = gistPrefix.ReplaceAllLiteralString(id, "")
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/gists/%v", id))
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	msg := make(map[string]interface{})
	json.Unmarshal(body, &msg)
	if files, has := msg["files"]; has {
		if filesMap, is := files.(map[string]interface{}); is {
			for _, file := range filesMap {
				if fileMap, is := file.(map[string]interface{}); is {
					if content, has := fileMap["content"]; has {
						if contentStr, is := content.(string); is {
							return ioutil.NopCloser(bytes.NewBufferString(contentStr)), nil
						}
					}
				}
			}
		}
	}
	return nil, errors.New(fmt.Sprintf("Unrecognized response format: %s", string(body)))
}

type GistPostFile struct {
	Content string `json:"content"`
}

type GistPostMsg struct {
	Description string `json:"description"`
	Public bool `json:"public"`
	Files map[string]*GistPostFile `json:"files"`
}

func (gh *ghandler) WritePaste(r io.Reader) (string, error) {
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	jsonMsg := &GistPostMsg{
		Description: gh.Description,
		Public: true,
		Files: make(map[string]*GistPostFile) }
	jsonMsg.Files[gh.Filename] = &GistPostFile{ Content: string(contents) }
	jsonData, err := json.Marshal(jsonMsg)
	if err != nil {
		return "", err
	}
	resp, err := http.Post("https://api.github.com/gists", "application/json",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	url := resp.Header.Get("Location")
	defer resp.Body.Close()
	if url == "" {
		return "", errors.New(fmt.Sprintf("Paste location missing from response header: %v",
				resp.Header))
	}
	fields := strings.Split(strings.TrimRight(url, "/"), "/")
	id := fields[len(fields)-1]
	if len(fields) < 2 {
		return "", errors.New(fmt.Sprintf("Invalid response: %s", id))
	}
	return fmt.Sprintf("%s:%s", gh.Prefix(), id), err
}
