/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/19/21 6:28 PM
* @Description: The file is for
***********************************************************************/

package simplepit

import (
	"errors"

	"github.com/azd1997/blockchain-consensus/defines"
)

type SimplePit struct {
	id    string
	seeds map[string]*defines.PeerInfo
	peers map[string]*defines.PeerInfo
}

func New(id string) (*SimplePit, error) {
	return &SimplePit{
		id:    id,
		seeds: map[string]*defines.PeerInfo{},
		peers: map[string]*defines.PeerInfo{},
	}, nil
}

func (s *SimplePit) Init() error {
	return nil
}

func (s *SimplePit) Inited() bool {
	return true
}

func (s *SimplePit) Close() error {
	return nil
}

func (s *SimplePit) Get(id string) (*defines.PeerInfo, error) {
	if pi, ok := s.seeds[id]; ok {
		return pi, nil
	}
	if pi, ok := s.peers[id]; ok {
		return pi, nil
	}
	return nil, errors.New("not found")
}

func (s *SimplePit) Set(info *defines.PeerInfo) error {
	if info == nil {
		return nil
	}
	if info.Duty == defines.PeerDuty_Seed {
		s.seeds[info.Id] = info
	} else if info.Duty == defines.PeerDuty_Peer {
		s.peers[info.Id] = info
	}
	return nil
}

func (s *SimplePit) Del(id string) error {
	delete(s.peers, id)
	delete(s.seeds, id)
	return nil
}

func (s *SimplePit) AddPeers(peers map[string]string) error {
	for id, addr := range peers {
		if info, ok := s.peers[id]; ok {
			info.Addr = addr
		} else {
			s.peers[id] = &defines.PeerInfo{
				Id:   id,
				Addr: addr,
				Duty: defines.PeerDuty_Peer,
			}
		}
	}
	return nil
}

func (s *SimplePit) AddSeeds(seeds map[string]string) error {
	for id, addr := range seeds {
		if info, ok := s.seeds[id]; ok {
			info.Addr = addr
		} else {
			s.seeds[id] = &defines.PeerInfo{
				Id:   id,
				Addr: addr,
				Duty: defines.PeerDuty_Seed,
			}
		}
	}
	return nil
}

func (s *SimplePit) Peers() map[string]*defines.PeerInfo {
	peerscopy := make(map[string]*defines.PeerInfo)
	for id := range s.peers {
		info := *(s.peers[id])
		peerscopy[id] = &info
	}
	return peerscopy
}

func (s *SimplePit) Seeds() map[string]*defines.PeerInfo {
	seedscopy := make(map[string]*defines.PeerInfo)
	for id := range s.seeds {
		info := *(s.seeds[id])
		seedscopy[id] = &info
	}
	return seedscopy
}

func (s *SimplePit) RangePeers(f func(peer *defines.PeerInfo) error) (total int, errs map[string]error) {
	errs = map[string]error{}
	var err error

	peers := s.Peers()
	for id, info := range peers {
		if id != s.id {
			total++
			err = f(info)
			if err != nil {
				errs[id] = err
			}
		}
	}
	return
}

func (s *SimplePit) RangeSeeds(f func(peer *defines.PeerInfo) error) (total int, errs map[string]error) {
	errs = map[string]error{}
	var err error

	seeds := s.Seeds()
	for id, info := range seeds {
		if id != s.id {
			total++
			err = f(info)
			if err != nil {
				errs[id] = err
			}
		}
	}
	return
}

func (s *SimplePit) IsSeed(id string) bool {
	if _, ok := s.seeds[id]; ok {
		return true
	}
	return false
}

func (s *SimplePit) IsPeer(id string) bool {
	if _, ok := s.peers[id]; ok {
		return true
	}
	return false
}

func (s *SimplePit) NSeed() int {
	return len(s.seeds)
}

func (s *SimplePit) NPeer() int {
	return len(s.peers)
}
