/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/26/20 5:16 PM
* @Description: Pot的handle相关方法
***********************************************************************/

package pot

//func TestPot_handleMsgWhenPreInited(t *testing.T) {
//	type fields struct {
//		id                  string
//		duty                defines.PeerDuty
//		epoch               int64
//		state               StateType
//		processes           *processTable
//		nWait               int
//		nWaitChan           chan int
//		msgin               chan *defines.Message
//		msgout              chan *defines.MessageWithError
//		clock               *Clock
//		potStartBeforeReady chan Moment
//		proofs              *proofTable
//		maybeNewBlock       *defines.Block
//		waitingNewBlock     *defines.Block
//		udbt                *undecidedBlockTable
//		txPool              requires.TransactionPool
//		bc                  requires.BlockChain
//		pit                 *peerinfo.PeerInfoTable
//		Logger              *log.Logger
//		done                chan struct{}
//	}
//	type args struct {
//		msg *defines.Message
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			p := &Pot{
//				id:                  tt.fields.id,
//				duty:                tt.fields.duty,
//				epoch:               tt.fields.epoch,
//				state:               tt.fields.state,
//				processes:           tt.fields.processes,
//				nWait:               tt.fields.nWait,
//				nWaitChan:           tt.fields.nWaitChan,
//				msgin:               tt.fields.msgin,
//				msgout:              tt.fields.msgout,
//				clock:               tt.fields.clock,
//				potStartBeforeReady: tt.fields.potStartBeforeReady,
//				proofs:              tt.fields.proofs,
//				maybeNewBlock:       tt.fields.maybeNewBlock,
//				waitingNewBlock:     tt.fields.waitingNewBlock,
//				udbt:                tt.fields.udbt,
//				txPool:              tt.fields.txPool,
//				bc:                  tt.fields.bc,
//				pit:                 tt.fields.pit,
//				Logger:              tt.fields.Logger,
//				done:                tt.fields.done,
//			}
//			if err := p.handleMsgWhenPreInited(tt.args.msg); (err != nil) != tt.wantErr {
//				t.Errorf("Pot.handleMsgWhenPreInited() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
