package account

import (
	"ceres/pkg/initialization/mysql"
	"ceres/pkg/initialization/redis"
	"ceres/pkg/initialization/utility"
	"ceres/pkg/model/account"
	"ceres/pkg/utility/jwt"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/gotomicro/ego/core/elog"
	uuid "github.com/satori/go.uuid"
)

const (
	expire = time.Second * 240
)

/// createNonce
/// create a new nonce
func createNonce() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06v", rand.Intn(1000000000))
}

/// GenerateWeb3LoginNonce
/// generate the login nonce help frontend to sign the login signature
func GenerateWeb3LoginNonce(address string) (response *account.WalletNonceResponse, err error) {
	nonce, err := redis.Client.Get(context.TODO(), address)
	if err != nil {
		elog.Debugf("NONCE test is %s", nonce)
		return
	}
	if nonce == "" {
		nonce = createNonce()
		err = redis.Client.Set(context.TODO(), address, nonce, expire)
		if err != nil {
			return
		}
	}
	response = &account.WalletNonceResponse{Nonce: nonce}
	return
}

/// VerifyEthSignatureAndLogin
/// verify the signature and login with the eth wallet
func VerifyEthSignatureAndLogin(address []byte, message []byte, signatue []byte, walletType int) (response *account.ComerLoginResponse, err error) {

	publicKey, err := secp256k1.RecoverPubkey(message, signatue)
	if err != nil {
		return
	}

	o := hex.EncodeToString(address)
	n := hex.EncodeToString(publicKey)

	if o != n {
		err = errors.New("illegal login request because the recover failed from the signature")
		return
	}

	// origin address passed from the frontend is the 0x prefix
	// but in this logic ceres will move the 0x predix in the router to do next

	comer, err := account.GetComerByAccoutOIN(mysql.DB, o)
	if err != nil {
		elog.Error(err.Error())
		return
	}

	if comer.ID == 0 {
		// create a new comer with the origin ID
		// create comer with account
		comer.UIN = utility.AccountSequnece.Next()
		comer.ComerID = uuid.Must(uuid.NewV4(), nil).String()
		comer.Avatar = comer.ComerID
		comer.Nick = "0x" + o
		outer := &account.Account{}
		outer.OIN = o
		outer.UIN = comer.UIN
		outer.IsMain = true
		outer.IsLinked = true
		outer.Nick = comer.Nick
		outer.Avatar = comer.Avatar
		outer.Category = account.EthAccount
		outer.Type = walletType
		// Create the account and comer within transaction
		err = account.CreateComerWithAccount(mysql.DB, &comer, outer)
		if err != nil {
			elog.Errorf("Comunion Eth login faild, because of %v", err)
			return
		}
	}

	// sign with jwt
	token := jwt.Sign(comer.UIN)

	response = &account.ComerLoginResponse{
		ComerID: comer.ComerID,
		Address: comer.Address,
		Nick:    comer.Nick,
		Avatar:  comer.Avatar,
		Token:   token,
		UIN:     comer.UIN,
	}
	return
}