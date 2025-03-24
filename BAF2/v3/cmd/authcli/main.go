package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/chaichis-network/v3/internal/auth"
	"github.com/chaichis-network/v3/internal/fabric"
	"github.com/chaichis-network/v3/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	logLevel        string
	configPath      string
	walletPath      string
	identityName    string
	clientID        string
	deviceID        string
	capabilities    []string
	sessionDir      string
	debugMode       bool // Added debug mode flag
	
	// Global variables
	log *logger.Logger
)

func init() {
	// Initialize logger
	log = logger.New("info")
	
	// Root command flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "config/connection-profile.json", "Path to connection profile")
	rootCmd.PersistentFlags().StringVar(&walletPath, "wallet", "wallet", "Path to wallet directory")
	rootCmd.PersistentFlags().StringVar(&identityName, "identity", "admin", "Identity name to use")
	rootCmd.PersistentFlags().StringVar(&sessionDir, "session-dir", "sessions", "Path to session directory")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug mode for Fabric client") // Added debug flag
	
	// Register client command flags
	registerClientCmd.Flags().StringVar(&clientID, "client-id", "", "Client ID to register")
	registerClientCmd.MarkFlagRequired("client-id")
	
	// Register device command flags
	registerDeviceCmd.Flags().StringVar(&deviceID, "device-id", "", "Device ID to register")
	registerDeviceCmd.Flags().StringSliceVar(&capabilities, "capabilities", []string{}, "Device capabilities (comma-separated)")
	registerDeviceCmd.MarkFlagRequired("device-id")
	
	// Authenticate command flags
	authenticateCmd.Flags().StringVar(&clientID, "client-id", "", "Client ID to authenticate")
	authenticateCmd.Flags().StringVar(&deviceID, "device-id", "", "Device ID to access")
	authenticateCmd.MarkFlagRequired("client-id")
	authenticateCmd.MarkFlagRequired("device-id")
	
	// Access device command flags
	accessDeviceCmd.Flags().StringVar(&clientID, "client-id", "", "Client ID requesting access")
	accessDeviceCmd.Flags().StringVar(&deviceID, "device-id", "", "Device ID to access")
	accessDeviceCmd.MarkFlagRequired("client-id")
	accessDeviceCmd.MarkFlagRequired("device-id")
	
	// Get device data command flags
	getDeviceDataCmd.Flags().StringVar(&deviceID, "device-id", "", "Device ID to query")
	getDeviceDataCmd.MarkFlagRequired("device-id")
	
	// Close session command flags
	closeSessionCmd.Flags().StringVar(&clientID, "client-id", "", "Client ID for the session")
	closeSessionCmd.Flags().StringVar(&deviceID, "device-id", "", "Device ID for the session")
	closeSessionCmd.MarkFlagRequired("client-id")
	closeSessionCmd.MarkFlagRequired("device-id")
	
	// List sessions command flags
	listSessionsCmd.Flags().StringVar(&clientID, "client-id", "", "Filter sessions by client ID (optional)")
	
	// Add subcommands to root command
	rootCmd.AddCommand(
		registerClientCmd,
		registerDeviceCmd,
		authenticateCmd,
		accessDeviceCmd,
		getDeviceDataCmd,
		closeSessionCmd,
		listSessionsCmd,
	)
}

var rootCmd = &cobra.Command{
	Use:   "authcli",
	Short: "Authentication Framework CLI",
	Long:  `Command-line interface for the Hyperledger Fabric Authentication Framework`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set log level
		log = logger.New(logLevel)
	},
}

var registerClientCmd = &cobra.Command{
	Use:   "register-client",
	Short: "Register a client with the Authentication Server",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create Fabric client
		fabricClient, err := fabric.NewClient(fabric.ClientOptions{
			ConfigPath:  configPath,
			WalletPath:  walletPath,
			Debug:       debugMode, // Enable debug mode based on flag
		})
		if err != nil {
			return fmt.Errorf("failed to create Fabric client: %v", err)
		}
		
		// Ensure identity exists in wallet
		if err := fabricClient.EnsureIdentity(identityName); err != nil {
			return fmt.Errorf("failed to ensure identity: %v", err)
		}
		
		// Create client manager
		clientManager, err := auth.NewClientManager(fabricClient, identityName)
		if err != nil {
			return fmt.Errorf("failed to create client manager: %v", err)
		}
		defer clientManager.Close()
		
		// Register client
		if err := clientManager.RegisterClient(clientID); err != nil {
			return fmt.Errorf("failed to register client: %v", err)
		}
		
		log.Infof("Client %s registered successfully", clientID)
		return nil
	},
}

