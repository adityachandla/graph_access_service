package bin_util

// The storage format is little endian
func ByteToUint(bytes []byte) uint32 {
	return uint32(bytes[0]) | (uint32(bytes[1]) << 8) |
		(uint32(bytes[2]) << 16) | (uint32(bytes[3]) << 24)
}

func ByteArrayToUintArray(bytes []byte) []uint32 {
	if len(bytes)%4 != 0 {
		panic("Invalid byte size")
	}
	arr := make([]uint32, len(bytes)/4)
	for i := 0; i < len(bytes); i += 4 {
		idx := i / 4
		arr[idx] = ByteToUint(bytes[i : i+4])
	}
	return arr
}

type pair struct {
	a, b uint32
}

func ByteArrayToPairArray(bytes []byte) []pair {
	if len(bytes)%8 != 0 {
		panic("Invalid byte size")
	}
	arr := make([]pair, len(bytes)/8)
	for i := 0; i < len(bytes); i += 8 {
		idx := i / 8
		arr[idx] = pair{
			a: ByteToUint(bytes[i : i+4]),
			b: ByteToUint(bytes[i+4 : i+8]),
		}
	}
	return arr
}
