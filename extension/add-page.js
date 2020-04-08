'use strict';

function createAddPage() {
  let addPageTemplate = document.getElementById("add-page-template");
  let page = addPageTemplate.cloneNode(true);

  let backButton = page.getElementsByClassName("add-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onAddPageBackButton(page, backButton);
  });

  let doneButton = page.getElementsByClassName("add-page-done-button")[0];
  doneButton.addEventListener("click", function() {
    onAddPageDoneButton(page, doneButton);
  });

  let sitename = page.getElementsByClassName("add-page-sitename")[0];
  sitename.addEventListener("input", function() {
    onAddPageSitenameChange(page, sitename);
  });

  let username = page.getElementsByClassName("add-page-username")[0];
  username.addEventListener("input", function() {
    onAddPageUsernameChange(page, username);
  });

  let pass = page.getElementsByClassName("add-page-password")[0];
  pass.addEventListener("input", function() {
    onAddPagePasswordChange(page, pass);
  });

  let repeat = page.getElementsByClassName("add-page-repeat-password")[0];
  repeat.addEventListener("input", function() {
    onAddPageRepeatPasswordChange(page, repeat);
  });

  let passType = page.getElementsByClassName("add-page-password-type")[0];
  passType.addEventListener("change", function() {
    onAddPagePasswordType(page, passType);
  });

  let passSize = page.getElementsByClassName("add-page-password-size")[0];
  passSize.addEventListener("wheel", function(event) {
    onAddPagePasswordSize(page, passSize, event);
  });

  let passToggle = page.getElementsByClassName("add-page-password-toggle")[0];
  passToggle.addEventListener("click", function() {
    onAddPagePasswordToggle(page, passToggle);
  });

  let copyButton = page.getElementsByClassName("add-page-password-copy")[0];
  copyButton.addEventListener("click", function() {
    onAddPageCopyButton(page, copyButton);
  });

  let generateButton = page.getElementsByClassName("add-page-password-generate")[0];
  generateButton.addEventListener("click", function() {
    onAddPageGenerateButton(page, generateButton);
  });

  return page;
}

function onAddPageDisplay(page) {
  let hostname = "";
  if (activeTab && activeTab.url) {
    var a = document.createElement('a');
    a.href = activeTab.url;

    hostname = a.hostname;
    if (a.hostname != a.host) {
      hostname = a.host; // a.host could include the port number
    }
    if (!hostname.includes(".")) {
      hostname = "";
    }
  }

  var sitename = page.getElementsByClassName("add-page-sitename")[0];
  sitename.value = hostname;

  // Move focus to the first empty element.
  if (sitename.value == "") {
    sitename.focus();
  } else {
    var username = page.getElementsByClassName("add-page-username")[0];
    username.focus();
  }
}

function onAddPageSitenameChange(page, sitenameInput) {
  autoAddPageDoneButton(page);
}

function onAddPageUsernameChange(page, usernameInput) {
  autoAddPageDoneButton(page);
}

function onAddPageBackButton(page, backButton) {
  let searchPage = createSearchPage();
  showPage(searchPage, "search-page", onSearchPageDisplay);
}

function onAddPageDoneButton(page, doneButton) {
  let sitename = page.getElementsByClassName("add-page-sitename")[0];
  let username = page.getElementsByClassName("add-page-username")[0];
  let password = getAddPagePassword(page);
  let moredata = page.getElementsByClassName("")[0];

  if (password == "" || username.value == "" || sitename.value == "") {
    return;
  }

  // FIXME: We must handle multiple user accounts per site.

  let req = {
    add_file: {
      file: sitename.value,
      sitename: sitename.value,
      username: username.value,
      password: password,
    },
  };

  // TODO: Also add moredata to the rest.

  // Issue add-password request through the background page.
  backgroundPage.addFile(req, function(resp) {
    onAddPageAddFileResponse(page, req, resp);
  });
}

function onAddPageAddFileResponse(page, req, resp) {
  if (!req) {
    console.log("error: add file request cannot be null");
    return;
  }
  if (!resp) {
    setOperationStatus("Could not issue add file request.");
    return;
  }
  if (resp.status != "") {
    setOperationStatus("Add file request has failed ("+resp.status+").");
    return;
  }

  let file = req.add_file.file;
  passwordFiles.push(file);
  persistentState.fileCountMap[file] = 1;
  backgroundPage.setLocalStorage({"persistentState": persistentState});

  let searchPage = createSearchPage();
  showPage(searchPage, "search-page", onSearchPageDisplay);
}

function onAddPageCopyButton(page, copyButton) {
  let password = getAddPagePassword(page);
  if (password == "") {
    return;
  }

  let whenCleared = function() {
    setOperationStatus("Password is cleared from the clipboard.");
  };

  if (backgroundPage.copyPassword(password, whenCleared)) {
    setOperationStatus("Password is copied to the clipboard.");
  } else {
    setOperationStatus("Cloud not copy the password to clipboard.");
  }
}

function onAddPageGenerateButton(page, elem) {
  generateAddPagePassword(page);
}

function onAddPagePasswordChange(page, pass) {
  autoAddPagePasswordColors(page);
  autoAddPageCopyButton(page);
  autoAddPageDoneButton(page);
}

function onAddPageRepeatPasswordChange(page, repeat) {
  autoAddPagePasswordColors(page);
  autoAddPageCopyButton(page);
  autoAddPageDoneButton(page);
}

function autoAddPageDoneButton(page, doneButton) {
  let sitename = page.getElementsByClassName("add-page-sitename")[0];
  let username = page.getElementsByClassName("add-page-username")[0];
  let password = getAddPagePassword(page);

  let disable = false;
  if (password == "" || username.value == "" || sitename.value == "") {
    disable = true;
  }

  if (!doneButton) {
    doneButton = page.getElementsByClassName("add-page-done-button")[0];
  }
  doneButton.disabled = disable;
}

