package match

import (
	"bufio"
	"bytes"
	"io"
)

// 判断匹配开头或结尾的方法
type MatchFunc func(data []byte) (int, int)

// 按前后标记拆包
type SplitMatcher struct {
	split bufio.SplitFunc
}

func NewSplitMatcher(split bufio.SplitFunc) *SplitMatcher {
	return &SplitMatcher{split: split}
}

// 解析字节流
func (m *SplitMatcher) Scanning(rd io.Reader, write func(data []byte)) error {
	scanner := bufio.NewScanner(rd)
	scanner.Split(m.split)
	for scanner.Scan() {
		if chunk := scanner.Bytes(); chunk != nil {
			write(chunk)
		}
	}
	return scanner.Err()
}

// 解析字节流
func (m *SplitMatcher) SplitStream(rd io.Reader, outch chan<- []byte) error {
	return m.Scanning(rd, func(data []byte) {
		outch <- data
	})
}

// 解析二进制数据
func (m *SplitMatcher) SplitBuffer(input []byte) (output [][]byte, err error) {
	err = m.Scanning(bytes.NewReader(input), func(data []byte) {
		output = append(output, data)
	})
	return
}

// 按前后标记分拆
type SplitCreator struct {
	StartToken []byte
	EndToken   []byte
}

func NewSplitCreator(start, end []byte) *SplitCreator {
	return &SplitCreator{StartToken: start, EndToken: end}
}

func (m *SplitCreator) GetSplit() bufio.SplitFunc {
	// 只有结尾标记
	if m.StartToken == nil {
		return SplitAfter(CreateMatchForward(m.EndToken))
	}
	matchStart := CreateMatchForward(m.StartToken)
	if bytes.Compare(m.StartToken, m.EndToken) == 0 {
		return SplitBoth(matchStart) // 前后标记相同
	} else {
		matchEnd := CreateMatchBackward(m.EndToken)
		return SplitBetween(matchStart, matchEnd) // 前后标记不同
	}
}

func CreateMatchForward(token []byte) MatchFunc {
	return func(data []byte) (int, int) {
		return bytes.Index(data, token), len(token)
	}
}

func CreateMatchBackward(token []byte) MatchFunc {
	return func(data []byte) (int, int) {
		return bytes.LastIndex(data, token), len(token)
	}
}

func MatchTwice(match MatchFunc, data []byte, atEOF bool) (int, int, int) {
	var (
		advance = -1
		i, m    int
	)
	if atEOF && len(data) == 0 {
		return advance, 0, 0 // 终止
	}
	if i, m = match(data); i >= 0 {
		advance = i
		if j, _ := match(data[i+m:]); j >= 0 {
			advance += j + m
		}
	}
	return advance, i, m
}

// 根据结尾分割，但是从前向后搜索
func SplitAfter(match MatchFunc) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (int, []byte, error) {
		if i, m := match(data); i >= 0 {
			advance := i + m
			return advance, data[:advance], nil
		}
		return len(data), nil, nil
	}
}

// 根据相同的开头和结尾分割
func SplitBoth(match MatchFunc) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (int, []byte, error) {
		advance, i, m := MatchTwice(match, data, atEOF)
		if advance < 0 {
			return 0, nil, nil
		}
		var token []byte
		if advance > i {
			advance += m
			token = data[i:advance]
		}
		return advance, token, nil
	}
}

// 根据不同的开头和结尾分割
func SplitBetween(matchStart, matchEnd MatchFunc) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (int, []byte, error) {
		advance, i, m := MatchTwice(matchStart, data, atEOF)
		if advance < 0 {
			return 0, nil, nil
		}
		var token []byte
		if advance > i {
			token = data[i:advance]
		} else {
			advance = len(data)
			token = data[i:]
		}
		if j, n := matchEnd(token[m:]); j >= 0 {
			token = token[:m+j+n]
		} else {
			token = nil
		}
		return advance, token, nil
	}
}

// 按定长分拆
type FixedSplitCreator struct {
	ByteSize int
}

func NewFixedSplitCreator(size int) *FixedSplitCreator {
	return &FixedSplitCreator{ByteSize: size}
}

func (m *FixedSplitCreator) GetSplit() bufio.SplitFunc {
	return func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if len(data) >= m.ByteSize {
			return m.ByteSize, data[:m.ByteSize], nil
		}
		var err error
		if atEOF {
			// ErrFinalToken，最后一段数据长度不够也可以结束
			err = bufio.ErrFinalToken
		}
		return len(data), data, err
	}
}
