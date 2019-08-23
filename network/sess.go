package network

import (
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid"
)

// 创建全局唯一的字符串ID
func RandomGUID() string {
	now := time.Now()
	source := rand.NewSource(now.UnixNano())
	entropy := ulid.Monotonic(rand.New(source), 0)
	randId := ulid.MustNew(ulid.Timestamp(now), entropy)
	return randId.String()
}

type SessionData interface {
	Get(key string) (interface{}, bool)
	Put(key string, value interface{})
	Init() error
	Reset() error
}

// 会话，拥有唯一的sid
type Session struct {
	sid  string
	data SessionData
}

func NewSession(safe bool) *Session {
	sess := &Session{sid: RandomGUID()}
	if safe {
		sess.data = new(SafeMapData)
	} else {
		sess.data = new(MapData)
	}
	_ = sess.data.Init()
	return sess
}

func (s *Session) Clear() {
	s.sid = ""
	_ = s.data.Reset()
}

func (s *Session) GetId() string {
	return s.sid
}

// 写入数据
func (s *Session) PutData(key string, value interface{}) {
	s.data.Put(key, value)
}

// 读取数据，可能不存在
func (s *Session) GetData(key string) (interface{}, bool) {
	return s.data.Get(key)
}

// 读取字符串数据
func (s *Session) GetString(key string) string {
	// 有可能value为nil发生，需单独处理
	if value, ok := s.data.Get(key); ok && value != nil {
		return value.(string)
	}
	return ""
}

// 会话数据，只使用string作为key
type MapData struct {
	data map[string]interface{}
}

func (d *MapData) Get(key string) (interface{}, bool) {
	if value, ok := d.data[key]; ok {
		return value, true
	}
	return nil, false
}

func (d *MapData) Put(key string, value interface{}) {
	d.data[key] = value
}

func (d *MapData) Init() error {
	d.data = make(map[string]interface{})
	return nil
}

func (d *MapData) Reset() error {
	return d.Init()
}

// 跨进程安全会话数据
type SafeMapData struct {
	sync.Map
}

func (d *SafeMapData) Get(key string) (interface{}, bool) {
	if value, ok := d.Load(key); ok {
		return value, true
	}
	return nil, false
}

func (d *SafeMapData) Put(key string, value interface{}) {
	d.Store(key, value)
}

func (d *SafeMapData) Init() error {
	return nil
}

func (d *SafeMapData) Reset() error {
	d.Range(func(key, value interface{}) bool {
		d.Delete(key)
		return true // 为true时继续执行下一个，否则中断
	})
	return nil
}
