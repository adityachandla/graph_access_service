package bin_util_test

import (
	"testing"
	"unsafe"

	"github.com/adityachandla/graph_access_service/bin_util"
)

func TestBoundsSingle(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("No panic occurred")
		}
	}()
	arr := []byte{1, 2, 3}
	bin_util.ByteToUint(arr)
}

func TestBoundsArray(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("No panic occurred")
		}
	}()
	arr := []byte{1, 2, 3, 4, 5, 6}
	bin_util.ByteArrayToUintArray(arr)
}

func TestBoundsPairArray(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("No panic occurred")
		}
	}()
	arr := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	bin_util.ByteArrayToPairArray(arr)
}

func TestSingleUintConversion(t *testing.T) {
	one := []byte{1, 0, 0, 0}
	res := bin_util.ByteToUint(one)
	if res != 1 {
		t.Fatalf("Conversion error")
	}
	two := []byte{0x01, 0x02, 0x03, 0x04}
	res = bin_util.ByteToUint(two)
	if res != 0x04030201 {
		t.Fatalf("Conversion error %x\n", res)
	}
}

func TestArrayUintConversion(t *testing.T) {
	bArr := []byte{0xfa, 0x2a, 0xba, 0xac, 0x12, 0x91, 0x1d, 0xef}
	arr := bin_util.ByteArrayToUintArray(bArr)
	if arr[0] != 0xacba2afa {
		t.Fail()
	}
	if arr[1] != 0xef1d9112 {
		t.Fail()
	}
}

type customPair struct {
	a, b uint32
}

func TestArrayToPairConversion(t *testing.T) {
	bArr := []byte{0xfa, 0x2a, 0xba, 0xac, 0x12, 0x91, 0x1d, 0xef}
	arr := bin_util.ByteArrayToPairArray(bArr)
	pairArray := *(*[]customPair)(unsafe.Pointer(&arr))
	if pairArray[0].a != 0xacba2afa {
		t.Fail()
	}
	if pairArray[0].b != 0xef1d9112 {
		t.Fail()
	}
}
