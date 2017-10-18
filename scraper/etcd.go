package scraper

import (
	"errors"
	"fmt"
	"sync"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/roscopecoltran/e3ch"
)

type EtcdConfig struct {
	Disabled bool `default:"false" help:"Disable etcd client" json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`

	// Clients
	Client *etcd.Client            `gorm:"-" json:"-" yaml:"-" toml:"-"`
	Once   *sync.Once              `gorm:"-" json:"-" yaml:"-" toml:"-"`
	E3ch   *client.EtcdHRCHYClient `gorm:"-" json:"-" yaml:"-" toml:"-"`

	// Errors
	OnceError error `gorm:"-" json:"-" yaml:"-" toml:"-"`
	InitCheck bool  `json:"init_check,omitempty" yaml:"init_check,omitempty" toml:"init_check,omitempty"`

	// Cluster endpoints
	Peers          []string      `gorm:"-" json:"peers,omitempty" yaml:"peers,omitempty" toml:"peers,omitempty"`
	Timeout        time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty" toml:"timeout,omitempty"`
	CommandTimeout time.Duration `json:"commandTimeout,omitempty" yaml:"commandTimeout,omitempty" toml:"commandTimeout,omitempty"`

	Routes []EtcdRoute `json:"routes" yaml:"routes" toml:"routes"`

	// Credentials
	Username         string `json:"username,omitempty" yaml:"username,omitempty" toml:"username,omitempty"`
	Password         string `json:"password,omitempty" yaml:"password,omitempty" toml:"password,omitempty"`
	PasswordFilePath string `json:"password_file,omitempty" yaml:"password_file,omitempty" toml:"password_file,omitempty"`

	// Secured Transport
	IsSecured     bool   `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
	CertFile      string `json:"cert_file,omitempty" yaml:"cert_file,omitempty" toml:"cert_file,omitempty"`
	KeyFile       string `json:"key_file,omitempty" yaml:"key_file,omitempty" toml:"key_file",omitempty`
	TrustedCAFile string `json:"trusted_ca_file,omitempty" yaml:"trusted_ca_file,omitempty" toml:"trusted_ca_file,omitempty"`

	// Root Key Config
	RootKey  string `json:"root_key,omitempty" yaml:"root_key,omitempty" toml:"root_ke,omitemptyy"`
	DirValue string `json:"dir_value,omitempty" yaml:"dir_value,omitempty" toml:"dir_value,omitempty"`

	// Debug activity
	Debug bool `help:"Enable debug output" json:"debug,omitempty" yaml:"debug,omitempty" toml:"debug,omitempty"`
}

func (ectl *EtcdConfig) NewE3Client(conf etcd.Config) (*etcd.Client, error) {
	var err error
	ectl.Client, err = etcd.New(etcd.Config{
		Endpoints:   []string{"etcd1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("failed to init initial etcd v3 client, error: ", err)
		return nil, err
	} else {
		fmt.Println("[SUCCESS] etcd v3 client connected !")
	}
	return ectl.Client, nil
}

// etcdConfig etcd.Config
func (ectl *EtcdConfig) NewE3chClient() (*client.EtcdHRCHYClient, error) {

	if len(ectl.Peers) == 0 {
		ectl.Peers = []string{"http://localhost:2379"}
	}

	if ectl.RootKey == "" {
		ectl.RootKey = "root"
	}

	etcdConfig := etcd.Config{
		Endpoints: ectl.Peers,
		Username:  ectl.Username,
		Password:  ectl.Password,
	}

	if ectl.IsSecured {
		tlsInfo := transport.TLSInfo{
			CertFile:      ectl.CertFile,
			KeyFile:       ectl.KeyFile,
			TrustedCAFile: ectl.TrustedCAFile,
		}
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return nil, err
		}
		etcdConfig.TLS = tlsConfig
	}

	if ectl.Client == nil {
		var cerr error
		ectl.Client, cerr = ectl.NewE3Client(etcdConfig)
		if cerr != nil {
			return nil, cerr
		}
	}

	client, err := client.New(ectl.Client, ectl.RootKey, ectl.DirValue)
	if err != nil {
		return nil, err
	}

	ectl.E3ch = client

	if ectl.InitCheck {
		err := ectl.CheckupE3ch()
		fmt.Println("[ERROR] failed to check all the E3CH client actions, error: ", err)
	}

	return client, client.FormatRootKey()
}

func (ectl *EtcdConfig) CheckupE3ch() error {
	if ectl.E3ch == nil {
		return errors.New("E3ch client not initialized")
	}
	var err error
	err = ectl.E3ch.FormatRootKey() // set the rootKey as directory
	if err != nil {
		fmt.Println("[ERROR] failed to  set the rootKey as directory, error: ", err)
		return err
	} else {
		fmt.Printf(" [SUCCESS] e3ch a set rootKey (%s) as directory !\n", "root")
	}

	// Quick Test
	err = ectl.E3ch.CreateDir("/dir1")
	if err != nil {
		fmt.Println(" [ERROR] failed to CreateDir '/dir1', error: ", err)
	}

	err = ectl.E3ch.Create("/dir1/key1", "")
	if err != nil {
		fmt.Println(" [ERROR] failed to Create '/dir1/key1', error: ", err)
	}

	err = ectl.E3ch.Create("/dir", "")
	if err != nil {
		fmt.Println(" [ERROR] failed to Create '/dir', error: ", err)
	}

	err = ectl.E3ch.Put("/dir1/key1", "value1")
	if err != nil {
		fmt.Println(" [ERROR] failed to Put: key='/dir1/key1', val='value1' , error: ", err)
	}

	_, err = ectl.E3ch.Get("/dir1/key1")
	if err != nil {
		fmt.Println(" [ERROR] failed to Get: key='/dir1/key1', error: ", err)
	}

	_, err = ectl.E3ch.List("/dir1")
	if err != nil {
		fmt.Println(" [ERROR] failed to List keys: key='/dir1', error: ", err)
	}

	err = ectl.E3ch.Delete("/dir")
	if err != nil {
		fmt.Println(" [ERROR] failed to Delete: key='/dir' , error: ", err)
	}

	_, err = ectl.E3ch.List("/")
	if err != nil {
		fmt.Println(" [ERROR] failed to List: key='/' , error: ", err)
	}
	return err
}

// Route configuration struct.
type EtcdRoute struct {
	Regexp string `json:"regexp" yaml:"regexp" toml:"regexp"`
	Schema string `json:"schema" yaml:"schema" toml:"schema"`
}

func CloneE3chClient(username, password string, client *client.EtcdHRCHYClient) (*client.EtcdHRCHYClient, error) {
	ectl, err := etcd.New(etcd.Config{
		Endpoints: client.EtcdClient().Endpoints(),
		Username:  username,
		Password:  password,
	})
	if err != nil {
		return nil, err
	}
	return client.Clone(ectl), nil
}
