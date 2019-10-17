package main

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/fletaio/fleta_testnet/core/pile"
	"github.com/fletaio/testnet_explorer/explorerservice"

	"github.com/fletaio/fleta_testnet/cmd/app"
	"github.com/fletaio/fleta_testnet/cmd/closer"
	"github.com/fletaio/fleta_testnet/cmd/config"
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/common/key"
	"github.com/fletaio/fleta_testnet/common/rlog"
	"github.com/fletaio/fleta_testnet/core/backend"
	_ "github.com/fletaio/fleta_testnet/core/backend/badger_driver"
	_ "github.com/fletaio/fleta_testnet/core/backend/buntdb_driver"
	"github.com/fletaio/fleta_testnet/core/chain"
	"github.com/fletaio/fleta_testnet/core/types"
	"github.com/fletaio/fleta_testnet/pof"
	"github.com/fletaio/fleta_testnet/process/admin"
	"github.com/fletaio/fleta_testnet/process/formulator"
	"github.com/fletaio/fleta_testnet/process/gateway"
	"github.com/fletaio/fleta_testnet/process/payment"
	"github.com/fletaio/fleta_testnet/process/vault"
	"github.com/fletaio/fleta_testnet/service/apiserver"
	"github.com/fletaio/fleta_testnet/service/p2p"
)

// Config is a configuration for the cmd
type Config struct {
	SeedNodeMap    map[string]string
	NodeKeyHex     string
	ObserverKeys   []string
	Port           int
	APIPort        int
	WebPort        int
	StoreRoot      string
	BackendVersion int
	RLogHost       string
	RLogPath       string
	UseRLog        bool
}

