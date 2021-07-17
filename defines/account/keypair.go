package account

import "crypto"

// KeyPair 密钥对
// 之所以做成接口而不是结构体，是为了支持自定义，选择何种数字签名算法、是否添加额外数据
// KeyPair 用于每个节点程序自身持有。
type KeyPair interface {
	/** 成员约定 **/
	// 尽管提供了获取公私钥的方法，但是在本库中，其实是希望避免公私钥的直接使用，只使用KeyPair和ID。

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

type EcdsaKeyPair struct {

}