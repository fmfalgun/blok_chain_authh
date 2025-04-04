package fabric

import (
	"os"
	"path/filepath"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/pkg/errors"
)

const (
	// DefaultConfigPath is the default path to the connection profile
	DefaultConfigPath = "config/connection-profile.json"
	
	// DefaultChannel is the default channel name
	DefaultChannel = "chaichis-channel"
)

// Client represents a Fabric client
type Client struct {
	configPath  string
	channelName string
	wallet      *Wallet
	gateway     *gateway.Gateway
	debug       bool
}

// ClientOptions contains options for creating a Fabric client
type ClientOptions struct {
	ConfigPath  string
	ChannelName string
	WalletPath  string
	Debug       bool
}

// NewClient creates a new Fabric client
func NewClient(options ClientOptions) (*Client, error) {
	// Set default options if not provided
	if options.ConfigPath == "" {
		options.ConfigPath = DefaultConfigPath
	}
	
	if options.ChannelName == "" {
		options.ChannelName = DefaultChannel
	}
	
	// Create wallet
	wallet, err := NewWallet(options.WalletPath)
	if err != nil {
		return nil, err
	}
	
	return &Client{
		configPath:  options.ConfigPath,
		channelName: options.ChannelName,
		wallet:      wallet,
		debug:       options.Debug,
	}, nil
}

// DefaultClient creates a client with default options
func DefaultClient() (*Client, error) {
	return NewClient(ClientOptions{})
}

// Connect connects to the Fabric network using the specified identity
func (c *Client) Connect(identity string) error {
	// Ensure identity exists in wallet
	if !c.wallet.Exists(identity) {
		return errors.Errorf("identity '%s' not found in wallet", identity)
	}
	
	// Ensure connection profile exists
	if _, err := os.Stat(c.configPath); os.IsNotExist(err) {
		return errors.Errorf("connection profile not found at '%s'", c.configPath)
	}
	
	// Load connection profile
	ccpPath, err := filepath.Abs(c.configPath)
	if err != nil {
		return errors.Wrap(err, "failed to get absolute path for connection profile")
	}
	
	if c.debug {
		fmt.Printf("Using connection profile at: %s\n", ccpPath)
	}
	
	configProvider := func() ([]core.ConfigBackend, error) {
		return config.FromFile(ccpPath)()
	}
	
	// Connect to gateway
	gw, err := gateway.Connect(
		gateway.WithConfig(configProvider),
		gateway.WithIdentity(c.wallet.wallet, identity),
	)
	if err != nil {
		return errors.Wrap(err, "failed to connect to gateway")
	}
	
	c.gateway = gw
	return nil
}

// GetNetwork returns the Fabric network
func (c *Client) GetNetwork() (*gateway.Network, error) {
	if c.gateway == nil {
		return nil, errors.New("not connected to gateway, call Connect() first")
	}
	
	network, err := c.gateway.GetNetwork(c.channelName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get network '%s'", c.channelName)
	}
	
	return network, nil
}

// GetContract returns a contract from the network
func (c *Client) GetContract(contractID string) (*gateway.Contract, error) {
	network, err := c.GetNetwork()
	if err != nil {
		return nil, err
	}
	
	if c.debug {
		fmt.Printf("Getting contract: %s\n", contractID)
	}
	
	contract := network.GetContract(contractID)
	return contract, nil
}

// Close closes the connection to the Fabric network
func (c *Client) Close() {
	if c.gateway != nil {
		c.gateway.Close()
		c.gateway = nil
	}
}

// GetWallet returns the client's wallet
func (c *Client) GetWallet() *Wallet {
	return c.wallet
}

// EnsureIdentity ensures that the specified identity exists in the wallet
func (c *Client) EnsureIdentity(identity string) error {
	return c.wallet.EnsureIdentity(identity)
}

// SetDebug enables or disables debug output
func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}
