// Copyright 2021 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

//go:build integration
// +build integration

package dot

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ChainSafe/gossamer/dot/core"
	"github.com/ChainSafe/gossamer/dot/state"
	"github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/internal/log"
	"github.com/ChainSafe/gossamer/lib/babe"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/crypto/sr25519"
	"github.com/ChainSafe/gossamer/lib/genesis"
	"github.com/ChainSafe/gossamer/lib/grandpa"
	"github.com/ChainSafe/gossamer/lib/keystore"
	"github.com/ChainSafe/gossamer/lib/runtime"
	"github.com/ChainSafe/gossamer/lib/runtime/wasmer"
	"github.com/ChainSafe/gossamer/lib/services"
	"github.com/ChainSafe/gossamer/lib/trie"
	"github.com/ChainSafe/gossamer/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestInitNode_Integration(t *testing.T) {
	cfg := NewTestConfig(t)
	require.NotNil(t, cfg)

	genFile := newTestGenesisRawFile(t, cfg)

	cfg.Init.Genesis = genFile

	err := InitNode(cfg)
	require.NoError(t, err)

	// confirm database was setup
	db, err := utils.SetupDatabase(cfg.Global.BasePath, false)
	require.NoError(t, err)
	require.NotNil(t, db)
}

func TestInitNode_GenesisSpec(t *testing.T) {
	cfg := NewTestConfig(t)
	require.NotNil(t, cfg)

	genFile := newTestGenesisFile(t, cfg)

	cfg.Init.Genesis = genFile

	err := InitNode(cfg)
	require.NoError(t, err)
	// confirm database was setup
	db, err := utils.SetupDatabase(cfg.Global.BasePath, false)
	require.NoError(t, err)
	require.NotNil(t, db)
}

func TestNodeInitializedIntegration(t *testing.T) {
	cfg := NewTestConfig(t)
	require.NotNil(t, cfg)

	genFile := newTestGenesisRawFile(t, cfg)

	cfg.Init.Genesis = genFile

	result := NodeInitialized(cfg.Global.BasePath)
	require.False(t, result)

	err := InitNode(cfg)
	require.NoError(t, err)

	result = NodeInitialized(cfg.Global.BasePath)
	require.True(t, result)
}

func TestNewNodeIntegration(t *testing.T) {
	cfg := NewTestConfig(t)
	require.NotNil(t, cfg)

	genFile := newTestGenesisRawFile(t, cfg)

	cfg.Init.Genesis = genFile

	err := InitNode(cfg)
	require.NoError(t, err)

	ks := keystore.NewGlobalKeystore()
	err = keystore.LoadKeystore("alice", ks.Gran)
	require.NoError(t, err)
	err = keystore.LoadKeystore("alice", ks.Babe)
	require.NoError(t, err)

	cfg.Core.Roles = types.FullNodeRole

	node, err := NewNode(cfg, ks)
	require.NoError(t, err)

	bp := node.Services.Get(&babe.Service{})
	require.IsType(t, &babe.Service{}, bp)
	fg := node.Services.Get(&grandpa.Service{})
	require.IsType(t, &grandpa.Service{}, fg)
}

func TestNewNode_Authority(t *testing.T) {
	cfg := NewTestConfig(t)
	require.NotNil(t, cfg)

	genFile := newTestGenesisRawFile(t, cfg)

	cfg.Init.Genesis = genFile

	err := InitNode(cfg)
	require.NoError(t, err)

	ks := keystore.NewGlobalKeystore()
	err = keystore.LoadKeystore("alice", ks.Gran)
	require.NoError(t, err)
	require.Equal(t, 1, ks.Gran.Size())
	err = keystore.LoadKeystore("alice", ks.Babe)
	require.NoError(t, err)
	require.Equal(t, 1, ks.Babe.Size())

	cfg.Core.Roles = types.AuthorityRole

	node, err := NewNode(cfg, ks)
	require.NoError(t, err)

	bp := node.Services.Get(&babe.Service{})
	require.NotNil(t, bp)
	fg := node.Services.Get(&grandpa.Service{})
	require.NotNil(t, fg)
}

