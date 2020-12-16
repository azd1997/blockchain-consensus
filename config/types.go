/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/17/20 10:55 AM
* @Description: The file is for
***********************************************************************/

package config

type TomlConfig struct {
	Account   AccountConfig   `toml:"account"`
	Consensus ConsensusConfig `toml:"consensus"`
	Pot       PotConfig       `toml:"pot"`  // 可选
	Pow       PowConfig       `toml:"pow"`  // 可选
	Raft      RaftConfig      `toml:"raft"` // 可选
	Store     StoreConfig     `toml:"store"`
	Bnet      BnetConfig      `toml:"bnet"`

	Seeds map[string]string `toml:"seeds"`
	Peers map[string]string `toml:"peers"`
}

type AccountConfig struct {
	Id   string `toml:"id"`
	Duty string `toml:"duty"`
	Addr string `toml:"addr"`
}

type ConsensusConfig struct {
	Type string `toml:"type"`
}

//////////////////

type PotConfig struct {
	TickMs int `toml:"tick_ms"`
}

type PowConfig struct {
}

type RaftConfig struct {
}

///////////////////

type StoreConfig struct {
	Engine   string `toml:"engine"`
	Database string `toml:"database"`
}

type BnetConfig struct {
	Protocol string `toml:"protocol"`
	Addr     string `toml:"addr"`
}
