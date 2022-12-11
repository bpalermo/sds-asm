package helper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTlsSecretFromBytes(t *testing.T) {
	name := "name"
	certificateChain := []byte{12, 56, 123}
	privateKey := []byte{112, 65, 13}

	tlsSecret := TlsSecretFromBytes(name, certificateChain, privateKey)

	assert.NotNil(t, tlsSecret)
	assert.Equal(t, name, tlsSecret.Name)
	assert.Equal(t, certificateChain, tlsSecret.GetTlsCertificate().CertificateChain.GetInlineBytes())
	assert.Equal(t, privateKey, tlsSecret.GetTlsCertificate().PrivateKey.GetInlineBytes())
}
