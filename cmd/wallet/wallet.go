/*
Copyright © 2022 Polygon <engineering@polygon.technology>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package wallet

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/maticnetwork/polygon-cli/hdwallet"
	"github.com/spf13/cobra"
)

var (
	inputWords               *int
	inputLang                *string
	inputKDFIterations       *uint
	inputPassword            *string
	inputPasswordFile        *string
	inputMnemonic            *string
	inputMnemonicFile        *string
	inputPath                *string
	inputAddressesToGenerate *uint
	inputUseRawEntropy       *bool
	inputRootOnly            *bool
)

// WalletCmd represents the wallet command
var WalletCmd = &cobra.Command{
	Use:   "wallet [create|inspect]",
	Short: "Create or inspect a bip39(ish) wallet",
	Long: `This command is meant to simplify the operations of creating wallets
across v1 and avail. This command can take a seed phrase and spit out
child accounts or generate new accmounts along with a seed phrase`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mode := args[0]
		var err error
		var mnemonic string
		if mode == "inspect" {
			// in the case of inspect, we'll partse a mnemonic and then continue
			mnemonic, err = getFileOrFlag(inputMnemonicFile, inputMnemonic)
			if err != nil {
				return err
			}
		} else {
			mnemonic, err = hdwallet.NewMnemonic(*inputWords, *inputLang)
			if err != nil {
				return err
			}
		}
		// mnemonic = "maid palace spring laptop shed when text taxi pupil movie athlete tag"
		// mnemonic = "crop cash unable insane eight faith inflict route frame loud box vibrant"
		// mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
		password, err := getFileOrFlag(inputPasswordFile, inputPassword)
		if err != nil {
			return err
		}
		pw, err := hdwallet.NewPolyWallet(mnemonic, password)
		if err != nil {
			return err
		}
		err = pw.SetPath(*inputPath)
		if err != nil {
			return err
		}
		err = pw.SetIterations(*inputKDFIterations)
		if err != nil {
			return err
		}
		err = pw.SetUseRawEntropy(*inputUseRawEntropy)
		if err != nil {
			return err
		}

		if *inputRootOnly {
			var key *hdwallet.PolyWalletExport
			key, err = pw.ExportRootAddress()
			if err != nil {
				return err
			}
			out, _ := json.MarshalIndent(key, " ", " ")
			fmt.Println(string(out))
			return nil
		}
		key, err := pw.ExportHDAddresses(int(*inputAddressesToGenerate))
		if err != nil {
			return err
		}
		// TODO support json vs txt out
		out, _ := json.MarshalIndent(key, " ", " ")
		fmt.Println(string(out))
		return nil
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expected exactly one argument: create or inspect")
		}
		if args[0] != "create" && args[0] != "inspect" {
			return fmt.Errorf("expected argument to be create or inspect. Got: %s", args[0])
		}
		return nil
	},
}

func getFileOrFlag(filename *string, flag *string) (string, error) {
	if filename == nil && flag == nil {
		return "", fmt.Errorf("both the filename and the flag pointers are nil")
	}
	if filename != nil && *filename != "" {
		filedata, err := os.ReadFile(*filename)
		if err != nil {
			return "", fmt.Errorf("could not open the specified file %s. Got error %s", *filename, err.Error())
		}
		return string(filedata), nil
	}
	if flag != nil {
		return *flag, nil
	}
	return "", fmt.Errorf("unable to determine flat or filename")
}

func init() {
	inputKDFIterations = WalletCmd.PersistentFlags().Uint("iterations", 2048, "Number of pbkdf2 iterations to perform")
	inputWords = WalletCmd.PersistentFlags().Int("words", 24, "The number of words to use in the mnemonic")
	inputAddressesToGenerate = WalletCmd.PersistentFlags().Uint("addresses", 10, "The number of addresses to generate")
	inputLang = WalletCmd.PersistentFlags().String("language", "english", "Which language to use [ChineseSimplified, ChineseTraditional, Czech, English, French, Italian, Japanese, Korean, Spanish]")
	// https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	// 0 - bitcoin
	// 60 - ether
	// 966 - matic
	inputPath = WalletCmd.PersistentFlags().String("path", "m/44'/60'/0'", "What would you like the derivation path to be")
	inputPassword = WalletCmd.PersistentFlags().String("password", "", "Password used along with the mnemonic")
	inputPasswordFile = WalletCmd.PersistentFlags().String("password-file", "", "Password stored in a file used along with the mnemonic")
	inputMnemonic = WalletCmd.PersistentFlags().String("mnemonic", "", "A mnemonic phrase used to generate entropy")
	inputMnemonicFile = WalletCmd.PersistentFlags().String("mnemonic-file", "", "A mneomonic phrase written in a file used to generate entropy")
	inputUseRawEntropy = WalletCmd.PersistentFlags().Bool("raw-entropy", false, "substrate and polkda dot don't follow strict bip39 and use raw entropy")
	inputRootOnly = WalletCmd.PersistentFlags().Bool("root-only", false, "don't produce HD accounts. Just produce a single wallet")
}
