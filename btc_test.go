package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/PowPool/btcpool/util"
)

// BlockHeader 定义区块头结构体
type BlockHeader struct {
	Version           int32
	PreviousBlockHash string
	MerkleRoot        string
	Timestamp         uint32
	Bits              uint32
	Nonce             uint32
}

// Serialize 将区块头序列化为字节数组
func (header *BlockHeader) Serialize() []byte {
	var buffer bytes.Buffer

	// 将各字段转换为小端字节序并写入buffer
	binary.Write(&buffer, binary.LittleEndian, header.Version)
	prevBlockHashBytes, _ := hex.DecodeString(header.PreviousBlockHash)
	buffer.Write(reverseBytes(prevBlockHashBytes))
	merkleRootBytes, _ := hex.DecodeString(header.MerkleRoot)
	buffer.Write(reverseBytes(merkleRootBytes))
	binary.Write(&buffer, binary.LittleEndian, header.Timestamp)
	binary.Write(&buffer, binary.LittleEndian, header.Bits)
	binary.Write(&buffer, binary.LittleEndian, header.Nonce)

	return buffer.Bytes()
}

// Hash 计算区块头的哈希值
func (header *BlockHeader) Hash() []byte {
	serializedHeader := header.Serialize()
	firstHash := sha256.Sum256(serializedHeader)
	secondHash := sha256.Sum256(firstHash[:])

	// 将哈希结果进行字节反转
	return reverseBytes(secondHash[:])
}

// reverseBytes 用于反转字节数组
func reverseBytes(data []byte) []byte {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
	return data
}

func TestHeaderHash(t *testing.T) {
	header := BlockHeader{
		Version:           539230468,
		PreviousBlockHash: "00000000000000db08c153a905b7ac11e0426ce40faaf1e4d91048bdb98bec87",
		MerkleRoot:        "17198d1474e8965345fb41fa40c961601980a5a1c470cfb7d576c9d236a9e8ac",
		Timestamp:         1724043472,
		Bits:              0x1a017dbf,
		Nonce:             0,
	}

	hash := header.Hash()
	fmt.Println("Block Hash:", hex.EncodeToString(hash))

	// 计算并打印难度值
	hashDiff := util.TargetHexToDiff(hex.EncodeToString(hash))
	fmt.Printf("Hash Difficulty: %s\n", hashDiff.String())
}

func TestHeaderHash2(t *testing.T) {
	// 00000000000000000002ae245707479783cc138d67cccdca954731d904883128
	header := BlockHeader{
		Version:           537157632,
		PreviousBlockHash: "00000000000000000000944279ce2f639b5b35227cfe605b1bd10c39d1f03196",
		MerkleRoot:        "434308e7317232bc97921e523db57c3a59d2f56398b2eb646f1bbe482d46a12a",
		Timestamp:         1725426384,
		Bits:              0x1703255b,
		Nonce:             1145160285,
	}

	hash := header.Hash()
	fmt.Println("Block Hash:", hex.EncodeToString(hash))

	// 计算并打印难度值
	hashDiff := util.TargetHexToDiff(hex.EncodeToString(hash))
	fmt.Printf("Hash Difficulty: %s\n", hashDiff.String())

	difficulty := bitsToDiff("1703255b")
	fmt.Printf("difficulty: %f\n", difficulty)
	//89471664776970.77000000
	//451050993881657845211416
}

// bitsToDiff 计算难度
func bitsToDiff(bits string) float64 {
	// 解析 bits
	bitsBytes, _ := hex.DecodeString(bits)
	exponent := bitsBytes[0]
	coefficient := new(big.Int).SetBytes(bitsBytes[1:])

	// 计算目标值
	target := new(big.Int).Mul(coefficient, new(big.Int).Exp(big.NewInt(256), big.NewInt(int64(exponent-3)), nil))

	// 最大目标值
	maxTarget := new(big.Int)
	maxTarget.SetString("00000000FFFF0000000000000000000000000000000000000000000000000000", 16)

	// 计算难度
	difficulty := new(big.Float).Quo(new(big.Float).SetInt(maxTarget), new(big.Float).SetInt(target))

	// 将难度转换为浮点数
	difficultyFloat, _ := difficulty.Float64()
	return difficultyFloat
}

// calculateNVersion computes the nVersion based on initialVersion, versionMask, and versionBits.
func calculateNVersion(initialVersion, versionMask, versionBits uint32) uint32 {
	//return (initialVersion & ^versionMask) | (versionBits & versionMask)
	return (initialVersion & ^versionMask) | (versionBits & versionMask)
}

func TestCalculateNVersion(t *testing.T) {
	// Test data
	initialVersion := uint32(0x20000004)
	versionMask := uint32(0x1fffe000)
	versionBits := uint32(0x1fff0000)
	expectedNVersion := uint32(0x3fffe004)

	// Call the function
	nVersion := calculateNVersion(initialVersion, versionMask, versionBits)
	fmt.Println(nVersion)

	// Check if the result matches the expected value
	if nVersion != expectedNVersion {
		t.Errorf("Expected nVersion to be 0x%x, but got 0x%x", expectedNVersion, nVersion)
	}
}