function autoAddPagePasswordColors(page, pass, repeat) {
  // TODO: change password field colors as necessary.
}

// autoAddPageGenerateButton enables or disables the generate password button.
function autoAddPageGenerateButton(page, generateButton) {
  let passType = currentAddPagePasswordType(page);
  let disable = true;
  if (passType != "") {
    disable = false;
  }

  if (!generateButton) {
    generateButton = page.getElementsByClassName("add-page-password-generate")[0];
  }
  generateButton.disabled = disable;
}

function autoAddPageCopyButton(page, copyButton) {
  let disable = true;
  let password = getAddPagePassword(page);
  if (password != "") {
    disable = false;
  }

  if (!copyButton) {
    copyButton = page.getElementsByClassName("add-page-password-copy")[0];
  }
  copyButton.disabled = disable;
}

function autoAddPagePasswordSize(page, sizeElem) {
  if (!page) {
    page = document.getElementById("page")
  }
  let passType = currentAddPagePasswordType(page);

  if (!sizeElem) {
    sizeElem = page.getElementsByClassName("add-page-password-size")[0];
  }
  if (passType == "") {
    sizeElem.style.setProperty("text-decoration", "line-through");
  } else {
    sizeElem.style.setProperty("text-decoration", "");
  }
}

function onAddPagePasswordToggle(page, toggleButton) {
  let pass = page.getElementsByClassName("add-page-password")[0];
  if (pass.type == "text") {
    pass.type = "password";
  } else {
    pass.type = "text";
  }

  let repeat = page.getElementsByClassName("add-page-repeat-password")[0];
  if (repeat.type == "text") {
    repeat.type = "password";
  } else {
    repeat.type = "text";
  }

  if (toggleButton.textContent == "visibility_off") {
    toggleButton.textContent = "visibility";
  } else {
    toggleButton.textContent = "visibility_off";
  }
}

// onAddPagePasswordSize is invoked on scroll wheel even on the password size.
function onAddPagePasswordSize(page, sizeElem, event) {
  if (!event || !event.deltaY) {
    return;
  }
  let passSize = page.getElementsByClassName("add-page-password-size")[0];
  let words = passSize.textContent.split(" ");
  let value = parseInt(words[0], 10);
  if (value < 32 && event.deltaY > 0) {
    value += 1;
  } else if (value > 3 && event.deltaY < 0) {
    value -= 1;
  }
  words[0] = value;
  passSize.textContent = words.join(" ");

  generateAddPagePassword(page);
  autoAddPageCopyButton(page);
  autoAddPageGenerateButton(page);
}

// onAddPagePasswordType is invoked when password type selection is changed.
function onAddPagePasswordType(page, typeSelect) {
  let disablePasswords = true;
  if (typeSelect.value == "") {
    disablePasswords = false;
  }

  // Toggle enabled/disabled flag on the password fields.
  var pass = page.getElementsByClassName("add-page-password")[0];
  pass.disabled = disablePasswords;

  var repeat = page.getElementsByClassName("add-page-repeat-password")[0];
  repeat.disabled = disablePasswords;

  if (disablePasswords) {
    generateAddPagePassword(page);
  } else {
    setAddPagePassword(page, "", "");
  }

  autoAddPageDoneButton(page);
  autoAddPageCopyButton(page);
  autoAddPagePasswordSize(page);
  autoAddPageGenerateButton(page);
}

function currentAddPagePasswordType(page, typeSelect) {
  if (!typeSelect) {
    typeSelect = page.getElementsByClassName("add-page-password-type")[0];
  }
  return typeSelect.value
}

function currentAddPagePasswordSize(page) {
  let size = page.getElementsByClassName("add-page-password-size")[0];
  let words = size.textContent.split(" ")
  return parseInt(words[0], 10);
}

function generateAddPagePassword(page) {
  let size = currentAddPagePasswordSize(page);
  let passType = currentAddPagePasswordType(page);
  if (passType == "") {
    autoAddPageCopyButton(page);
    autoAddPageGenerateButton(page);
    return;
  }

  let numbers = "0123456789";
  let letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";
  let symbols = "`!#$%&'()*+,\-./:;<=>?@[]^_{|}~" + '"';
  let base64std = letters+"+/";

  let password = "";
  if (passType == "numbers") {
    password = pwgen(numbers, size);
  } else if (passType == "letters") {
    password = pwgen(letters, size)
  } else if (passType == "letters_numbers") {
    password = pwgen(letters+numbers, size)
  } else if (passType == "letters_numbers_symbols") {
    password = pwgen(letters+numbers+symbols, size);
  } else if (passType == "base64std") {
    password = pwgen(base64std, size);
  } else {
    console.log("error: unhandled password generation type "+passType);
    return;
  }

  setAddPagePassword(page, password, password);
  autoAddPageCopyButton(page);
  autoAddPageGenerateButton(page);
}

function getAddPagePassword(page, pass, repeat) {
  if (!pass) {
    pass = page.getElementsByClassName("add-page-password")[0];
  }
  if (!repeat) {
    repeat = page.getElementsByClassName("add-page-repeat-password")[0];
  }
  if (pass.value != repeat.value) {
    return "";
  }
  return pass.value;
}

function setAddPagePassword(page, first, second) {
  let pass = page.getElementsByClassName("add-page-password")[0];
  pass.value = first;

  let repeatPass = page.getElementsByClassName("add-page-repeat-password")[0];
  repeatPass.value = second;
}

function pwgen(charset, length) {
  let result = "";
  for (var i = 0, n = charset.length; i < length; ++i) {
    result += charset.charAt(Math.floor(Math.random() * n));
  }
  return result;
}
