package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr/musig2"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/schnorr"
	"github.com/dominant-strategies/go-quai/common"
	"github.com/dominant-strategies/go-quai/crypto"
)

func TestSingleSigner(t *testing.T) {

	// ECDSA key
	key, err := crypto.HexToECDSA("345debf66bc68724062b236d3b0a6eb30f051e725ebb770f1dc367f2c569f003")
	if err != nil {
		fmt.Println(err)
	}
	addr := crypto.PubkeyToAddress(key.PublicKey)

	b, err := hex.DecodeString("345debf66bc68724062b236d3b0a6eb30f051e725ebb770f1dc367f2c569f003")
	if err != nil {
		fmt.Println(err)
	}

	// btcec key for schnorr use
	btcecKey, _ := btcec.PrivKeyFromBytes(b)

	// Check that btcecPubkey is the same as the ECDSA key
	btcecPubkey := btcecKey.PubKey()
	if !bytes.Equal(btcecPubkey.SerializeUncompressed(), crypto.FromECDSAPub(&key.PublicKey)) {
		t.Fatalf("btcec pubkey does not match ECDSA pubkey")
	}

	outpointHash := common.HexToHash("000000000000000000000000000000000000000000000000000012")
	outpointIndex := uint32(0)

	// key = hash(blockHash, index)
	// Find hash / index for originUtxo / imagine this is block hash
	prevOut := *NewOutPoint(&outpointHash, outpointIndex)

	in := TxIn{
		PreviousOutPoint: prevOut,
		PubKey:           btcecPubkey.SerializeUncompressed(),
	}

	newOut := TxOut{
		Value:   10000000,
		Address: addr.Bytes(),
	}

	utxo := &UtxoTx{
		TxIn:  []TxIn{in},
		TxOut: []TxOut{newOut},
	}

	tx := NewTx(utxo)
	txHash := tx.Hash()

	sig, err := schnorr.Sign(btcecKey, txHash[:])
	if err != nil {
		t.Fatalf("schnorr signing failed!")
	}

	// Finally we'll combined all the nonces, and ensure that it validates
	// as a single schnorr signature.
	if !sig.Verify(txHash[:], btcecPubkey) {
		t.Fatalf("final sig is invalid!")
	}
}

