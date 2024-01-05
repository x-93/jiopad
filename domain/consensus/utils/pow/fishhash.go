package pow

import (
	"github.com/edsrzf/mmap-go"
	"github.com/karlsen-network/karlsend/domain/consensus/model/externalapi"
	"golang.org/x/crypto/sha3"

	//"crypto/sha3"
	"encoding/binary"
	"fmt"
	"os"
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

/*
type hash512 struct {
    data [64]byte
}

func (h *hash512) word32s() []uint32 {
    word32s := make([]uint32, 16)
    for i := range word32s {
        word32s[i] = binary.LittleEndian.Uint32(h.data[i*4 : (i+1)*4])
    }
    return word32s
}

func (h *hash512) setWord32s(index int, value uint32) {
    if index < 0 || index >= 16 {
        panic("index out of range")
    }
    start := index * 4
    binary.LittleEndian.PutUint32(h.data[start:start+4], value)
}
*/

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
	/*
		for i := 0; i < len(result); i++ {
			//result[i] = byte(fnv1(uint32(u[i]), uint32(v[i])))
			gxghsgh
		}
	*/

	for j := 0; j < 16; j++ {
		//fetch1[j] = byte(fnv1(uint32(mix[j]), uint32(fetch1[j])))
		//fetch2[j] = mix[j] ^ fetch2[j]
		binary.LittleEndian.PutUint32(result[4*j:], fnv1(binary.LittleEndian.Uint32(u[4*j:]), binary.LittleEndian.Uint32(v[4*j:])))
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
	//state.mix[0] ^= byte(state.seed)
	binary.LittleEndian.PutUint32(state.mix[0:], binary.LittleEndian.Uint32(state.mix[0:])^state.seed)

	//hash := sha3.New512()
	hash := sha3.NewLegacyKeccak512()
	hash.Write(state.mix[:])
	copy(state.mix[:], hash.Sum(nil))

	return state
}

func (state *itemState) update(round uint32) {
	numWords := len(state.mix) / 4
	//t := fnv1(state.seed^round, uint32(state.mix[round%uint32(numWords)]))
	t := fnv1(state.seed^round, binary.LittleEndian.Uint32(state.mix[4*(round%uint32(numWords)):]))
	parentIndex := t % uint32(state.numCacheItems)
	state.mix = fnv1Hash512(state.mix, *state.cache[parentIndex])
}

/*
func (state *itemState) updateD(round uint32) {
	numWords := len(state.mix) / 4
	//t := fnv1(state.seed^round, uint32(state.mix[round%uint32(numWords)]))
	//t := fnv1(state.seed^round, binary.BigEndian.Uint32(state.mix[round%uint32(numWords):]))
	t := fnv1(state.seed^round, binary.LittleEndian.Uint32(state.mix[4*(round%uint32(numWords)):]))
	parentIndex := t % uint32(state.numCacheItems)

	if round == 1 {
		fmt.Printf("updateD state.seed is : %d\n", state.seed)
		fmt.Printf("updateD round is : %d\n", round)
		fmt.Printf("updateD round-uint32(numWords) is : %d\n", round%uint32(numWords))
		fmt.Printf("updateD state.mix is : %x\n", state.mix[4*(round%uint32(numWords)):])
		fmt.Printf("updateD LittleEndian is : %d\n", binary.LittleEndian.Uint32(state.mix[4*(round%uint32(numWords)):]))

		fmt.Printf("updateD numWords is : %d\n", numWords)
		fmt.Printf("updateD t is : %d\n", t)
		fmt.Printf("updateD parentIndex is : %d\n", parentIndex)
		fmt.Printf("updateD state.mix BEFORE is : %x\n", state.mix)
	}
	state.mix = fnv1Hash512(state.mix, *state.cache[parentIndex])

	if round == 1 {
		fmt.Printf("updateD cache parentindex is : %x\n", *state.cache[parentIndex])
		fmt.Printf("updateD state.mix FINAL is : %x\n", state.mix)
	}

	//fmt.Printf("updateD round is : %d", round)
	//fmt.Printf("mix : %x\n", state.mix)
}
*/

func (state *itemState) final() hash512 {
	//hash := sha3.New512()
	hash := sha3.NewLegacyKeccak512()
	hash.Write(state.mix[:])
	copy(state.mix[:], hash.Sum(nil))
	return state.mix
}

func calculateDatasetItem1024(ctx *fishhashContext, index uint32) hash1024 {
	item0 := newItemState(ctx, int64(index)*2)
	item1 := newItemState(ctx, int64(index)*2+1)

	/*
		if index < 3 {
			fmt.Printf("calculateDatasetItem1024 item0.mix is : %x\n", item0.mix)
			fmt.Printf("calculateDatasetItem1024 item1.mix is : %x\n", item1.mix)
		}

		if index < 3 {
			for j := uint32(0); j < fullDatasetItemParents; j++ {
				item0.updateD(j)
				item1.updateD(j)
			}
			fmt.Printf("calculateDatasetItem1024 item0.mix update is : %x\n", item0.mix)
			fmt.Printf("calculateDatasetItem1024 item1.mix update is : %x\n", item1.mix)
		} else {
	*/
	for j := uint32(0); j < fullDatasetItemParents; j++ {
		item0.update(j)
		item1.update(j)
	}
	//}

	/*
		if index < 3 {
			fmt.Printf("calculateDatasetItem1024 item0.mix is : %x\n", item0.mix)
			fmt.Printf("calculateDatasetItem1024 item1.mix is : %x\n", item1.mix)
		}
	*/

	it0 := item0.final()
	it1 := item1.final()

	/*
		if index < 3 {
			fmt.Printf("calculateDatasetItem1024 it0 is : %x\n", it0)
			fmt.Printf("calculateDatasetItem1024 it1 is : %x\n", it1)
			fmt.Printf("calculateDatasetItem1024 merge is : %x\n", mergeHashes(it0, it1))
		}
	*/

	return mergeHashes(it0, it1)
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

		p0 := binary.LittleEndian.Uint32(mix[0:]) % indexLimit
		p1 := binary.LittleEndian.Uint32(mix[4*4:]) % indexLimit
		p2 := binary.LittleEndian.Uint32(mix[8*4:]) % indexLimit

		//fmt.Printf("The words is : %d - %d - %d\n", mix[0], mix[4], mix[8])
		//fmt.Printf("The words lg is : %x - %x - %x\n", mix[0:4], mix[4:8], mix[8:12])
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
			binary.LittleEndian.PutUint32(
				fetch1[4*j:],
				fnv1(binary.LittleEndian.Uint32(mix[4*j:4*j+4]), binary.LittleEndian.Uint32(fetch1[4*j:4*j+4])))
			binary.LittleEndian.PutUint32(
				fetch2[4*j:],
				binary.LittleEndian.Uint32(mix[4*j:4*j+4])^binary.LittleEndian.Uint32(fetch2[4*j:4*j+4]))
		}

		//fmt.Printf("The NEW fetch1 is : %x \n", fetch1)
		//fmt.Printf("The NEW fetch2 is : %x \n", fetch2)

		for j := 0; j < 16; j++ {
			//mix[j] = fetch0[j]*fetch1[j] + fetch2[j]
			binary.LittleEndian.PutUint64(
				mix[8*j:],
				binary.LittleEndian.Uint64(fetch0[8*j:8*j+8])*binary.LittleEndian.Uint64(fetch1[8*j:8*j+8])+binary.LittleEndian.Uint64(fetch2[8*j:8*j+8]))
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
		/*
			h1 := fnv1(binary.LittleEndian.Uint32(mix[j:j+4]), binary.LittleEndian.Uint32(mix[(j+1):(j+1)+4]))
			h2 := fnv1(h1, binary.LittleEndian.Uint32(mix[(j+2):(j+2)+4]))
			h3 := fnv1(h2, binary.LittleEndian.Uint32(mix[(j+3):(j+3)+4]))
			binary.LittleEndian.PutUint32(mixHash[i:], h3)
		*/
		h1 := fnv1(binary.LittleEndian.Uint32(mix[j:]), binary.LittleEndian.Uint32(mix[j+4:]))
		h2 := fnv1(h1, binary.LittleEndian.Uint32(mix[j+8:]))
		h3 := fnv1(h2, binary.LittleEndian.Uint32(mix[j+12:]))
		binary.LittleEndian.PutUint32(mixHash[i:], h3)
	}

	//fmt.Printf("The COLLAPSED mix is : %x \n", mixHash)

	return mixHash
}

