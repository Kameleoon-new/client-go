package kameleoon

import (
	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/puzpuzpuz/xsync/v3"
)

var KameleoonClientFactory = newKameleoonClientFactory()

type kameleoonClientFactory struct {
	clients *xsync.MapOf[string, *kameleoonClient]
}

func newKameleoonClientFactory() *kameleoonClientFactory {
	return &kameleoonClientFactory{
		clients: xsync.NewMapOf[string, *kameleoonClient](),
	}
}

func (cf *kameleoonClientFactory) Create(siteCode string, cfg *KameleoonClientConfig) (KameleoonClient, error) {
	logging.Info("CALL: KameleoonClientFactory.Create(siteCode: %s, config: %s)", siteCode, cfg)
	client, err := cf.createWithConfigSource(siteCode, func() (*KameleoonClientConfig, error) {
		return cfg, nil
	})
	logging.Info("RETURN: KameleoonClientFactory.Create(siteCode: %s, config: %s) -> (client, error: %s)",
		siteCode, cfg, err)
	return client, err
}

func (cf *kameleoonClientFactory) CreateFromFile(siteCode string, cfgPath string) (KameleoonClient, error) {
	logging.Info("CALL: KameleoonClientFactory.CreateFromFile(siteCode: %s, configPath: %s)", siteCode, cfgPath)
	client, err := cf.createWithConfigSource(siteCode, func() (*KameleoonClientConfig, error) {
		return LoadConfig(cfgPath)
	})
	logging.Info(
		"RETURN: KameleoonClientFactory.CreateFromFile(siteCode: %s, configPath: %s) -> (client, error: %s)",
		siteCode, cfgPath, err)
	return client, err
}

func (cf *kameleoonClientFactory) createWithConfigSource(siteCode string,
	cfgSrc func() (*KameleoonClientConfig, error)) (KameleoonClient, error) {
	var err error
	client, _ := cf.clients.Compute(siteCode,
		func(former *kameleoonClient, loaded bool) (*kameleoonClient, bool) {
			if loaded {
				return former, false
			}
			var client *kameleoonClient
			cfg, cerr := cfgSrc()
			if cerr == nil {
				client, cerr = newClient(siteCode, cfg)
			}
			err = cerr
			return client, err != nil // a failed creation leaves no entry behind
		})
	return client, err
}

func (cf *kameleoonClientFactory) Forget(siteCode string) {
	logging.Info("CALL: KameleoonClientFactory.Forget(siteCode: %s)", siteCode)
	cf.clients.Compute(siteCode, func(client *kameleoonClient, loaded bool) (*kameleoonClient, bool) {
		if loaded {
			client.close()
		}
		return client, true
	})
	logging.Info("RETURN: KameleoonClientFactory.Forget(siteCode: %s)", siteCode)
}
