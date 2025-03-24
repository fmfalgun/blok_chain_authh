package main

import (
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// extractCertificatesFromFile extracts all certificates from a PEM file
func extractCertificatesFromFile(filePath string) ([]*x509.Certificate, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s", err)
	}

	var certs []*x509.Certificate
	for len(data) > 0 {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %s", err)
		}

		certs = append(certs, cert)
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found in file")
	}

	return certs, nil
}

// convertToPEM converts a certificate to PEM format
func convertToPEM(cert *x509.Certificate) string {
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
	return string(pemData)
}

// findCertificates recursively finds all certificate files in a directory
func findCertificates(dir string) ([]string, error) {
	var certFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Look for common certificate file extensions and names
		fileName := strings.ToLower(info.Name())
		if strings.HasSuffix(fileName, ".crt") ||
			strings.HasSuffix(fileName, ".pem") ||
			strings.HasSuffix(fileName, ".cert") ||
			strings.Contains(fileName, "ca.") ||
			strings.Contains(fileName, "ca-cert") ||
			strings.Contains(fileName, "tlsca") {
			certFiles = append(certFiles, path)
		}

		return nil
	})

	return certFiles, err
}

// validateTLSCertificate validates a TLS certificate
func validateTLSCertificate(certPath string) error {
	certs, err := extractCertificatesFromFile(certPath)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d certificates in %s\n", len(certs), certPath)
	for i, cert := range certs {
		fmt.Printf("Certificate %d:\n", i+1)
		fmt.Printf("  Subject: %s\n", cert.Subject)
		fmt.Printf("  Issuer: %s\n", cert.Issuer)
		fmt.Printf("  Valid from: %s\n", cert.NotBefore)
		fmt.Printf("  Valid until: %s\n", cert.NotAfter)
		fmt.Printf("  Serial number: %s\n", cert.SerialNumber)
		fmt.Printf("  Key usage: %v\n", cert.KeyUsage)
		fmt.Printf("  Is CA: %t\n", cert.IsCA)
		fmt.Printf("  DNS names: %v\n", cert.DNSNames)
	}

	return nil
}

// convertCertificate converts a certificate to the correct format for Fabric
func convertCertificate(certPath, outputPath string) error {
	certs, err := extractCertificatesFromFile(certPath)
	if err != nil {
		return err
	}

	var pemData string
	for _, cert := range certs {
		pemData += convertToPEM(cert)
	}

	if err := ioutil.WriteFile(outputPath, []byte(pemData), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %s", err)
	}

	fmt.Printf("Converted %d certificates to %s\n", len(certs), outputPath)
	return nil
}

// createConnectionProfileWithCerts creates a connection profile with embedded certificates
func createConnectionProfileWithCerts(profilePath, ordererCertPath, peerCertPath, outputPath string) error {
	// Read the connection profile
	profileData, err := ioutil.ReadFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to read connection profile: %s", err)
	}

	// Read the orderer certificate
	ordererCert, err := ioutil.ReadFile(ordererCertPath)
	if err != nil {
		return fmt.Errorf("failed to read orderer certificate: %s", err)
	}

	// Read the peer certificate
	peerCert, err := ioutil.ReadFile(peerCertPath)
	if err != nil {
		return fmt.Errorf("failed to read peer certificate: %s", err)
	}

	// Convert orderer certificate to base64 if needed
	// ordererCertBase64 := base64.StdEncoding.EncodeToString(ordererCert)

	// Convert peer certificate to base64 if needed
	// peerCertBase64 := base64.StdEncoding.EncodeToString(peerCert)

	// Replace certificate paths with actual certificate data
	profileString := string(profileData)
	
	// Find and replace the orderer certificate path
	ordererPathPattern := `"path": ".*orderer.*"`
	if strings.Contains(profileString, `"path":`) {
		ordererCertBlock := fmt.Sprintf(`"pem": %q`, string(ordererCert))
		profileString = strings.Replace(profileString, ordererPathPattern, ordererCertBlock, -1)
	}

	// Find and replace the peer certificate path
	peerPathPattern := `"path": ".*peer.*"`
	if strings.Contains(profileString, `"path":`) {
		peerCertBlock := fmt.Sprintf(`"pem": %q`, string(peerCert))
		profileString = strings.Replace(profileString, peerPathPattern, peerCertBlock, -1)
	}

	// Write the modified connection profile
	if err := ioutil.WriteFile(outputPath, []byte(profileString), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %s", err)
	}

	fmt.Printf("Created connection profile with embedded certificates at %s\n", outputPath)
	return nil
}

func main() {
	// Define command line flags
	validateCmd := flag.Bool("validate", false, "Validate a TLS certificate")
	convertCmd := flag.Bool("convert", false, "Convert a certificate to the correct format")
	findCmd := flag.Bool("find", false, "Find certificates in a directory")
	embedCmd := flag.Bool("embed", false, "Create a connection profile with embedded certificates")

	certPath := flag.String("cert", "", "Path to the certificate file")
	dirPath := flag.String("dir", "", "Path to the directory to search for certificates")
	outputPath := flag.String("output", "", "Path to the output file")
	profilePath := flag.String("profile", "", "Path to the connection profile")
	ordererCertPath := flag.String("orderer-cert", "", "Path to the orderer certificate")
	peerCertPath := flag.String("peer-cert", "", "Path to the peer certificate")

	flag.Parse()

	// Validate command line arguments
	if *validateCmd && *certPath == "" {
		fmt.Println("Error: -cert is required for -validate")
		flag.Usage()
		os.Exit(1)
	}

	if *convertCmd && (*certPath == "" || *outputPath == "") {
		fmt.Println("Error: -cert and -output are required for -convert")
		flag.Usage()
		os.Exit(1)
	}

	if *findCmd && *dirPath == "" {
		fmt.Println("Error: -dir is required for -find")
		flag.Usage()
		os.Exit(1)
	}

	if *embedCmd && (*profilePath == "" || *ordererCertPath == "" || *peerCertPath == "" || *outputPath == "") {
		fmt.Println("Error: -profile, -orderer-cert, -peer-cert, and -output are required for -embed")
		flag.Usage()
		os.Exit(1)
	}

	// Execute the requested command
	var err error
	if *validateCmd {
		err = validateTLSCertificate(*certPath)
	} else if *convertCmd {
		err = convertCertificate(*certPath, *outputPath)
	} else if *findCmd {
		var certFiles []string
		certFiles, err = findCertificates(*dirPath)
		if err == nil {
			fmt.Printf("Found %d certificate files:\n", len(certFiles))
			for _, file := range certFiles {
				fmt.Printf("  %s\n", file)
			}
		}
	} else if *embedCmd {
		err = createConnectionProfileWithCerts(*profilePath, *ordererCertPath, *peerCertPath, *outputPath)
	} else {
		fmt.Println("No command specified")
		flag.Usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
