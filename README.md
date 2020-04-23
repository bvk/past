OVERVIEW
--------

This package reimplements the password-store management command `pass` under a
new name `past`. Additionally, it comes with in-built support for an browser
extension for Google Chrome and Chromium. As of now, only Mac OS X and
GNU/Linux are supported.

You should already be familiar with password-store package; otherwise, this
package may not be useful to you. See
[passwordstore.org](https://passwordstore.org) to learn about password-stores.

DEPENDENCIES
------------

Past requires `git` and `gpg` tools (GPGTools on Mac OS X) for normal
operations. Please make sure that `git` and `gpg` tools are available in the
`PATH` directories `$HOME/bin:/usr/local/bin:/usr/bin:/bin`.

You also need to have `go` version `1.12` or above for installation.

Browser extension does not use any external javascript libraries so that
security footprint is minimized. However, it uses external Google's Material
Icons library for buttons.

INSTALLATION
------------

Use the following instructions to install this package.

```
$ go get github.com/bvk/past
$ ~/go/bin/past install
```

First command compiles and installs the `past` command-line tool, and second
command configures the browser extension backend and opens the chrome web store
URL where users can install the extension.

USAGE
-----

Past is similar to `pass`, but is not a drop-in replacement. The following
subcommands are supported.

```
  edit        Updates an existing password-file with external editor.
  generate    Inserts a new password-file with an auto-generated password.
  git         Runs git(1) command on the password-store repository.
  import      Imports passwords from other password managers' data files.
  init        Creates or re-encrypts a password-store with GPG keys.
  insert      Inserts a password to the in a new password-file.
  install     Installs the backend for browser extension.
  keys        Prints GPG public keys information.
  list        Prints the names of all password-files.
  scan        Decrypts all files to search for a string or regexp.
  show        Decrypts a password-file and prints it's content.
```

Browser extension enables most of the password-store operations and a few GPG
keyring operations. Following is the list of operations browser extension can
perform:

1. Initialize the GPG keyring, including create, delete, import and export
encryption keys.

2. Create a new password-store or import an exiting password-store from remote
git repositories.

3. Search, view, add, remove and edit password-store entries.

4. Add or remote GPG keys to password-stores

5. Sync local password-store changes to the remote or sync changes from the
remote password-store.

Also, note that passwords copied into the clipboard are cleared after 10
seconds automatically.

SCREENSHOTS
-----------

![Search Passwords Page](extras/search-passwords.png?raw=true)
![Add  Password Page](extras/add-password.png?raw=true)
