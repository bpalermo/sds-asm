package snapshot

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tlsV3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"math/big"
	"net"
	"time"
)

func GenerateSnapshot() (*cache.Snapshot, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"Company, INC."},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivateKey, _ := rsa.GenerateKey(rand.Reader, 4096)

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization:  []string{"Company, INC."},
			Country:       []string{"BR"},
			Province:      []string{""},
			Locality:      []string{"Rio de Janeiro"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"22610142"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivateKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	certBytes, _ := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivateKey.PublicKey, caPrivateKey)

	tlsSecret, _ := anypb.New(&tlsV3.Secret{
		Name: "example_com",
		Type: &tlsV3.Secret_TlsCertificate{
			TlsCertificate: &tlsV3.TlsCertificate{
				CertificateChain: &v3.DataSource{
					Specifier: &v3.DataSource_InlineBytes{
						InlineBytes: certBytes,
					},
				},
				PrivateKey: &v3.DataSource{
					Specifier: &v3.DataSource_InlineBytes{
						InlineBytes: x509.MarshalPKCS1PrivateKey(certPrivateKey),
					},
				},
			},
		},
	})

	return cache.NewSnapshot("1",
		map[resource.Type][]types.Resource{
			resource.SecretType: {
				tlsSecret,
			},
		},
	)
}