func TestMultiSigners(t *testing.T) {

	// ECDSA key
	key1, err := crypto.HexToECDSA("345debf66bc68724062b236d3b0a6eb30f051e725ebb770f1dc367f2c569f003")
	if err != nil {
		fmt.Println(err)
	}

	b1, err := hex.DecodeString("345debf66bc68724062b236d3b0a6eb30f051e725ebb770f1dc367f2c569f003")
	if err != nil {
		fmt.Println(err)
	}

	// btcec key for schnorr use
	btcecKey1, _ := btcec.PrivKeyFromBytes(b1)
	btcecPubkey1 := btcecKey1.PubKey()
	if !bytes.Equal(btcecPubkey1.SerializeUncompressed(), crypto.FromECDSAPub(&key1.PublicKey)) {
		t.Fatalf("btcec pubkey does not match ECDSA pubkey")
	}

	key2, err := crypto.HexToECDSA("000000f66bc68724062b236d3b0a6eb30f051e725ebb770f1dc367f2c569f003")
	if err != nil {
		fmt.Println(err)
	}

	b2, err := hex.DecodeString("000000f66bc68724062b236d3b0a6eb30f051e725ebb770f1dc367f2c569f003")
	if err != nil {
		fmt.Println(err)
	}

	btcecKey2, _ := btcec.PrivKeyFromBytes(b2)
	btcecPubkey2 := btcecKey2.PubKey()
	if !bytes.Equal(btcecPubkey2.SerializeUncompressed(), crypto.FromECDSAPub(&key2.PublicKey)) {
		t.Fatalf("btcec pubkey does not match ECDSA pubkey")
	}

	outpointIndex := uint32(0)

	outpointHash1 := common.HexToHash("00000000000000000000000000000000000000000000000000000")
	outpointHash2 := common.HexToHash("00000000000000000000000000000000000000000000000000001")

	// key = hash(blockHash, index)
	// Find hash / index for originUtxo / imagine this is block hash
	prevOut1 := *NewOutPoint(&outpointHash1, outpointIndex)
	prevOut2 := *NewOutPoint(&outpointHash2, outpointIndex)

	in1 := TxIn{
		PreviousOutPoint: prevOut1,
		PubKey:           btcecKey1.PubKey().SerializeUncompressed(),
	}

	in2 := TxIn{
		PreviousOutPoint: prevOut2,
		PubKey:           btcecKey2.PubKey().SerializeUncompressed(),
	}

	addr1 := common.HexToAddress("0x1400000000000000000000000000000000000001")
	addr2 := common.HexToAddress("0x1500000000000000000000000000000000000001")

	newOut1 := TxOut{
		Value:   50000,
		Address: addr1.Bytes(),
	}

	newOut2 := TxOut{
		Value:   10000,
		Address: addr2.Bytes(),
	}

	utxo := &UtxoTx{
		TxIn:  []TxIn{in1, in2},
		TxOut: []TxOut{newOut1, newOut2},
	}

	tx := NewTx(utxo)
	txHash := sha256.Sum256(tx.Hash().Bytes())

	keys := []*btcec.PrivateKey{btcecKey1, btcecKey2}
	signSet := []*btcec.PublicKey{btcecKey1.PubKey(), btcecKey2.PubKey()}

	var combinedKey *btcec.PublicKey
	var ctxOpts []musig2.ContextOption

	ctxOpts = append(ctxOpts, musig2.WithKnownSigners(signSet))

	// Now that we have all the signers, we'll make a new context, then
	// generate a new session for each of them(which handles nonce
	// generation).
	signers := make([]*musig2.Session, len(keys))
	for i, signerKey := range keys {
		signCtx, err := musig2.NewContext(
			signerKey, false, ctxOpts...,
		)
		if err != nil {
			t.Fatalf("unable to generate context: %v", err)
		}

		if combinedKey == nil {
			combinedKey, err = signCtx.CombinedKey()
			if err != nil {
				t.Fatalf("combined key not available: %v", err)
			}
		}

		session, err := signCtx.NewSession()
		if err != nil {
			t.Fatalf("unable to generate new session: %v", err)
		}
		signers[i] = session
	}

	// Next, in the pre-signing phase, we'll send all the nonces to each
	// signer.
	var wg sync.WaitGroup
	for i, signCtx := range signers {
		signCtx := signCtx

		wg.Add(1)
		go func(idx int, signer *musig2.Session) {
			defer wg.Done()

			for j, otherCtx := range signers {
				if idx == j {
					continue
				}

				nonce := otherCtx.PublicNonce()
				haveAll, err := signer.RegisterPubNonce(nonce)
				if err != nil {
					t.Fatalf("unable to add public nonce")
				}

				if j == len(signers)-1 && !haveAll {
					t.Fatalf("all public nonces should have been detected")
				}
			}
		}(i, signCtx)
	}

	wg.Wait()

	// In the final step, we'll use the first signer as our combiner, and
	// generate a signature for each signer, and then accumulate that with
	// the combiner.
	combiner := signers[0]
	for i := range signers {
		signer := signers[i]
		partialSig, err := signer.Sign(txHash)
		if err != nil {
			t.Fatalf("unable to generate partial sig: %v", err)
		}

		// We don't need to combine the signature for the very first
		// signer, as it already has that partial signature.
		if i != 0 {
			haveAll, err := combiner.CombineSig(partialSig)
			if err != nil {
				t.Fatalf("unable to combine sigs: %v", err)
			}

			if i == len(signers)-1 && !haveAll {
				t.Fatalf("final sig wasn't reconstructed")
			}
		}
	}

	aggKey, _, _, _ := musig2.AggregateKeys(
		signSet, false,
	)

	fmt.Println("aggKey", aggKey.FinalKey)
	fmt.Println("combinedKey", combinedKey)

	if !aggKey.FinalKey.IsEqual(combinedKey) {
		t.Fatalf("aggKey is invalid!")
	}

	// Finally we'll combined all the nonces, and ensure that it validates
	// as a single schnorr signature.
	finalSig := combiner.FinalSig()
	if !finalSig.Verify(txHash[:], combinedKey) {
		t.Fatalf("final sig is invalid!")
	}
}

