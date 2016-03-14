package main

import (
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"hash/fnv"
	"io"
	"os"
	"sort"
	"strconv"
)

var (
	Postfix = [3]string{"_X1", "_2X", "_K3"}
)

const (
	gIDPrefix = "GROUPID"
	c2 uint32 = 0x27d4eb2d
	workerCnt int = 150
)

type WorkerHashKey struct {
	ID  string
	Key uint32
	Index int
}

// Robert Jenkins's algorithm
func uint32Hash(a uint32) uint32 {
	a = (a+0x7ed55d16) + (a<<12);
	a = (a^0xc761c23c) ^ (a>>19);
	a = (a+0x165667b1) + (a<<5);
	a = (a+0xd3a2646c) ^ (a<<9);
	a = (a+0xfd7046c5) + (a<<3);
	a = (a^0xb55a4f09) ^ (a>>16);
	return a
}

func hash32shiftmult(key uint32) uint32 {
	key = (key ^ 61) ^ (key >> 16);
	key = key + (key << 3);
	key = key ^ (key >> 4);
	key = key * c2;
	key = key ^ (key >> 15);
	return key;
}

type WorkerHashKeyList []WorkerHashKey

func (w WorkerHashKeyList) Len() int {
	return len(w)
}

func (w WorkerHashKeyList) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

func (w WorkerHashKeyList) Less(i, j int) bool {
	return w[i].Key < w[j].Key
}

func findWorkerID(gID int, ring WorkerHashKeyList) (string, error) {
	hashVal := hash32shiftmult(uint32(gID))
	fmt.Fprintf(os.Stderr, "gID hash val=%v\n", hashVal)
	findPos := func(i int) bool {
		return ring[i].Key >= hashVal
	}
	pos := sort.Search(len(ring), findPos)
	if pos == len(ring) {
		fmt.Fprintf(os.Stderr, "pos=%v\n", ring[0].Index)
		return "", errors.New("fail to find")
	}
	fmt.Fprintf(os.Stderr, "pos=%v\n", ring[pos].Index)
	return ring[pos].ID, nil
}

func insertNewNode(workerID string, index int, ring *WorkerHashKeyList) error {
	h := fnv.New32a()
	var ringID [2]WorkerHashKey
	for i := 0; i < 2; i++ {
		h.Reset()
		ringID[i].ID = workerID
		io.WriteString(h, workerID+Postfix[i])
		io.WriteString(h, workerID+strconv.Itoa(index))
		ringID[i].Key = hash32shiftmult(h.Sum32())
		ringID[i].Index = index
		*ring = append(*ring, ringID[i])
	}
	sort.Sort(ring)
	/*
		length := len(*ring)
		for i := 0; i < length; i++ {
			fmt.Fprintf(os.Stderr, "ring[%v]=%v\n", i, (*ring)[i])
		}
	*/
	return nil
}

func main() {
	var ring WorkerHashKeyList

	var workerIDs [workerCnt]string
	for i := 0; i < workerCnt; i++ {
		workerIDs[i] = uuid.NewV4().String()
		insertNewNode(workerIDs[i], i, &ring)
	}
	/*
	var length = len(ring)
	for i := 0; i < length; i++ {
		fmt.Fprintf(os.Stderr, "ring[%v]=%v\n", i, ring[i])
	}
	*/

	for i := 0; i < 10 ; i++ {
		findWorkerID(i, ring)
	}

}
