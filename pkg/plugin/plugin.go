package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/pkg/conf"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	// TypeServiceDiscovery serviceDiscovery plugin
	TypeServiceDiscovery = "service-discovery"

	// URLPrefix url prefix
	URLPrefix = "/plugins/"

	// POST post method
	POST = "POST"
	// GET get method
	GET = "GET"
	// PUT put method
	PUT = "PUT"
	// DELETE delete method
	DELETE = "DELETE"
)

var (
	// ErrPluginNotFound Plugin not found
	ErrPluginNotFound = errors.New("Plugin not found")
)

// Plugin plugin define
type Plugin struct {
	Type    string `json:"type"`
	Address string `json:"address"`
}

// Marshal marshal
func (p *Plugin) Marshal() []byte {
	v, _ := json.Marshal(p)
	return v
}

// RegistryCenter plugin registry center
// The plugin registry center will scan plugin dir to find usable plugins. Than plugin registry center do regist spec plugin use a http get access
type RegistryCenter struct {
	cnf        *conf.Conf
	httpClient *util.FastHTTPClient
	plugins    map[string]*Plugin
}

// NewRegistryCenter get a RegistryCenter
func NewRegistryCenter(cnf *conf.Conf, httpClient *util.FastHTTPClient) *RegistryCenter {
	return &RegistryCenter{
		cnf:        cnf,
		httpClient: httpClient,
		plugins:    make(map[string]*Plugin),
	}
}

// Load load plugin from plugin_dir
func (c *RegistryCenter) Load() error {
	if c.cnf.PluginDir == "" {
		log.Info("Load from empty plugin dir.")
		return nil
	}

	err := filepath.Walk(c.cnf.PluginDir, c.process)

	if err != nil {
		log.ErrorErrorf(err, "Load plugins fialure at <%s>.", c.cnf.PluginDir)
		return err
	}

	return nil
}

func (c *RegistryCenter) process(path string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}

	if f.IsDir() {
		return nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	plugin := &Plugin{}
	err = json.Unmarshal(data, plugin)
	if err != nil {
		log.WarnErrorf(err, "Load plugin failure at <%s>", filepath.Join(path, f.Name()))
		return err
	}

	c.registry(plugin)

	return nil
}

func (c *RegistryCenter) registry(plugin *Plugin) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(fmt.Sprintf("http://%s%s%s", plugin.Address, URLPrefix, plugin.Type))
	req.Header.SetMethod(POST)
	req.SetBody(plugin.Marshal())

	resp, err := c.httpClient.Do(req, plugin.Address)
	defer func() {
		if resp != nil {
			fasthttp.ReleaseResponse(resp)
		}
	}()

	if err != nil {
		log.WarnErrorf(err, "Plugin<%s, %s> added failure", plugin.Type, plugin.Address)
		return
	}

	c.plugins[plugin.Type] = plugin
	log.Infof("Plugin<%s, %s> added", plugin.Type, plugin.Address)
}

// DoPost do post
func (c *RegistryCenter) DoPost(pluginType string, action string, data []byte) ([]byte, error) {
	return c.do(pluginType, action, POST, data)
}

// DoPut do put
func (c *RegistryCenter) DoPut(pluginType string, action string, data []byte) ([]byte, error) {
	return c.do(pluginType, action, PUT, data)
}

// DoDelete do delete
func (c *RegistryCenter) DoDelete(pluginType string, action string, data []byte) ([]byte, error) {
	return c.do(pluginType, action, DELETE, data)
}

// DoGet do get
func (c *RegistryCenter) DoGet(pluginType string, action string) ([]byte, error) {
	return c.do(pluginType, action, GET, nil)
}

func (c *RegistryCenter) do(pluginType string, action string, method string, data []byte) ([]byte, error) {
	p, ok := c.plugins[pluginType]

	if !ok {
		return nil, ErrPluginNotFound
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(fmt.Sprintf("http://%s%s%s%s", p.Address, URLPrefix, p.Type, action))
	req.Header.SetMethod(method)
	req.SetBody(data)

	resp, err := c.httpClient.Do(req, p.Address)
	defer func() {
		if resp != nil {
			fasthttp.ReleaseResponse(resp)
		}
	}()

	if err != nil {
		return nil, err
	}

	return resp.Body(), nil
}