func TestVerify(t *testing.T) {
	// Test that we mark a signature as valid when it is and invalid when it is not valid
	testCases := []struct {
		name           string
		privateKey     string
		publicKey      string
		messageDigest  string
		signature      string
		expectedResult bool
	}{
		{ // TEST CASE #1:
			name:           "Confirms a Valid Signature",
			privateKey:     "c90fdaa22168c234c4c6628b80dc1cd129024e088a67cc74020bbea63b14e5c7",
			publicKey:      "04fac2114c2fbb091527eb7c64ecb11f8021cb45e8e7809d3c0938e4b8c0e5f84bc655c2105c3c5c380f2c8b8ce2c0c25b0d57062d2d28187254f0deb802b8891f",
			messageDigest:  "4df3c3f68fcc83b27e9d42c90431a72499f17875c81a599b566c9889b9696703",
			signature:      "5364b58801791a30ee9f2dfb16bdfb543800eccddb514c56c7b8d75e30d25149ba273d4e61d2bb29f6e9e8a29bc7a31f6653a53bd81cf6994df07e58b1cb768b",
			expectedResult: true,
		},
		{ // TEST CASE #2:
			name:           "Confirms a Valid Signature",
			privateKey:     "c90fdaa22168c234c4c6628b80dc1cd129024e088a67cc74020bbea63b14e5c7",
			publicKey:      "04fac2114c2fbb091527eb7c64ecb11f8021cb45e8e7809d3c0938e4b8c0e5f84bc655c2105c3c5c380f2c8b8ce2c0c25b0d57062d2d28187254f0deb802b8891f",
			messageDigest:  "0000000000000000000000000000000000000000000000000000000000000000",
			signature:      "f6b70ab3159a32120c3af5ada42625c08f5f3d412179d8763347a3b3a133b73389b5163772dd8f8c02ea513e81eb244508a81dc4a11495c1ee458a226e178a1a",
			expectedResult: true,
		},
		{ // TEST CASE #3:
			name:           "Fails signature with public key not on the curve",
			publicKey:      "03eefdea4cdb677750a420fee807eacf21eb9898ae79b9768766e4faa04a2d4a34",
			messageDigest:  "4df3c3f68fcc83b27e9d42c90431a72499f17875c81a599b566c9889b9696703",
			signature:      "00000000000000000000003b78ce563f89a0ed9414f5aa28ad0d96d6795f9c6302a8dc32e64e86a333f20ef56eac9ba30b7246d6d25e22adb8c6be1aeb08d49d",
			expectedResult: false,
		},
		{ // TEST CASE #4: FAILING
			name:           "Fails signature with incorrect R residuosity",
			privateKey:     "c90fdaa22168c234c4c6628b80dc1cd129024e088a67cc74020bbea63b14e5c7",
			publicKey:      "04fac2114c2fbb091527eb7c64ecb11f8021cb45e8e7809d3c0938e4b8c0e5f84bc655c2105c3c5c380f2c8b8ce2c0c25b0d57062d2d28187254f0deb802b8891f",
			messageDigest:  "243f6a8885a308d313198a2e03707344a4093822299f31d0082efa98ec4e6c89",
			signature:      "48a215e87777e4fa800d5d2a3d7b858414401727063f2c189355853b9d0f9a87f468606087da7f2373befefa1259e71cccbdc9bd75eadd1a73e346420fa75cf7",
			expectedResult: false,
		},
		{ // TEST CASE #5:
			name:           "Fails signature with negated message",
			privateKey:     "c90fdaa22168c234c4c6628b80dc1cd129024e088a67cc74020bbea63b14e5c7",
			publicKey:      "04fac2114c2fbb091527eb7c64ecb11f8021cb45e8e7809d3c0938e4b8c0e5f84bc655c2105c3c5c380f2c8b8ce2c0c25b0d57062d2d28187254f0deb802b8891f",
			messageDigest:  "5e2d58d8b3bcdf1abadec7829054f90dda9805aab56c77333024b9d0a508b75c",
			signature:      "00da9b08172a9b6f0466a2defd817f2d7ab437e0d253cb5395a963866b3574bed092f9d860f1776a1f7412ad8a1eb50daccc222bc8c0e26b2056df2f273efdec",
			expectedResult: false,
		},
		{ // TEST CASE #6:
			name:           "Fails signature with negated s value",
			privateKey:     "c90fdaa22168c234c4c6628b80dc1cd129024e088a67cc74020bbea63b14e5c7",
			publicKey:      "0279be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798",
			messageDigest:  "0000000000000000000000000000000000000000000000000000000000000000",
			signature:      "b75dea1788881b057ff2a2d5c2847a7bebbfe8d8f9c0d3e76caa7ac462f065780b979f9f782580dc8c410105eda618e3334236428a1522e58c1cb9bdf058a308",
			expectedResult: false,
		},
	}

	var signature [64]byte
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			publicBytes, _ := hex.DecodeString(testCase.publicKey)
			msgBytes, _ := hex.DecodeString(testCase.messageDigest)
			sig, _ := hex.DecodeString(testCase.signature)
			copy(signature[:], sig)
			public, err := btcec.ParsePubKey(publicBytes)
			if err != nil {
				errorMessage := err.Error()
				if strings.Contains(errorMessage, "invalid public key") && testCase.expectedResult == false {
					return
				}
				t.Fatal(err)
			}

			sigFormatted, _ := schnorr.ParseSignature(signature[:])
			result := sigFormatted.Verify(msgBytes, public)

			if result != testCase.expectedResult { // || err != nil
				t.Fatalf("Did not confirm/deny validity of signature as expected: Want: %t    Got: %t   Error: %s", testCase.expectedResult, result, err)
			} else {
				t.Logf("SUCCESS: Expected verify result. Schnorr Signature is valid: %t    Error: %s", result, err)
			}
		})
	}
}

