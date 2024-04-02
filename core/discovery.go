package core

import (
	"errors"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect SelectMode = iota
	RoundRobinSelect
)

type DisCovery interface {
	Refresh()
	Update([]string) error
	Get(SelectMode) (string, error)
	GetAll() ([]string, error)
}

type MulitServerDisCovery struct {
	addrs []string
	r     *rand.Rand
	mu    sync.Mutex
	index int
}

func NewMulitServerDisCovery(addrs []string) *MulitServerDisCovery {
	d := &MulitServerDisCovery{
		addrs: addrs,
		r:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}

func (m *MulitServerDisCovery) Refresh() {
	return
}

func (m *MulitServerDisCovery) Update(addrs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addrs = addrs
	return nil
}

func (m *MulitServerDisCovery) Get(mode SelectMode) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	addrNum := len(m.addrs)
	if addrNum == 0 {
		log.Fatalln("没有服务注册")
		return "", errors.New("没有服务注册")
	}
	index := 0
	switch mode {
	case RandomSelect:
		index = m.r.Intn(addrNum)
	case RoundRobinSelect:
		m.index++
		m.index = m.index % addrNum
		index = m.index
	default:
		return "", errors.New("选择模式异常")
	}
	return m.addrs[index], nil
}

func (m *MulitServerDisCovery) GetAll() ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	addrNum := len(m.addrs)
	if addrNum == 0 {
		return nil, errors.New("没有服务注册")
	}
	addr := make([]string, len(m.addrs), len(m.addrs))
	_ = copy(addr, m.addrs)
	return addr, nil
}
