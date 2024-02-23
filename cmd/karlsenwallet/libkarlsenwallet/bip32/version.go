package bip32

import "github.com/pkg/errors"

// BitcoinMainnetPrivate is the version that is used for
// bitcoin mainnet bip32 private extended keys.
// Ecnodes to xprv in base58.
var BitcoinMainnetPrivate = [4]byte{
	0x04,
	0x88,
	0xad,
	0xe4,
}

// BitcoinMainnetPublic is the version that is used for
// bitcoin mainnet bip32 public extended keys.
// Ecnodes to xpub in base58.
var BitcoinMainnetPublic = [4]byte{
	0x04,
	0x88,
	0xb2,
	0x1e,
}

// KarlsenMainnetPrivate is the version that is used for
// karlsen mainnet bip32 private extended keys.
// Ecnodes to xprv in base58.
var KarlsenMainnetPrivate = [4]byte{
	0x03,
	0x8f,
	0x2e,
	0xf4,
}

// KarlsenMainnetPublic is the version that is used for
// karlsen mainnet bip32 public extended keys.
// Ecnodes to kpub in base58.
var KarlsenMainnetPublic = [4]byte{
	0x03,
	0x8f,
	0x33,
	0x2e,
}

// KarlsenTestnetPrivate is the version that is used for
// karlsen testnet bip32 public extended keys.
// Ecnodes to ktrv in base58.
var KarlsenTestnetPrivate = [4]byte{
	0x03,
	0x90,
	0x9e,
	0x07,
}

// KarlsenTestnetPublic is the version that is used for
// karlsen testnet bip32 public extended keys.
// Ecnodes to ktub in base58.
var KarlsenTestnetPublic = [4]byte{
	0x03,
	0x90,
	0xa2,
	0x41,
}

// KarlsenDevnetPrivate is the version that is used for
// karlsen devnet bip32 public extended keys.
// Ecnodes to kdrv in base58.
var KarlsenDevnetPrivate = [4]byte{
	0x03,
	0x8b,
	0x3d,
	0x80,
}

// KarlsenDevnetPublic is the version that is used for
// karlsen devnet bip32 public extended keys.
// Ecnodes to xdub in base58.
var KarlsenDevnetPublic = [4]byte{
	0x03,
	0x8b,
	0x41,
	0xba,
}

// KarlsenSimnetPrivate is the version that is used for
// karlsen simnet bip32 public extended keys.
// Ecnodes to ksrv in base58.
var KarlsenSimnetPrivate = [4]byte{
	0x03,
	0x90,
	0x42,
	0x42,
}

// KarlsenSimnetPublic is the version that is used for
// karlsen simnet bip32 public extended keys.
// Ecnodes to xsub in base58.
var KarlsenSimnetPublic = [4]byte{
	0x03,
	0x90,
	0x46,
	0x7d,
}

func toPublicVersion(version [4]byte) ([4]byte, error) {
	switch version {
	case BitcoinMainnetPrivate:
		return BitcoinMainnetPublic, nil
	case KarlsenMainnetPrivate:
		return KarlsenMainnetPublic, nil
	case KarlsenTestnetPrivate:
		return KarlsenTestnetPublic, nil
	case KarlsenDevnetPrivate:
		return KarlsenDevnetPublic, nil
	case KarlsenSimnetPrivate:
		return KarlsenSimnetPublic, nil
	}

	return [4]byte{}, errors.Errorf("unknown version %x", version)
}

func isPrivateVersion(version [4]byte) bool {
	switch version {
	case BitcoinMainnetPrivate:
		return true
	case KarlsenMainnetPrivate:
		return true
	case KarlsenTestnetPrivate:
		return true
	case KarlsenDevnetPrivate:
		return true
	case KarlsenSimnetPrivate:
		return true
	}

	return false
}
