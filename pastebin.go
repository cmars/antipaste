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
)

var pbApiKey = flag.String("pb-api", "89f37b01f7f599990fef3e94fe7a570d", "Pastebin API key")
var pbPrefix = regexp.MustCompile("^pb:")

type PastebinHandler struct {
	ApiKey string
}

func init() {
	pbh := &PastebinHandler{ ApiKey: *pbApiKey }
	protocolHandlers[pbh.Prefix()] = pbh
}

func (pbh *PastebinHandler) Prefix() string {
	return "pb"
}

func (pbh *PastebinHandler) ReadPaste(url string) (io.ReadCloser, error) {
	url = strings.Trim(url, "/")
	fields := strings.Split(url, "/")
	if len(fields) == 0 {
		return nil, errors.New(fmt.Sprintf("Invalid pastebin paste URL %v", url))
	}
	pbId := fields[len(fields)-1]
	pbId = pbPrefix.ReplaceAllLiteralString(pbId, "")
	resp, err := http.Get(fmt.Sprintf("http://pastebin.com/raw.php?i=%s", pbId))
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (pbh *PastebinHandler) WritePaste(r io.Reader) (string, error) {
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	resp, err := http.PostForm("http://pastebin.com/api/api_post.php",
		url.Values{
			"api_option": {"paste"},
			"api_dev_key": {*pbApiKey},
			"api_paste_code": {string(contents)}})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	url, err := ioutil.ReadAll(resp.Body)
	fields := strings.Split(string(url), "/")
	id := fields[len(fields)-1]
	if len(fields) < 2 {
		return "", errors.New(fmt.Sprintf("Invalid response: %s", id))
	}
	return fmt.Sprintf("%s:%s", pbh.Prefix(), id), err
}
