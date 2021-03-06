package cs1a

import (
	"crypto/elliptic"
	"crypto/rand"
	"math/big"

	"github.com/telehash/gogotelehash/e3x/cipherset"
	"github.com/telehash/gogotelehash/e3x/cipherset/cs1a/eccp"
	"github.com/telehash/gogotelehash/e3x/cipherset/cs1a/secp160r1"
	"github.com/telehash/gogotelehash/internal/util/base32util"
)

type key struct {
	pub struct{ x, y *big.Int }
	prv struct{ d []byte }
}

func decodeKeyBytes(pub, prv []byte) (*key, error) {
	var (
		k = &key{}
	)

	if len(pub) != 0 {
		k.pub.x, k.pub.y = eccp.Unmarshal(secp160r1.P160(), pub)
		if k.pub.x == nil || k.pub.y == nil {
			return nil, cipherset.ErrInvalidKey
		}
	}

	if len(prv) != 0 {
		k.prv.d = prv
	}

	return k, nil
}

func generateKey() (*key, error) {
	var (
		k   = &key{}
		err error
	)

	k.prv.d, k.pub.x, k.pub.y, err = elliptic.GenerateKey(secp160r1.P160(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return k, nil
}

func (k *key) CSID() uint8 { return 0x1a }

func (k *key) Public() []byte {
	if k == nil || k.pub.x == nil || k.pub.y == nil {
		return nil
	}

	return eccp.Marshal(secp160r1.P160(), k.pub.x, k.pub.y)
}

func (k *key) Private() []byte {
	if k == nil || k.prv.d == nil {
		return nil
	}

	buf := make([]byte, len(k.prv.d))
	copy(buf, k.prv.d)
	return buf
}

func (k *key) String() string {
	return base32util.EncodeToString(k.Public())
}

func (k *key) CanSign() bool {
	return k != nil && k.prv.d != nil
}

func (k *key) CanEncrypt() bool {
	return k != nil && k.pub.x != nil && k.pub.y != nil
}
