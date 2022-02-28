// Copyright (c) 2019 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package util

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringToRau(t *testing.T) {
	require := require.New(t)
	inputString := []string{"123", "123.321", ".123", "0.123", "123.", "+123.0", "0.",
		".0", "00.00", ".00", "00."}
	expectedString := []string{"123000000000000", "123321000000000", "123000000000",
		"123000000000", "123000000000000", "123000000000000", "0", "0", "0", "0", "0"}
	invalidString := []string{"  .123", "123. ", "0.12345678912345678900000",
		"1..2", "1.+2", "-1.2", ".", ". ", " .", " . ", ""}
	for i, teststring := range inputString {
		res, err := StringToRau(teststring, GasPriceDecimalNum)
		require.NoError(err)
		require.Equal(res.String(), expectedString[i])
	}
	for _, teststring := range invalidString {
		_, err := StringToRau(teststring, GasPriceDecimalNum)
		require.Error(err)
	}
}

func TestRauToString(t *testing.T) {
	require := require.New(t)
	inputString := []string{"1", "0", "1000000000000", "200000000000", "30000000000",
		"1004000000000", "999999999999999999999939987", "100090907000030000100"}
	IotxString := []string{"0.000000000000000001", "0", "0.000001", "0.0000002", "0.00000003", "0.000001004",
		"999999999.999999999999939987", "100.0909070000300001"}
	GasString := []string{"0.000000000001", "0", "1", "0.2", "0.03", "1.004",
		"999999999999999.999999939987", "100090907.0000300001"}
	for i, testString := range inputString {
		testBigInt, ok := new(big.Int).SetString(testString, 10)
		require.True(ok)
		res := RauToString(testBigInt, IotxDecimalNum)
		require.Equal(IotxString[i], res)
		res = RauToString(testBigInt, GasPriceDecimalNum)
		require.Equal(GasString[i], res)
		res = RauToString(testBigInt, 0)
		require.Equal(testBigInt.String(), res)
	}
}

