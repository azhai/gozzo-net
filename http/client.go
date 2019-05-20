package http

import (
	orig "net/http"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
)

// 读取Response的状态码和文本内容
func ReadResponse(resp *orig.Response, err error) (int, string) {
	defer resp.Body.Close()
	code := resp.StatusCode
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return code, ""
	}
	return code, string(body)
}

// http或https客户端
type Client struct {
	Prefix  string
	Headers map[string]string
	*orig.Client
}

func NewClient(prefix string) *Client {
	return &Client{
		Prefix:  prefix,
		Headers: make(map[string]string),
		Client: orig.DefaultClient,
	}
}

// 为https设置客户端证书
func (c *Client) SetPemCert(pemfile string) error {
	cert, err := ioutil.ReadFile(pemfile)
	if err != nil {
		return err
	}
	conf := &tls.Config{RootCAs: x509.NewCertPool()}
	conf.RootCAs.AppendCertsFromPEM(cert)
	c.Transport = &orig.Transport{
		TLSClientConfig: conf,
	}
	return nil
}

func (c *Client) AddHeader(key, value string) {
	c.Headers[key] = value
}

func (c *Client) SetContentType(ctype string) {
	c.AddHeader("content-type", ctype)
}

// 通用操作
func (c *Client) Do(method, url string, reader io.Reader) (int, string) {
	req, _ := orig.NewRequest(method, c.Prefix+url, reader)
	for key, value := range c.Headers {
		if value != "" {
			req.Header.Set(key, value)
		}
	}
	return ReadResponse(c.Client.Do(req))
}

// GET操作
func (c *Client) Get(url, query string) (int, string) {
	if query != "" {
		url += "?" + query
	}
	if len(c.Headers) == 0 {
		resp, err := c.Client.Get(c.Prefix + url)
		return ReadResponse(resp, err)
	}
	return c.Do("GET", url, nil)
}

// HEAD操作
func (c *Client) Head(url, query string) (int, string) {
	if query != "" {
		url += "?" + query
	}
	return c.Do("HEAD", url, nil)
}

// DELETE操作
func (c *Client) Delete(url, query string) (int, string) {
	if query != "" {
		url += "?" + query
	}
	return c.Do("DELETE", url, nil)
}

// PUT操作
func (c *Client) Put(url string, reader io.Reader) (int, string) {
	return c.Do("PUT", url, reader)
}

// POST操作
func (c *Client) Post(url string, reader io.Reader) (int, string) {
	return c.Do("POST", url, reader)
}

// 提交表单
func (c *Client) PostForm(url, data string) (int, string) {
	c.SetContentType("application/x-www-form-urlencoded")
	return c.Post(url, strings.NewReader(data))
}

// 提交JSON数据
func (c *Client) PostJson(url string, obj interface{}) (int, string) {
	c.SetContentType("application/json")
	data, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return c.Post(url, bytes.NewReader(data))
}
