package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/PlakarKorp/kloset/encryption"
	"github.com/PlakarKorp/kloset/repository"
	"github.com/PlakarKorp/kloset/storage"
	"github.com/PlakarKorp/kloset/versioning"
	"github.com/PlakarKorp/plakar/appcontext"
	"github.com/PlakarKorp/plakar/utils"
)

var ErrCantUnlock = errors.New("failed to unlock repository")

func getpassphrase(ctx *appcontext.AppContext) ([]byte, error) {
	if ctx.KeyFromFile != "" {
		return []byte(ctx.KeyFromFile), nil
	}

	if pass, ok := os.LookupEnv("KAPSULE_PASSPHRASE"); ok {
		return []byte(pass), nil
	}

	return nil, nil
}

func setupEncryption(ctx *appcontext.AppContext, config *storage.Configuration, params map[string]string) error {
	if config.Encryption == nil {
		return nil
	}

	secret, err := getpassphrase(ctx)
	if err != nil {
		return err
	}

	if secret != nil {
		key, err := encryption.DeriveKey(config.Encryption.KDFParams,
			secret)
		if err != nil {
			return err
		}

		if !encryption.VerifyCanary(config.Encryption, key) {
			return ErrCantUnlock
		}
		ctx.SetSecret(key)
		return nil
	}

	// fall back to prompting
	for range 3 {
		passphrase, err := utils.GetPassphrase("repository")
		if err != nil {
			return err
		}

		key, err := encryption.DeriveKey(config.Encryption.KDFParams,
			passphrase)
		if err != nil {
			return err
		}
		if encryption.VerifyCanary(config.Encryption, key) {
			ctx.SetSecret(key)
			return nil
		}
	}

	return ErrCantUnlock
}

func openKapsule(ctx *appcontext.AppContext, kapsule string) (*repository.Repository, error) {
	if strings.HasPrefix(kapsule, "http://") ||
		strings.HasPrefix(kapsule, "https://") {
		kapsule = "ptar+" + kapsule
	}

	store, serializedConfig, err := storage.Open(ctx.GetInner(), map[string]string{
		"location": kapsule,
	})
	if err != nil {
		fmt.Println("Error opening kapsule:", err)
		return nil, err
	}

	repoConfig, err := storage.NewConfigurationFromWrappedBytes(serializedConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", flag.CommandLine.Name(), err)
		return nil, err
	}

	if repoConfig.Version != versioning.FromString(storage.VERSION) {
		fmt.Fprintf(os.Stderr, "%s: incompatible repository version: %s != %s\n",
			flag.CommandLine.Name(), repoConfig.Version, storage.VERSION)
		return nil, err
	}

	if err := setupEncryption(ctx, repoConfig, map[string]string{}); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", flag.CommandLine.Name(), err)
		return nil, err
	}

	repo, err := repository.New(ctx.GetInner(), ctx.GetSecret(), store, serializedConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", flag.CommandLine.Name(), err)
		return nil, err
	}

	return repo, nil
}
