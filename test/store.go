/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/14/20 10:45 AM
* @Description: 测试用的Store实现
***********************************************************************/

package test

import (
	"bytes"
	"log"
	"strings"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/requires"
)

// 测试用的requires.Store实现
type Store struct {
	Cfs map[requires.CF]bool
	Kvs map[string]string
}

func (s *Store) Open() error {
	log.Printf("testStore open ...\n")
	return nil
}

func (s *Store) Close() error {
	log.Printf("testStore close ...\n")
	return nil
}

func (s *Store) Get(cf requires.CF, key []byte) ([]byte, error) {
	k := string(cf[:]) + string(key)
	v := new(defines.PeerInfo)
	v.Decode(bytes.NewReader([]byte(s.Kvs[k])))
	log.Printf("testStore Get: {key: %s, value: %s}\n", string(key), v.String())
	//time.Sleep(50 * time.Millisecond)
	return []byte(s.Kvs[k]), nil
}

func (s *Store) Set(cf requires.CF, key, value []byte) error {
	k := string(cf[:]) + string(key)
	s.Kvs[k] = string(value)
	v := new(defines.PeerInfo)
	v.Decode(bytes.NewReader(value))
	log.Printf("testStore Set: {key: %s, value: %s}\n", string(key), v.String())
	//time.Sleep(300 * time.Millisecond)
	return nil
}

func (s *Store) Del(cf requires.CF, key []byte) error {
	delete(s.Kvs, string(cf[:]) + string(key))
	log.Printf("testStore Del: {key: %s}\n", string(key))
	//time.Sleep(300 * time.Millisecond)
	return nil
}

func (s *Store) RegisterCF(cf requires.CF) error {
	s.Cfs[cf] = true
	return nil
}

func (s *Store) RangeCF(cf requires.CF, f func(key, value []byte) error) error {
	var firstErr, err error
	for k, v := range s.Kvs {
		if strings.HasPrefix(k, string(cf[:])) {
			err = f([]byte(k[requires.CFLen:]), []byte(v))
			if err != nil {
				if firstErr == nil {
					firstErr = err
				} else {
					continue
				}
			}
		}
	}

	return firstErr
}