/*
	func hash(output, header []byte, ctx *fishhashContext) {
		seed := hash512{}

		//hasher := sha3.New512()
		hasher := sha3.NewLegacyKeccak512()
		hasher.Write(header)
		copy(seed[:], hasher.Sum(nil))

		//we by pass for testing
		seed = hash512{0x95, 0x32, 0xc2, 0x3a, 0x1f, 0x0e, 0x71, 0x22,
			0xd9, 0x53, 0xf5, 0xe4, 0x17, 0xe3, 0x0e, 0x95,
			0xec, 0x4f, 0x8f, 0x49, 0x56, 0x8c, 0x56, 0x9f,
			0xd8, 0x62, 0xe3, 0x05, 0xa5, 0x18, 0x39, 0xd9}

		fmt.Printf("The B3-1 hash is : %x \n", seed)

		mixHash := fishhashKernel(ctx, seed)

		fmt.Printf("The kernel hash is : %x \n", mixHash)

		finalData := make([]byte, len(seed)+len(mixHash))
		copy(finalData[:len(seed)], seed[:])
		copy(finalData[len(seed):], mixHash[:])

		fmt.Printf("The finalData hash is : %x \n", finalData)

		hasher.Reset()
		hasher.Write(finalData)
		copy(output, hasher.Sum(nil))

		fmt.Printf("The B3-2 fishhash is : %x \n", output)

		os.Exit(42)
	}
*/
func fishHash(ctx *fishhashContext, hashin *externalapi.DomainHash) *externalapi.DomainHash {
	/*
		output := make([]byte, 32)
		hash(output, hashin.ByteSlice(), ctx)
		outputArray := [32]byte{}
		copy(outputArray[:], output)
	*/
	seed := hash512{}
	//output := hash256{}
	//output := make([]byte, 32)
	copy(seed[:], hashin.ByteSlice())

	//we by pass for testing
	/*
		seed = hash512{0x95, 0x32, 0xc2, 0x3a, 0x1f, 0x0e, 0x71, 0x22,
			0xd9, 0x53, 0xf5, 0xe4, 0x17, 0xe3, 0x0e, 0x95,
			0xec, 0x4f, 0x8f, 0x49, 0x56, 0x8c, 0x56, 0x9f,
			0xd8, 0x62, 0xe3, 0x05, 0xa5, 0x18, 0x39, 0xd9}
		fmt.Printf("The B3-1 FORCED hash is : %x \n", seed)
	*/

	output := fishhashKernel(ctx, seed)
	outputArray := [32]byte{}
	copy(outputArray[:], output[:])
	return externalapi.NewDomainHashFromByteArray(&outputArray)
}