func TestStartNode(t *testing.T) {
	cfg := NewTestConfig(t)
	require.NotNil(t, cfg)

	genFile := newTestGenesisRawFile(t, cfg)

	cfg.Init.Genesis = genFile
	cfg.Core.GrandpaAuthority = false

	err := InitNode(cfg)
	require.NoError(t, err)

	ks := keystore.NewGlobalKeystore()
	err = keystore.LoadKeystore("alice", ks.Gran)
	require.NoError(t, err)
	err = keystore.LoadKeystore("alice", ks.Babe)
	require.NoError(t, err)

	cfg.Core.Roles = types.FullNodeRole

	node, err := NewNode(cfg, ks)
	require.NoError(t, err)

	err = node.Start()
	require.NoError(t, err)
	// TODO (ed): this isn't getting started signal, and is hanging here.
	<-node.started
	node.Stop()
}

func TestInitNode_LoadGenesisData(t *testing.T) {
	// TODO (ed): this is failing with sigabrt
	cfg := NewTestConfig(t)
	require.NotNil(t, cfg)

	genPath := newTestGenesisAndRuntime(t)

	cfg.Init.Genesis = genPath
	cfg.Core.GrandpaAuthority = false

	err := InitNode(cfg)
	require.NoError(t, err)

	config := state.Config{
		Path: cfg.Global.BasePath,
	}
	stateSrvc := state.NewService(config)

	gen, err := genesis.NewGenesisFromJSONRaw(genPath)
	require.NoError(t, err)

	genTrie, err := genesis.NewTrieFromGenesis(gen)
	require.NoError(t, err)

	genesisHeader, err := types.NewHeader(common.NewHash([]byte{0}),
		genTrie.MustHash(), trie.EmptyHash, big.NewInt(0), types.NewDigest())
	require.NoError(t, err)

	err = stateSrvc.Initialise(gen, genesisHeader, genTrie)
	require.NoError(t, err)

	err = stateSrvc.Start()
	require.NoError(t, err)

	t.Cleanup(func() {
		err = stateSrvc.Stop()
		require.NoError(t, err)
	})

	gendata, err := stateSrvc.BaseState().LoadGenesisData()
	require.NoError(t, err)

	testGenesis := newTestGenesis(t)

	expected := &genesis.Data{
		Name:       testGenesis.Name,
		ID:         testGenesis.ID,
		Bootnodes:  common.StringArrayToBytes(testGenesis.Bootnodes),
		ProtocolID: testGenesis.ProtocolID,
	}
	require.Equal(t, expected, gendata)

	genesisHeader, err = stateSrvc.BlockState().BestBlockHeader()
	require.NoError(t, err)

	stateRoot := genesisHeader.StateRoot
	expectedHeader, err := types.NewHeader(common.NewHash([]byte{0}),
		stateRoot, trie.EmptyHash, big.NewInt(0), types.NewDigest())
	require.NoError(t, err)
	require.Equal(t, expectedHeader.Hash(), genesisHeader.Hash())
}

func TestInitNode_LoadStorageRoot(t *testing.T) {
	cfg := NewTestConfig(t)
	require.NotNil(t, cfg)

	genPath := newTestGenesisAndRuntime(t)

	cfg.Core.Roles = types.FullNodeRole
	cfg.Core.BabeAuthority = false
	cfg.Core.GrandpaAuthority = false
	cfg.Init.Genesis = genPath

	gen, err := genesis.NewGenesisFromJSONRaw(genPath)
	require.NoError(t, err)

	err = InitNode(cfg)
	require.NoError(t, err)

	ks := keystore.NewGlobalKeystore()
	ed25519Keyring, _ := keystore.NewEd25519Keyring()
	ks.Gran.Insert(ed25519Keyring.Alice())
	sr25519Keyring, _ := keystore.NewSr25519Keyring()
	ks.Babe.Insert(sr25519Keyring.Alice())
	node, err := NewNode(cfg, ks)
	require.NoError(t, err)

	expected := &trie.Trie{}
	err = expected.LoadFromMap(gen.GenesisFields().Raw["top"])
	require.NoError(t, err)

	expectedRoot, err := expected.Hash()
	require.NoError(t, err)

	coreServiceInterface := node.Services.Get(&core.Service{})

	coreSrvc, ok := coreServiceInterface.(*core.Service)
	require.True(t, ok, "could not find core service")
	require.NotNil(t, coreSrvc)

	stateRoot, err := coreSrvc.StorageRoot()
	require.NoError(t, err)
	require.Equal(t, expectedRoot, stateRoot)
}

