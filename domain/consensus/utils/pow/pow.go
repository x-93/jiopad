package pow

import (
	"github.com/karlsen-network/karlsend/domain/consensus/model/externalapi"
	"github.com/karlsen-network/karlsend/domain/consensus/utils/consensushashing"
	"github.com/karlsen-network/karlsend/domain/consensus/utils/hashes"
	"github.com/karlsen-network/karlsend/domain/consensus/utils/serialization"
	"github.com/karlsen-network/karlsend/util/difficulty"
	"github.com/pkg/errors"

	"fmt"
	"math/big"
)

// State is an intermediate data structure with pre-computed values to speed up mining.
type State struct {
	mat        matrix
	Timestamp  int64
	Nonce      uint64
	Target     big.Int
	prePowHash externalapi.DomainHash
	//cache 	   cache
	context fishhashContext
}

// var context *fishhashContext
var sharedContext *fishhashContext

// NewState creates a new state with pre-computed values to speed up mining
// It takes the target from the Bits field
func NewState(header externalapi.MutableBlockHeader, generatedag bool) *State {
	target := difficulty.CompactToBig(header.Bits())
	// Zero out the time and nonce.
	timestamp, nonce := header.TimeInMilliseconds(), header.Nonce()
	header.SetTimeInMilliseconds(0)
	header.SetNonce(0)
	prePowHash := consensushashing.HeaderHash(header)
	header.SetTimeInMilliseconds(timestamp)
	header.SetNonce(nonce)

	if sharedContext != nil {
		log.Debugf("NewState object 12345 : %x\n", sharedContext.FullDataset[12345])
	}

	return &State{
		Target:     *target,
		prePowHash: *prePowHash,
		//will remove matrix opow
		//mat:       *generateMatrix(prePowHash),
		Timestamp: timestamp,
		Nonce:     nonce,
		context:   *getContext(generatedag),
	}
}

func (state *State) IsContextReady() bool {
	fmt.Printf("IsContextReady -- log0 %+v \n", state)
	fmt.Printf("IsContextReady -- log1 %+v \n", &state)
	fmt.Printf("IsContextReady -- log2 %+v \n", &state.context)
	if state != nil && &state.context != nil {
		return state.context.ready
	} else {
		return false
	}
}

// CalculateProofOfWorkValue hashes the internal header and returns its big.Int value
func (state *State) CalculateProofOfWorkValue() *big.Int {
	// PRE_POW_HASH || TIME || 32 zero byte padding || NONCE
	writer := hashes.NewPoWHashWriter()
	writer.InfallibleWrite(state.prePowHash.ByteSlice())
	err := serialization.WriteElement(writer, state.Timestamp)
	if err != nil {
		panic(errors.Wrap(err, "this should never happen. Hash digest should never return an error"))
	}

	zeroes := [32]byte{}
	writer.InfallibleWrite(zeroes[:])
	err = serialization.WriteElement(writer, state.Nonce)
	if err != nil {
		panic(errors.Wrap(err, "this should never happen. Hash digest should never return an error"))
	}
	//log.Debugf("Hash prePowHash %x\n", state.prePowHash.ByteSlice())
	//fmt.Printf("Hash prePowHash %x\n", state.prePowHash.ByteSlice())
	powHash := writer.Finalize()
	//middleHash := state.mat.HeavyHash(powHash)
	//log.Debugf("Hash b3-1: %x\n", powHash.ByteSlice())
	//fmt.Printf("Hash b3-1: %x\n", powHash.ByteSlice())
	middleHash := fishHash(&state.context, powHash)
	//log.Debugf("Hash fish: %x\n", middleHash.ByteSlice())
	//fmt.Printf("Hash fish: %x\n", middleHash.ByteSlice())

	/*
		writer2 := hashes.NewHeavyHashWriter()
		writer2.InfallibleWrite(heavyHash.ByteSlice())
		finalHash := writer2.Finalize()
	*/

	writer2 := hashes.NewPoWHashWriter()
	writer2.InfallibleWrite(middleHash.ByteSlice())
	finalHash := writer2.Finalize()

	//log.Debugf("Hash b3-2: %x\n", finalHash.ByteSlice())
	//fmt.Printf("Hash b3-2: %x\n", finalHash.ByteSlice())
	return toBig(finalHash)
	//return toBig(heavyHash)
}

// IncrementNonce the nonce in State by 1
func (state *State) IncrementNonce() {
	state.Nonce++
}

// CheckProofOfWork check's if the block has a valid PoW according to the provided target
// it does not check if the difficulty itself is valid or less than the maximum for the appropriate network
func (state *State) CheckProofOfWork() bool {
	// The block pow must be less than the claimed target
	powNum := state.CalculateProofOfWorkValue()

	// The block hash must be less or equal than the claimed target.
	return powNum.Cmp(&state.Target) <= 0
}

// CheckProofOfWorkByBits check's if the block has a valid PoW according to its Bits field
// it does not check if the difficulty itself is valid or less than the maximum for the appropriate network
func CheckProofOfWorkByBits(header externalapi.MutableBlockHeader) bool {
	return NewState(header, false).CheckProofOfWork()
}

// ToBig converts a externalapi.DomainHash into a big.Int treated as a little endian string.
func toBig(hash *externalapi.DomainHash) *big.Int {
	// We treat the Hash as little-endian for PoW purposes, but the big package wants the bytes in big-endian, so reverse them.
	buf := hash.ByteSlice()
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf)
}

// BlockLevel returns the block level of the given header.
func BlockLevel(header externalapi.BlockHeader, maxBlockLevel int) int {
	// Genesis is defined to be the root of all blocks at all levels, so we define it to be the maximal
	// block level.
	if len(header.DirectParents()) == 0 {
		return maxBlockLevel
	}

	proofOfWorkValue := NewState(header.ToMutable(), false).CalculateProofOfWorkValue()
	level := maxBlockLevel - proofOfWorkValue.BitLen()
	// If the block has a level lower than genesis make it zero.
	if level < 0 {
		level = 0
	}
	return level
}
