OVERVIEW
--------

This package reimplements the password-store management command `pass` under a
new name `past`. Additionally, it comes with in-built support for a browser
extension for Google Chrome or Chromium. As of now, only Mac OS X and GNU/Linux
are supported.

You should already be a password-store user to use this package; otherwise,
functionality provided in this package may not be useful to you. See
[passwordstore.org](https://passwordstore.org) to learn about password-store.

DEPENDENCIES
------------

Past requires `git` and `gpg` tools for normal operations. You need to have
`go` version `1.12` or above for installation.

Please make sure that `git` and `gpg` tools are available in one of the default
`PATH` directories (`/bin:/usr/bin:/usr/local/bin`).

INSTALLATION
------------

Use the following instructions to install this package.

```
$ go get github.com/bvk/past
$ ~/go/bin/past install
```

First command compiles and installs the `past` command-line tool, and second
command configures the backend necessary for browser extension and opens the
chrome web store URL where users can install the extension.

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

Browser extension is designed to be as minimal as possible. As of now, users
cannot create password-stores or new entries using the extension, but can only
copy passwords to the clipboard. They are cleared after 10 seconds
automatically.

SCREENSHOT
----------

![Extension Popup](extras/screenshot.png?raw=true)
