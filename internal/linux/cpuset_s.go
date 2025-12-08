package linux

import (
	"math/bits"
	"os"
	"strconv"
	"strings"
)

const _NCPUBITS = 64

type cpuMask uint64

type CPUSetS []cpuMask

var _possibleCPUs int = 0

func possibleCPUs() int {
	if _possibleCPUs == 0 {
		_possibleCPUs = getLastPossibleCPU() + 1
	}
	return _possibleCPUs
}

func getLastPossibleCPU() int {
	data, err := os.ReadFile("/sys/devices/system/cpu/possible")
	if err != nil {
		return 0
	}
	s := strings.TrimSpace(string(data))
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return 0
	}
	last, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	return last
}

// cpuBitsIndex returns the index in the CPUSetS slice
// for the given cpu. It also ensures that the slice is
// large enough to hold the cpu.
func (s *CPUSetS) cpuBitsIndex(cpu int) int {
	index := cpu / 64
	if len(*s) <= index {
		*s = append(*s, make([]cpuMask, index-len(*s)+1)...)
	}
	return index
}

// Zero clears the set s, so that it contains no CPUs.
func (s *CPUSetS) Zero() {
	*s = make([]cpuMask, len(*s))
}

// Fill adds all possible CPU bits to the set s.
func (s *CPUSetS) Fill() {
	for i := range s.cpuBitsIndex(possibleCPUs()-1) + 1 {
		(*s)[i] = ^cpuMask(0)
	}
	// If possibleCPUs is not divisible by 64, there are now extra
	// 1s in s. Clear the last 1s from the last index so that
	// number of ones will match possible CPUs, and the last 1
	// will be at possibleCPUs()-1.
	extraOnes := len(*s)*_NCPUBITS - possibleCPUs()
	if extraOnes > 0 {
		(*s)[len(*s)-1] &^= ^cpuMask(0) << (uint(_NCPUBITS) - uint(extraOnes))
	}
}

func cpuBitsMask(cpu int) cpuMask {
	return cpuMask(1 << (uint(cpu) % _NCPUBITS))
}

// Set adds cpu to the set s.
func (s *CPUSetS) Set(cpu int) {
	i := s.cpuBitsIndex(cpu)
	(*s)[i] |= cpuBitsMask(cpu)
}

// Clear removes cpu from the set s.
func (s *CPUSetS) Clear(cpu int) {
	i := s.cpuBitsIndex(cpu)
	(*s)[i] &^= cpuBitsMask(cpu)
}

// IsSet reports whether cpu is in the set s.
func (s *CPUSetS) IsSet(cpu int) bool {
	if cpu > len(*s)*_NCPUBITS {
		return false
	}
	i := s.cpuBitsIndex(cpu)
	if i < len(*s) {
		return (*s)[i]&cpuBitsMask(cpu) != 0
	}
	return false
}

// Count returns the number of CPUs in the set s.
func (s *CPUSetS) Count() int {
	c := 0
	for _, b := range *s {
		c += bits.OnesCount64(uint64(b))
	}
	return c
}
