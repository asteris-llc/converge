---
title: "Module Verification"
date: "2016-09-13T00:13:59-05:00"

menu:
  main:
    parent: "converge"
    weight: 20
---

In the future, converge will require modules to be signed using a gpg detached signature.
The following steps will walk you through the creation of a gpg keypair suitable for signing a module.
If you have an existing gpg signing key skip to the [Signing modules](#signing-modules) step.

## Generate a gpg signing key

Create a file named `gpg-batch` with the following content.

```
%echo Generating a default key
Key-Type: RSA
Key-Length: 2048
Subkey-Type: RSA
Subkey-Length: 2048
Name-Real: Test Asteris
Name-Comment: Test Key
Name-Email: test@aster.is
Expire-Date: 0
Passphrase: asteris
%pubring test.pub
%secring test.sec
%commit
%echo done
```

We can use this to quickly generate a key pair using batch mode.

```
$ gpg --batch --gen-key gpg-batch
```

We can verify this worked by listing the keys.

```
$ gpg --no-default-keyring --secret-keyring ./test.sec --keyring ./test.pub --list-keys
./test.pub

----------
pub   2048R/475CC928 2016-08-07
uid       [ unknown] Test Asteris (Test signing key) <test@aster.is>
sub   2048R/1327D89C 2016-08-07
```

We can tell from the output above, that the level of trust for the signing key is unknown. This will cause the following warning if we attempt to validate a module signed with this key using the gpg cli:

```
gpg: WARNING: This key is not certified with a trusted signature!
```

Since we know exactly where this key came from let's trust it:

```
$ gpg --no-default-keyring --secret-keyring ./test.sec --keyring ./test.pub --edit-key 475CC928

gpg (GnuPG) 1.4.20; Copyright (C) 2015 Free Software Foundation, Inc.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

Secret key is available.

pub  2048R/475CC928  created: 2016-08-07  expires: never       usage: SCEA
                     trust: unknown      validity: unknown
sub  2048R/1327D89C  created: 2016-08-07  expires: never       usage: SEA
[ unknown] (1). Test Asteris (Test signing key) <test@aster.is>

Please decide how far you trust this user to correctly verify other users' keys
(by looking at passports, checking fingerprints from different sources, etc.)

  1 = I don't know or won't say
  2 = I do NOT trust
  3 = I trust marginally
  4 = I trust fully
  5 = I trust ultimately
  m = back to the main menu

Your decision? 5
Do you really want to set this key to ultimate trust? (y/N) y

pub  2048R/475CC928  created: 2016-08-07  expires: never       usage: SCEA
                     trust: ultimate      validity: ultimate
sub  2048R/1327D89C  created: 2016-08-07  expires: never       usage: SEA
[ultimate] (1). Test Asteris (Test signing key) <test@aster.is>
Please note that the shown key validity is not necessarily correct
unless you restart the program.

gpg> quit
```

## Signing modules

Now you can start signing modules with the key. The following command will produce a signature file called `basic.hcl.asc`.

```
$ gpg --no-default-keyring --armor --secret-keyring ./test.sec --keyring ./test.pub --output basic.hcl.asc --detach-sig basic.hcl
```

This file should be shipped along side the module so that the converge tool can download it and use it to verify that the module has not been modified after the signature was created.

## Public keystore

In order to verify a module's signature against its signature file, converge needs access to our public key. This can be exported with the following command.

```
$ gpg --no-default-keyring --armor --secret-keyring ./test.sec --keyring ./test.pub --export test@aster.is > pubkeys.gpg
```

Now we must add the key to the converge's public key database. The following directories make up the default converge keystore layout:

```
sytem: /usr/lib/converge/trustedkeys/
user: ~/.converge/trustedkeys/
local: $(pwd)/trustedkeys/
```

The system path is designed to be used by system administrators. The user path is where converge stores keys that are added through the `converge key trust` command. Finally, the local path can be used for keys that you do not want stored globally or managed by converge.

Trusted keys are saved in the desired directory named after the fingerprint of the public key. For global and local keys, we will need to manually create this file.

We do this by downloading the key, capturing its fingerprint, and storing it in the database using the fingerprint as the filename.

```
$ curl -O https://example.com/pubkeys.gpg

$ gpg --no-default-keyring --with-fingerprint pubkeys.gpg
pub  2048R/475CC928 2016-08-07 Test Asteris (Test signing key) <test@aster.is>
     Key fingerprint = 74FD F669 F18D 59F9 2B0A  ACCD 7203 51FF 475C C928
	 sub  2048R/1327D89C 2016-08-07

$ echo "74FD F669 F18D 59F9 2B0A  ACCD 7203 51FF 475C C928" | tr -d "[:space:]" | tr '[:upper:]' '[:lower:]'
74fdf669f18d59f92b0aaccd720351ff475cc928

mkdir -p trustedkeys
mv pubkeys.gpg trustedkeys/74fdf669f18d59f92b0aaccd720351ff475cc928
```

You can disable a key stored in the global system path by creating an empty file in the user or local paths with the same name. Keys stored in the local path will also mask keys in the user path.

## Trusting keys

There is an easier way to add a key to the user keystore, using the `converge key trust` subcommand.

```
$ converge key trust pubkeys.gpg
```

The command will ask you to verify that the fingerprint matches the fingerprint you expected for the key.

```
The gpg key fingerprint is 74fdf669f18d59f92b0aaccd720351ff475cc928
Are you sure you want to trust this key (yes/no)? yes
Trusting key "https://example.com/pubkeys.gpg".
```

## Converge integration

Now let's walk through the steps converge takes when fetching modules. For now, converge will not attempt to download the detached signature and verify the module. You can enable module verification with the `--verify-modules` flag.

```
$ converge plan --verify-modules https://example.com/modules/basic.hcl
```

This will result in converge retrieving the following URIs.

```
https://example.com/modules/basic.hcl
https://example.com/modules/basic.hcl.asc
```

Then it verifies the signature of the module using the public keys in the key database.
