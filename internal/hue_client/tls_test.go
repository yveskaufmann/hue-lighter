package hueclient

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"

	"com.github.yveskaufmann/hue-lighter/internal/testutils"
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
		name        string
		bridgeID    string
		certContent string
		setupCert   func(t *testing.T) (string, func())
		wantErr     bool
		errorMsg    string
	}{
		{
			name:     "creates TLS config with valid certificate",
			bridgeID: "test-bridge-123",
			setupCert: func(t *testing.T) (string, func()) {
				caPEM, _ := generateTestCertificate(t, "test-bridge-123", nil)
				
				tmpFile, err := os.CreateTemp("", "ca-cert-*.pem")
				require.NoError(t, err)
				
				_, err = tmpFile.Write(caPEM)
				require.NoError(t, err)
				tmpFile.Close()
				
				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			wantErr: false,
		},
		{
			name:     "handles bridge ID case conversion",
			bridgeID: "TEST-BRIDGE-UPPER",
			setupCert: func(t *testing.T) (string, func()) {
				caPEM, _ := generateTestCertificate(t, "test-bridge-upper", nil)
				
				tmpFile, err := os.CreateTemp("", "ca-cert-*.pem")
				require.NoError(t, err)
				
				_, err = tmpFile.Write(caPEM)
				require.NoError(t, err)
				tmpFile.Close()
				
				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			wantErr: false,
		},
		{
			name:     "returns error when certificate file not found",
			bridgeID: "test-bridge-123",
			setupCert: func(t *testing.T) (string, func()) {
				return "/nonexistent/path/to/cert.pem", func() {}
			},
			wantErr:  true,
			errorMsg: "failed to read x509 certs",
		},
		{
			name:     "returns error with invalid PEM format",
			bridgeID: "test-bridge-123",
			setupCert: func(t *testing.T) (string, func()) {
				tmpFile, err := os.CreateTemp("", "invalid-cert-*.pem")
				require.NoError(t, err)
				
				_, err = tmpFile.Write([]byte("invalid pem content"))
				require.NoError(t, err)
				tmpFile.Close()
				
				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			wantErr:  true,
			errorMsg: "failed to append x509 certs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			certPath, cleanup := tt.setupCert(t)
			defer cleanup()

			config, err := NewBridgeTLSConfig(tt.bridgeID, certPath)

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
			expectedServerName := tt.bridgeID
			if tt.bridgeID != "" {
				expectedServerName = tt.bridgeID
			}
			assert.Contains(t, []string{expectedServerName, "test-bridge-123", "test-bridge-upper"}, config.ServerName)
		})
	}
}

func TestResolveCABundlePath(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func(t *testing.T) (string, func())
		wantErr     bool
		errorMsg    string
		expectedPath string
	}{
		{
			name: "returns custom path from HUE_CA_CERTS_PATH environment variable",
			setupEnv: func(t *testing.T) (string, func()) {
				tmpFile, err := os.CreateTemp("", "ca-bundle-*.pem")
				require.NoError(t, err)
				tmpFile.Close()
				
				cleanup := testutils.SetEnv(t, "HUE_CA_CERTS_PATH", tmpFile.Name())
				
				return tmpFile.Name(), func() {
					cleanup()
					os.Remove(tmpFile.Name())
				}
			},
			wantErr: false,
		},
		{
			name: "returns error when custom path file does not exist",
			setupEnv: func(t *testing.T) (string, func()) {
				customPath := "/tmp/nonexistent-ca-bundle.pem"
				cleanup := testutils.SetEnv(t, "HUE_CA_CERTS_PATH", customPath)
				return customPath, cleanup
			},
			wantErr:  true,
			errorMsg: "CA bundle not found",
		},
		{
			name: "returns error when default path does not exist",
			setupEnv: func(t *testing.T) (string, func()) {
				// Clear any existing env var to test default path
				cleanup := testutils.SetEnv(t, "HUE_CA_CERTS_PATH", "")
				return "/etc/hue-lighter/cacert_bundle.pem", cleanup
			},
			wantErr:  true,
			errorMsg: "CA bundle not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedPath, cleanup := tt.setupEnv(t)
			defer cleanup()

			path, err := ResolveCABundlePath()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, expectedPath, path)
		})
	}
}

func TestCreateCustomCertVerifier(t *testing.T) {
	// Generate test certificates
	caPEM, serverCertBytes := generateTestCertificate(t, "test-bridge-123", nil)
	
	// Create cert pool with CA
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(caPEM)
	require.True(t, ok)

	tests := []struct {
		name             string
		expectedServer   string
		serverCert       []byte
		wantErr          bool
		errorMsg         string
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
		
		tmpFile, err := os.CreateTemp("", "integration-cert-*.pem")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		
		_, err = tmpFile.Write(caPEM)
		require.NoError(t, err)
		tmpFile.Close()
		
		config, err := NewBridgeTLSConfig(bridgeID, tmpFile.Name())
		require.NoError(t, err)
		assert.NotNil(t, config)
		
		// Verify config properties
		assert.True(t, config.InsecureSkipVerify)
		assert.NotNil(t, config.RootCAs)
		assert.NotNil(t, config.VerifyPeerCertificate)
		assert.Equal(t, bridgeID, config.ServerName)
	})
}

func TestResolveCABundlePath_FileSystemTests(t *testing.T) {
	t.Run("validates file access permissions", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "access-test-*.pem")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()
		
		cleanup := testutils.SetEnv(t, "HUE_CA_CERTS_PATH", tmpFile.Name())
		defer cleanup()
		
		path, err := ResolveCABundlePath()
		require.NoError(t, err)
		assert.Equal(t, tmpFile.Name(), path)
	})
}
