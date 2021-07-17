package account

var kpgRegisterMap = map[string]KeyPairGenerator{

}

// KeyPairGenerator KeyPair生成器
// 每一种定义的生成器唯一实例都需要进行注册，注册是注册到kpgRegisterMap
type KeyPairGenerator interface {
	// Type 返回密钥对生成器类型
	Type() string
	// Register
	Register()
	Generate() KeyPair
}
