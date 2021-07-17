/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/19/21 2:13 PM
* @Description: The file is for
***********************************************************************/

package pitable

import (
	"errors"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/pitable/simplepit"
)

type PitType uint8

const (
	PitType_SimplePit PitType = iota
)

// Pit 内部的PeerInfoTable接口
type Pit interface {
	Init() error
	Inited() bool
	Close() error

	Get(id string) (*defines.PeerInfo, error)
	Set(info *defines.PeerInfo) error
	Del(id string) error

	AddPeers(peers map[string]string) error
	AddSeeds(seeds map[string]string) error

	Peers() map[string]*defines.PeerInfo
	Seeds() map[string]*defines.PeerInfo

	RangePeers(f func(peer *defines.PeerInfo) error) (total int, errs map[string]error)
	RangeSeeds(f func(peer *defines.PeerInfo) error) (total int, errs map[string]error)

	IsSeed(id string) bool
	IsPeer(id string) bool

	NSeed() int
	NPeer() int
}

// New pitable.New(PitType_SimplePit, your_id)
func New(pittype PitType, id string) (Pit, error) {
	switch pittype {
	case PitType_SimplePit:
		return simplepit.New(id)
	default:
		return nil, errors.New("unsupport pit type")
	}
}
