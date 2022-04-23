package math

const (
	bitsize       = 32 << (^uint(0) >> 63) // 64
	maxintHeadBit = 1 << (bitsize - 2)
)

// IsPowerOfTwo 判断是否是2的幂次方
func IsPowerOfTwo(n int) bool {
	return n&(n-1) == 0
}

// CeilToPowerOfTwo 返回数轴右侧离n最近的2的幂次方（大于等于n）
func CeilToPowerOfTwo(n int) int {
	if n&maxintHeadBit != 0 && n > maxintHeadBit {
		panic("argument is too large")
	}
	if n <= 2 {
		return 2
	}
	n--
	n = fillBits(n)
	n++
	return n
}

// FloorToPowerOfTwo 返回数轴左侧离n最近的2的幂次方(小于等于n)
func FloorToPowerOfTwo(n int) int {
	if n <= 2 {
		return 2
	}
	n = fillBits(n)
	n >>= 1
	n++
	return n
}

//
func fillBits(n int) int {
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n
}
