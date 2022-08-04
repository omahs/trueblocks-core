// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package chunksPkg

import (
	"errors"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/config"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/validate"
)

func (opts *ChunksOptions) validateChunks() error {
	opts.testLog()

	if opts.BadFlag != nil {
		return opts.BadFlag
	}

	if len(opts.Mode) == 0 {
		return validate.Usage("Please choose at least one of {0}.", "[stats|manifest|pins|blooms|index|addresses|appearances]")
	}

	err := validate.ValidateEnum("mode", opts.Mode, "[stats|manifest|pins|blooms|index|addresses|appearances]")
	if err != nil {
		return err
	}

	if opts.Globals.LogLevel > 0 && opts.Mode == "addresses" && opts.Globals.Format == "json" {
		return validate.Usage("You may not use --format json and log_level > 0 in addresses mode")
	}

	if opts.Publish {
		return validate.Usage("The {0} option is not yet enabled", "--publish")
	}

	if opts.Mode != "manifest" {
		if opts.PinRemote || opts.Publish {
			return validate.Usage("The {0} and {1} options are available only in {2} mode.", "--pin_remote", "--publish", "manifest")
		}
		if opts.Clean {
			return validate.Usage("The {0} option is available only in {1} mode.", "--clean", "manifest")
		}
		if opts.Check {
			return validate.Usage("The {0} option is available only in {1} mode.", "--check", "manifest")
		}
	} else {
		settings := config.GetBlockScrapeSettings(opts.Globals.Chain)
		key, secret := settings.Pinata_api_key, settings.Pinata_secret_api_key
		if opts.PinRemote {
			if len(key) == 0 {
				return validate.Usage("The {0} option requires {1}", "--pin_remote", "a pinata_api_key")
			}
			if len(secret) == 0 {
				return validate.Usage("The {0} option requires {1}", "--pin_remote", "a pinata_secret_api_key")
			}
		}
		if opts.Publish {
			if len(key) == 0 {
				return validate.Usage("The {0} option requires {1}", "--pin_remote", "a pinata_api_key")
			}
			if len(secret) == 0 {
				return validate.Usage("The {0} option requires {1}", "--pin_remote", "a pinata_secret_api_key")
			}
		}
	}

	if opts.Belongs {
		if opts.Mode != "addresses" && opts.Mode != "blooms" {
			return validate.Usage("The --belongs option is only available with either the addresses or blooms mode")
		}
		if len(opts.Addrs) == 0 {
			return validate.Usage("You must specifiy at least one address with the --belongs option")
		}
		if len(opts.Blocks) == 0 {
			return validate.Usage("You must specifiy at least one block identifier with the --belongs option")
		}
	}

	err = validate.ValidateIdentifiers(
		opts.Globals.Chain,
		opts.Blocks,
		validate.ValidBlockIdWithRangeAndDate,
		1,
		&opts.BlockIds,
	)
	if err != nil {
		if invalidLiteral, ok := err.(*validate.InvalidIdentifierLiteralError); ok {
			return invalidLiteral
		}
		if errors.Is(err, validate.ErrTooManyRanges) {
			return validate.Usage("Specify only a single block range at a time.")
		}
		return err
	}

	if opts.Repair {
		if opts.Mode != "manifest" {
			return validate.Usage("The --repair option is only available in index mode")
		}

		if len(opts.BlockIds) != 1 {
			return validate.Usage("You must supply exactly one block number with the --repair option")
		}

		blockNums, err := opts.BlockIds[0].ResolveBlocks(opts.Globals.Chain)
		if err != nil {
			return err
		}
		if len(blockNums) != 1 {
			return validate.Usage("You must supply exactly one block number with the --repair option")
		}

		if opts.Globals.TestMode {
			return validate.Usage("The --repair option is not available in test mode")
		}
	}

	if len(opts.Addrs) > 0 && !opts.Belongs {
		return validate.Usage("You may only specify an address with the --belongs option")
	}

	if opts.Details && opts.Belongs {
		return validate.Usage("Choose either {0} or {1}, not both.", "--details", "--belongs")
	}

	if !opts.Details && opts.Globals.ToFile {
		return validate.Usage("You may not use the {0} option without {1}.", "--to_file", "--details")
	}

	// Note this does not return if a migration is needed
	// TODO: BOGUS - DO WE WANT TO DISALLOW INVESTIGATION OF OLDER INSTALLATIONS?
	// migrate.CheckBackLevelIndex(opts.Globals.Chain)

	if opts.Remote {
		if opts.Mode != "pins" && opts.Mode != "manifest" {
			return validate.Usage("The {0} option is only available {1}.", "--remote", "in pins or manifest mode")
		}
	}

	return opts.Globals.Validate()
}
