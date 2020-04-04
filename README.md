OVERVIEW
--------

This package reimplements the password-store management command `pass` under a
new name `past`. Additionally, it comes with in-built support for a browser
extension for Google Chrome. As of now, only GNU/Linux is supported.

You should already be a password-store user to use this package; otherwise, the
functionality provided in this package is of no use to you. See
passwordstore.org to learn more.

DEPENDENCIES
------------

Past requires `git` and `gpg` tools for normal operations. You need to have
`go` version `1.12` or above for installation.

INSTALLATION
------------

Use the following instructions to install this package.

```
$ go get github.com/bvk/past
$ ~/go/bin/past install
```

First command compiles and installs the `past` command-line tool and second
command configures the backend necessary for browser extension and opens the
chrome web store URL where users can install the extension.

USAGE
-----

Past is similar to `pass`, but is not drop-in replacement. The following
subcommands are supported.

```
  edit        Updates an existing password-file with external editor.
  generate    Inserts a new password-file with an auto-generated password.
  init        Creates or re-encrypts a password-store with GPG keys.
  insert      Inserts a password to the in a new password-file.
  install     Installs the backend for browser extension.
  keys        Prints GPG public keys information.
  list        Prints the names of all password-files.
  scan        Decrypts all files to search for a string or regexp.
  show        Decrypts a password-file and prints it's content.
  git         Runs git(1) command on the password-store repository.
```

Browser extension is designed to be as minimal as possible. Users cannot create new
password-store entries through the extension, but can only copy the password to
the clipboard -- which is cleared after 10 seconds.

SCREENSHOT
----------

![Extension Popup](extras/screenshot.png?raw=true)
