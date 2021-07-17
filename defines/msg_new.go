package defines

////////////////////////////////////////////////////////////////////


func NewMessage_NewBlock(from, to string, epoch int64, nb *Block) (*Message, error) {
	// 构建区块消息
	blockBytes, err := nb.Encode()
	if err != nil {
		return nil, err
	}

	// 构建msg
	msg := &Message{
		Version:   CodeVersion,
		Type:      MessageType_NewBlock,
		From:      from,
		To:        to,
		Epoch: epoch,
		Base:      nb.PrevHash,
		BaseIndex: nb.Index - 1,
		Data:      [][]byte{blockBytes},
	}
	if err := msg.WriteDesc("type", MessageType_NewBlock.String()); err != nil {
		panic(err)
	}

	return msg, nil
}

func NewMessageAndSign_NewBlock(from, to string, epoch int64, nb *Block) (*Message, error) {
	msg, err := NewMessage_NewBlock(from, to, epoch, nb)
	if err != nil {
		return nil, err
	}
	err = msg.Sign()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func NewMessage_Txs(from, to string, epoch int64, txs []*Transaction) (*Message, error) {

	// 编码
	txBytes := make([][]byte, len(txs))
	for i := 0; i < len(txs); i++ {
		if txs[i] != nil {
			enced, err := txs[i].Encode()
			if err != nil {
				return nil, err
			}
			txBytes[i] = enced
		}
	}

	// 构建msg
	msg := &Message{
		Version: CodeVersion,
		Type:    MessageType_Txs,
		From:    from,
		To:      to,
		Epoch: epoch,
		Data:    txBytes,
	}
	if err := msg.WriteDesc("type", MessageType_Txs.String()); err != nil {
		return nil, err
	}

	return msg, nil
}

func NewMessageAndSign_Txs(from, to string, epoch int64, txs []*Transaction) (*Message, error) {

	msg, err := NewMessage_Txs(from, to, epoch, txs)
	if err != nil {
		return nil, err
	}
	err = msg.Sign()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// TODO: 把其他消息的构建方法也完成

func NewMessage_GenTxFaster(from, to string) (*Message, error) {
	msg := &Message{
		Version: CodeVersion,
		Type:    MessageType_GenTxFaster,
		From:    from,
		To:      to,
	}
	if err := msg.WriteDesc("type", MessageType_GenTxFaster.String()); err != nil {
		return nil, err
	}

	return msg, nil
}

func NewMessageAndSign_GenTxFaster(from, to string) (*Message, error) {

	msg, err := NewMessage_GenTxFaster(from, to)
	if err != nil {
		return nil, err
	}
	err = msg.Sign()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func NewMessage_MonitorPong(from, to string) (*Message, error) {
	msg := &Message{
		Version: CodeVersion,
		Type:    MessageType_MonitorPong,
		From:    from,
		To:      to,
	}
	if err := msg.WriteDesc("type", MessageType_MonitorPong.String()); err != nil {
		return nil, err
	}

	return msg, nil
}

func NewMessageAndSign_MonitorPong(from, to string) (*Message, error) {

	msg, err := NewMessage_MonitorPong(from, to)
	if err != nil {
		return nil, err
	}
	err = msg.Sign()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func NewMessage_StopNode(from, to string) (*Message, error) {
	msg := &Message{
		Version: CodeVersion,
		Type:    MessageType_StopNode,
		From:    from,
		To:      to,
	}
	if err := msg.WriteDesc("type", MessageType_StopNode.String()); err != nil {
		return nil, err
	}

	return msg, nil
}

func NewMessageAndSign_StopNode(from, to string) (*Message, error) {

	msg, err := NewMessage_StopNode(from, to)
	if err != nil {
		return nil, err
	}
	err = msg.Sign()
	if err != nil {
		return nil, err
	}
	return msg, nil
}