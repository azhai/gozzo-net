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
func ReadResponse(resp *orig.Response, err error) (code int, body []byte) {
	if resp != nil {
		defer resp.Body.Close()
		code = resp.StatusCode
	}
	if err != nil {
		body = []byte(err.Error())
		return
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return
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
		Client:  orig.DefaultClient,
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
func (c *Client) Do(method, url string, reader io.Reader) (int, []byte) {
	req, _ := orig.NewRequest(method, c.Prefix+url, reader)
	for key, value := range c.Headers {
		if value != "" {
			req.Header.Set(key, value)
		}
	}
	return ReadResponse(c.Client.Do(req))
}

func (c *Client) DoStr(method, url string, reader io.Reader) (int, string) {
	code, body := c.Do(method, url, reader)
	return code, string(body)
}

// 提交JSON数据
func (c *Client) Json(method, url string, obj, res interface{}) (int, error) {
	var (
		data []byte
		err error
	)
	if obj != nil {
		data, err = json.Marshal(obj)
	}
	c.SetContentType("application/json")
	code, body := c.Do(method, url, bytes.NewReader(data))
	if code == 200 && body != nil {
		err = json.Unmarshal(body, &res)
	}
	return code, err
}

// GET操作
func (c *Client) Get(url, query string) (int, string) {
	if query != "" {
		url += "?" + query
	}
	if len(c.Headers) > 0 {
		return c.DoStr("GET", url, nil)
	}
	resp, err := c.Client.Get(c.Prefix + url)
	code, body := ReadResponse(resp, err)
	return code, string(body)
}

// HEAD操作
func (c *Client) Head(url, query string) (int, string) {
	if query != "" {
		url += "?" + query
	}
	return c.DoStr("HEAD", url, nil)
}

// DELETE操作
func (c *Client) Delete(url, query string) (int, string) {
	if query != "" {
		url += "?" + query
	}
	return c.DoStr("DELETE", url, nil)
}

// PUT操作
func (c *Client) Put(url string, reader io.Reader) (int, string) {
	return c.DoStr("PUT", url, reader)
}

// POST操作
func (c *Client) Post(url string, reader io.Reader) (int, string) {
	return c.DoStr("POST", url, reader)
}

// 提交表单
func (c *Client) PostForm(url, data string) (int, string) {
	c.SetContentType("application/x-www-form-urlencoded")
	return c.Post(url, strings.NewReader(data))
}