func bitwiseXOR(x, y hash512) hash512 {
	/*var result hash512
	for i := 0; i < len(result); i++ {
		result[i] = x[i] ^ y[i]
	}
	return result*/
	var result hash512
	for i := 0; i < 8; i++ {
		//binary.BigEndian.PutUint64(result[4*i:], binary.BigEndian.Uint64(x[4*i:])^binary.BigEndian.Uint64(y[4*i:]))
		//binary.LittleEndian.PutUint64(result[4*i:], binary.LittleEndian.Uint64(x[4*i:])^binary.LittleEndian.Uint64(y[4*i:]))
		binary.LittleEndian.PutUint64(result[8*i:], binary.LittleEndian.Uint64(x[8*i:])^binary.LittleEndian.Uint64(y[8*i:]))
	}
	return result
}

func buildLightCache(cache []*hash512, numItems int, seed hash256) {

	println("GENERATING LIGHT CACHE ===============================================\n")
	item := hash512{}
	//hash := sha3.New512()
	hash := sha3.NewLegacyKeccak512()
	hash.Write(seed[:])
	copy(item[:], hash.Sum(nil))
	//cache[0] = &item
	cache[0] = &item

	/*
		fmt.Printf("buildLightCache seed : %x\n", seed) //ok
		fmt.Printf("buildLightCache item : %x\n", item) //ok
		fmt.Printf("buildLightCache cache[0] : %x\n", cache[0])
		fmt.Printf("buildLightCache *cache[0] : %x\n", *cache[0])
	*/

	for i := 1; i < numItems; i++ {
		hash.Reset()
		//hash.Write(item[:])
		//copy(item[:], hash.Sum(nil))
		//*cache[i] = item
		hash.Write(cache[i-1][:])
		newitem := hash512{}
		copy(newitem[:], hash.Sum(nil))
		cache[i] = &newitem
		//copy(cache[i][:], hash.Sum(nil))
	}
	/*
		fmt.Printf("buildLightCache cache[0] : %x\n", *cache[0])
		fmt.Printf("buildLightCache cache[42] : %x\n", *cache[42])
		fmt.Printf("buildLightCache cache[100] : %x\n", *cache[100])
	*/

	for q := 0; q < lightCacheRounds; q++ {
		for i := 0; i < numItems; i++ {
			indexLimit := uint32(numItems)
			//t := uint32(cache[i][0])
			//t := binary.BigEndian.Uint32(cache[i][0:])
			t := binary.LittleEndian.Uint32(cache[i][0:])
			v := t % indexLimit
			w := uint32(numItems+(i-1)) % indexLimit
			x := bitwiseXOR(*cache[v], *cache[w])

			if i == 0 && q == 0 {
				/*
					fmt.Printf("light_cache_rounds:%d num_items:%d index_limit:%d t:%d v:%d w:%d \n", lightCacheRounds, numItems, indexLimit, t, v, w)
					fmt.Printf("x : %x\n", x)
					fmt.Printf("buildLightCache cache[i] : %x\n", *cache[i])
					fmt.Printf("buildLightCache cache[v] : %x\n", *cache[v])
					fmt.Printf("buildLightCache cache[w] : %x\n", *cache[w])
				*/

				var result hash512
				for k := 0; k < 8; k++ {
					//binary.BigEndian.PutUint64(result[4*i:], binary.BigEndian.Uint64(x[4*i:])^binary.BigEndian.Uint64(y[4*i:]))
					binary.LittleEndian.PutUint64(result[8*k:], binary.LittleEndian.Uint64(cache[v][8*k:])^binary.LittleEndian.Uint64(cache[w][8*k:]))
					//fmt.Printf("result[4*i:]:%d cache[v][4*k:]:%d cache[w][4*k:]:%d  \n", binary.LittleEndian.Uint64(result[8*k:]), binary.LittleEndian.Uint64(cache[v][8*k:]), binary.LittleEndian.Uint64(cache[w][8*k:]))

				}

			}

			hash.Reset()
			hash.Write(x[:])
			copy(cache[i][:], hash.Sum(nil))

		}
	}
	/*
		fmt.Printf("buildLightCache cache[0] - 2 : %x\n", *cache[0])
		fmt.Printf("buildLightCache cache[42] - 2 : %x\n", *cache[42])
		fmt.Printf("buildLightCache cache[100] - 2 : %x\n", *cache[100])
	*/
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

	fmt.Printf("getContext object 0 : %x\n", lightCache[0])
	fmt.Printf("getContext object 42 : %x\n", lightCache[42])
	fmt.Printf("getContext object 100 : %x\n", lightCache[100])

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

func mapHashesToFile(hashes []hash1024, filename string) error {
	// Create or open fila
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// hash1024 table size (128 per object)
	size := len(hashes) * 128

	// file size setup
	err = file.Truncate(int64(size))
	if err != nil {
		return err
	}

	// Mapping the file in memory
	mmap, err := mmap.Map(file, mmap.RDWR, 0)
	if err != nil {
		return err
	}
	defer mmap.Unmap()

	// Copy data from memory to file
	for i, hash := range hashes {
		copy(mmap[i*128:(i+1)*128], hash[:])
	}

	// Sync data
	err = mmap.Flush()
	if err != nil {
		return err
	}

	return nil
}

func loadmappedHashesFromFile(filename string) ([]hash1024, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Mapping file in memory
	mmap, err := mmap.Map(file, mmap.RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer mmap.Unmap()

	// Get the nb of hash1028 (128bytes)
	numHashes := len(mmap) / 128
	hashes := make([]hash1024, numHashes)

	// Read data and convert in hash1024
	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], mmap[i*128:(i+1)*128])
	}

	return hashes, nil
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

	// dag file name (hardcoded for debug)
	filename := "hashes.dat"

	fmt.Printf("Verifying DAG local storage file \n")
	hashes, err := loadmappedHashesFromFile(filename)
	if err == nil {
		fmt.Printf("DAG loaded succesfully from local storage \n")
		ctx.FullDataset = hashes

		fmt.Printf("debug DAG hash[10] : %x\n", ctx.FullDataset[10])
		fmt.Printf("debug DAG hash[42] : %x\n", ctx.FullDataset[42])
		fmt.Printf("debug DAG hash[12345] : %x\n", ctx.FullDataset[12345])

		fmt.Printf("DAG context ready \n")
		ctx.ready = true
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

	fmt.Printf("getContext object 10 : %x\n", ctx.FullDataset[10])
	fmt.Printf("getContext object 42 : %x\n", ctx.FullDataset[42])
	fmt.Printf("getContext object 12345 : %x\n", ctx.FullDataset[12345])

	fmt.Printf("Saving dataset to file \n")
	//err = saveHashesToFile(ctx.FullDataset, filename)
	err = mapHashesToFile(ctx.FullDataset, filename)

	if err != nil {
		panic(err)
	}

	log.Debugf("DATASET GENERATED ===============================================\n")
	ctx.ready = true
}
