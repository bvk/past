OVERVIEW
--------

This package reimplements the password-store management script `pass` in Go
with new name `past` and with in-built support for a browser-extension for
Google Chrome.

As of now, only GNU/Linux is supported.

DEPENDENCIES
------------

Past requires `git` and `gpg` tools for normal operations. You need to have
`go` version `1.12` or above for installation.

INSTALLATION
------------

Use the following instructions to install this package.

```
$ go get -u github.com/bvk/past@v1.0.0
$ sudo ~/go/bin/past install https://github.com/bvk/past/past-1.0.0.crx
```

Above command will install the browser extension for all users. You may need to
restart google-chrome instance if it is already running.

USAGE
-----

Past is similar to `pass`, but is not functionally equivalent. The following
subcommands are supported, but running `past -h` will gives you more help.

```
         init - Creates or re-encrypts a password-store with GPG keys.
         list - Prints the names of all password files.
         show - Decrypt a password file to print the password.
       insert - Inserts a password to the store in a new password file.
     generate - Inserts a new password file with auto-generated password.
         scan - Decrypts all files to search for a string or regexp.

          git - Runs git(1) command on the password store.
```