var registerDeviceCmd = &cobra.Command{
	Use:   "register-device",
	Short: "Register an IoT device with the ISV",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create Fabric client
		fabricClient, err := fabric.NewClient(fabric.ClientOptions{
			ConfigPath:  configPath,
			WalletPath:  walletPath,
			Debug:       debugMode, // Enable debug mode based on flag
		})
		if err != nil {
			return fmt.Errorf("failed to create Fabric client: %v", err)
		}
		
		// Ensure identity exists in wallet
		if err := fabricClient.EnsureIdentity(identityName); err != nil {
			return fmt.Errorf("failed to ensure identity: %v", err)
		}
		
		// Create device manager
		deviceManager, err := auth.NewDeviceManager(fabricClient, identityName)
		if err != nil {
			return fmt.Errorf("failed to create device manager: %v", err)
		}
		
		// Register device
		if err := deviceManager.RegisterDevice(deviceID, capabilities); err != nil {
			return fmt.Errorf("failed to register device: %v", err)
		}
		
		log.Infof("Device %s registered successfully with capabilities: %s", deviceID, strings.Join(capabilities, ", "))
		return nil
	},
}

var authenticateCmd = &cobra.Command{
	Use:   "authenticate",
	Short: "Authenticate a client for device access",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create Fabric client
		fabricClient, err := fabric.NewClient(fabric.ClientOptions{
			ConfigPath:  configPath,
			WalletPath:  walletPath,
			Debug:       debugMode, // Enable debug mode based on flag
		})
		if err != nil {
			return fmt.Errorf("failed to create Fabric client: %v", err)
		}
		
		// Ensure identity exists in wallet
		if err := fabricClient.EnsureIdentity(identityName); err != nil {
			return fmt.Errorf("failed to ensure identity: %v", err)
		}
		
		// Create client manager
		clientManager, err := auth.NewClientManager(fabricClient, identityName)
		if err != nil {
			return fmt.Errorf("failed to create client manager: %v", err)
		}
		defer clientManager.Close()
		
		// Authenticate client
		if err := clientManager.Authenticate(clientID, deviceID); err != nil {
			return fmt.Errorf("failed to authenticate: %v", err)
		}
		
		log.Infof("Authentication successful for client %s to access device %s", clientID, deviceID)
		return nil
	},
}

var accessDeviceCmd = &cobra.Command{
	Use:   "access-device",
	Short: "Access an IoT device",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create Fabric client
		fabricClient, err := fabric.NewClient(fabric.ClientOptions{
			ConfigPath:  configPath,
			WalletPath:  walletPath,
			Debug:       debugMode, // Enable debug mode based on flag
		})
		if err != nil {
			return fmt.Errorf("failed to create Fabric client: %v", err)
		}
		
		// Ensure identity exists in wallet
		if err := fabricClient.EnsureIdentity(identityName); err != nil {
			return fmt.Errorf("failed to ensure identity: %v", err)
		}
		
		// Create device manager
		deviceManager, err := auth.NewDeviceManager(fabricClient, identityName)
		if err != nil {
			return fmt.Errorf("failed to create device manager: %v", err)
		}
		
		// Access device
		session, err := deviceManager.AccessDevice(clientID, deviceID)
		if err != nil {
			return fmt.Errorf("failed to access device: %v", err)
		}
		
		// Create session manager
		sessionManager := auth.NewSessionManager(sessionDir)
		
		// Save session
		if err := sessionManager.SaveSession(session); err != nil {
			return fmt.Errorf("failed to save session: %v", err)
		}
		
		log.Infof("Access granted to device %s for client %s", deviceID, clientID)
		log.Infof("Session ID: %s", session.SessionID)
		return nil
	},
}

