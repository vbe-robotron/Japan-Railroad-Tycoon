package core

import (
	"encoding/binary"
	"hash/fnv"
	"math/rand"
)

const (
	nanoIDAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_-"
	nanoIDLength   = 21
)

// IDGenerator は決定論的に NanoID を生成する。
//
// 同じ Seed と同じ Counter の列からは、常に同じ ID 列が得られる。
// これにより seed + アクション列 によるリプレイが可能になる。
// JSON にシリアライズして復元しても Counter が保たれるため、
// 続きから同じ ID 列を生成できる。
type IDGenerator struct {
	Seed    int64 `json:"seed"`
	Counter int64 `json:"counter"`
}

// NewIDGenerator は指定した seed の IDGenerator を生成する。
func NewIDGenerator(seed int64) *IDGenerator {
	return &IDGenerator{Seed: seed}
}

// Next は次の ID を生成する。
//
// Seed と Counter を LittleEndian で FNV-1a ハッシュし、その値をシードに
// 乱数を生成して NanoID を作る。グローバル乱数は使わないため、
// 同じ Seed と Counter の列からは常に同じ ID が得られる。
func (g *IDGenerator) Next() string {
	g.Counter++
	h := fnv.New64a()
	_ = binary.Write(h, binary.LittleEndian, g.Seed)
	_ = binary.Write(h, binary.LittleEndian, g.Counter)
	seed := int64(h.Sum64())
	r := rand.New(rand.NewSource(seed))
	return generateNanoID(r)
}

// generateNanoID は指定した乱数源から NanoID を生成する。
//
// 外部依存を持たず純粋に決定論を保つため、math/rand の局所乱数源のみを
// 使い、nanoIDAlphabet から nanoIDLength 文字を選ぶ。
func generateNanoID(r *rand.Rand) string {
	b := make([]byte, nanoIDLength)
	for i := range b {
		b[i] = nanoIDAlphabet[r.Intn(len(nanoIDAlphabet))]
	}
	return string(b)
}
