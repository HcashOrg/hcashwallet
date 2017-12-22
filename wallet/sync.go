// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wallet

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/HcashOrg/bitset"
	"github.com/HcashOrg/hcashutil/hdkeychain"
	"github.com/HcashOrg/hcashwallet/chain"
	"github.com/HcashOrg/hcashwallet/wallet/udb"
	"github.com/HcashOrg/hcashwallet/walletdb"
	"github.com/HcashOrg/hcashutil"
)

const MaxAccountForTestNet = 16

type result struct {
	used    bool
	account uint32
	acctype uint8
	err     error
}
func (w *Wallet) findLastUsedAccount(client *chain.RPCClient, coinTypeXpriv *hdkeychain.ExtendedKey) (uint32, uint32, map[uint32]result, error) {
	const scanLen = 100

	var lastRecorded  uint32
	//var err error
	err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		ns := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		var err error
		lastRecorded, err = w.Manager.LastAccount(ns)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return  0, 0, nil, err
	}

	var (
		lastUsed uint32
		lo, hi   uint32 = lastRecorded, hdkeychain.HardenedKeyStart / scanLen
	)

	requestAccount := make(map[uint32] result, 0)

Bsearch:
	for lo <= hi {
		mid := (hi + lo) / 2

		var results [scanLen]result
		var wg sync.WaitGroup
		for i := scanLen - 1; i >= 0; i-- {
			var wgs sync.WaitGroup
			i := i
			account := mid*scanLen + uint32(i)
			if account >= hdkeychain.HardenedKeyStart {
				continue
			}

			wgs.Add(2)
			usedType := 0
			go func() {
				used, err := w.newAcctAndUsed(client, coinTypeXpriv, account, udb.AcctypeEc)
				if used {
					results[i] = result{used, account, udb.AcctypeEc, err}
					usedType = 1
				}
				wgs.Done()
			}()
			go func() {
				used, err := w.newAcctAndUsed(client, coinTypeXpriv, account, udb.AcctypeBliss)
				if used {
					results[i] = result{used, account, udb.AcctypeBliss, err}
					usedType = 2
				}
				wgs.Done()
			}()

			//TODO
			/*
			coinTypeXpriv.SetAlgType(udb.AcctypeLms)
			go func() {
				used, err := w.newAcctAndUsed(client, coinTypeXpriv, account)
				results[i] = result{used, account, udb.AcctypeLms, err}
				wg.Done()
			}()
			*/
			wgs.Wait()
			if usedType == 0 {
				results[i] = result{false, account, udb.AcctypeEc, nil}
			}

		}
		wg.Wait()
		for i := scanLen - 1; i >= 0; i-- {
			if results[i].err != nil {
				return 0, 0, nil, results[i].err
			}
			if results[i].used {
				lastUsed = results[i].account
				lo = mid + 1
				continue Bsearch
			}
		}
		if mid == lastRecorded {
			break
		}
		hi = mid - 1
	}

	for i := lastRecorded + 1 ; i <= lastUsed ; i ++ {
		var wg sync.WaitGroup
		wg.Add(2)

		account := i
		usedType := 0

		var AcctResult result
		go func() {
			used, err := w.newAcctAndUsed(client, coinTypeXpriv, account, udb.AcctypeEc)
			if used {
				AcctResult = result{used, account, udb.AcctypeEc, err}
				usedType = 1
			}
			wg.Done()
		}()

		go func() {
			used, err := w.newAcctAndUsed(client, coinTypeXpriv, account, udb.AcctypeBliss)
			if used {
				AcctResult = result{used, account, udb.AcctypeBliss, err}
				usedType = 2
			}
			wg.Done()
		}()

		//TODO
		/*
		coinTypeXpriv.SetAlgType(udb.AcctypeLms)
		go func() {
			used, err := w.newAcctAndUsed(client, coinTypeXpriv, account)
			results[i] = result{used, account, udb.AcctypeLms, err}
			wg.Done()
		}()
		*/

		wg.Wait()
		if usedType == 0 {
			//return  0, 0, nil, hdkeychain.ErrUnknownAlg
			continue
		}
		requestAccount[account] = AcctResult
	}
	return lastRecorded, lastUsed, requestAccount, nil
}


