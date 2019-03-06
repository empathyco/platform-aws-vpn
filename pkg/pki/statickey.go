package pki

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const (
	kStaticKeyBits = 2048
)

type StaticKey []byte

func NewStaticKey() StaticKey {
	buf := make([]byte, kStaticKeyBits/8)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return StaticKey(buf)
}

func (k StaticKey) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("#\n# %d bit OpenVPN static key\n#\n", 8*len(k)))
	buf.WriteString("-----BEGIN OpenVPN Static key V1-----\n")

	hexEnc := hex.NewEncoder(&buf)
	for i := 0; i < len(k); i += 16 {
		if _, err := hexEnc.Write(k[i : i+16]); err != nil {
			panic(err)
		}
		buf.WriteByte('\n')
	}

	buf.WriteString("-----END OpenVPN Static key V1-----\n")

	return buf.String()
}
