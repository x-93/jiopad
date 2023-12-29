package pow

import (
	"github.com/karlsen-network/karlsend/domain/consensus/model/externalapi"
	"golang.org/x/crypto/sha3"

	//"crypto/sha3"
	"encoding/binary"
	"sync"
)

// ============================================================================

const (
	fnvPrime               = 0x01000193
	fullDatasetItemParents = 512
	numDatasetAccesses     = 32
	lightCacheRounds       = 3
	lightCacheNumItems     = 1179641
	fullDatasetNumItems    = 37748717
)

var (
	// we use the same seed as fish hash for debug reasons
	seed = [32]byte{
		0xeb, 0x01, 0x63, 0xae, 0xf2, 0xab, 0x1c, 0x5a,
		0x66, 0x31, 0x0c, 0x1c, 0x14, 0xd6, 0x0f, 0x42,
		0x55, 0xa9, 0xb3, 0x9b, 0x0e, 0xdf, 0x26, 0x53,
		0x98, 0x44, 0xf1, 0x17, 0xad, 0x67, 0x21, 0x19,
	}

	//sharedContext     *fishhashContext
	sharedContextLock sync.Mutex
)

type hash256 [32]byte
type hash512 [64]byte
type hash1024 [128]byte

type fishhashContext struct {
	ready               bool
	LightCacheNumItems  int
	LightCache          []*hash512
	FullDatasetNumItems uint32
	FullDataset         []hash1024
}

func fnv1(u, v uint32) uint32 {
	return (u * fnvPrime) ^ v
}

func fnv1Hash512(u, v hash512) hash512 {
	var result hash512
	for i := 0; i < len(result); i++ {
		result[i] = byte(fnv1(uint32(u[i]), uint32(v[i])))
	}
	return result
}

type itemState struct {
	cache         []*hash512
	numCacheItems int64
	seed          uint32
	mix           hash512
}

func newItemState(ctx *fishhashContext, index int64) *itemState {
	state := &itemState{
		cache:         ctx.LightCache,
		numCacheItems: int64(ctx.LightCacheNumItems),
		seed:          uint32(index),
	}

	state.mix = *state.cache[index%state.numCacheItems]
	state.mix[0] ^= byte(state.seed)

	hash := sha3.New512()
	hash.Write(state.mix[:])
	copy(state.mix[:], hash.Sum(nil))

	return state
}

func (state *itemState) update(round uint32) {
	numWords := len(state.mix) / 4
	//t := fnv1(state.seed^round, uint32(state.mix[round%uint32(numWords)]))
	t := fnv1(state.seed^round, binary.BigEndian.Uint32(state.mix[round%uint32(numWords):]))
	parentIndex := t % uint32(state.numCacheItems)
	state.mix = fnv1Hash512(state.mix, *state.cache[parentIndex])
}

func (state *itemState) final() hash512 {
	hash := sha3.New512()
	hash.Write(state.mix[:])
	copy(state.mix[:], hash.Sum(nil))
	return state.mix
}

func calculateDatasetItem1024(ctx *fishhashContext, index uint32) hash1024 {
	item0 := newItemState(ctx, int64(index)*2)
	item1 := newItemState(ctx, int64(index)*2+1)

	for j := uint32(0); j < fullDatasetItemParents; j++ {
		item0.update(j)
		item1.update(j)
	}
	return mergeHashes(item0.final(), item1.final())
}

func lookup(ctx *fishhashContext, index uint32) hash1024 {
	if ctx.FullDataset != nil {
		item := &ctx.FullDataset[index]
		log.Debugf(" %d - ", index)

		if item[0] == 0 {
			*item = calculateDatasetItem1024(ctx, index)
		}

		return *item
	}

	return calculateDatasetItem1024(ctx, index)
}

func mergeHashes(hash1, hash2 hash512) (result hash1024) {
	copy(result[:len(hash1)], hash1[:])
	copy(result[len(hash1):], hash2[:])
	return
}

