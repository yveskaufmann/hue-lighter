package hueclient

import (
	"crypto/tls"
	"crypto/x509"
	"embed"
	"fmt"
	"os"
	"strings"
)

// VerifyPeerCertificate defines the signature for custom certificate verification functions.
// It matches the signature required by tls.Config's VerifyPeerCertificate field.
type VerifyPeerCertificate func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error

// caBundleFile is the name of the Philips Hue CA bundle within the embedded certs directory.
// The bundle is not part of the repository; `make build` copies it from configs/certs.
const caBundleFile = "certs/cacert_bundle.pem"

//go:embed all:certs
var embeddedCerts embed.FS

// NewBridgeTLSConfig creates a tls.Config for connecting to a Philips Hue bridge to
// support accessing its API over HTTPS.
//
// It loads the Philips CA certificate, sets up the root CA pool, and
// configures custom certificate verification to allow CN fallback if SAN is missing.
//
// Parameters:
//   - bridgeId: the expected bridge identifier (CN/SAN).
//   - caPEM: the PEM encoded CA bundle, see LoadCABundle.
func NewBridgeTLSConfig(bridgeId string, caPEM []byte) (*tls.Config, error) {
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("tlsConfig creation error: failed to get system cert pool: %v", err)
	}

	if ok := caCertPool.AppendCertsFromPEM(caPEM); !ok {
		return nil, fmt.Errorf("tlsConfig creation error: failed to append x509 certs to cert pool")
	}

	// Philips Hue API is providing the bridge ID in uppercase, but within certificates it is lowercased.
	bridgeId = strings.ToLower(bridgeId)

	config := &tls.Config{
		// Standard verification must be disabled here; otherwise, our custom verification logic will not be used.
		InsecureSkipVerify:    true,
		RootCAs:               caCertPool,
		ServerName:            bridgeId,
		VerifyPeerCertificate: createCustomCertVerifier(bridgeId, caCertPool),
	}

	return config, nil
}

// LoadCABundle returns the Philips Hue CA bundle and a description of where it
// was loaded from. A file set via `HUE_CA_CERTS_PATH` takes precedence over the
// bundle embedded into the binary at build time.
func LoadCABundle() (pem []byte, source string, err error) {
	if certPath := os.Getenv("HUE_CA_CERTS_PATH"); certPath != "" {
		pem, err := os.ReadFile(certPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, "", fmt.Errorf("CA bundle not found at %s (set via HUE_CA_CERTS_PATH)", certPath)
			}
			return nil, "", fmt.Errorf("failed to read CA bundle %s: %v", certPath, err)
		}
		return pem, certPath, nil
	}

	pem, err = embeddedCerts.ReadFile(caBundleFile)
	if err != nil {
		return nil, "", fmt.Errorf(
			"CA bundle not found: this binary was built without an embedded CA bundle. Obtain the " +
				"Philips Hue CA bundle from " +
				"https://developers.meethue.com/develop/application-design-guidance/using-https/, " +
				"place it at configs/certs/cacert_bundle.pem and rebuild with `make build`, or point " +
				"HUE_CA_CERTS_PATH to it, see README.md for instructions",
		)
	}

	return pem, "embedded", nil
}

// createCustomCertVerifier returns VerifyPeerCertificate function that validates
// the server certificate against the provided root CAs and allows CN fallback
// if SAN is missing.
func createCustomCertVerifier(expectedServerName string, rootCAs *x509.CertPool) VerifyPeerCertificate {
	// The cert provided by the Hue Bridge uses a self-signed certificate and
	// is missing proper SAN entries. They are signed with CN set to the bridge ID only.
	// However, Go's TLS library requires SAN to be set for hostname verification - every certificate
	// without SAN is considered invalid and rejected.
	//
	// See: https://golang.org/pkg/crypto/x509/#Certificate.Verify
	//
	// Therefore we implement custom TLS verification logic here to accept the certificate
	// if the CN matches the expected bridge ID, or if the SAN contains the expected bridge ID.
	// We still verify the certificate chain against the system cert pool + the CA certs provided
	// by Philips.

	// Verify the rawCerts here according to your needs
	// For example, you can parse the certificates and check their fields
	// or compare them against known good certificates.

	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {

		if len(rawCerts) == 0 {
			return fmt.Errorf("no server certificate provided")
		}

		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return fmt.Errorf("failed to parse server certificate: %v", err)
		}

		// Validate the chain
		opts := x509.VerifyOptions{
			Roots:         rootCAs,
			Intermediates: x509.NewCertPool(),
		}
		for _, ic := range rawCerts[1:] {
			if c, err := x509.ParseCertificate(ic); err == nil {
				opts.Intermediates.AddCert(c)
			}
		}
		if _, err := cert.Verify(opts); err != nil {
			return fmt.Errorf("certificate verification failed: %v", err)
		}

		if len(cert.DNSNames) > 0 {
			found := false
			for _, name := range cert.DNSNames {
				if name == expectedServerName {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("server name %s not found in certificate SANs", expectedServerName)
			}
		} else if cert.Subject.CommonName != "" {
			if cert.Subject.CommonName != expectedServerName {
				return fmt.Errorf("server certificate CN %s does not match expected %s", cert.Subject.CommonName, expectedServerName)
			}
		} else {
			return fmt.Errorf("certificate has neither SANs nor CN for hostname verification")
		}
		return nil
	}
}
