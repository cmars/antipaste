package antipaste

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var ubuntuPoster = flag.String("ubuntu-poster", "anonymous", "Ubuntu poster name")
var ubuntuPrefix = regexp.MustCompile("^ubuntu:")

type UbuntuHandler struct {
	Poster string
}

func init() {
	uph := &UbuntuHandler{ Poster: *ubuntuPoster }
	protocolHandlers[uph.Prefix()] = uph
}

func (uph *UbuntuHandler) Prefix() string {
	return "ubuntu"
}

func (uph *UbuntuHandler) ReadPaste(url string) (io.ReadCloser, error) {
	url = strings.Trim(url, "/")
	fields := strings.Split(url, "/")
	if len(fields) == 0 {
		return nil, errors.New(fmt.Sprintf("Invalid ubuntu paste URL %v", url))
	}
	id := fields[len(fields)-1]
	id = dpPrefix.ReplaceAllLiteralString(id, "")
	resp, err := http.Get(fmt.Sprintf("http://paste.ubuntu.com/%s/", id))
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	contentStr := fmt.Sprintf("%s\n", string(pgpBlockRE.Find(contents)))
	return ioutil.NopCloser(bytes.NewBufferString(contentStr)), nil
}

func (uph *UbuntuHandler) WritePaste(r io.Reader) (string, error) {
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	resp, err := http.PostForm("http://paste.ubuntu.com/",
		url.Values{
			"poster": {uph.Poster},
			"syntax": {"text"},
			"content": {string(contents)}})
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
	return fmt.Sprintf("%s:%s", uph.Prefix(), id), err
}