func fishhashKernel(ctx *fishhashContext, seed hash512) hash256 {
	indexLimit := uint32(ctx.FullDatasetNumItems)
	//seedInit := uint32(seed[0])
	//mix := hash1024{seed, seed}
	mix := mergeHashes(seed, seed)

	//fmt.Printf("The index_limit is : %d \n", indexLimit)
	//fmt.Printf("The seed is : %x \n", seed)
	//fmt.Printf("The mix is : %x \n", mix)

	log.Debugf("lookup matrix : ")
	for i := uint32(0); i < numDatasetAccesses; i++ {
		/*
			p0 := uint32(mix[0]) % indexLimit
			p1 := uint32(mix[4]) % indexLimit
			p2 := uint32(mix[8]) % indexLimit
		*/

		p0 := binary.BigEndian.Uint32(mix[0:4]) % indexLimit
		p1 := binary.BigEndian.Uint32(mix[4:8]) % indexLimit
		p2 := binary.BigEndian.Uint32(mix[8:12]) % indexLimit

		//fmt.Printf("The words is : %d - %d - %d\n", mix[0], mix[4], mix[8])
		//fmt.Printf("The words lg is : %d - %d - %d\n", mix[0:4], mix[4:8], mix[8:12])
		//fmt.Printf("The words32 is : %d - %d - %d\n", uint32(mix[0]), uint32(mix[4]), uint32(mix[8]))
		//fmt.Printf("The words32 lg is : %d - %d - %d\n", binary.BigEndian.Uint32(mix[0:4]), binary.BigEndian.Uint32(mix[4:8]), binary.BigEndian.Uint32(mix[8:12]))
		//fmt.Printf("The indexes is : %d - %d - %d\n", p0, p1, p2)

		fetch0 := lookup(ctx, p0)
		fetch1 := lookup(ctx, p1)
		fetch2 := lookup(ctx, p2)

		//fmt.Printf("The fetch0 is : %x \n", fetch0)
		//fmt.Printf("The fetch1 is : %x \n", fetch1)
		//fmt.Printf("The fetch2 is : %x \n", fetch2)

		for j := 0; j < 32; j++ {
			//fetch1[j] = byte(fnv1(uint32(mix[j]), uint32(fetch1[j])))
			//fetch2[j] = mix[j] ^ fetch2[j]
			binary.BigEndian.PutUint32(
				fetch1[4*j:],
				fnv1(binary.BigEndian.Uint32(mix[4*j:4*j+4]), binary.BigEndian.Uint32(fetch1[4*j:4*j+4])))
			binary.BigEndian.PutUint32(
				fetch2[4*j:],
				binary.BigEndian.Uint32(mix[4*j:4*j+4])^binary.BigEndian.Uint32(fetch2[4*j:4*j+4]))
		}

		//fmt.Printf("The NEW fetch1 is : %x \n", fetch1)
		//fmt.Printf("The NEW fetch2 is : %x \n", fetch2)

		for j := 0; j < 16; j++ {
			//mix[j] = fetch0[j]*fetch1[j] + fetch2[j]
			binary.BigEndian.PutUint64(
				mix[8*j:],
				binary.BigEndian.Uint64(fetch0[8*j:8*j+8])*binary.BigEndian.Uint64(fetch1[8*j:8*j+8])+binary.BigEndian.Uint64(fetch2[8*j:8*j+8]))
		}
		log.Debugf("\n")
	}

	//fmt.Printf("The FINAL mix is : %x \n", mix)

	mixHash := hash256{}
	for i := 0; i < (len(mix) / 4); i += 4 {
		//h1 := fnv1(uint32(mix[i]), uint32(mix[i+1]))
		//h2 := fnv1(h1, uint32(mix[i+2]))
		//h3 := fnv1(h2, uint32(mix[i+3]))
		//mixHash[i/4] = byte(h3)
		j := 4 * i
		h1 := fnv1(binary.BigEndian.Uint32(mix[j:j+4]), binary.BigEndian.Uint32(mix[(j+1):(j+1)+4]))
		h2 := fnv1(h1, binary.BigEndian.Uint32(mix[(j+2):(j+2)+4]))
		h3 := fnv1(h2, binary.BigEndian.Uint32(mix[(j+3):(j+3)+4]))
		binary.BigEndian.PutUint32(mixHash[i:], h3)

	}

	return mixHash
}

func hash(output, header []byte, ctx *fishhashContext) {
	seed := hash512{}

	hasher := sha3.New512()
	hasher.Write(header)
	copy(seed[:], hasher.Sum(nil))

	mixHash := fishhashKernel(ctx, seed)

	finalData := make([]byte, len(seed)+len(mixHash))
	copy(finalData[:len(seed)], seed[:])
	copy(finalData[len(seed):], mixHash[:])

	hasher.Reset()
	hasher.Write(finalData)
	copy(output, hasher.Sum(nil))
}

