---
title: "Module Verification"
date: "2016-09-13T00:13:59-05:00"

menu:
  main:
    parent: "converge"
    weight: 20
---

This guide will walk you through signing, distributing, and verifying a converge module.

## Signing Modules

In the future, converge will require modules to be signed using a gpg detached signature.
The following steps will walk you through the creation of a gpg keypair suitable for signing a module.
If you have an existing gpg signing key skip to the [Signing the Module](#signing-the-module) step. There is a test key `samples/pubkey.gpg` that is used with testing converge.

### Generate a gpg signing key

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

#### Generate the key using batch mode

```
$ gpg --batch --gen-key gpg-batch
```

#### List the keys

```
$ gpg --no-default-keyring --secret-keyring ./test.sec --keyring ./test.pub --list-keys
./test.pub
```

```
----------
pub   2048R/475CC928 2016-08-07
uid       [ unknown] Test Asteris (Test signing key) <test@aster.is>
sub   2048R/1327D89C 2016-08-07
```

From the output above, the level of trust for the signing key is unknown.
This will cause the following warning if we attempt to validate a module signed with this key using the gpg cli:

```
gpg: WARNING: This key is not certified with a trusted signature!
```

Since we know exactly where this key came from let's trust it:

```
$ gpg --no-default-keyring --secret-keyring ./test.sec --keyring ./test.pub --edit-key 475CC928
```

```
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

#### Export the public key

```
$ gpg --no-default-keyring --armor --secret-keyring ./test.sec --keyring ./test.pub --export test@aster.is > pubkeys.gpg
```

### Signing the module

```
$ gpg --no-default-keyring --armor --secret-keyring ./test.sec --keyring ./test.pub --output base.hcl.asc --detach-sig basic.hcl
```

#### Verify the module using gpg

```
$ gpg --no-default-keyring --secret-keyring ./test.sec --keyring ./test.pub --verify basic.hcl.asc basic.hcl
```

```
gpg: Signature made Sun Aug  7 14:02:17 2016 CDT using RSA key ID 1327D89C
gpg: Good signature from "Test Asteris (Test signing key) <test@aster.is>"
```

At this point you should have the following three files:

```
basic.hcl.asc
basic.hcl
pubkeys.gpg
```

## Distributing Modules

Serve the following files.

```
https://example.com/modules/basic.hcl.asc
https://example.com/modules/basic.hcl
https://example.com/pubkeys.gpg
```

### converge Integration

Let's walk through the steps converge takes when fetching modules.

The following converge command:

```
$ converge plan https://example.com/modules/basic.hcl
```

results in converge retrieving the following URIs:

```
https://example.com/modules/basic.hcl
https://example.com/modules/basic.hcl.asc
```

Then it verifies the signature of the module with the public keys in its database.

## Verifying modules with converge

### Establishing Trust

By default converge does not trust any signing keys.
Trust is established by storing public keys in the converge keystore.
This can be done using `converge trust` or manually, using the procedures described in the next section.

The following directories make up the default converge keystore layout:

```
/urs/lib/converge/trustedkeys/
~/.converge/trustedkeys/
$(pwd)/trustedkeys/
```

Trusted keys are saved in the desired directory named after the fingerprint of the public key.
The `/usr/lib/converge` path is designed to be used by the OS distribution.
You can "disable" a trusted key by writing an empty file under `~/.converge/trutedkeys` or `$(pwd)/trustedkeys)`.
For example, if your OS distribution shipped with the following trusted key:

```
/usr/lib/converge/trustedkeys/74fdf669f18d59f92b0aaccd720351ff475cc928
```

you can disable it by writing the following empty file:

```
~/.converge/trustedkeys/74fdf669f18d59f92b0aaccd720351ff475cc928
```

### Trusting a key

As an example, let's look at how we can trust a key used to sign modules.

#### Using converge trust

The easiest way to trust a key is to use the `converge trust` subcommand.
In this case, we directly pass it the URI containing the public key we wish to trust:

```
$ converge trust https://example.com/pubkeys.gpg
```

```
The gpg key fingerprint is 74fdf669f18d59f92b0aaccd720351ff475cc928
Are you sure you want to trust this key (yes/no)? yes
Trusting key "https://example.com/pubkeys.gpg".
```

#### Manually adding keys

An alternative to using `converge trust` is to manually trust keys by adding them to converge's database.
We do this by downloading the key, capturing its fingerprint, and storing it in the database using the fingerprint as filename

##### Download the public key

```
$ curl -O https://example.com/pubkeys.gpg
```

###### Capture the public key fingerprint

```
$ gpg --no-default-keyring --with-fingerprint pubkeys.gpg
```

```
pub  2048R/475CC928 2016-08-07 Test Asteris (Test signing key) <test@aster.is>
     Key fingerprint = 74FD F669 F18D 59F9 2B0A  ACCD 7203 51FF 475C C928
sub  2048R/1327D89C 2016-08-07
```

Remove white spaces and convert to lowercase:

```
$ echo "74FD F669 F18D 59F9 2B0A  ACCD 7203 51FF 475C C928" | tr -d "[:space:]" | tr '[:upper:]' '[:lower:]'
```

```
74fdf669f18d59f92b0aaccd720351ff475cc928
```

##### Trust the key globally

```
mkdir -p ~/.converge/trustedkeys/
mv pubkeys.gpg ~/.converge/trustedkeys/74fdf669f18d59f92b0aaccd720351ff475cc928
```

### Example Usage

#### Download, verify and plan a module

For now, converge will not attempt to download the detached signature and verify the module. You can enable module verification with the `--verify-modules` flag.

```
$ converge apply --verify-modules https://example.com/modules/basic.hcl
```