//Todo: Only for testnet
func (w *Wallet) findLastUsedAccountForTest(client *chain.RPCClient, coinTypeXpriv *hdkeychain.ExtendedKey) (uint32, uint32, map[uint32]result, bool, error) {
	var lastRecorded  uint32
	var accountinfo *udb.AccountProperties
	var havebliss bool
	err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		ns := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		var err error
		lastRecorded, err = w.Manager.LastAccount(ns)
		if err != nil {
			return err
		}
		for i := uint32(0); i <= lastRecorded ; i++ {
			accountinfo, err = w.Manager.AccountProperties(ns, i)
			if err != nil {
				return err
			}
			if accountinfo.AccountType == udb.AcctypeBliss {
				havebliss = true
			}
		}


		return nil
	})
	if err != nil {
		return  0, 0, nil, false, err
	}
	var (
		lastUsed uint32
		lo, hi   uint32 = lastRecorded, MaxAccountForTestNet
	)

	requestAccount := make(map[uint32] result, 0)

Bsearch:
	for lo <= hi {
		mid := (hi + lo) / 2
		var r result
		var wgs sync.WaitGroup
		account := mid
		wgs.Add(2)
		usedType := 0
		go func() {
			used, err := w.newAcctAndUsed(client, coinTypeXpriv, account, udb.AcctypeEc)
			if used {
				r = result{used, account, udb.AcctypeEc, err}
				usedType = 1
			}
			wgs.Done()
		}()
		go func() {
			used, err := w.newAcctAndUsed(client, coinTypeXpriv, account, udb.AcctypeBliss)
			if used {
				r = result{used, account, udb.AcctypeBliss, err}
				usedType = 2
			}
			wgs.Done()
			}()
		wgs.Wait()
		if usedType == 0 {
			r = result{false, account, udb.AcctypeEc, nil}
		}

		if r.err != nil {
			return 0, 0, nil, false, r.err
		}
		if r.used {
			lastUsed = r.account
			lo = mid + 1
			continue Bsearch
		}
		if mid == lastRecorded {
			break
		}
		hi = mid - 1
	}

	for i := lastRecorded + 1 ; i <= lastUsed ; i ++ {
		var wg sync.WaitGroup
		wg.Add(2)

		account := i
		usedType := 0

		var AcctResult result
		go func() {
			used, err := w.newAcctAndUsed(client, coinTypeXpriv, account, udb.AcctypeEc)
			if used {
				AcctResult = result{used, account, udb.AcctypeEc, err}
				usedType = 1
			}
			wg.Done()
		}()

		go func() {
			used, err := w.newAcctAndUsed(client, coinTypeXpriv, account, udb.AcctypeBliss)
			if used {
				AcctResult = result{used, account, udb.AcctypeBliss, err}
				usedType = 2
				havebliss = true
			}
			wg.Done()
		}()
		wg.Wait()
		if usedType == 0 {
			//return  0, 0, nil, hdkeychain.ErrUnknownAlg
			continue
		}
		requestAccount[account] = AcctResult
	}
	return lastRecorded, lastUsed, requestAccount, havebliss, nil
}

func (w *Wallet) newAcctAndUsed(client *chain.RPCClient, coinTypeXpriv *hdkeychain.ExtendedKey, account uint32, acctype uint8) (bool, error){
	xpriv, err := coinTypeXpriv.SwitchChild(hdkeychain.HardenedKeyStart + account, acctype)
	if err != nil {
		return false, err
	}
	xpub, err := xpriv.Neuter()
	if err != nil {
		xpriv.Zero()
		return  false, err
	}

	used, err := w.accountUsed(client, xpub, xpriv)
	xpriv.Zero()
	if err != nil {
		return  false, err
	}

	return used, nil
}

