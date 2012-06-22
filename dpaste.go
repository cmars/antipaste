package antipaste

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"strconv"
)

var dpExpire = flag.String("dpaste-expire", "3600", "dpaste expiration (seconds)")
var dpLexer = flag.String("dpaste-lexer", "text", "dpaste lexer")
var dpTitle = flag.String("dpaste-title", "", "dpaste title")
var dpPrefix = regexp.MustCompile("^dpaste:")

type DpasteHandler struct {
	Expire int
	Lexer string
	Title string
}

func init() {
	ttl, err := strconv.ParseInt(*dpExpire, 10, 32)
	if err != nil {
		ttl = 3600
	}
	dph := &DpasteHandler{
		Expire: int(ttl),
		Lexer: *dpLexer,
		Title: *dpTitle }
	protocolHandlers[dph.Prefix()] = dph
}

func (dph *DpasteHandler) Prefix() string {
	return "dpaste"
}

func (dph *DpasteHandler) ReadPaste(url string) (io.ReadCloser, error) {
	url = strings.Trim(url, "/")
	fields := strings.Split(url, "/")
	if len(fields) == 0 {
		return nil, errors.New(fmt.Sprintf("Invalid dpaste paste URL %v", url))
	}
	id := fields[len(fields)-1]
	id = dpPrefix.ReplaceAllLiteralString(id, "")
	resp, err := http.Get(fmt.Sprintf("http://dpaste.org/%s/raw/", id))
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (dph *DpasteHandler) WritePaste(r io.Reader) (string, error) {
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	resp, err := http.PostForm("http://dpaste.org/",
		url.Values{
			"content": {string(contents)},
			"lexer": {dph.Lexer},
			"expire_options": {fmt.Sprintf("%d", dph.Expire)},
			"title": {dph.Title}})
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
	return fmt.Sprintf("%s:%s", dph.Prefix(), id), err
}
