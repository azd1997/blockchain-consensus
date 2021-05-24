/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 5/22/21 12:29 AM
* @Description: The file is for
***********************************************************************/

package conn_net

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestConnStatus_DisableRecv(t *testing.T) {
	tests := []struct {
		name string
		cs   ConnStatus
		should ConnStatus
	}{
		{"ex1", 0x00, 0x00},
		{"ex2", 0x01, 0x01},
		{"ex3", 0x02, 0x00},
		{"ex4", 0x03, 0x01},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := tt.cs
			tt.cs.DisableRecv()
			if tt.cs != tt.should {
				t.Errorf("old=%x, res=%x, should=%x\n", old, tt.cs, tt.should)
			}
		})
	}
}

func TestConnStatus_DisableSend(t *testing.T) {
	tests := []struct {
		name string
		cs   ConnStatus
		should ConnStatus
	}{
		{"ex1", 0x00, 0x00},
		{"ex2", 0x01, 0x00},
		{"ex3", 0x02, 0x02},
		{"ex4", 0x03, 0x02},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := tt.cs
			tt.cs.DisableSend()
			if tt.cs != tt.should {
				t.Errorf("old=%x, res=%x, should=%x\n", old, tt.cs, tt.should)
			}
		})
	}
}

func TestConnStatus_EnableRecv(t *testing.T) {
	tests := []struct {
		name string
		cs   ConnStatus
		should ConnStatus
	}{
		{"ex1", 0x00, 0x02},
		{"ex2", 0x01, 0x03},
		{"ex3", 0x02, 0x02},
		{"ex4", 0x03, 0x03},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := tt.cs
			tt.cs.EnableRecv()
			if tt.cs != tt.should {
				t.Errorf("old=%x, res=%x, should=%x\n", old, tt.cs, tt.should)
			}
		})
	}
}

func TestConnStatus_EnableSend(t *testing.T) {
	tests := []struct {
		name string
		cs   ConnStatus
		should ConnStatus
	}{
		{"ex1", 0x00, 0x01},
		{"ex2", 0x01, 0x01},
		{"ex3", 0x02, 0x03},
		{"ex4", 0x03, 0x03},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := tt.cs
			tt.cs.EnableSend()
			if tt.cs != tt.should {
				t.Errorf("old=%x, res=%x, should=%x\n", old, tt.cs, tt.should)
			}
		})
	}
}

func TestConnStatus_String(t *testing.T) {
	tests := []struct {
		name string
		cs   ConnStatus
		want string
	}{
		{"ex1", 0x00, "Closed"},
		{"ex2", 0x01, "OnlySend"},
		{"ex3", 0x02, "OnlyRecv"},
		{"ex4", 0x03, "SendRecv"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSizeof(t *testing.T) {
	sz1 := unsafe.Sizeof(ConnStatus_Closed)
	sz2 := unsafe.Sizeof(uint8(ConnStatus_Closed))
	fmt.Println(sz1, sz2)
}

func TestConnStatus_CanSend(t *testing.T) {
	tests := []struct {
		name string
		cs   ConnStatus
		can bool
	}{
		{"ex1", 0x00, false},
		{"ex2", 0x01, true},
		{"ex3", 0x02, false},
		{"ex4", 0x03, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := tt.cs
			canSend := tt.cs.CanSend()
			if canSend != tt.can {
				t.Errorf("old=%x, canSend=%v, can=%v\n", old, canSend, tt.can)
			}
		})
	}
}

func TestConnStatus_CanRecv(t *testing.T) {
	tests := []struct {
		name string
		cs   ConnStatus
		can bool
	}{
		{"ex1", 0x00, false},
		{"ex2", 0x01, false},
		{"ex3", 0x02, true},
		{"ex4", 0x03, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := tt.cs
			canRecv := tt.cs.CanRecv()
			if canRecv != tt.can {
				t.Errorf("old=%x, canRecv=%v, can=%v\n", old, canRecv, tt.can)
			}
		})
	}
}