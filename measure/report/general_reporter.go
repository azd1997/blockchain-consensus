package report

import (
	"errors"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/bnet"
)

type GeneralReporter struct {
	monitorId, monitorAddr string
	net bnet.BNet
}

func NewGeneralReporter(monitorId, monitorAddr string, net bnet.BNet) (*GeneralReporter, error) {
	return &GeneralReporter{
		monitorId: monitorId,
		monitorAddr: monitorAddr,
		net: net,
	}, nil
}

func (r *GeneralReporter) ReportNewBlock(nb *defines.Block) error {
	if r.net == nil {return errors.New("GeneralReporter.ReportNewBlock fail: client == nil")}

	// 构建区块消息
	msg, err := defines.NewMessage_NewBlock(r.net.ID(), r.monitorId, 0, nb)
	if err != nil {
		return err
	}
	// 发送区块消息
	err = r.net.Send(r.monitorId, r.monitorAddr, msg)
	if err != nil {
		return err
	}
	return nil
}

func (r *GeneralReporter) ReportNewTx(tx ...*defines.Transaction) error {
	if r.net == nil {return errors.New("GeneralReporter.ReportNewTx fail: client == nil")}

	// 构建交易消息
	msg, err := defines.NewMessage_Txs(r.net.ID(), r.monitorId, 0, tx)
	if err != nil {
		return err
	}
	// 发送交易消息
	err = r.net.Send(r.monitorId, r.monitorAddr, msg)
	if err != nil {
		return err
	}
	return nil
}
