package generation

import (
	"chunkymonkey/chunkstore"
	"chunkymonkey/generation/perlin"
	"chunkymonkey/nbt"
	. "chunkymonkey/types"
)

const SeaLevel = 64

// ChunkData implements chunkstore.IChunkReader.
type ChunkData struct {
	loc        ChunkXz
	blocks     []byte
	blockData  []byte
	blockLight []byte
	skyLight   []byte
	heightMap  []byte
}

func newChunkData(loc ChunkXz) *ChunkData {
	return &ChunkData{
		loc:        loc,
		blocks:     make([]byte, ChunkSizeH*ChunkSizeH*ChunkSizeY),
		blockData:  make([]byte, (ChunkSizeH*ChunkSizeH*ChunkSizeY)>>1),
		skyLight:   make([]byte, (ChunkSizeH*ChunkSizeH*ChunkSizeY)>>1),
		blockLight: make([]byte, (ChunkSizeH*ChunkSizeH*ChunkSizeY)>>1),
		heightMap:  make([]byte, ChunkSizeH*ChunkSizeH),
	}
}

func (data *ChunkData) ChunkLoc() *ChunkXz {
	return &data.loc
}

func (data *ChunkData) Blocks() []byte {
	return data.blocks
}

func (data *ChunkData) BlockData() []byte {
	return data.blockData
}

func (data *ChunkData) BlockLight() []byte {
	return data.blockLight
}

func (data *ChunkData) SkyLight() []byte {
	return data.skyLight
}

func (data *ChunkData) HeightMap() []byte {
	return data.heightMap
}

func (data *ChunkData) GetRootTag() *nbt.NamedTag {
	return nil
}


// TestGenerator implements chunkstore.IChunkStore.
type TestGenerator struct {
	heightSource Source
}

func NewTestGenerator(seed int64) *TestGenerator {
	perlin := perlin.NewPerlinNoise(seed)

	inputs := []Source{
		&Scale{
			Wavelength: 200,
			Amplitude:  30,
			Source:     perlin,
		},
		&Turbulence{
			Dx:     perlin,
			Dy:     &Offset{10, 0, perlin},
			Factor: 0.25,
			Source: &Mult{
				A: &Scale{
					Wavelength: 30,
					Amplitude:  20,
					Source:     perlin,
				},
				// Local steepness.
				B: &Scale{
					Wavelength: 200,
					Amplitude:  1,
					Source:     &Add{perlin, 0.6},
				},
			},
		},
		&Scale{
			Wavelength: 5,
			Amplitude:  2,
			Source:     perlin,
		},
	}

	return &TestGenerator{
		heightSource: NewSumStack(inputs),
	}
}

func (gen *TestGenerator) LoadChunk(chunkLoc *ChunkXz) (result <-chan chunkstore.ChunkResult) {
	resultChan := make(chan chunkstore.ChunkResult)
	result = resultChan
	go gen.generate(*chunkLoc, resultChan)
	return
}

func (gen *TestGenerator) generate(loc ChunkXz, result chan<- chunkstore.ChunkResult) {
	baseBlockXyz := loc.GetChunkCornerBlockXY()

	baseX, baseZ := baseBlockXyz.X, baseBlockXyz.Z

	data := newChunkData(loc)

	baseIndex := BlockIndex(0)
	heightMapIndex := 0
	for x := 0; x < ChunkSizeH; x++ {
		for z := 0; z < ChunkSizeH; z++ {
			xf, zf := float64(x)+float64(baseX), float64(z)+float64(baseZ)
			height := int(SeaLevel + gen.heightSource.At2d(xf, zf))

			if height < 0 {
				height = 0
			} else if height >= ChunkSizeY {
				height = ChunkSizeY - 1
			}

			skyLightHeight := gen.setBlockStack(
				height,
				data.blocks[baseIndex:baseIndex+ChunkSizeY])

			lightBase := baseIndex >> 1

			gen.setSkyLightStack(
				skyLightHeight,
				data.blocks[baseIndex:baseIndex+ChunkSizeY],
				data.skyLight[lightBase:lightBase+ChunkSizeY/2])

			data.heightMap[heightMapIndex] = byte(skyLightHeight)

			heightMapIndex++
			baseIndex += ChunkSizeY
		}
	}

	result <- chunkstore.ChunkResult{
		Reader: data,
		Err:    nil,
	}
}

func (gen *TestGenerator) setBlockStack(height int, blocks []byte) (skyLightHeight int) {
	var topBlockType byte
	if height < SeaLevel+1 {
		skyLightHeight = SeaLevel + 1

		for y := SeaLevel; y > height; y-- {
			blocks[y] = 9 // stationary water
		}
		blocks[height] = 12 // sand
		topBlockType = 12
	} else {

		if height <= SeaLevel+1 {
			blocks[height] = 12 // sand
			topBlockType = 12
		} else {
			blocks[height] = 2 // grass
			topBlockType = 3   // dirt
		}
	}

	for y := height - 1; y > height-3 && y > 0; y-- {
		blocks[y] = topBlockType // dirt
	}
	for y := height - 3; y > 0; y-- {
		blocks[y] = 1 // stone
	}

	if skyLightHeight < 0 {
		skyLightHeight = 0
	}

	return
}

func (gen *TestGenerator) setSkyLightStack(skyLightHeight int, blocks []byte, skyLight []byte) {
	for y := ChunkSizeY - 1; y >= skyLightHeight; y-- {
		BlockIndex(y).SetBlockData(skyLight, 15)
	}

	lightLevel := 15

	for y := skyLightHeight; y >= 0 && lightLevel > 0; y-- {
		// TODO Use real block data in here.
		if blocks[y] == 9 {
			lightLevel -= 3
		} else if blocks[y] == 0 {
			// air
		} else {
			lightLevel -= 15
		}
		if lightLevel < 0 {
			lightLevel = 0
		}

		BlockIndex(y).SetBlockData(skyLight, byte(lightLevel))
	}
}
