package util

import (
	"fmt"
	"math/big"
	"testing"
)

func TestTargetHash256StratumFormat(t *testing.T) {
	//d37214c56841e40e6e7b504605d83c034011f715279b3d500000001600000000
	blockHashHex := "0000000000000016279b3d504011f71505d83c036e7b50466841e40ed37214c5"
	blockHashStratumHex, _ := TargetHash256StratumFormat(blockHashHex)
	fmt.Println("blockHashStratumHex:", blockHashStratumHex)
}

func TestHash256StratumFormat(t *testing.T) {
	//533eddad5a8998679ce9d8ec8692b9a37ecff0fb0ca5f028e6ac85a0f6a38b87
	txIdHex := "878ba3f6a085ace628f0a50cfbf0cf7ea3b99286ecd8e99c6798895aaddd3e53"
	txIdStratumHex, _ := Hash256StratumFormat(txIdHex)
	fmt.Println("txIdStratumHex:", txIdStratumHex)
}

func TestIsValidBTCAddress(t *testing.T) {
	//bc1qkyz0zhxe6aktl35uzp6rwkr2aht4wrnpvlvr03
	//bc1pzf9qwcjr0c4j0p97yrt2mmqehsarfkc98czxsqauwflcv95h37rqn0tfwz
	address := "bc1pzf9qwcjr0c4j0p97yrt2mmqehsarfkc98czxsqauwflcv95h37rqn0tfwz"
	ok := IsValidBTCAddress(address)
	fmt.Println("IsValidBTCAddress:", ok)
}

func TestDifficultyHash(t *testing.T) {

	/*
	   [D] 2024/09/09 16:55:49.131540 miner.go:263: Target Hex: 00000000bceb08d48adc84b2fba2072e0babda860c1d79ce3216a27058c1af04
	   [D] 2024/09/09 16:55:49.131540 miner.go:267: hashDiff: 5820043751
	   [D] 2024/09/09 16:55:49.132229 miner.go:247: blockHeader.Version: 536870916
	*/

	// 假设你已经得到了区块的哈希值
	blockHash := "000000000000000000014175a87189eb28e429220e3e5492640b3b563c0749a8"
	//962749310902516297335596
	//89473030027062.73731999
	//89471664776971
	// 计算区块哈希值对应的难度
	hashDiff := TargetHexToDiff(blockHash)

	fmt.Println(hashDiff)

	target := nBitsToTarget(uint32(386082139))
	difficulty := targetToDifficulty(target)

	fmt.Printf("Target: %s\n", target.Text(16))
	fmt.Printf("Difficulty: %s\n", difficulty.Text('f', 8))

}

// 创世区块的目标难度
var genesisTarget = new(big.Int).Exp(big.NewInt(2), big.NewInt(224), nil)

// 从 nBits 计算目标难度
func nBitsToTarget(nBits uint32) *big.Int {
	exponent := nBits >> 24
	coefficient := nBits & 0xFFFFFF

	target := new(big.Int).Lsh(big.NewInt(int64(coefficient)), uint(8*(exponent-3)))
	return target
}

// 从目标难度计算难度
func targetToDifficulty(target *big.Int) *big.Float {
	difficulty := new(big.Float).Quo(new(big.Float).SetInt(genesisTarget), new(big.Float).SetInt(target))
	return difficulty
}
