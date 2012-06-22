package antipaste

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"code.google.com/p/gorilla/mux"
	"code.google.com/p/gorilla/securecookie"
)

// These are used to create a secure cookie for the passphrase
var hashKey []byte
var blockKey []byte
var secureCookie *securecookie.SecureCookie

func init() {
	hashKey = securecookie.GenerateRandomKey(64)
	blockKey = securecookie.GenerateRandomKey(64)
	secureCookie = securecookie.New(hashKey, blockKey)
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("%v", err)))
}

func Index(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		writeError(w, err)
		return
	}
	p := r.Form.Get("p")
	if p != "" {
		Show(w, r, p)
	} else {
		ids, err := ListIdentities()
		if err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		indexTemplate.Execute(w, &indexArgs{
			PageName: "Open Paste",
			Identities: ids })
	}
}

/* Paste submission. */
func Paste(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		writeError(w, err)
		return
	}
	recipient_fps, has := r.Form["recipient"]
	if !has || len(recipient_fps) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required parameter: recipient"))
		return
	}
	contents := r.Form.Get("contents")
	if contents == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required parameter: contents"))
	}
	identities, err := MapIdentities()
	if err != nil {
		writeError(w, err)
		return
	}
	recipients := []*Identity{}
	for _, recipient_fp := range recipient_fps {
		if recipient_fp == "" {
			continue
		}
		id, has := identities[recipient_fp]
		if !has {
			writeError(w, errors.New(
				fmt.Sprintf("Recipient %v not found in keyring", recipient_fp)))
			return
		} else {
			recipients = append(recipients, id)
		}
	}
	ciphertext, err := Encrypt(contents, recipients)
	if err != nil {
		writeError(w, err)
		return
	}
	pastebin := &Pastebin{}
	pastebinUrl, err := pastebin.NewPaste(ciphertext)
	if err != nil {
		writeError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/?p=%s", url.QueryEscape(pastebinUrl)), http.StatusFound)
}

/* Show a paste. */
func Show(w http.ResponseWriter, r *http.Request, p string) {
	// TODO: decode special anti-paste id forms
	// Like: PB:abc123 is a pastebin.com ID
	//       DP:abc123 could be dpaste.com, etc.
	pastebin := &Pastebin{}
	ciphertext, err := pastebin.GetPaste(p)
	if err != nil {
		writeError(w, err)
		return
	}
	plaintext, err := Decrypt(ciphertext)
	if err == ErrMissingPassphrase {
		// Try to read the passphrase out of the secure cookie
		if cookie, ck_err := r.Cookie("secpp"); ck_err == nil {
			var passphrase string
			if ck_err = secureCookie.Decode("secpp", cookie.Value, &passphrase); ck_err == nil {
				plaintext, err = DecryptWithPassphrase(passphrase, ciphertext)
			}
		}
		// If that didn't work out, prompt for the passphrase
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("/askpp?p=%s", p), http.StatusFound)
		}
	} else if err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	showTemplate.Execute(w, &showArgs{
		PageName: "Paste",
		Url: p,
		Paste: plaintext })
}

func AskPassphrase(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, err)
		return
	}
	p := r.Form.Get("p")
	passphrase := r.Form.Get("pp")
	// If method is POST and we have the passphrase, redirect to Show or home
	if r.Method == "POST" && passphrase != "" {
		// Store the passphrase in a secure cookie
		if encoded, err := secureCookie.Encode("secpp", passphrase); err != nil {
			cookie := &http.Cookie{
				Name: "secpp",
				Value: encoded,
				Path: "/"}
			http.SetCookie(w, cookie)
		}
		if p != "" {
			http.Redirect(w, r, fmt.Sprintf("/?p=%s", p), http.StatusFound)
		} else {
			http.Redirect(w, r, "/", http.StatusFound)
		}
	}
	askppTemplate.Execute(w, &askppArgs{
		PageName: "Passphrase Required",
		Url: p })
}

func Run() {
	r := mux.NewRouter()
	r.HandleFunc("/paste", Paste)
	r.HandleFunc("/askpp", AskPassphrase)
	if devpath := os.Getenv("ANTIPASTE_DEVPATH"); devpath != "" {
		r.Handle("/static/{path:.*}", http.FileServer(http.Dir(devpath)))
	} else {
		staticParent, err := ioutil.TempDir("", "antipaste")
		if err != nil {
			log.Fatal(err)
		}
		err = tarExtract(staticParent, staticArchive())
		r.Handle("/static/{path:.*}", http.FileServer(http.Dir(staticParent)))
		defer os.RemoveAll(staticParent)
	}
	r.HandleFunc("/", Index)
	http.Handle("/", r)
	err := http.ListenAndServe("127.0.0.1:12345", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func tarExtract(parent string, archive []byte) error {
	buf := bytes.NewBuffer(archive)
	gunzip, err := gzip.NewReader(buf)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gunzip)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		filename := filepath.Join(append([]string{parent}, strings.Split(hdr.Name, "/")...)...)
		if hdr.Typeflag == tar.TypeReg || hdr.Typeflag == tar.TypeRegA {
			outF, err := os.OpenFile(filename, os.O_CREATE | os.O_WRONLY | os.O_EXCL,
					os.FileMode(int32(hdr.Mode)))
			if err != nil {
				return err
			}
			_, err = io.Copy(outF, tr)
		} else if hdr.Typeflag == tar.TypeDir {
			os.MkdirAll(filename, os.FileMode(int32(hdr.Mode)))
		}
		if err != nil {
			return err
		}
	}
	return nil
}
