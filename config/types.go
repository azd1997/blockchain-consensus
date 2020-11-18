/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/17/20 10:55 AM
* @Description: The file is for
***********************************************************************/

package config

type tomlConfig struct {
	Account accountConfig `toml:"account"`
	Consensus consensusConfig `toml:"consensus"`
	Pot potConfig `toml:"pot"`	// 可选
	Pow powConfig `toml:"pow"`	// 可选
	Raft raftConfig `toml:"raft"`	// 可选
	Store storeConfig `toml:"store"`
	Bnet bnetConfig `toml:"bnet"`

	Seeds map[string]string `toml:"seeds"`
	Peers map[string]string `toml:"peers"`
}

type accountConfig struct {
	Id string `toml:"id"`
	Duty string `toml:"duty"`
}

type consensusConfig struct {
	Type string `toml:"type"`
}

//////////////////

type potConfig struct {
	TickMs int `toml:"tick_ms"`
}

type powConfig struct {

}

type raftConfig struct {

}

///////////////////

type storeConfig struct {
	Engine string `toml:"engine"`
	Database string `toml:"database"`
}

type bnetConfig struct {
	Protocol string `toml:"protocol"`
	Addr string `toml:"addr"`
}