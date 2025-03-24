package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ConnectionProfile represents a Fabric connection profile
type ConnectionProfile struct {
	Name          string                   `json:"name"`
	Version       string                   `json:"version"`
	Client        ClientConfig             `json:"client"`
	Channels      map[string]ChannelConfig `json:"channels"`
	Organizations map[string]OrgConfig     `json:"organizations"`
	Orderers      map[string]OrdererConfig `json:"orderers"`
	Peers         map[string]PeerConfig    `json:"peers"`
	CAs           map[string]CAConfig      `json:"certificateAuthorities"`
}

// ClientConfig represents the client section of a connection profile
type ClientConfig struct {
	Organization string           `json:"organization"`
	Connection   ConnectionConfig `json:"connection"`
}

// ConnectionConfig represents connection timeouts
type ConnectionConfig struct {
	Timeout TimeoutConfig `json:"timeout"`
}

// TimeoutConfig represents timeout settings
type TimeoutConfig struct {
	Peer    PeerTimeoutConfig `json:"peer"`
	Orderer string            `json:"orderer"`
}

// PeerTimeoutConfig represents peer timeout settings
type PeerTimeoutConfig struct {
	Endorser string `json:"endorser"`
}

// ChannelConfig represents a channel configuration
type ChannelConfig struct {
	Orderers []string                 `json:"orderers"`
	Peers    map[string]PeerRoleConfig `json:"peers"`
}

// PeerRoleConfig represents peer roles for a channel
type PeerRoleConfig struct {
	EndorsingPeer  bool `json:"endorsingPeer"`
	ChaincodeQuery bool `json:"chaincodeQuery"`
	LedgerQuery    bool `json:"ledgerQuery"`
	EventSource    bool `json:"eventSource"`
}

// OrgConfig represents an organization configuration
type OrgConfig struct {
	MSPID                  string   `json:"mspid"`
	Peers                  []string `json:"peers"`
	CertificateAuthorities []string `json:"certificateAuthorities"`
}

// OrdererConfig represents an orderer configuration
type OrdererConfig struct {
	URL          string            `json:"url"`
	GRPCOptions  map[string]interface{} `json:"grpcOptions"`
	TLSCACerts   CertConfig        `json:"tlsCACerts"`
}

// PeerConfig represents a peer configuration
type PeerConfig struct {
	URL          string            `json:"url"`
	GRPCOptions  map[string]interface{} `json:"grpcOptions"`
	TLSCACerts   CertConfig        `json:"tlsCACerts"`
}

// CAConfig represents a certificate authority configuration
type CAConfig struct {
	URL         string            `json:"url"`
	CAName      string            `json:"caName"`
	TLSCACerts  CertConfig        `json:"tlsCACerts"`
	HTTPOptions map[string]interface{} `json:"httpOptions"`
}

// CertConfig represents TLS certificate configuration
type CertConfig struct {
	Path string `json:"path"`
}