// ModifyRSig takes a Schnorr signature and modifies its R component.
// signatureHex is the original signature in hexadecimal format.
// It returns the modified signature also in hexadecimal format.
func TestModifyRSig(t *testing.T) {
	signatureHex := "b75dea1788881b057ff2a2d5c2847a7bebbfe8d8f9c0d3e76caa7ac462f06578f468606087da7f2373befefa1259e71cccbdc9bd75eadd1a73e346420fa75cf7"
	// Decode the signature from hex format
	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		log.Fatalf("Failed to parse signature: %s", err)
	}

	// Check if the signature length is even
	if len(signatureBytes)%2 != 0 {
		log.Fatalf("Failed to parse signature: %s", err)
	}

	// Calculate the length of R and S components
	halfLength := len(signatureBytes) / 2

	// Modify the R component
	// Here, we simply invert the bytes of the R component
	// You can replace this logic with any other modification you need
	for i := 0; i < halfLength; i++ {
		signatureBytes[i] = ^signatureBytes[i]
	}

	// Encode the modified signature back to hex format
	modifiedSigHex := hex.EncodeToString(signatureBytes)
	fmt.Println("Modified Signature: ", modifiedSigHex)
}

// ModifySSig takes a Schnorr signature and modifies its S component.
// signatureHex is the original signature in hexadecimal format.
// It returns the modified signature also in hexadecimal format.
func TestModifySSig(t *testing.T) {
	signatureHex := "b75dea1788881b057ff2a2d5c2847a7bebbfe8d8f9c0d3e76caa7ac462f06578f468606087da7f2373befefa1259e71cccbdc9bd75eadd1a73e346420fa75cf7"
	// Decode the signature from hex format
	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		log.Fatalf("Failed to parse signature: %s", err)
	}

	// Check if the signature length is even
	if len(signatureBytes)%2 != 0 {
		log.Fatalf("Signature length is not even: %s", err)
	}

	// Calculate the length of R and S components
	halfLength := len(signatureBytes) / 2

	// Modify the S component
	// Here, we simply invert the bytes of the S component
	// You can replace this logic with any other modification you need
	for i := halfLength; i < len(signatureBytes); i++ {
		signatureBytes[i] = ^signatureBytes[i]
	}

	// Encode the modified signature back to hex format
	modifiedSigHex := hex.EncodeToString(signatureBytes)
	fmt.Println("Modified Signature: ", modifiedSigHex)
}
