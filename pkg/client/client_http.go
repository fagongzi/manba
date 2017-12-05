package client

import (
	"encoding/json"
	"fmt"

	"github.com/fagongzi/gateway/pkg/api"
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/valyala/fasthttp"
)

type httpAPIClient struct {
	server     string
	apiVersion APIVersion
	rawClient  *util.FastHTTPClient
}

// NewHTTPClient return a Client with default option
func NewHTTPClient(server string, version APIVersion) (Client, error) {
	return NewHTTPClientWithOption(server, version, util.DefaultHTTPOption())
}

// NewHTTPClientWithOption returns a Client with http option for manager gateway meta data
func NewHTTPClientWithOption(server string, version APIVersion, option *util.HTTPOption) (Client, error) {
	return &httpAPIClient{
		server:     server,
		apiVersion: version,
		rawClient:  util.NewFastHTTPClientOption(option),
	}, nil
}

// AddCluster add clueter
func (c *httpAPIClient) AddCluster(cluster *model.Cluster) (string, error) {
	err := cluster.Validate()
	if err != nil {
		return "", err
	}

	v, err := c.do(apiForResources(c.apiVersion, apiServer), httpPOST, cluster)
	if err != nil {
		return "", err
	}

	return v.(string), nil
}

func (c *httpAPIClient) DeleteCluster(id string) error {
	_, err := c.do(apiForResource(c.apiVersion, apiCluster, id), httpDELETE, nil)
	return err
}

func (c *httpAPIClient) GetCluster(id string) (*model.Cluster, error) {
	result := &model.Cluster{}
	data, err := c.do(apiForResource(c.apiVersion, apiCluster, id), httpGET, nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(util.MustMarshal(data), result)
	if err != nil {
		return nil, err
	}

	if result.ID == "" {
		return nil, nil
	}

	return result, nil
}

func (c *httpAPIClient) GetClusters() ([]*model.Cluster, error) {
	var result []*model.Cluster
	data, err := c.do(apiForResources(c.apiVersion, apiCluster), httpGET, nil)
	if err != nil {
		return nil, err
	}

	return result, json.Unmarshal(util.MustMarshal(data), &result)
}

func (c *httpAPIClient) AddServer(server *model.Server) (string, error) {
	err := server.Validate()
	if err != nil {
		return "", err
	}

	v, err := c.do(apiForResources(c.apiVersion, apiServer), httpPOST, server)
	if err != nil {
		return "", err
	}

	return v.(string), nil
}

func (c *httpAPIClient) do(path string, method string, data interface{}) (interface{}, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	url := c.getURL(path)
	req.SetRequestURI(url)
	req.Header.SetMethod(method)
	if nil != data {
		body, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		req.SetBody(body)
	}

	resp, err := c.rawClient.Do(req, c.server, nil)
	if err != nil {
		return nil, err
	}

	defer fasthttp.ReleaseResponse(resp)

	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("resp code: %d", resp.StatusCode())
	}

	result := &api.Result{}
	err = util.Unmarshal(result, resp.Body())
	if err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("code: %d, error:%s", result.Code, result.Error)
	}

	return result.Value, nil
}

func (c *httpAPIClient) getURL(path string) string {
	return fmt.Sprintf("http://%s/%s", c.server, path)
}