func main() {
	// Parse command line arguments
	connectionProfile := flag.String("profile", "config/connection-profile-simple.json", "Path to connection profile")
	flag.Parse()
	
	fmt.Printf("Validating connection profile: %s\n", *connectionProfile)
	
	// Check if the connection profile exists
	if _, err := os.Stat(*connectionProfile); os.IsNotExist(err) {
		fmt.Printf("Error: Connection profile not found: %s\n", *connectionProfile)
		os.Exit(1)
	}
	
	// Read the connection profile
	content, err := ioutil.ReadFile(*connectionProfile)
	if err != nil {
		fmt.Printf("Failed to read connection profile: %s\n", err)
		os.Exit(1)
	}
	
	// Parse the connection profile
	var profile ConnectionProfile
	if err := json.Unmarshal(content, &profile); err != nil {
		fmt.Printf("Failed to parse connection profile: %s\n", err)
		fmt.Printf("This indicates a JSON syntax error. Please check the file format.\n")
		os.Exit(1)
	}
	
	fmt.Printf("✓ Connection profile JSON syntax is valid\n")
	
	// Validate the connection profile structure
	fmt.Printf("Validating connection profile structure...\n")
	
	// Check client section
	if profile.Client.Organization == "" {
		fmt.Printf("✗ Missing or empty client.organization\n")
	} else {
		fmt.Printf("✓ Client organization: %s\n", profile.Client.Organization)
	}
	
	// Check channels section
	if len(profile.Channels) == 0 {
		fmt.Printf("✗ No channels defined\n")
	} else {
		fmt.Printf("✓ Channels defined: %d\n", len(profile.Channels))
		for channelName, channel := range profile.Channels {
			fmt.Printf("  - Channel: %s\n", channelName)
			
			if len(channel.Orderers) == 0 {
				fmt.Printf("    ✗ No orderers defined for channel\n")
			} else {
				fmt.Printf("    ✓ Orderers: %v\n", channel.Orderers)
			}
			
			if len(channel.Peers) == 0 {
				fmt.Printf("    ✗ No peers defined for channel\n")
			} else {
				fmt.Printf("    ✓ Peers: %d\n", len(channel.Peers))
			}
		}
	}
	
	// Check organizations section
	if len(profile.Organizations) == 0 {
		fmt.Printf("✗ No organizations defined\n")
	} else {
		fmt.Printf("✓ Organizations defined: %d\n", len(profile.Organizations))
		for orgName, org := range profile.Organizations {
			fmt.Printf("  - Organization: %s\n", orgName)
			
			if org.MSPID == "" {
				fmt.Printf("    ✗ Missing MSPID\n")
			} else {
				fmt.Printf("    ✓ MSPID: %s\n", org.MSPID)
			}
			
			if len(org.Peers) == 0 {
				fmt.Printf("    ✗ No peers defined for organization\n")
			} else {
				fmt.Printf("    ✓ Peers: %v\n", org.Peers)
			}
		}
	}
	
	// Check orderers section
	if len(profile.Orderers) == 0 {
		fmt.Printf("✗ No orderers defined\n")
	} else {
		fmt.Printf("✓ Orderers defined: %d\n", len(profile.Orderers))
		for ordererName, orderer := range profile.Orderers {
			fmt.Printf("  - Orderer: %s\n", ordererName)
			
			if orderer.URL == "" {
				fmt.Printf("    ✗ Missing URL\n")
			} else {
				fmt.Printf("    ✓ URL: %s\n", orderer.URL)
			}
			
			// Check TLS certificate
			if orderer.TLSCACerts.Path == "" {
				fmt.Printf("    ✗ Missing TLS CA certificate path\n")
			} else {
				certPath := orderer.TLSCACerts.Path
				if filepath.IsAbs(certPath) {
					fmt.Printf("    ✓ TLS CA certificate path: %s\n", certPath)
					
					// Check if the certificate file exists
					if _, err := os.Stat(certPath); os.IsNotExist(err) {
						fmt.Printf("    ✗ TLS CA certificate file not found\n")
					} else {
						fmt.Printf("    ✓ TLS CA certificate file exists\n")
					}
				} else {
					fmt.Printf("    ⚠ TLS CA certificate path is not absolute: %s\n", certPath)
					
					// Try to resolve relative to the connection profile
					absProfilePath, err := filepath.Abs(*connectionProfile)
					if err == nil {
						profileDir := filepath.Dir(absProfilePath)
						absCertPath := filepath.Join(profileDir, certPath)
						fmt.Printf("    ⚠ Attempting to resolve relative to profile: %s\n", absCertPath)
						
						if _, err := os.Stat(absCertPath); os.IsNotExist(err) {
							fmt.Printf("    ✗ TLS CA certificate file not found at resolved path\n")
						} else {
							fmt.Printf("    ✓ TLS CA certificate file exists at resolved path\n")
						}
					}
				}
			}
			
			// Check gRPC options
			if orderer.GRPCOptions == nil {
				fmt.Printf("    ✗ Missing gRPC options\n")
			} else {
				fmt.Printf("    ✓ gRPC options defined\n")
				
				// Check for ssl-target-name-override
				if sslOverride, ok := orderer.GRPCOptions["ssl-target-name-override"]; ok {
					fmt.Printf("    ✓ ssl-target-name-override: %v\n", sslOverride)
				} else {
					fmt.Printf("    ⚠ Missing ssl-target-name-override\n")
				}
				
				// Check for allow-insecure
				if allowInsecure, ok := orderer.GRPCOptions["allow-insecure"]; ok {
					fmt.Printf("    ✓ allow-insecure: %v\n", allowInsecure)
				} else {
					fmt.Printf("    ⚠ Missing allow-insecure\n")
				}
			}
		}
	}
	
	// Check peers section
	if len(profile.Peers) == 0 {
		fmt.Printf("✗ No peers defined\n")
	} else {
		fmt.Printf("✓ Peers defined: %d\n", len(profile.Peers))
		for peerName, peer := range profile.Peers {
			fmt.Printf("  - Peer: %s\n", peerName)
			
			if peer.URL == "" {
				fmt.Printf("    ✗ Missing URL\n")
			} else {
				fmt.Printf("    ✓ URL: %s\n", peer.URL)
			}
			
			// Check TLS certificate
			if peer.TLSCACerts.Path == "" {
				fmt.Printf("    ✗ Missing TLS CA certificate path\n")
			} else {
				certPath := peer.TLSCACerts.Path
				if filepath.IsAbs(certPath) {
					fmt.Printf("    ✓ TLS CA certificate path: %s\n", certPath)
					
					// Check if the certificate file exists
					if _, err := os.Stat(certPath); os.IsNotExist(err) {
						fmt.Printf("    ✗ TLS CA certificate file not found\n")
					} else {
						fmt.Printf("    ✓ TLS CA certificate file exists\n")
					}
				} else {
					fmt.Printf("    ⚠ TLS CA certificate path is not absolute: %s\n", certPath)
					
					// Try to resolve relative to the connection profile
					absProfilePath, err := filepath.Abs(*connectionProfile)
					if err == nil {
						profileDir := filepath.Dir(absProfilePath)
						absCertPath := filepath.Join(profileDir, certPath)
						fmt.Printf("    ⚠ Attempting to resolve relative to profile: %s\n", absCertPath)
						
						if _, err := os.Stat(absCertPath); os.IsNotExist(err) {
							fmt.Printf("    ✗ TLS CA certificate file not found at resolved path\n")
						} else {
							fmt.Printf("    ✓ TLS CA certificate file exists at resolved path\n")
						}
					}
				}
			}
		}
	}
	
	fmt.Printf("\nConnection profile validation complete.\n")
}
