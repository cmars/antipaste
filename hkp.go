package antipaste

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"github.com/cmars/go.crypto/openpgp"
)

type Hkp struct {
	Hostname string
	Port int
}

type HkpResult struct {
	KeyId string
	Algo int
	KeyLen int
	CreationDate uint64
	ExpirationDate uint64
	Flags string
	Uids []*HkpUserId
}

type HkpUserId struct {
	Uid string
	CreationDate uint64
	ExpirationDate uint64
	Flags string
}

func NewHkp(hostname string, port int) *Hkp {
	if port == 0 {
		port = 11371
	}
	return &Hkp{ Hostname: hostname, Port: port }
}

func ParseHkpUri(uri string) (*Hkp, error) {
	hkpFields := strings.Split(uri, ":")
	if len(hkpFields) < 1 {
		return nil, errors.New(fmt.Sprintf("Invalid Hkp Uri: %s", uri))
	}
	hkp := &Hkp{ Hostname: hkpFields[0] }
	if len(hkpFields) > 1 {
		hkpPort, err := strconv.ParseUint(hkpFields[1], 10, 32)
		if err != nil {
			return nil, err
		}
		hkp.Port = int(hkpPort)
	}
	return hkp, nil
}

func (hkp *Hkp) BaseUrl() string {
	return fmt.Sprintf("http://%s:%d", hkp.Hostname, hkp.Port)
}

func (hkp *Hkp) Lookup(value string) (results []*HkpResult, err error) {
	resp, err := http.Get(fmt.Sprintf(
			"%s/pks/lookup?op=index&search=%s&options=mr",
			hkp.BaseUrl(), value))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rdr := bufio.NewReaderSize(resp.Body, 1024)
	var current *HkpResult
	var line []byte
	for line, _, err = rdr.ReadLine(); ; line, _, err = rdr.ReadLine() {
		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return nil, err
		}
		fields := strings.Split(string(line), ":")
		switch fields[0] {
		case "info":
			;
		case "pub":
			current = &HkpResult{
				KeyId: fields[1], Flags: fields[6] }
			algo, err := strconv.ParseUint(fields[2], 10, 32)
			if err != nil {
				return results, err
			} else {
				current.Algo = int(algo)
			}
			keylen, err := strconv.ParseUint(fields[3], 10, 32)
			if err != nil {
				return results, err
			} else {
				current.KeyLen = int(keylen)
			}
			if fields[4] == "" {
				current.CreationDate = 0
			} else {
				current.CreationDate, err = strconv.ParseUint(fields[4], 10, 64)
				if err != nil {
					return results, err
				}
			}
			if fields[5] == "" {
				current.ExpirationDate = 0xFFFFFFFFFFFFFFFF
			} else {
				current.ExpirationDate, err = strconv.ParseUint(fields[5], 10, 64)
				if err != nil {
					return results, err
				}
			}
			results = append(results, current)
		case "uid":
			if current == nil {
				return results, errors.New("Invalid response from server: 'uid' record before 'pub'")
			}
			uid := &HkpUserId{ Uid: fields[1], Flags: fields[4] }
			if fields[2] == "" {
				uid.CreationDate = 0
			} else {
				uid.CreationDate, err = strconv.ParseUint(fields[2], 10, 64)
				if err != nil {
					return results, err
				}
			}
			if fields[3] == "" {
				uid.ExpirationDate = 0xFFFFFFFFFFFFFFFF
			} else {
				uid.ExpirationDate, err = strconv.ParseUint(fields[3], 10, 64)
				if err != nil {
					return results, err
				}
			}
			current.Uids = append(current.Uids, uid)
		}
	}
	return results, err
}

func (hkp *Hkp) Get(keyid string) (*openpgp.Entity, error) {
	resp, err := http.Get(fmt.Sprintf("%s/pks/lookup?op=get&search=0x%s",
			hkp.BaseUrl(), keyid))
	entities, err := openpgp.ReadArmoredKeyRing(resp.Body)
	if err != nil {
		return nil, err
	}
	for _, entity := range entities {
		return entity, nil
	}
	return nil, errors.New("Key not found")
}
