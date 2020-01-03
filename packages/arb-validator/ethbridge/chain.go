/*
 * Copyright 2019, Offchain Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ethbridge

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ArbAddresses struct {
	ChainFactory       string `json:"ChainFactory"`
	ChannelFactory     string `json:"ChannelFactory"`
	GlobalPendingInbox string `json:"GlobalPendingInbox"`
	OneStepProof       string `json:"OneStepProof"`
}

func waitForReceipt(ctx context.Context, client *ethclient.Client, auth *bind.TransactOpts, tx *types.Transaction, methodName string) (*types.Receipt, error) {
	for {
		select {
		case _ = <-time.After(time.Second):
			receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
			if err != nil {
				if err.Error() == ethereum.NotFound.Error() {
					continue
				}
				return nil, err
			}
			if receipt.Status != 1 {
				data, err := receipt.MarshalJSON()
				if err != nil {
					return nil, errors.New("Failed unmarshalling receipt")
				}
				callMsg := ethereum.CallMsg{
					From:     auth.From,
					To:       tx.To(),
					Gas:      tx.Gas(),
					GasPrice: tx.GasPrice(),
					Value:    tx.Value(),
					Data:     tx.Data(),
				}
				_, err = client.CallContract(ctx, callMsg, receipt.BlockNumber)
				if err != nil {
					return nil, fmt.Errorf("Transaction %v failed with error %v", methodName, err)
				}
				return nil, fmt.Errorf("Transaction %v failed with tx %v", methodName, string(data))
			}
			return receipt, nil
		case _ = <-ctx.Done():
			return nil, errors.New("Receipt not found")
		}
	}
}