func main() {
	var cfg Config
	if err := config.LoadFile("./config.toml", &cfg); err != nil {
		panic(err)
	}
	if len(cfg.StoreRoot) == 0 {
		cfg.StoreRoot = "./ndata"
	}
	if cfg.WebPort == 0 {
		cfg.WebPort = 80
	}
	if len(cfg.RLogHost) > 0 && cfg.UseRLog {
		if len(cfg.RLogPath) == 0 {
			cfg.RLogPath = "./ndata_rlog"
		}
		rlog.SetRLogHost(cfg.RLogHost)
		rlog.Enablelogger(cfg.RLogPath)
	}

	var ndkey key.Key
	if len(cfg.NodeKeyHex) > 0 {
		if bs, err := hex.DecodeString(cfg.NodeKeyHex); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			panic(err)
		} else {
			ndkey = Key
		}
	} else {
		if bs, err := ioutil.ReadFile("./ndkey.key"); err != nil {
			k, err := key.NewMemoryKey()
			if err != nil {
				panic(err)
			}

			fs, err := os.Create("./ndkey.key")
			if err != nil {
				panic(err)
			}
			fs.Write(k.Bytes())
			fs.Close()
			ndkey = k
		} else {
			if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
				panic(err)
			} else {
				ndkey = Key
			}
		}
	}

	ObserverKeys := []common.PublicHash{}
	for _, k := range cfg.ObserverKeys {
		pubhash, err := common.ParsePublicHash(k)
		if err != nil {
			panic(err)
		}
		ObserverKeys = append(ObserverKeys, pubhash)
	}
	SeedNodeMap := map[common.PublicHash]string{}
	for k, netAddr := range cfg.SeedNodeMap {
		pubhash, err := common.ParsePublicHash(k)
		if err != nil {
			panic(err)
		}
		SeedNodeMap[pubhash] = netAddr
	}
	SeedNodeMap[common.MustParsePublicHash("4YjmYcLVvBSmtjh4Z7frRZhWgdEAYTSABCoqqzhKEJa")] = "199.247.2.136:41000"
	SeedNodeMap[common.MustParsePublicHash("27n37VV3ebGWSNH5r9wX3ZhUwzxC2heY34UvXjizLDK")] = "199.247.0.226:41000"
	SeedNodeMap[common.MustParsePublicHash("4GzTnuP7Hky1Dye1AJMLzEXTX2a5kEka5h9AJVvZyTD")] = "80.240.18.208:41000"
	SeedNodeMap[common.MustParsePublicHash("4ew8HQEwwSqeepMDCnwN9PiYg1uvoeZXyudqdQZBCb3")] = "149.28.240.38:41000"
	SeedNodeMap[common.MustParsePublicHash("VbMwA5AwSfn93ks8HMv7vvSx4THuzfeefTWVoANEha")] = "149.28.249.207:41000"
	SeedNodeMap[common.MustParsePublicHash("8eDJ3h8DLW8RSovYUjxmcDi1QNvo7UW64MQxGZ9dnS")] = "144.202.69.204:41000"
	SeedNodeMap[common.MustParsePublicHash("3ZdKaqaCbGSQ5xmAphzVTeEF1eGzX6iU4LLGD2ox2g9")] = "45.77.59.252:41000"
	SeedNodeMap[common.MustParsePublicHash("3UHQyJwSSHHCw29fB5xiGk9W7GNf1DjGC284WhW6jpD")] = "108.61.172.231:41000"
	SeedNodeMap[common.MustParsePublicHash("v3GwqbQehcqNVYbRzDk3TDJ7yJ19DgwoamZnMJZuVg")] = "78.141.196.120:41000"
	SeedNodeMap[common.MustParsePublicHash("3HhrC3gPR951SjnxjnHpfhRSWH1iR3SbCSwtCHvTLuC")] = "45.32.173.38:41000"
	SeedNodeMap[common.MustParsePublicHash("4Ei1HSF3KtDfGrdzHCWfRf4NSTZ2oYCT1CNGFkjV1WB")] = "149.28.106.61:41000"
	SeedNodeMap[common.MustParsePublicHash("3u6v76WAknSq1j86Pfb6p31FsBAJztPdVmY1kkw4k66")] = "207.246.69.195:41000"
	SeedNodeMap[common.MustParsePublicHash("CV3cNk8UZxJcsLYjSgMdKuMf7VbDnbHXyqvb2rSE4y")] = "140.82.63.172:41000"
	SeedNodeMap[common.MustParsePublicHash("38qmoMNCuBht1ihjCKVV5nTWvfiDU7NBNeeHWhB7eT7")] = "45.77.76.27:41000"
	SeedNodeMap[common.MustParsePublicHash("25waPgmJrY3Wy3zoB8yiPA6YtJdGR9ci5mbj5vfwBTN")] = "45.63.1.207:41000"
	SeedNodeMap[common.MustParsePublicHash("EMLGsnW7RvSWTtmArG7aJuASvR7iFwg7uy59FmAwT2")] = "140.82.52.163:41000"
	SeedNodeMap[common.MustParsePublicHash("3Uo6d6w1Xrebq1j42Nm2TguHn42R5MgZTMHBwP4HfrX")] = "95.179.209.187:41000"
	SeedNodeMap[common.MustParsePublicHash("3EqB9DUVdx6Z9HW8RvbWdg5ybxSaRsdLzc5zT2d3rKE")] = "140.82.55.177:41000"

	MaxBlocksPerFormulator := uint32(10)
	ChainID := uint8(0x01)
	Name := "FLETA Testnet"
	Version := uint16(0x0001)

	cm := closer.NewManager()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		cm.CloseAll()
	}()
	defer cm.CloseAll()

	var back backend.StoreBackend
	var cdb *pile.DB
	switch cfg.BackendVersion {
	case 0:
		contextDB, err := backend.Create("badger", cfg.StoreRoot)
		if err != nil {
			panic(err)
		}
		back = contextDB
	case 1:
		contextDB, err := backend.Create("buntdb", cfg.StoreRoot+"/context")
		if err != nil {
			panic(err)
		}
		chainDB, err := pile.Open(cfg.StoreRoot + "/chain")
		if err != nil {
			panic(err)
		}
		chainDB.SetSyncMode(true)
		back = contextDB
		cdb = chainDB
	}
	st, err := chain.NewStore(back, cdb, ChainID, Name, Version)
	if err != nil {
		panic(err)
	}
	cm.Add("store", st)

	if st.Height() > 0 {
		if _, err := cdb.GetData(st.Height(), 0); err != nil {
			panic(err)
		}
	}

	cs := pof.NewConsensus(MaxBlocksPerFormulator, ObserverKeys)
	app := app.NewFletaApp()
	cn := chain.NewChain(cs, app, st)
	cn.MustAddProcess(admin.NewAdmin(1))
	cn.MustAddProcess(vault.NewVault(2))
	cn.MustAddProcess(formulator.NewFormulator(3))
	cn.MustAddProcess(gateway.NewGateway(4))
	cn.MustAddProcess(payment.NewPayment(5))
	as := apiserver.NewAPIServer()
	cn.MustAddService(as)
	e, err := explorerservice.NewBlockExplorer("_explorer", cs, cfg.WebPort)
	if err != nil {
		panic(err)
	}
	cn.MustAddService(e)
	if err := cn.Init(); err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("chain", cn)

	if err := st.IterBlockAfterContext(func(b *types.Block) error {
		if cm.IsClosed() {
			return chain.ErrStoreClosed
		}
		if err := cn.ConnectBlock(b); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if err == chain.ErrStoreClosed {
			return
		}
		panic(err)
	}

	nd := p2p.NewNode(ndkey, SeedNodeMap, cn, cfg.StoreRoot+"/peer")
	if err := nd.Init(); err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("node", nd)

	go nd.Run(":" + strconv.Itoa(cfg.Port))
	go as.Run(":" + strconv.Itoa(cfg.APIPort))

	cm.Wait()
}
