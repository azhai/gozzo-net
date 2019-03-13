package network

import (
	"math/rand"
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

// 会话，拥有唯一的sid
type Session struct {
	sid  string
	data map[string]interface{}
}

func NewSession() *Session {
	return &Session{
		sid:  RandomGUID(),
		data: make(map[string]interface{}),
	}
}

func (s *Session) GetId() string {
	return s.sid
}

// 写入数据
func (s *Session) PutData(key string, value interface{}) {
	s.data[key] = value
}

// 读取数据，可能不存在
func (s *Session) GetData(key string) (interface{}, bool) {
	if value, ok := s.data[key]; ok {
		return value, true
	}
	return nil, false
}

// 读取字符串数据
func (s *Session) GetString(key string) string {
	if value, ok := s.data[key]; ok {
		return value.(string)
	}
	return ""
}
