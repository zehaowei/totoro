package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

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

// measure cpu usage of the whole machine
func GetMachineCpuUsage() float64 {
	idle1, total1 := getCPUSample()
	time.Sleep(2 * time.Second)
	idle2, total2 := getCPUSample()
	idleTicks := float64(idle2 - idle1)
	totalTicks := float64(total2 - total1)
	cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks
	return cpuUsage
}

func getCPUSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}
