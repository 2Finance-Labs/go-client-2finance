package wallet_manager

type WalletManager struct {
}

type IWalletManager interface {
	Lock()
	Unlock()
}

func (w *WalletManager) Lock() {

}

func (w *WalletManager) Unlock() {

}
