package hueclient

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to generate a test certificate
func generateTestCertificate(t *testing.T, cn string, dnsNames []string) ([]byte, []byte) {
	// Generate CA private key
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create CA certificate template
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test CA",
			Organization: []string{"Test CA Org"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Self-sign CA certificate
	caCertBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)

	// Parse the CA certificate
	caCert, err := x509.ParseCertificate(caCertBytes)
	require.NoError(t, err)

	// Generate server private key
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create server certificate template
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: cn,
		},
		NotBefore:   time.Now().Add(-1 * time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	// Add DNS names if provided
	if len(dnsNames) > 0 {
		serverTemplate.DNSNames = dnsNames
	}

	// Sign server certificate with CA
	serverCertBytes, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	require.NoError(t, err)

	// Encode CA certificate to PEM
	caPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertBytes,
	})

	return caPEM, serverCertBytes
}

func TestNewBridgeTLSConfig(t *testing.T) {
	tests := []struct {
		name      string
		bridgeID  string
		setupCert func(t *testing.T) []byte
		wantErr   bool
		errorMsg  string
	}{
		{
			name:     "creates TLS config with valid certificate",
			bridgeID: "test-bridge-123",
			setupCert: func(t *testing.T) []byte {
				caPEM, _ := generateTestCertificate(t, "test-bridge-123", nil)
				return caPEM
			},
			wantErr: false,
		},
		{
			name:     "handles bridge ID case conversion",
			bridgeID: "TEST-BRIDGE-UPPER",
			setupCert: func(t *testing.T) []byte {
				caPEM, _ := generateTestCertificate(t, "test-bridge-upper", nil)
				return caPEM
			},
			wantErr: false,
		},
		{
			name:     "returns error with empty CA bundle",
			bridgeID: "test-bridge-123",
			setupCert: func(t *testing.T) []byte {
				return nil
			},
			wantErr:  true,
			errorMsg: "failed to append x509 certs",
		},
		{
			name:     "returns error with invalid PEM format",
			bridgeID: "test-bridge-123",
			setupCert: func(t *testing.T) []byte {
				return []byte("invalid pem content")
			},
			wantErr:  true,
			errorMsg: "failed to append x509 certs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewBridgeTLSConfig(tt.bridgeID, tt.setupCert(t))

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, config)
			assert.True(t, config.InsecureSkipVerify) // Required for custom verification
			assert.NotNil(t, config.RootCAs)
			assert.NotNil(t, config.VerifyPeerCertificate)

			// Verify bridge ID is lowercased
			assert.Equal(t, strings.ToLower(tt.bridgeID), config.ServerName)
		})
	}
}

func TestLoadCABundle(t *testing.T) {
	t.Run("reads the file set via HUE_CA_CERTS_PATH", func(t *testing.T) {
		caPEM, _ := generateTestCertificate(t, "test-bridge-123", nil)
		certPath := filepath.Join(t.TempDir(), "ca-bundle.pem")
		require.NoError(t, os.WriteFile(certPath, caPEM, 0600))
		t.Setenv("HUE_CA_CERTS_PATH", certPath)

		pem, source, err := LoadCABundle()

		require.NoError(t, err)
		assert.Equal(t, caPEM, pem)
		assert.Equal(t, certPath, source)
	})

	t.Run("returns error when HUE_CA_CERTS_PATH file does not exist", func(t *testing.T) {
		t.Setenv("HUE_CA_CERTS_PATH", filepath.Join(t.TempDir(), "nonexistent-ca-bundle.pem"))

		_, _, err := LoadCABundle()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "CA bundle not found")
	})

	t.Run("falls back to the embedded CA bundle", func(t *testing.T) {
		t.Setenv("HUE_CA_CERTS_PATH", "")

		pem, source, err := LoadCABundle()

		// Whether a bundle is embedded depends on the build (the bundle is not
		// part of the repository), so both outcomes are valid here.
		if _, statErr := embeddedCerts.ReadFile(caBundleFile); statErr != nil {
			require.Error(t, err)
			assert.Contains(t, err.Error(), "built without an embedded CA bundle")
			return
		}
		require.NoError(t, err)
		assert.NotEmpty(t, pem)
		assert.Equal(t, "embedded", source)
	})
}

func TestCreateCustomCertVerifier(t *testing.T) {
	// Generate test certificates
	caPEM, serverCertBytes := generateTestCertificate(t, "test-bridge-123", nil)

	// Create cert pool with CA
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(caPEM)
	require.True(t, ok)

	tests := []struct {
		name           string
		expectedServer string
		serverCert     []byte
		wantErr        bool
		errorMsg       string
	}{
		{
			name:           "verifies certificate with matching CN",
			expectedServer: "test-bridge-123",
			serverCert:     serverCertBytes,
			wantErr:        false,
		},
		{
			name:           "returns error with mismatched CN",
			expectedServer: "different-bridge-id",
			serverCert:     serverCertBytes,
			wantErr:        true,
			errorMsg:       "does not match expected",
		},
		{
			name:           "returns error with no certificate provided",
			expectedServer: "test-bridge-123",
			serverCert:     nil,
			wantErr:        true,
			errorMsg:       "no server certificate provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := createCustomCertVerifier(tt.expectedServer, certPool)
			require.NotNil(t, verifier)

			var rawCerts [][]byte
			if tt.serverCert != nil {
				rawCerts = [][]byte{tt.serverCert}
			}

			err := verifier(rawCerts, nil)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCreateCustomCertVerifier_WithSAN(t *testing.T) {
	// Generate certificate with SAN entries
	dnsNames := []string{"test-bridge-san"}
	caPEM, serverCertBytes := generateTestCertificate(t, "test-bridge-cn", dnsNames)

	// Create cert pool with CA
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(caPEM)
	require.True(t, ok)

	tests := []struct {
		name           string
		expectedServer string
		wantErr        bool
		errorMsg       string
	}{
		{
			name:           "verifies certificate with matching SAN",
			expectedServer: "test-bridge-san",
			wantErr:        false,
		},
		{
			name:           "returns error with mismatched SAN",
			expectedServer: "wrong-bridge-id",
			wantErr:        true,
			errorMsg:       "not found in certificate SANs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := createCustomCertVerifier(tt.expectedServer, certPool)
			require.NotNil(t, verifier)

			rawCerts := [][]byte{serverCertBytes}
			err := verifier(rawCerts, nil)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCreateCustomCertVerifier_InvalidCertificate(t *testing.T) {
	certPool := x509.NewCertPool()
	verifier := createCustomCertVerifier("test-bridge", certPool)

	tests := []struct {
		name     string
		rawCerts [][]byte
		errorMsg string
	}{
		{
			name:     "returns error with invalid certificate data",
			rawCerts: [][]byte{[]byte("invalid certificate bytes")},
			errorMsg: "failed to parse server certificate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifier(tt.rawCerts, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestNewBridgeTLSConfig_Integration(t *testing.T) {
	// Integration test: Create a complete valid certificate and config
	t.Run("full integration with valid certificate", func(t *testing.T) {
		bridgeID := "ecb5fafffe123456"
		caPEM, _ := generateTestCertificate(t, bridgeID, nil)

		config, err := NewBridgeTLSConfig(bridgeID, caPEM)
		require.NoError(t, err)
		assert.NotNil(t, config)

		// Verify config properties
		assert.True(t, config.InsecureSkipVerify)
		assert.NotNil(t, config.RootCAs)
		assert.NotNil(t, config.VerifyPeerCertificate)
		assert.Equal(t, bridgeID, config.ServerName)
	})
}
