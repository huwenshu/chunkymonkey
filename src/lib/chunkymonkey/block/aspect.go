package block

import (
	"rand"

	"chunkymonkey/item"
	"chunkymonkey/itemtype"
	. "chunkymonkey/types"
)

// The distance from the edge of a block that items spawn at in fractional
// blocks.
const blockItemSpawnFromEdge = 4.0 / PixelsPerBlock

// IBlockPlayer defines the interactions that a block aspect may have upon a
// player.
type IBlockPlayer interface {
	OpenWindow(invTypeId InvTypeId)
}

// The interface required of a chunk by block behaviour.
type IChunkBlock interface {
	GetRand() *rand.Rand
	GetItemType(itemTypeId ItemTypeId) (itemType *itemtype.ItemType, ok bool)
	AddItem(item *item.Item)
}

// BlockInstance represents the instance of a block within a chunk.
type BlockInstance struct {
	Chunk    IChunkBlock
	BlockLoc BlockXyz
	SubLoc   SubChunkXyz
	// TODO decide if *BlockType belongs in here as well.
	// Note that only the lower nibble of data is stored.
	Data byte
}

// Defines the behaviour of a block.
type IBlockAspect interface {
	Name() string
	Hit(instance *BlockInstance, player IBlockPlayer, digStatus DigStatus) (destroyed bool)
	Interact(instance *BlockInstance, player IBlockPlayer)
}