func (w *Wallet) accountUsed(client *chain.RPCClient, xpub, xpriv *hdkeychain.ExtendedKey) (bool, error) {
	var err error
	var extKey, intKey, intKeypriv, extKeypriv *hdkeychain.ExtendedKey
	if xpub.GetAlgType() == udb.AcctypeEc {
		extKey, intKey, err = deriveBranches(xpub)
	} else if xpub.GetAlgType() == udb.AcctypeBliss {
		intKeypriv, err = xpriv.Child(udb.InternalBranch)
		intKey, err = intKeypriv.Neuter()
		extKeypriv, err = xpriv.Child(udb.ExternalBranch)
		extKey, err = extKeypriv.Neuter()
		xpriv.Zero()
	}

	if err != nil {
		return false, err
	}
	type result struct {
		used bool
		err  error
	}
	results := make(chan result, 2)
	merge := func(used bool, err error) {
		results <- result{used, err}
	}
	go func() { merge(w.branchUsed(client, extKey, extKeypriv)) }()
	go func() { merge(w.branchUsed(client, intKey, intKeypriv)) }()
	for i := 0; i < 2; i++ {
		r := <-results
		if r.err != nil {
			return false, err
		}
		if r.used {
			return true, nil
		}
	}
	if xpub.GetAlgType() == udb.AcctypeBliss {
		intKeypriv.Zero()
		extKeypriv.Zero()
	}
	return false, nil
}

func (w *Wallet) branchUsed(client *chain.RPCClient, branchXpub, branchXpriv *hdkeychain.ExtendedKey) (bool, error) {
	var err error
	addrs := make([]hcashutil.Address, 0, w.gapLimit)
	if branchXpub.GetAlgType() == udb.AcctypeEc {
		addrs, err = deriveChildAddresses(branchXpub, 0, uint32(2*w.gapLimit), w.chainParams)
	} else if branchXpub.GetAlgType() == udb.AcctypeBliss {
		addrs, err = deriveBlissAddresses(branchXpriv, 0, uint32(2*w.gapLimit), w.chainParams)
	}
	if err != nil {
		return false, err
	}
	existsBitsHex, err := client.ExistsAddresses(addrs)
	if err != nil {
		return false, err
	}
	for _, r := range existsBitsHex {
		if r != '0' {
			return true, nil
		}
	}
	return false, nil
}

// findLastUsedAddress returns the child index of the last used child address
// derived from a branch key.  If no addresses are found, ^uint32(0) is
// returned.
func (w *Wallet) findLastUsedAddress(client *chain.RPCClient, branchkey *hdkeychain.ExtendedKey, account, branch uint32) (uint32, error) {
	var (
		lastUsed        = ^uint32(0)
		scanLen         = uint32(w.gapLimit)
		segments        = hdkeychain.HardenedKeyStart / scanLen
		lo, hi   uint32 = 0, segments - 1
	)
Bsearch:
	for lo <= hi {
		mid := (hi + lo) / 2
		addrs := make([]hcashutil.Address, 0)
		var err error
		if branchkey.GetAlgType() == udb.AcctypeEc {
			addrs, err = deriveChildAddresses(branchkey, mid*scanLen, scanLen, w.chainParams)
		} else if branchkey.GetAlgType() == udb.AcctypeBliss {
			addrs, err = deriveBlissAddresses(branchkey, mid*scanLen, scanLen*2, w.chainParams)
		}
		if err != nil {
			return 0, err
		}
		existsBitsHex, err := client.ExistsAddresses(addrs)
		if err != nil {
			return 0, err
		}
		existsBits, err := hex.DecodeString(existsBitsHex)
		if err != nil {
			return 0, err
		}
		for i := len(addrs) - 1; i >= 0; i-- {
			if bitset.Bytes(existsBits).Get(i) {
				lastUsed = mid*scanLen + uint32(i)
				lo = mid + 1
				continue Bsearch
			}
		}
		if mid == 0 {
			break
		}
		hi = mid - 1
	}
	return lastUsed, nil
}

