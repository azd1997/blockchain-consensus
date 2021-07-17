package account

// ID 账户ID，也是KeyPair中公钥等信息的编码版本。
// ID用于在集群中分享，其具备验证签名与非对称加密能力
type ID interface {
	// Len 返回ID的长度
	Len() int
	// String 返回ID字符串
	String() string
	// ValueOf ID允许注册和查取自定义的KV对
	ValueOf(key string) (value string)

	// VerifySign 验证该ID对应的账户的签名。例如，有A,B两个账户，B可以使用A的ID来对其签名进行验证。
	// 对于ID.VerifySign来说，验证签名实质是先转换回PublicKey，再使用PublicKey验证签名。
	VerifySign(target []byte, sig []byte) (err error)

	// Encrypt 使用该ID对某段内容进行非对称加密。例如，有A,B两个账户，B可以使用A的ID来加密一段数据，只有A能够利用其KeyPair进行解密。
	// 对于ID.Encrypt来说，实质是先转换回PublicKey，再使用PublicKey进行加密。
	Encrypt(target []byte) (encrypted []byte, err error)
}

// BaseID 最基础的ID实现，除了String以外都是空实现
type BaseID string

func (id BaseID) Len() int {return len(string(id))}

func (id BaseID) String() string {return string(id)}

func (id BaseID) ValueOf(key string) (value string) {return ""}

func (id BaseID) VerifySign(target []byte, sig []byte) (err error) {return nil}

func (id BaseID) Encrypt(target []byte) (encrypted []byte, err error) {return nil, nil}