func TestTrimHexPrefix(t *testing.T) {
	require := require.New(t)
	tests := []string{"0xsjkdfhu238fhjk", "0x756c7d7aa7cfdb1c7447ffa13f4dd1eff04052a7addc6ac9ac29e0b234d088c2",
		"832947sd", "Ox38jj8j32j89", "00xasd98", "jsd8f9h0x", "0xx00xx0", "0X79a7hHIY^&?<><||0x)X", "~0@@x~", "0x0x0x0x",
		"0x608060405234801561001057600080fd5b50610504806100206000396000f3006080604052600436106100405763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041663e3b48f488114610045575b600080fd5b6040805160206004803580820135838102808601850190965280855261010495369593946024949385019291829185019084908082843750506040805187358901803560208181028481018201909552818452989b9a99890198929750908201955093508392508501908490808284375050604080516020601f89358b018035918201839004830284018301909452808352979a9998810197919650918201945092508291508401838280828437509497506101069650505050505050565b005b600080600061012c8651111515156101a557604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602760248201527f6e756d626572206f6620726563697069656e7473206973206c6172676572207460448201527f68616e2033303000000000000000000000000000000000000000000000000000606482015290519081900360840190fd5b845186511461021557604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601460248201527f706172616d6574657273206e6f74206d61746368000000000000000000000000604482015290519081900360640190fd5b60009250600091505b855182101561025057848281518110151561023557fe5b9060200190602002015183019250818060010192505061021e565b348311156102bf57604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f6e6f7420656e6f75676820746f6b656e00000000000000000000000000000000604482015290519081900360640190fd5b8234039050600091505b85518210156103cc5785828151811015156102e057fe5b9060200190602002015173ffffffffffffffffffffffffffffffffffffffff166108fc868481518110151561031157fe5b602090810290910101516040518115909202916000818181858888f19350505050158015610343573d6000803e3d6000fd5b507f69ca02dd4edd7bf0a4abb9ed3b7af3f14778db5d61921c7dc7cd545266326de2868381518110151561037357fe5b90602001906020020151868481518110151561038b57fe5b60209081029091018101516040805173ffffffffffffffffffffffffffffffffffffffff9094168452918301528051918290030190a16001909101906102c9565b600081111561043757604051339082156108fc029083906000818181858888f19350505050158015610402573d6000803e3d6000fd5b506040805182815290517f2e1897b0591d764356194f7a795238a87c1987c7a877568e50d829d547c92b979181900360200190a15b7f53a85291e316c24064ff2c7668d99f35ecbb40ef4e24794ff9d8abe901c7e62c846040518080602001828103825283818151815260200191508051906020019080838360005b8381101561049657818101518382015260200161047e565b50505050905090810190601f1680156104c35780820380516001836020036101000a031916815260200191505b509250505060405180910390a15050505050505600a165627a7a72305820a5345ea18c711d66438a38447bcdcef66494250ebc60bebae3181446e68291390029"}
	expects := []string{"sjkdfhu238fhjk", "756c7d7aa7cfdb1c7447ffa13f4dd1eff04052a7addc6ac9ac29e0b234d088c2",
		"832947sd", "Ox38jj8j32j89", "00xasd98", "jsd8f9h0x", "x00xx0", "0X79a7hHIY^&?<><||0x)X", "~0@@x~", "0x0x0x",
		"608060405234801561001057600080fd5b50610504806100206000396000f3006080604052600436106100405763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041663e3b48f488114610045575b600080fd5b6040805160206004803580820135838102808601850190965280855261010495369593946024949385019291829185019084908082843750506040805187358901803560208181028481018201909552818452989b9a99890198929750908201955093508392508501908490808284375050604080516020601f89358b018035918201839004830284018301909452808352979a9998810197919650918201945092508291508401838280828437509497506101069650505050505050565b005b600080600061012c8651111515156101a557604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602760248201527f6e756d626572206f6620726563697069656e7473206973206c6172676572207460448201527f68616e2033303000000000000000000000000000000000000000000000000000606482015290519081900360840190fd5b845186511461021557604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601460248201527f706172616d6574657273206e6f74206d61746368000000000000000000000000604482015290519081900360640190fd5b60009250600091505b855182101561025057848281518110151561023557fe5b9060200190602002015183019250818060010192505061021e565b348311156102bf57604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f6e6f7420656e6f75676820746f6b656e00000000000000000000000000000000604482015290519081900360640190fd5b8234039050600091505b85518210156103cc5785828151811015156102e057fe5b9060200190602002015173ffffffffffffffffffffffffffffffffffffffff166108fc868481518110151561031157fe5b602090810290910101516040518115909202916000818181858888f19350505050158015610343573d6000803e3d6000fd5b507f69ca02dd4edd7bf0a4abb9ed3b7af3f14778db5d61921c7dc7cd545266326de2868381518110151561037357fe5b90602001906020020151868481518110151561038b57fe5b60209081029091018101516040805173ffffffffffffffffffffffffffffffffffffffff9094168452918301528051918290030190a16001909101906102c9565b600081111561043757604051339082156108fc029083906000818181858888f19350505050158015610402573d6000803e3d6000fd5b506040805182815290517f2e1897b0591d764356194f7a795238a87c1987c7a877568e50d829d547c92b979181900360200190a15b7f53a85291e316c24064ff2c7668d99f35ecbb40ef4e24794ff9d8abe901c7e62c846040518080602001828103825283818151815260200191508051906020019080838360005b8381101561049657818101518382015260200161047e565b50505050905090810190601f1680156104c35780820380516001836020036101000a031916815260200191505b509250505060405180910390a15050505050505600a165627a7a72305820a5345ea18c711d66438a38447bcdcef66494250ebc60bebae3181446e68291390029"}
	for i, s := range tests {
		require.Equal(expects[i], TrimHexPrefix(s))
	}
}

func TestParseHdwPath(t *testing.T) {
	r := require.New(t)

	tests := []struct {
		addressOrAlias string
		a, b, c        uint32
		err            string
	}{
		{"hdw::0/1/2", 0, 1, 2, ""},
		{"hdw::0/1", 0, 0, 1, ""},
		{"hdw::2/0", 0, 2, 0, ""},
		{"hdw::1/2/0", 1, 2, 0, ""},
		{"hdw::0", 0, 0, 0, "derivation path error"},
		{"hdw::", 0, 0, 0, "derivation path error"},
		{"hdw::0/1/2/3", 0, 0, 0, "derivation path error"},
		{"hdw::a/3", 0, 0, 0, "must be integer value"},
		{"hdw::a/b", 0, 0, 0, "must be integer value"},
		{"hdw::1/23b", 0, 0, 0, "must be integer value"},
	}
	for _, v := range tests {
		a, b, c, err := ParseHdwPath(v.addressOrAlias)
		r.Equal(a, v.a)
		r.Equal(b, v.b)
		r.Equal(c, v.c)
		if err != nil {
			r.Contains(err.Error(), v.err)
		}
	}
}