func (w *Wallet) FindactiveAddressesForBliss(account, branch, start, count uint32)([]hcashutil.Address, error){
	addrs := make([]hcashutil.Address, count)
	err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		var err error
		addrmgrNs := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		addrs, err = w.Manager.LoadBlissAddrs(addrmgrNs, account, branch, start, count)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return addrs, nil
}
// DiscoverActiveAddresses accesses the consensus RPC server to discover all the
// addresses that have been used by an HD keychain stemming from this wallet. If
// discoverAccts is true, used accounts will be discovered as well.  This
// feature requires the wallet to be unlocked in order to derive hardened
// account extended pubkeys.
//
// A transaction filter (re)load and rescan should be performed after discovery.
func (w *Wallet) DiscoverActiveAddresses(chainClient *chain.RPCClient, discoverAccts bool) error {
	// Start by rescanning the accounts and determining what the
	// current account index is. This scan should only ever be
	// performed if we're restoring our wallet from seed.
	if discoverAccts {
		log.Infof("Discovering used accounts")
		var coinTypePrivKey *hdkeychain.ExtendedKey
		defer func() {
			if coinTypePrivKey != nil {
				coinTypePrivKey.Zero()
			}
		}()
		err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
			var err error
			coinTypePrivKey, err = w.Manager.CoinTypePrivKey(tx)
			return err
		})
		if err != nil {
			return err
		}
		//lastRecorded, lastUsed, requestAccounts, err := w.findLastUsedAccount(chainClient, coinTypePrivKey)
		lastRecorded, lastUsed, requestAccounts, havebliss, err := w.findLastUsedAccountForTest(chainClient, coinTypePrivKey)
		if err != nil {
			return err
		}
		if lastRecorded <= lastUsed {
			acctXpubs := make(map[uint32]*hdkeychain.ExtendedKey)
			acctXprivs := make(map[uint32]*hdkeychain.ExtendedKey)
			w.addressBuffersMu.Lock()
			err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
				ns := tx.ReadWriteBucket(waddrmgrNamespaceKey)
				if !havebliss {
					lastUsed++
				}
				for acct := lastRecorded + 1; acct <= lastUsed; acct++ {
					var err error
					if acct == lastUsed  && !havebliss {
						acct, err = w.Manager.NewAccount(ns, fmt.Sprintf("postquantum"), udb.AcctypeBliss)
						if err != nil {
							return err
						}
						err = udb.PutLastAccount(ns, acct)
						if err != nil {
							return err
						}
						err = udb.CreateBlissBucket(ns)
					} else {
						if requestAccounts[acct].acctype == udb.AcctypeBliss{
							acct, err = w.Manager.NewAccount(ns, fmt.Sprintf("postquantum"), requestAccounts[acct].acctype)
						} else {
							acct, err = w.Manager.NewAccount(ns, fmt.Sprintf("account-%d", acct), requestAccounts[acct].acctype)
						}
					}
					if err != nil {
						return err
					}
					xpub, err := w.Manager.AccountExtendedPubKey(tx, acct)
					if err != nil {
						return err
					}
					acctXpubs[acct] = xpub
					if xpub.GetAlgType() == udb.AcctypeBliss {
						xpriv, err := w.Manager.AccountExtendedPrivKey(tx, acct)
						if err != nil {
							return err
						}
						acctXprivs[acct] = xpriv
					}
				}
				return nil
			})
			if err != nil {
				w.addressBuffersMu.Unlock()
				return err
			}
			for acct := lastRecorded + 1; acct <= lastUsed; acct++ {
				_, ok := w.addressBuffers[acct]
				if !ok {
					var err error
					var extKey, intKey *hdkeychain.ExtendedKey
					if acctXpubs[acct].GetAlgType() == udb.AcctypeEc {
						extKey, intKey, err = deriveBranches(acctXpubs[acct])
						if err != nil {
							w.addressBuffersMu.Unlock()
							return err
						}
					} else if acctXpubs[acct].GetAlgType() == udb.AcctypeBliss {
						intKeypriv, err := acctXprivs[acct].Child(udb.InternalBranch)
						if err != nil {
							w.addressBuffersMu.Unlock()
							return err
						}
						intKey, err = intKeypriv.Neuter()
						if err != nil {
							w.addressBuffersMu.Unlock()
							return err
						}
						extKeypriv, err := acctXprivs[acct].Child(udb.ExternalBranch)
						if err != nil {
							w.addressBuffersMu.Unlock()
							return err
						}
						extKey, err = extKeypriv.Neuter()
						if err != nil {
							w.addressBuffersMu.Unlock()
							return err
						}
						intKeypriv.Zero()
						extKeypriv.Zero()
						acctXprivs[acct].Zero()
					}
					w.addressBuffers[acct] = &bip0044AccountData{
						albExternal: addressBuffer{branchXpub: extKey},
						albInternal: addressBuffer{branchXpub: intKey},
					}
				}
			}
			w.addressBuffersMu.Unlock()
		}
	}

	var lastAcct uint32
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		ns := tx.ReadBucket(waddrmgrNamespaceKey)
		var err error
		lastAcct, err = w.Manager.LastAccount(ns)
		return err
	})
	if err != nil {
		return err
	}


	log.Infof("Discovering used addresses for %d account(s)", lastAcct+1)

	// Rescan addresses for the both the internal and external
	// branches of the account.
	errs := make(chan error, lastAcct+1)
	var wg sync.WaitGroup
	wg.Add(int(lastAcct + 1))
	for acct := uint32(0); acct <= lastAcct; acct++ {
		// Address usage discovery for each account can be performed
		// concurrently.
		acct := acct
		go func() {
			defer wg.Done()
			// Do this for both external (0) and internal (1) branches.
			for branch := uint32(0); branch < 2; branch++ {
				var branchkey *hdkeychain.ExtendedKey
				err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
					var err error
					if w.Manager.IsLocked(){
						return fmt.Errorf("wallet is locked")
					}
					branchkey, err = w.Manager.AccountBranchExtendedPubKey(tx, acct, branch)
					//TODO
					if branchkey.GetAlgType() == udb.AcctypeBliss {
						branchkey, err = w.Manager.AccountBranchExtendedPrivKey(tx, acct, branch)
					}
					return err
				})
				if err != nil {
					errs <- err
					return
				}

				lastUsed, err := w.findLastUsedAddress(chainClient, branchkey, acct, branch)
				if err != nil {
					errs <- err
					return
				}

				// Save discovered addresses for the account plus additional
				// addresses that may be used by other wallets sharing the same
				// seed.
				err = walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
					ns := tx.ReadWriteBucket(waddrmgrNamespaceKey)

					// SyncAccountToAddrIndex never removes derived addresses
					// from an account, and can be called with just the
					// discovered last used child index, plus the gap limit.
					// Cap it to the highest child index.
					//
					// If no addresses were used for this branch, lastUsed is
					// ^uint32(0) and adding the gap limit it will sync exactly
					// gapLimit number of addresses (e.g. 0-19 when the gap
					// limit is 20).
					gapLimit := uint32(w.gapLimit)
					err := w.Manager.SyncAccountToAddrIndex(ns, acct,
						minUint32(lastUsed+gapLimit, hdkeychain.HardenedKeyStart-1),
						branch)
					if err != nil {
						return err
					}
					if lastUsed < hdkeychain.HardenedKeyStart {
						err = w.Manager.MarkUsedChildIndex(tx, acct, branch, lastUsed)
						if err != nil {
							return err
						}
					}

					props, err := w.Manager.AccountProperties(ns, acct)
					if err != nil {
						return err
					}
					lastReturned := props.LastReturnedExternalIndex

					w.addressBuffersMu.Lock()
					acctData := w.addressBuffers[acct]
					buf := &acctData.albExternal
					if branch == udb.InternalBranch {
						buf = &acctData.albInternal
						lastReturned = props.LastReturnedInternalIndex
					}
					buf.lastUsed = lastUsed
					buf.cursor = lastReturned - lastUsed
					w.addressBuffersMu.Unlock()

					// Unfortunately if the cursor is equal to or greater than
					// the gap limit, the next child index isn't completely
					// known.  Depending on the gap limit policy being used, the
					// next address could be the index after the last returned
					// child or the child may wrap around to a lower value.
					log.Infof("Synchronized account %d branch %d to next child index %v",
						acct, branch, lastReturned+1)
					return nil
				})
				if err != nil {
					errs <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	select {
	case err := <-errs:
		// Drain remaining
		go func() {
			for {
				select {
				case <-errs:
				default:
					return
				}
			}
		}()
		return err
	default:
		log.Infof("Finished address discovery")
		return nil
	}
}