func bitwiseXOR(x, y hash512) hash512 {
	var result hash512
	for i := 0; i < len(result); i++ {
		result[i] = x[i] ^ y[i]
	}
	return result
}

func buildLightCache(cache []*hash512, numItems int, seed hash256) {

	println("GENERATING LIGHT CACHE ===============================================\n")
	item := hash512{}
	hash := sha3.New512()
	hash.Write(seed[:])
	copy(item[:], hash.Sum(nil))
	cache[0] = &item

	for i := 1; i < numItems; i++ {
		hash.Reset()
		hash.Write(item[:])
		copy(item[:], hash.Sum(nil))
		cache[i] = &item
	}

	for q := 0; q < lightCacheRounds; q++ {
		for i := 0; i < numItems; i++ {
			indexLimit := uint32(numItems)
			t := uint32(cache[i][0])
			v := t % indexLimit
			w := uint32(numItems+(i-1)) % indexLimit
			x := bitwiseXOR(*cache[v], *cache[w])

			hash.Reset()
			hash.Write(x[:])
			copy(cache[i][:], hash.Sum(nil))
		}
	}
}

func buildDatasetSegment(ctx *fishhashContext, start, end uint32) {
	for i := start; i < end; i++ {
		ctx.FullDataset[i] = calculateDatasetItem1024(ctx, i)
	}
}

func getContext(full bool) *fishhashContext {
	sharedContextLock.Lock()
	defer sharedContextLock.Unlock()

	if sharedContext != nil {
		if !full || sharedContext.FullDataset != nil {
			log.Debugf("log0 getContext ====\n")
			return sharedContext
		}
		log.Debugf("log1 getContext ==== going to build dataset\n")
	}

	// DIABLE LIGHT CACHE FOR THE MOMENT

	lightCache := make([]*hash512, lightCacheNumItems)
	log.Debugf("getContext ==== building light cache\n")
	buildLightCache(lightCache, lightCacheNumItems, seed)
	log.Debugf("getContext ==== light cache done\n")

	log.Debugf("getContext fullDatasetNumItems - 2.0 : %d\n", fullDatasetNumItems)
	fullDataset := make([]hash1024, fullDatasetNumItems)

	sharedContext = &fishhashContext{
		ready:               false,
		LightCacheNumItems:  lightCacheNumItems,
		LightCache:          lightCache,
		FullDatasetNumItems: fullDatasetNumItems,
		FullDataset:         fullDataset,
	}

	log.Debugf("getContext object 12345 : %x\n", fullDataset[12345])

	//test
	if full {
		log.Debugf("getContext ==== building full dataset\n")
		prebuildDataset(sharedContext, 8)
		log.Debugf("getContext ==== full dataset built\n")
	}

	return sharedContext
}

func prebuildDataset(ctx *fishhashContext, numThreads uint32) {
	log.Debugf("prebuildDataset ==================================================\n")

	if ctx.FullDataset == nil {
		return
	}

	if ctx.ready == true {
		log.Debugf("dataset already generated\n")
		return
	}
	println("GENERATING DATASET ===============================================\n")

	if numThreads > 1 {
		log.Debugf("prebuildDataset multi thread")
		batchSize := ctx.FullDatasetNumItems / numThreads
		var wg sync.WaitGroup

		for i := uint32(0); i < numThreads; i++ {
			start := i * batchSize
			end := start + batchSize
			if i == numThreads-1 {
				end = ctx.FullDatasetNumItems
			}

			wg.Add(1)
			go func(ctx *fishhashContext, start, end uint32) {
				defer wg.Done()
				buildDatasetSegment(ctx, start, end)
			}(ctx, start, end)
		}

		wg.Wait()
	} else {
		log.Debugf("prebuildDataset solo thread\n")
		buildDatasetSegment(ctx, 0, ctx.FullDatasetNumItems)
	}

	log.Debugf("DATASET GENERATED ===============================================\n")
	ctx.ready = true
}

func fishHash(ctx *fishhashContext, hashin *externalapi.DomainHash) *externalapi.DomainHash {
	output := make([]byte, 32)
	hash(output, hashin.ByteSlice(), ctx)
	outputArray := [32]byte{}
	copy(outputArray[:], output)
	return externalapi.NewDomainHashFromByteArray(&outputArray)
}