var getDeviceDataCmd = &cobra.Command{
	Use:   "get-device-data",
	Short: "Get data for an IoT device",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create Fabric client
		fabricClient, err := fabric.NewClient(fabric.ClientOptions{
			ConfigPath:  configPath,
			WalletPath:  walletPath,
			Debug:       debugMode, // Enable debug mode based on flag
		})
		if err != nil {
			return fmt.Errorf("failed to create Fabric client: %v", err)
		}
		
		// Ensure identity exists in wallet
		if err := fabricClient.EnsureIdentity(identityName); err != nil {
			return fmt.Errorf("failed to ensure identity: %v", err)
		}
		
		// Create device manager
		deviceManager, err := auth.NewDeviceManager(fabricClient, identityName)
		if err != nil {
			return fmt.Errorf("failed to create device manager: %v", err)
		}
		
		// Get device data
		device, err := deviceManager.GetDeviceData(deviceID)
		if err != nil {
			return fmt.Errorf("failed to get device data: %v", err)
		}
		
		// Display device information
		fmt.Printf("Device Information for %s:\n", deviceID)
		fmt.Printf("  Status: %s\n", device.Status)
		fmt.Printf("  Capabilities: %s\n", strings.Join(device.Capabilities, ", "))
		if device.LastSeen != "" {
			fmt.Printf("  Last Seen: %s\n", device.LastSeen)
		}
		if device.RegisteredAt != "" {
			fmt.Printf("  Registered At: %s\n", device.RegisteredAt)
		}
		
		return nil
	},
}

var closeSessionCmd = &cobra.Command{
	Use:   "close-session",
	Short: "Close an active session with an IoT device",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create session manager
		sessionManager := auth.NewSessionManager(sessionDir)
		
		// Get session
		_, err := sessionManager.GetSession(clientID, deviceID)
		if err != nil {
			return fmt.Errorf("failed to get session: %v", err)
		}
		
		// Create Fabric client
		fabricClient, err := fabric.NewClient(fabric.ClientOptions{
			ConfigPath:  configPath,
			WalletPath:  walletPath,
			Debug:       debugMode, // Enable debug mode based on flag
		})
		if err != nil {
			return fmt.Errorf("failed to create Fabric client: %v", err)
		}
		
		// Ensure identity exists in wallet
		if err := fabricClient.EnsureIdentity(identityName); err != nil {
			return fmt.Errorf("failed to ensure identity: %v", err)
		}
		
		// Create device manager
		deviceManager, err := auth.NewDeviceManager(fabricClient, identityName)
		if err != nil {
			return fmt.Errorf("failed to create device manager: %v", err)
		}
		
		// Close session
		if err := deviceManager.CloseSession(clientID, deviceID); err != nil {
			return fmt.Errorf("failed to close session: %v", err)
		}
		
		// Remove session
		if err := sessionManager.RemoveSession(clientID, deviceID); err != nil {
			return fmt.Errorf("failed to remove session: %v", err)
		}
		
		log.Infof("Session closed for client %s and device %s", clientID, deviceID)
		return nil
	},
}

var listSessionsCmd = &cobra.Command{
	Use:   "list-sessions",
	Short: "List active sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create session manager
		sessionManager := auth.NewSessionManager(sessionDir)
		
		var sessions []*auth.Session
		var err error
		
		// List sessions (filtered by client if provided)
		if clientID != "" {
			sessions, err = sessionManager.GetActiveSessionsForClient(clientID)
			if err != nil {
				return fmt.Errorf("failed to get sessions for client %s: %v", clientID, err)
			}
		} else {
			sessions, err = sessionManager.ListActiveSessions()
			if err != nil {
				return fmt.Errorf("failed to list sessions: %v", err)
			}
		}
		
		// Display sessions
		if len(sessions) == 0 {
			fmt.Println("No active sessions found")
			return nil
		}
		
		fmt.Printf("Active Sessions (%d):\n", len(sessions))
		for i, session := range sessions {
			fmt.Printf("%d. Client: %s, Device: %s, Session ID: %s\n", i+1, session.ClientID, session.DeviceID, session.SessionID)
			fmt.Printf("   Status: %s\n", session.Status)
			if session.EstablishedAt != "" {
				fmt.Printf("   Established At: %s\n", session.EstablishedAt)
			}
			if session.ExpiresAt != "" {
				fmt.Printf("   Expires At: %s\n", session.ExpiresAt)
			}
			fmt.Println()
		}
		
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
