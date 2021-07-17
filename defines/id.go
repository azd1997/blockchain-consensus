/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/9/20 9:06 AM
* @Description: id在整个系统中应该是定长的
***********************************************************************/

package defines

import (
	"crypto"
)

const (
	IdLen = 20
)

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

// KeyPair 密钥对
// 之所以做成接口而不是结构体，是为了支持自定义，选择何种数字签名算法、是否添加额外数据
// KeyPair 用于每个节点程序自身持有。
type KeyPair interface {
	/** 成员约定 **/

	// PublicKey 返回密钥对中的公钥
	PublicKey() crypto.PublicKey
	// PrivateKey 返回密钥对中的私钥
	PrivateKey() crypto.PrivateKey
	// ID 返回生成的ID项
	ID() ID

	/** 方法约定 **/

	// Sign 私钥签名
	Sign(target []byte) (sig []byte, err error)
	// VerifySign 公钥验证签名。ID也具备该能力
	VerifySign(target []byte, sig []byte) (err error)

	// Encrypt 公钥加密。 ID也具备该能力
	Encrypt(target []byte) (encrypted []byte, err error)
	// Decrypt 私钥解密
	Decrypt(target []byte) (decrypted []byte, err error)

	// Encode 编码为[]byte，可用于存储
	Encode() (encoded []byte, err error)
	// Decode 解码为KeyPair，可用于从存储加载。调用之前须new一个KeyPair的实现结构体的实例。
	Decode() (err error)
}

// BaseKeyPair 基础的密钥对，其实是空密钥对
type BaseKeyPair struct {
	id string
}

func (kp BaseKeyPair) PublicKey() crypto.PublicKey {return nil}

func (kp BaseKeyPair) PrivateKey() crypto.PrivateKey {return nil}

func (kp BaseKeyPair) ID() ID {return BaseID(kp.id)}

func (kp BaseKeyPair) Sign(target []byte) (sig []byte, err error) {return nil, nil}

func (kp BaseKeyPair) VerifySign(target []byte, sig []byte) (err error) {return nil}

func (kp BaseKeyPair) Encrypt(target []byte) (encrypted []byte, err error) {return nil, nil}

func (kp BaseKeyPair) Decrypt(target []byte) (decrypted []byte, err error) {return nil, nil}

func (kp BaseKeyPair) Encode() (encoded []byte, err error) {return nil, nil}

func (kp BaseKeyPair) Decode() (err error) {return nil}