func balanceKey(t *testing.T, publicKey [32]byte) (storageTrieKey []byte) {
	accountKey := append([]byte("balance:"), publicKey[:]...)
	hash, err := common.Blake2bHash(accountKey)
	require.NoError(t, err)
	return hash[:]
}

func TestInitNode_LoadBalances(t *testing.T) {
	cfg := NewTestConfig(t)
	require.NotNil(t, cfg)

	genPath := newTestGenesisAndRuntime(t)

	cfg.Core.Roles = types.FullNodeRole
	cfg.Core.BabeAuthority = false
	cfg.Core.GrandpaAuthority = false
	cfg.Init.Genesis = genPath

	err := InitNode(cfg)
	require.NoError(t, err)

	ks := keystore.NewGlobalKeystore()
	ed25519Keyring, _ := keystore.NewEd25519Keyring()
	ks.Gran.Insert(ed25519Keyring.Alice())

	stateConfig := state.Config{LogLevel: log.Critical}
	stateService := state.NewService(stateConfig)
	node, err := NewNode(cfg, ks)
	require.NoError(t, err)
	mgr := node.Services.Get(stateService)

	stateSrv, ok := mgr.(state.Service)
	require.True(t, ok, "could not find core service")
	require.NotNil(t, stateSrv)

	kr, _ := keystore.NewSr25519Keyring()
	alice := kr.Alice().Public().(*sr25519.PublicKey).AsBytes()

	bal, err := stateSrv.StorageState().GetStorage(nil, balanceKey(t, alice))
	require.NoError(t, err)

	const genesisBalance = "0x0000000000000001"
	expected, err := common.HexToBytes(genesisBalance)
	require.NoError(t, err)
	require.Equal(t, expected, bal)
}

func TestNode_StopFunc(t *testing.T) {
	// TODO (ed): fix this
	t.Skip()
	testvar := "before"

	node := &Node{
		Services: services.NewServiceRegistry(log.New(log.SetWriter(io.Discard))),
		wg:       sync.WaitGroup{},
	}
	node.wg.Add(1)

	node.Stop()
	require.Equal(t, testvar, "after")
}

func TestNode_PersistGlobalName_WhenInitialize(t *testing.T) {
	globalName := RandomNodeName()

	cfg := NewTestConfig(t)
	cfg.Global.Name = globalName

	cfg.Core.Roles = types.FullNodeRole
	cfg.Core.BabeAuthority = false
	cfg.Core.GrandpaAuthority = false
	cfg.Init.Genesis = newTestGenesisAndRuntime(t)

	err := InitNode(cfg)
	require.NoError(t, err)

	storedName, err := LoadGlobalNodeName(cfg.Global.BasePath)
	require.NoError(t, err)
	require.Equal(t, globalName, storedName)
}

// newTestGenesisAndRuntime create a new test runtime and a new test genesis
// file with the test runtime stored in raw data and returns the genesis file
func newTestGenesisAndRuntime(t *testing.T) (filename string) {
	_ = wasmer.NewTestInstance(t, runtime.NODE_RUNTIME)
	runtimeFilePath := runtime.GetAbsolutePath(runtime.NODE_RUNTIME_FP)

	runtimeData, err := os.ReadFile(filepath.Clean(runtimeFilePath))
	require.Nil(t, err)

	gen := NewTestGenesis(t)
	hex := hex.EncodeToString(runtimeData)

	gen.Genesis.Raw = map[string]map[string]string{}
	if gen.Genesis.Raw["top"] == nil {
		gen.Genesis.Raw["top"] = make(map[string]string)
	}
	gen.Genesis.Raw["top"]["0x3a636f6465"] = "0x" + hex
	gen.Genesis.Raw["top"]["0xcf722c0832b5231d35e29f319ff27389f5032bfc7bfc3ba5ed7839f2042fb99f"] = "0x0000000000000001"

	genData, err := json.Marshal(gen)
	require.NoError(t, err)

	filename = filepath.Join(t.TempDir(), "genesis.json")
	err = os.WriteFile(filename, genData, os.ModePerm)
	require.NoError(t, err)

	return filename
}
