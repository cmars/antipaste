package antipaste

import (
	"encoding/hex"
	"errors"
	"flag"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"github.com/cmars/go.crypto/openpgp"
	"github.com/cmars/go.crypto/openpgp/packet"
)

var homeFlag = flag.String("homedir", "", "Antipaste keyring home directory")

var pgpBlockRE = regexp.MustCompile(`(?ms:-----BEGIN [^-]+-----.*-----END [^-]+-----)`)

type Keyserver struct {
	Hostname string
	Port int
}

type Pgp struct {
	SecRing openpgp.EntityList
	PubRing openpgp.EntityList
	Keyservers []Keyserver
}

func FpToString(fp [20]byte) (string, error) {
	if len(fp) != 20 {
		return "", errors.New("Invalid fingerprint")
	}
	return hex.EncodeToString(fp[:]), nil
}

func StringToFp(fpStr string) ([20]byte, error) {
	result := [20]byte{}
	slice, err := hex.DecodeString(fpStr)
	if err != nil {
		return result, err
	}
	if len(slice) != 20 {
		return result, errors.New("Invalid fingerprint")
	}
	for i, b := range slice {
		result[i] = b
	}
	return result, nil
}

func (pgp *Pgp) encrypt(ciphertext io.Writer,
		recipients []*openpgp.Entity) (plaintext io.WriteCloser, err error) {
	return openpgp.Encrypt(ciphertext, recipients, nil, nil, nil)
}

// Decrypt content using a private key in our keyring.
func (pgp *Pgp) decrypt(r io.Reader) (io.Reader, error) {
	md, err := openpgp.ReadMessage(r, pgp.SecRing, nil, nil)
	if err != nil {
		return nil, err
	}
	return md.UnverifiedBody, nil
}

// Resolve a recipient by key ID, email address, etc.
func (pgp *Pgp) resolveRecipient(id string) *openpgp.Entity {
	id = strings.ToLower(id)
	for _, entity := range pgp.PubRing {
		fp, _ := FpToString(entity.PrimaryKey.Fingerprint)
		if strings.HasSuffix(fp, id) {
			return entity
		}
	}
	return nil
}

func homeDir() (string, error) {
	if *homeFlag != "" {
		return *homeFlag, nil
	}
	luser, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(luser.HomeDir, ".antipaste"), nil
}

func keyFiles() (pubFile string, secFile string, err error) {
	basepath, err := homeDir()
	os.MkdirAll(basepath, 0700)
	if err == nil {
		pubFile = filepath.Join(basepath, "pubring.gpg")
		secFile = filepath.Join(basepath, "secring.gpg")
	}
	return pubFile, secFile, err
}

func (pgp *Pgp) Load() error {
	pubFile, secFile, err := keyFiles()
	if err != nil {
		return err
	}
	// Read the public key ring
	_, err = os.Stat(pubFile)
	if err == os.ErrNotExist {
		pgp.PubRing = []*openpgp.Entity{}
	} else if err == nil {
		pubReader, err := os.Open(pubFile)
		if err != nil {
			return err
		}
		defer pubReader.Close()
		pgp.PubRing, err = openpgp.ReadKeyRing(pubReader)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	// Read the private key ring
	_, err = os.Stat(secFile)
	if err == os.ErrNotExist {
		pgp.SecRing = []*openpgp.Entity{}
	} else if err == nil {
		secReader, err := os.Open(secFile)
		if err != nil {
			return err
		}
		defer secReader.Close()
		pgp.SecRing, err = openpgp.ReadKeyRing(secReader)
	}
	return err
}

func (pgp *Pgp) Save() error {
	pubFile, secFile, err := keyFiles()
	if err != nil {
		return err
	}
	// Write public key ring
	pubWriter, err := os.OpenFile(pubFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer pubWriter.Close()
	for _, e := range pgp.PubRing {
		err = e.Serialize(pubWriter)
		if err != nil {
			return err
		}
	}
	// Write private key ring
	secWriter, err := os.OpenFile(secFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer secWriter.Close()
	for _, e := range pgp.SecRing {
		err = e.SerializePrivate(secWriter, nil)
		if err != nil {
			return err
		}
	}
	return err
}

func (pgp *Pgp) GenKey(name string, email string, comment string) (*openpgp.Entity, error) {
	config := &packet.Config{}
	entity, err := openpgp.NewEntity(name, comment, email, config)
	if err != nil {
		return entity, err
	}
	// Self-sign each identity
	for _, ident := range entity.Identities {
		err = ident.SelfSignature.SignUserId(ident.UserId.Id, entity.PrimaryKey, entity.PrivateKey, config)
		if err != nil {
			return entity, err
		}
	}
	pubEntity := &openpgp.Entity{
		PrimaryKey: entity.PrimaryKey,
		Identities: entity.Identities,
		Subkeys: []openpgp.Subkey{} }
	// Self-sign each subkey
	for _, subkey := range entity.Subkeys {
		err = subkey.Sig.SignKey(subkey.PublicKey, entity.PrivateKey, config)
		pubSubkey := &openpgp.Subkey{ PublicKey: subkey.PublicKey, Sig: subkey.Sig }
		pubEntity.Subkeys = append(pubEntity.Subkeys, *pubSubkey)
	}
	pgp.SecRing = append(pgp.SecRing, entity)
	pgp.PubRing = append(pgp.PubRing, entity)
	return pubEntity, err
}
