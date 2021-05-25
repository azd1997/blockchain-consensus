package report

import (
	"errors"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/bnet/cs_net"
)

type EchartsReporter struct {
	client *cs_net.Client
}

func NewEchartsReporter(serverId, serverListenAddr, clientId, clientListenAddr string) (*EchartsReporter, error) {
	msgout := make(chan *defines.Message)	// 不会使用
	client, err := cs_net.Dial(serverId, serverListenAddr, clientId, clientListenAddr, msgout)
	if err != nil {
		return nil, err
	}
	return &EchartsReporter{client: client}, nil
}

func (r *EchartsReporter) ReportNewBlock(nb *defines.Block) error {
	if r.client == nil {return errors.New("EchartsReporter.ReportNewBlock fail: client == nil")}

	// 构建区块消息
	msg, err := defines.NewMessage_NewBlock("", "", 0, nb)
	if err != nil {
		return err
	}
	// 发送区块消息
	err = r.client.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func (r *EchartsReporter) ReportNewTx(tx ...*defines.Transaction) error {
	if r.client == nil {return errors.New("EchartsReporter.ReportNewTx fail: client == nil")}

	// 构建交易消息
	msg, err := defines.NewMessage_Txs("", "", 0, tx)
	if err != nil {
		return err
	}
	// 发送交易消息
	err = r.client.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

