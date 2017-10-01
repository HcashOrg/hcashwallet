# hcashwallet


Hcashwallet is a daemon handling hcash wallet functionality for a single user.  

It acts as both an RPC client to hcashd and an RPC server for wallet clients and legacy RPC applications. It manages all of your accounts, addresses, and transactions and allows users to participate in Proof-of-Stake voting.

Hcashwallet is not an SPV client and requires connecting to a local or remote hcashd instance for asynchronous blockchain queries and notifications over websockets. Full hcashd installation instructions can be found [here](https://github.com/HcashOrg/hcashd).

## Build and Installation

The installation of hcashwallet requires Go 1.7 or newer.
* Glide
	
	Glide is used to manage project dependencies and provide reproducible builds. To install:
	```
	go get -u github.com/Masterminds/glide
	```
* Build and Installation
	
	For a first time installation, the project and dependency sources can be obtained manually with git and glide (create directories as needed):
	```
	git clone https://github.com/HcashOrg/hcashwallet $GOPATH/src/github.com/HcashOrg/hcashwallet
	cd $GOPATH/src/github.com/HcashOrg/hcashwallet
	glide install
	go install $(glide nv)
	```
    To update an existing source tree, pull the latest changes and install the matching dependencies:
    ```
	cd $GOPATH/src/github.com/HcashOrg/hcashwallet
	git pull
	glide install
	go install $(glide nv)
    ```

## Getting started

The following instructions detail how to get started with hcashwallet connecting to a localhost hcashd. The command should be run in the console.

* Creating

The folloing instruction detail how to create a new wallet when first starting the wallet. (PS: This instruction is only for initailize your new wallet, thus it's not recommanded when you already had wallet)

```
hcashwallet --create
```

During this process, you’ll set a private passphrase, optionally set a public passphrase, and record your seed. 

* Configuring

After creating the wallet for the first time, it's necessary to configure your wallet before launching. It's recommanded to copy the sample hcashd and hcashwallet configurations and update with your RPC username and password.

```
$ cp $GOPATH/src/github.com/HcashOrg/hcashd/sample-hcashd.conf ~/.hcashd/hcashd.conf
$ cp $GOPATH/src/github.com/HcashOrg/hcashwallet/sample-hcashwallet.conf ~/.hcashwallet/hcashwallet.conf
```
After copy the sample configuration file to working directory, you need to update your RPC username and password. In addtion, if you want to participate the PoS consensus process, you need to purchase ticket and set the following parameter in your hcashwallet.conf.
```
enablevoting=1
```  
The detailed information of configuration will be released soon.

* Lauching

Before launching hcashwallet, it's necessary to start hcashd first. Detailed information can be found [here](https://github.com/HcashOrg/hcashd). 

```
hcashwallet
```
If already set enable voting before , you need to type in your private phrase later.

You can run hcashctl.exe and type in the following common instructions to gain detailed inforamtion of your current state. 

PS: Hcashctl will be installed with the installation of hcashd.
```
hcashctl getinfo    //Displays the basic info about the network including current block number and network difficulty.
hcashctl --wallet getnewaddress   //Get a new address in the given account.
hcashctl --wallet getbalance      //Get the spendable balance in the given account. 
hcashctl --wallet getstakeinfo    //Get info about the current status of the PoS pool. 
hcashctl --wallet sendtoaddress "[address]" [amount]  //Send hcash from your account to the wanted address
hcashctl --wallet purchaseticket "[fromaccount]" spendlimit minconf "ticketaddress" "[numtickets]")   
//Purchase tickets to participate in PoS process. Spendlimit denotes the limit on the amount to spend on ticket, minconf denotes the minimal required confirmation of the transaction(e.g. 1)
```

## License

hcashwallet is licensed under the [copyfree](http://copyfree.org) ISC License.