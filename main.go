package main

import (
	"fmt"
	"log"
	"os"

	"github.com/iwat/etherezor/internal/etherscan"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/usbwallet"
	"github.com/howeyc/gopass"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env", ".env.local")

	hub, err := usbwallet.NewTrezorHub()
	if err != nil {
		log.Fatalln("newhub:", err)
	}

	log.Printf("newhub: Got TrezorHub %p", hub)

	sink := make(chan accounts.WalletEvent)
	subscription := hub.Subscribe(sink)

	listWallets(hub)

	go func() {
		for e := range sink {
			log.Println("event:", e, walletEventTypeString(e.Kind))
			switch e.Kind {
			case accounts.WalletArrived:
				openWallet(e.Wallet)
			case accounts.WalletOpened:
				log.Println("accounts:", e.Wallet.Accounts())

				a, err := e.Wallet.Derive(accounts.DefaultBaseDerivationPath, true)
				if err != nil {
					log.Println("derive:", err)
				} else {
					log.Println("derive:", a.Address.Hex())

					eth, err := etherscan.BalanceOf(a.Address.Hex())
					if err != nil {
						log.Println("balance:", err)
					} else {
						log.Println("balance:", fromWei(eth, "ether"), "ETH")
					}
				}

				if err := e.Wallet.Close(); err != nil {
					log.Println("close:", err)
				}
			}
		}
	}()

	for {
		for err := range subscription.Err() {
			log.Println("subscription:", err)
			subscription.Unsubscribe()
		}
	}
}

func openWallet(w accounts.Wallet) {
	status, err := w.Status()
	if err != nil {
		log.Println("list:", w.URL().TerminalString(), err)
		return
	}

	log.Println("list:", w.URL().TerminalString(), status)

	pass := ""
	for i := 0; i < 3; i++ {
		log.Println("open:", w.URL().TerminalString(), pass)
		err := w.Open(pass)
		if err == usbwallet.ErrTrezorPINNeeded {
			i--
			log.Println("7|8|9")
			log.Println("4|5|6")
			log.Println("1|2|3")
			p, e := gopass.GetPasswdPrompt(err.Error()+": ", true, os.Stdin, os.Stdout)
			if e != nil {
				log.Println("pass:", e)
				break
			}

			pass = string(p)
			continue
		} else if err != nil {
			log.Println("open:", err)
			pass = ""
			continue
		}
		break
	}

	status, err = w.Status()
	if err != nil {
		log.Println("list:", w.URL().TerminalString(), err)
		return
	}

	log.Println("list:", w.URL().TerminalString(), status)
}

func listWallets(h *usbwallet.Hub) {
	for _, w := range h.Wallets() {
		openWallet(w)
	}
}

func walletEventTypeString(e accounts.WalletEventType) string {
	switch e {
	case accounts.WalletArrived:
		return "Arrived"
	case accounts.WalletOpened:
		return "Opened"
	case accounts.WalletDropped:
		return "Dropped"
	default:
		return fmt.Sprintf("unknown %d", int(e))
	}
}
