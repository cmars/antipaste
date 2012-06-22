package antipaste

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"github.com/cmars/go.crypto/openpgp"
	"github.com/cmars/go.crypto/openpgp/armor"
)

const (
	UndefinedAction = 0
	GetAction = iota
	PutAction
)

var getUri = flag.String("get", "", "Get paste")
var putProtocol = flag.String("put", "", "Put paste")
var newKey = flag.Bool("new", false, "New key")
var findKey = flag.String("find", "", "Find key")
var keyserver = flag.String("hkp", "", "Keyserver")
var importKey = flag.String("import", "", "Import key fingerprint")
var usageError = errors.New("Usage: antipaste -get uri | -put <dest> <file> [id1[,id2,...]] | ...")

type App struct {
	pgp *Pgp
	Action int
	Protocol string
	Handler ProtocolHandler
	// getAction arguments
	getTarget string
	// putAction arguments
	putFileName string
	putRecipients []*openpgp.Entity
}

func NewApp() *App {
	app := &App{}
	app.pgp = &Pgp{}
	app.pgp.Load()
	return app
}

func (app *App) Run() error {
	// Parse general command line flags
	flag.Parse()
	args := flag.Args()
	if *getUri != "" {
		if protocol, uri, err := parseUri(*getUri); err == nil {
			// Ok, we found a uri.
			app.Protocol = protocol
			app.getTarget = uri
			return app.runGet()
		} else {
			return err
		}
	} else if *putProtocol != "" {
		// Assume its a paste, we'll check it...
		app.Protocol = *putProtocol
		// Parse the rest of the paste command line:
		// <file> <recipients...>
		var err error
		var putRecipients []string
		app.putFileName, putRecipients, err = parsePut(args)
		if err != nil {
			return err
		}
		if err = app.resolveRecipients(putRecipients); err == nil {
			return app.runPut()
		} else {
			return err
		}
	} else if *newKey {
		if len(args) == 3 {
			return app.runNewKey(args[0], args[1], args[2])
		} else {
			return errors.New("Usage: -new <name> <email> <comment>")
		}
	} else if *findKey != "" {
		return app.runFindKey(*findKey, *keyserver)
	} else if *importKey != "" {
		return app.runImportKey(*importKey, *keyserver)
	}
	return usageError
}

func (app *App) runGet() (err error) {
	r, err := app.Handler.ReadPaste(app.getTarget)
	if err != nil {
		return err
	}
	defer r.Close()
	block, err := armor.Decode(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ASCII-armor failed: %v\n", err)
		return err
	}
	decOut, err := app.pgp.decrypt(block.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Decrypt failed: %v\n", err)
		return err
	}
	_, err = io.Copy(os.Stdout, decOut)
	return err
}

func (app *App) runPut() (err error) {
	// Create a pipe between the encryption and the protocol handler
	pipeReader, pipeWriter := io.Pipe()
	// Open the plaintext input we're encrypting
	var srcIn io.Reader
	if app.putFileName == "-" {
		srcIn = os.Stdin
	} else {
		srcF, err := os.Open(app.putFileName)
		if err != nil {
			return err
		}
		defer srcF.Close()
		srcIn = srcF
	}
	// Start writing to the pipe
	go func(){
		encOut, err := armor.Encode(pipeWriter, "ANTIPASTE", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ASCII-armor failed: %v\n", err)
			return
		}
		plainOut, err := app.pgp.encrypt(encOut, app.putRecipients)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Encrypt failed: %v\n", err)
			return
		}
		//fmt.Fprintf(os.Stderr, "Writing to output...\n")
		_, err = io.Copy(plainOut, srcIn)
		//fmt.Fprintf(os.Stderr, "Plaintext output written\n")
		err = plainOut.Close()
		//fmt.Fprintf(os.Stderr, "Plaintext output closed\n")
		encOut.Close()
		//fmt.Fprintf(os.Stderr, "Encrypted output closed\n")
		pipeWriter.Close()
		//fmt.Fprintf(os.Stderr, "Pipe output closed\n")
	}()
	// Protocol handler reads the encrypted content from the pipe
	pasteUrl, err := app.Handler.WritePaste(pipeReader)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "%v\n", pasteUrl)
	return nil
}

func (app *App) resolveRecipients(putRecipients []string) error {
	recipients := make(map[string]*openpgp.Entity)
	for _, recipient := range putRecipients {
		entity := app.pgp.resolveRecipient(recipient)
		if entity == nil {
			return errors.New(fmt.Sprintf("Recipient not found: %s", recipient))
		}
		fingerprint, _ := FpToString(entity.PrimaryKey.Fingerprint)
		recipients[fingerprint] = entity
	}
	// Clean up recipient list, make unique just in case there were collisions
	result := []*openpgp.Entity{}
	for _, entity := range recipients {
		result = append(result, entity)
	}
	app.putRecipients = result
	return nil
}

func parsePut(args []string) (fileName string, recipients []string, err error) {
	recipients = []string{}
	for i, arg := range(args) {
		switch i {
		case 0:
			fileName = arg
		default:
			recipients = append(recipients, arg)
		}
	}
	switch {
	case fileName == "" || len(recipients) == 0:
		err = errors.New("Too few arguments")
	case fileName == "-":
		;
	default:
		info, statErr := os.Stat(fileName)
		if statErr != nil {
			err = statErr
		} else if info.IsDir() {
			err = errors.New(fmt.Sprintf("Not a valid input file: %s", fileName))
		}
	}
	return
}

// Parse a URI into protocol, parameter URI to that plugin used to fetch content.
// Return an error if it's not a valid antipaste URI.
func parseUri(uri string) (string, string, error) {
	parts := strings.SplitN(uri, ":", 2)
	if len(parts) > 1 {
		if parts[0] == "http" {
			return "http", uri, nil
		} else if _, has := protocolHandlers[parts[0]]; has {
			return parts[0], parts[1], nil
		}
	} else if info, err := os.Stat(uri); err == nil && !info.IsDir() {
		return "file", uri, nil
	}
	return "", "", errors.New(fmt.Sprintf("Not an antipaste URI: %s", uri))
}

func (app *App) runNewKey(name string, email string, comment string) error {
	_, err := app.pgp.GenKey(name, email, comment)
	if err != nil {
		return err
	}
	return app.pgp.Save()
}

func (app *App) runFindKey(findKey string, keyserver string) error {
	if keyserver == "" {
		keyserver = "pgp.mit.edu"
	}
	hkp, err := ParseHkpUri(keyserver)
	if err != nil {
		return err
	}
	results, err := hkp.Lookup(findKey)
	if err != nil {
		return err
	}
	for _, result := range results {
		fmt.Fprintf(os.Stderr, "%v\n", *result)
	}
	return nil
}

func (app *App) runImportKey(keyid string, keyserver string) error {
	if keyserver == "" {
		keyserver = "pgp.mit.edu"
	}
	hkp, err := ParseHkpUri(keyserver)
	if err != nil {
		return err
	}
	result, err := hkp.Get(keyid)
	if err != nil {
		return err
	}
	app.pgp.PubRing = append(app.pgp.PubRing, result)
	app.pgp.Save()
	return nil
}
