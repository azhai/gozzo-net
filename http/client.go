package http

import (
	orig "net/http"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
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

type DecodeFunc func(body []byte) error

func NewJsonDecoder(res interface{}) DecodeFunc {
	return func(body []byte) error {
		return json.Unmarshal(body, &res)
	}
}

type Resource struct {
	HttpMethod string
	ContentType string
	MimeType string
	Data interface{}
}

func NewUpload(filename string) *Resource {
	return &Resource{
		HttpMethod: "Post",
		ContentType: "multipart/form-data",
		MimeType: "file",
		Data: filename,
	}
}

func NewJsonReq(method, mime string, obj interface{}) *Resource {
	if mime == "" {
		mime = "json"
	}
	return &Resource{
		HttpMethod: method,
		ContentType: "application/json",
		MimeType: mime,
		Data: obj,
	}
}

func (r *Resource) GetMethod() (string, string) {
	if r.HttpMethod == "" {
		r.HttpMethod = "GET"
	}
	return r.HttpMethod, r.ContentType
}

func (r *Resource) GetReader() io.Reader {
	if r.Data == nil {
		return nil
	}
	switch r.MimeType {
	case "bytes", "Bytes":
		return bytes.NewReader(r.Data.([]byte))
	case "string", "String":
		return strings.NewReader(r.Data.(string))
	case "file", "File", "filename", "FileName":
		if fp, err := os.Open(r.Data.(string)); err == nil {
			return fp
		}
	case "json", "JSON", "struct", "Struct":
		if json, err := json.Marshal(r.Data); err == nil {
			return bytes.NewReader(json)
		}
	}
	return nil
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

// 封装IO
func (c *Client) DoWrap(url string, req *Resource, dec DecodeFunc) (int, error) {
	var err error
	method, reqtype := req.GetMethod()
	if reqtype != "" {
		c.SetContentType(reqtype)
	}
	code, body := c.Do(method, url, req.GetReader())
	if code == 200 && dec != nil {
		err = dec(body)
	}
	return code, err
}

// GET操作
func (c *Client) Get(url, query string) (code int, body string) {
	if query != "" {
		url += "?" + query
	}
	if len(c.Headers) > 0 {
		code, body := c.Do("GET", url, nil)
		return code, string(body)
	} else {
		resp, err := c.Client.Get(c.Prefix + url)
		code, body := ReadResponse(resp, err)
		return code, string(body)
	}
}

// HEAD操作
func (c *Client) Head(url, query string) (int, string) {
	if query != "" {
		url += "?" + query
	}
	code, body := c.Do("HEAD", url, nil)
	return code, string(body)
}

// DELETE操作
func (c *Client) Delete(url, query string) (int, string) {
	if query != "" {
		url += "?" + query
	}
	code, body := c.Do("DELETE", url, nil)
	return code, string(body)
}

// PUT操作
func (c *Client) Put(url string, data []byte) (int, []byte) {
	return c.Do("PUT", url, bytes.NewReader(data))
}

// POST操作
func (c *Client) Post(url string, data []byte) (int, []byte) {
	return c.Do("POST", url, bytes.NewReader(data))
}

// 提交表单
func (c *Client) PostForm(url, data string) (int, string) {
	c.SetContentType("application/x-www-form-urlencoded")
	code, body := c.Post(url, []byte(data))
	return code, string(body)
}

// 提交JSON数据
func (c *Client) PostJson(url string, obj, res interface{}) (int, error) {
	req := NewJsonReq("Post", "json", obj)
	dec := NewJsonDecoder(res)
	return c.DoWrap(url, req, dec)
}
