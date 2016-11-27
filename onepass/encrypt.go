package onepass

import (
	"hash"
	"crypto/cipher"
	"fmt"
	"encoding/hex"
	"strings"
)

type encrypt struct {
	d hash.Hash
	Ya cipher.Block
}

func (c *encrypt)encryptPayload (payload string) Payload {
	payloadBytes := []byte(payload)
	iv := codec.generateRandomBytesArray(16)

	mode := cipher.NewCBCEncrypter(c.Ya, iv)
	makeup := strings.Repeat("\t", mode.BlockSize() - (len(payload) % mode.BlockSize()))
	payloadBytes = append(payloadBytes, []byte(makeup)...)
	mode.CryptBlocks(payloadBytes, payloadBytes)

	payloadStr := codec.fromBits(payloadBytes, true, true)
	ivStr := codec.fromBits(iv, true, true)

	c.d.Reset()
	c.d.Write([]byte(ivStr + payloadStr))
	mac := codec.fromBits(c.d.Sum(nil), true, true)

	modeDec := cipher.NewCBCDecrypter(c.Ya, iv)
	modeDec.CryptBlocks(payloadBytes, payloadBytes)

	return Payload {
		Alg: 	"aead-cbchmac-256",
		Iv: 	ivStr,
		Data: 	payloadStr,
		Hmac: 	mac,
	}
}

func (c *encrypt)decryptPayload (payload string, iv string, hmac string) (string) {
	payloadBytes := codec.toBits(payload, true)
	ivBytes := codec.toBits(iv, true)
	hmacBytes := codec.toBits(hmac, true)

	c.d.Reset()
	c.d.Write([]byte(iv + payload))
	mac := c.d.Sum(nil)

	if (hex.EncodeToString(hmacBytes) == hex.EncodeToString(mac)) {
		mode := cipher.NewCBCDecrypter(c.Ya, ivBytes)
		mode.CryptBlocks(payloadBytes, payloadBytes)
		return fmt.Sprintf("%s", payloadBytes)
	}

	return ""
}

/*
function ka(a, c) {
    function b(a, b, c) {
        if (void 0 === a)
            throw new sjcl.exception.invalid('iv is required');
        if (void 0 === b && void 0 === c)
            throw new sjcl.exception.invalid('Either ciphertext or adata is required for hmac calculation.');
        if ('string' !== typeof a || void 0 !== b && 'string' !== typeof b || void 0 !== c && 'string' !== typeof c)
            throw new sjcl.exception.invalid('Invalid input: ' + typeof a + '/' + typeof b + '/' + typeof c);
        a = [a];
        void 0 !== b && a.push(b);
        void 0 !== c && a.push(c);
        return d.encrypt(a.join(''))
    }
    var d;
    this.Ya = new sjcl.cipher.aes(a);
    d = new sjcl.misc.hmac(c,sjcl.hash.sha256);
    this.aa = 'aead-cbchmac-256';
    this.encryptPayload = this.vb = function(a) {
        var c, d;
        a = 'object' === typeof a ? JSON.stringify(a) : a;
        c = sjcl.codec.utf8String.toBits(a);
        a = crypto.getRandomValues(new Uint8Array(16));
        d = sjcl.codec.bytes.toBits(a);
        a = r.A(d);
        c = sjcl.mode.cbc.encrypt(this.Ya, c, d);
        d = r.A(c);
        c = {
            alg: this.aa,
            iv: a,
            data: d
        };
        a = b(a, d, void 0);
        a = sjcl.bitArray.clamp(a, 96);
        c.hmac = r.A(a);
        return c
    }
    ;
    this.decryptPayload = this.ub = function(a) {
        var c, d, k, m, p, l, t, y;
        if (a.alg !== this.aa)
            throw Error('Mismatched payload algorithm: <' + a.alg + '/' + this.aa + '>');
        c = a.iv;
        d = a.data;
        k = a.adata;
        try {
            p = a.hmac;
            m = r.H(p);
            a = k;
            var u, D;
            void 0 === m && (m = a,
            a = void 0);
            D = sjcl.bitArray.bitLength(m);
            if (96 > D)
                throw new sjcl.exception.invalid('The supplied hmac value is invalid.');
            u = b(c, d, a);
            sjcl.bitArray.bitLength(u) > D && (u = sjcl.bitArray.bitSlice(u, 0, D));
            if (!sjcl.bitArray.equal(u, m))
                throw new sjcl.exception.corrupt('Failed to validate payload hmac.');
            l = sjcl.mode.cbc.decrypt(this.Ya, r.H(d), r.H(c));
            t = sjcl.codec.utf8String.fromBits(l);
            y = JSON.parse(t)
        } catch (W) {
            console.error(W)
        } finally {
            return y
        }
    }
}
 */