package util

import "os"

const (
	MaxUint32 = 4294967295
	DefaultUUIDCache = 128
)

// generate unique ID
type UUIDGenerator struct {
	Prefix       string
	idGen        uint32
	internalChan chan uint32
}

func MakeUUIDGenerator(prefix string) *UUIDGenerator {
	gen := &UUIDGenerator{
		Prefix:       prefix,
		idGen:        0,
		internalChan: make(chan uint32, DefaultUUIDCache),
	}
	gen.startGen()
	return gen
}

// start goroutine, put UUID into chan
func (ug *UUIDGenerator) startGen() {
	go func() {
		for {
			if ug.idGen == MaxUint32 {
				ug.idGen = 1
			} else {
				ug.idGen += 1
			}
			ug.internalChan <- ug.idGen
		}
	}()
}

// get UUID
func (ug *UUIDGenerator) Get() string {
	idgen := <-ug.internalChan
	return ug.Prefix + string(idgen)
}

func CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
