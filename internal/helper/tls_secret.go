package helper

import (
	v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tlsV3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
)

func TlsSecretFromBytes(name string, certChain []byte, privateKey []byte) *tlsV3.Secret {
	return &tlsV3.Secret{
		Name: name,
		Type: &tlsV3.Secret_TlsCertificate{
			TlsCertificate: &tlsV3.TlsCertificate{
				CertificateChain: &v3.DataSource{
					Specifier: &v3.DataSource_InlineBytes{
						InlineBytes: certChain,
					},
				},
				PrivateKey: &v3.DataSource{
					Specifier: &v3.DataSource_InlineBytes{
						InlineBytes: privateKey,
					},
				},
			},
		},
	}
}
