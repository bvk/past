'use strict';

// Edit page instance is used for both New Password and Edit Password pages.
//
// In the New-Password view will have created with page empty data and user can
// edit the website name, but in the Edit-Password view website name is marked
// read-only.
//
// Parameter params is expected to be in the following format:
//
// params = {
//   username: "user",
//   password: "pass",
//   sitename: "website.com",
//   data: "key1:value1\nk2:value2",
// }
function createEditPage(params) {
  let editPageTemplate = document.getElementById("edit-page-template");
  let page = editPageTemplate.cloneNode(true);
  page.setAttribute("page-params", "{}");

  // Save the page parameters in the page instance as a string.
  if (params) {
    console.log("params", params);
    page.setAttribute("page-params", JSON.stringify(params));
  }

  console.log(page.getAttribute("page-params"));

  let title = page.getElementsByClassName("edit-page-title")[0];
  title.textContent = (params ? "Edit Password" : "New Password");

  let backButton = page.getElementsByClassName("edit-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onEditPageBackButton(page, backButton);
  });

  let closeButton = page.getElementsByClassName("edit-page-close-button")[0];
  closeButton.addEventListener("click", function() {
    onEditPageCloseButton(page, closeButton);
  });

  let undoButton = page.getElementsByClassName("edit-page-undo-button")[0];
  undoButton.addEventListener("click", function() {
    onEditPageUndoButton(page, undoButton);
  });

  let doneButton = page.getElementsByClassName("edit-page-done-button")[0];
  doneButton.addEventListener("click", function() {
    onEditPageDoneButton(page, doneButton);
  });

  let sitename = page.getElementsByClassName("edit-page-sitename")[0];
  sitename.addEventListener("input", function() {
    onEditPageSitenameChange(page, sitename);
  });

  let username = page.getElementsByClassName("edit-page-username")[0];
  username.addEventListener("input", function() {
    onEditPageUsernameChange(page, username);
  });

  let pass = page.getElementsByClassName("edit-page-password")[0];
  pass.addEventListener("input", function() {
    onEditPagePasswordChange(page, pass);
  });

  let repeat = page.getElementsByClassName("edit-page-repeat-password")[0];
  repeat.addEventListener("input", function() {
    onEditPageRepeatPasswordChange(page, repeat);
  });

  let data = page.getElementsByClassName("edit-page-data")[0];
  data.addEventListener("input", function() {
    onEditPageDataChange(page, data);
  });

  let passType = page.getElementsByClassName("edit-page-password-type")[0];
  passType.addEventListener("change", function() {
    onEditPagePasswordType(page, passType);
  });

  let passSize = page.getElementsByClassName("edit-page-password-size")[0];
  passSize.addEventListener("wheel", function(event) {
    onEditPagePasswordSize(page, passSize, event);
  });

  let passToggle = page.getElementsByClassName("edit-page-password-toggle")[0];
  passToggle.addEventListener("click", function() {
    onEditPagePasswordToggle(page, passToggle);
  });

  let copyButton = page.getElementsByClassName("edit-page-password-copy")[0];
  copyButton.addEventListener("click", function() {
    onEditPageCopyButton(page, copyButton);
  });

  let generateButton = page.getElementsByClassName("edit-page-password-generate")[0];
  generateButton.addEventListener("click", function() {
    onEditPageGenerateButton(page, generateButton);
  });

  return page;
}

function onEditPageDisplay(page) {
  onEditPageUndoButton(page);

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

  var sitename = page.getElementsByClassName("edit-page-sitename")[0];
  if (sitename.value == "") {
    sitename.value = hostname;
  }

  // Move focus to the first empty element.
  if (sitename.value == "") {
    sitename.focus();
  } else {
    var username = page.getElementsByClassName("edit-page-username")[0];
    username.focus();
  }

  autoEditPageUndoButton(page);
}

function onEditPageSitenameChange(page, sitenameInput) {
  autoEditPageDoneButton(page);
  autoEditPageUndoButton(page);
}

function onEditPageUsernameChange(page, usernameInput) {
  autoEditPageDoneButton(page);
  autoEditPageUndoButton(page);
}

function onEditPageBackButton(page, backButton) {
  let searchPage = createSearchPage();
  showPage(searchPage, "search-page", onSearchPageDisplay);
}

function onEditPageCloseButton(page, closeButton) {
  window.close();
}

function onEditPageUndoButton(page, undoButton) {
  let username = "";
  let sitename = "";
  let password = "";
  let data = "";
  let pageParams = page.getAttribute("page-params");
  if (pageParams != "{}") {
    let params = JSON.parse(pageParams);
    username = params.username;
    password = params.password;
    sitename = params.sitename;
    data = params.data;
  }

  page.getElementsByClassName("edit-page-sitename")[0].value = sitename;
  page.getElementsByClassName("edit-page-username")[0].value = username;
  page.getElementsByClassName("edit-page-password-type")[0].value = "";
  page.getElementsByClassName("edit-page-password")[0].value = password;
  page.getElementsByClassName("edit-page-repeat-password")[0].value = password;
  page.getElementsByClassName("edit-page-data")[0].value = data;

  if (!undoButton) {
    undoButton = page.getElementsByClassName("edit-page-undo-button")[0];
  }
  undoButton.disabled = true;

  autoEditPageDoneButton(page);
  autoEditPageCopyButton(page);
  autoEditPagePasswordSize(page);
  autoEditPageGenerateButton(page);
}

function onEditPageDoneButton(page, doneButton) {
  let sitename = page.getElementsByClassName("edit-page-sitename")[0];
  let username = page.getElementsByClassName("edit-page-username")[0];
  let password = getEditPagePassword(page);
  let data = page.getElementsByClassName("edit-page-data")[0];
  if (password == "" || username.value == "" || sitename.value == "") {
    return;
  }

  // FIXME: We must handle multiple user accounts per site.

  let pageParams = page.getAttribute("page-params");
  let origFile = "";
  if (pageParams != "{}") {
    origFile = JSON.parse(pageParams).sitename;
  }

  let req = {};
  if (origFile != "") {
    req = {
      edit_file: {
        file: sitename.value,
        orig_file: origFile,

        sitename: sitename.value,
        username: username.value,
        password: password,
        data: data.value,
      },
    };
  } else {
    req = {
      add_file: {
        file: sitename.value,
        sitename: sitename.value,
        username: username.value,
        password: password,
        data: data.value,
      },
    };
  };

  // TODO: Also add moredata to the rest.

  // Issue add-password request through the background page.
  backgroundPage.addFile(req, function(resp) {
    if (!req) {
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
    onEditPageAddFileResponse(page, req, resp);
  });
}

function onEditPageAddFileResponse(page, req, resp) {
  if (req.add_file) {
    let file = req.add_file.file;
    passwordFiles.push(file);
    persistentState.fileCountMap[file] = 1;
    backgroundPage.setLocalStorage({"persistentState": persistentState});
  }

  let searchPage = createSearchPage();
  showPage(searchPage, "search-page", onSearchPageDisplay);
}

function onEditPageCopyButton(page, copyButton) {
  let password = getEditPagePassword(page);
  if (password == "") {
    return;
  }

  let whenCleared = function() {
    setOperationStatus("Password is cleared from the clipboard.");
  };

  if (backgroundPage.copyString(password, 10, whenCleared)) {
    setOperationStatus("Password is copied to the clipboard.");
  } else {
    setOperationStatus("Cloud not copy the password to clipboard.");
  }
}

function onEditPageGenerateButton(page, elem) {
  generateEditPagePassword(page);
}

function onEditPagePasswordChange(page, pass) {
  autoEditPagePasswordColors(page);
  autoEditPageCopyButton(page);
  autoEditPageDoneButton(page);
  autoEditPageUndoButton(page);
}

function onEditPageRepeatPasswordChange(page, repeat) {
  autoEditPagePasswordColors(page);
  autoEditPageCopyButton(page);
  autoEditPageDoneButton(page);
  autoEditPageUndoButton(page);
}

function onEditPageDataChange(page, data) {
  autoEditPagePasswordColors(page);
  autoEditPageCopyButton(page);
  autoEditPageDoneButton(page);
  autoEditPageUndoButton(page);
}

function autoEditPageUndoButton(page, undoButton) {
  if (!undoButton) {
    undoButton = page.getElementsByClassName("edit-page-undo-button")[0];
  }

  let initUsername = "";
  let initSitename = "";
  let initPassword = "";
  let initData = "";
  let pageParams = page.getAttribute("page-params");
  if (pageParams != "{}") {
    let params = JSON.parse(pageParams);
    initUsername = params.username;
    initPassword = params.password;
    initSitename = params.sitename;
    initData = params.data;
  }

  let sitename = page.getElementsByClassName("edit-page-sitename")[0];
  let username = page.getElementsByClassName("edit-page-username")[0];
  let password = page.getElementsByClassName("edit-page-password")[0];
  let repeated = page.getElementsByClassName("edit-page-repeat-password")[0];
  let data = page.getElementsByClassName("edit-page-data")[0];

  let disable = true;
  if (password.value != initPassword || sitename.value != initSitename ||
      username.value != initUsername || repeated.value != initPassword ||
      data.value != initData) {
    disable = false;
  }
  undoButton.disabled = disable;
}

function autoEditPageDoneButton(page, doneButton) {
  let sitename = page.getElementsByClassName("edit-page-sitename")[0];
  let username = page.getElementsByClassName("edit-page-username")[0];
  let password = getEditPagePassword(page);
  let data = page.getElementsByClassName("edit-page-data")[0];

  let disable = false;
  if (password == "" || username.value == "" || sitename.value == "") {
    disable = true;
  }

  if (!disable) {
    let pageParams = page.getAttribute("page-params");
    if (pageParams != "{}") {
      let params = JSON.parse(pageParams);
      if (password == params.password && sitename.value == params.sitename &&
          username.value == params.username && data.value == params.data) {
        disable = true;
      }
    }
  }

  if (!doneButton) {
    doneButton = page.getElementsByClassName("edit-page-done-button")[0];
  }
  doneButton.disabled = disable;
}

function autoEditPagePasswordColors(page, pass, repeat) {
  // TODO: change password field colors as necessary.
}

// autoEditPageGenerateButton enables or disables the generate password button.
function autoEditPageGenerateButton(page, generateButton) {
  let passType = currentEditPagePasswordType(page);
  let disable = true;
  if (passType != "") {
    disable = false;
  }

  if (!generateButton) {
    generateButton = page.getElementsByClassName("edit-page-password-generate")[0];
  }
  generateButton.disabled = disable;
}

function autoEditPageCopyButton(page, copyButton) {
  let disable = true;
  let password = getEditPagePassword(page);
  if (password != "") {
    disable = false;
  }

  if (!copyButton) {
    copyButton = page.getElementsByClassName("edit-page-password-copy")[0];
  }
  copyButton.disabled = disable;
}

function autoEditPagePasswordSize(page, sizeElem) {
  if (!page) {
    page = document.getElementById("page")
  }
  let passType = currentEditPagePasswordType(page);

  if (!sizeElem) {
    sizeElem = page.getElementsByClassName("edit-page-password-size")[0];
  }
  if (passType == "") {
    sizeElem.style.setProperty("text-decoration", "line-through");
  } else {
    sizeElem.style.setProperty("text-decoration", "");
  }
}

function onEditPagePasswordToggle(page, toggleButton) {
  let pass = page.getElementsByClassName("edit-page-password")[0];
  if (pass.type == "text") {
    pass.type = "password";
  } else {
    pass.type = "text";
  }

  let repeat = page.getElementsByClassName("edit-page-repeat-password")[0];
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

// onEditPagePasswordSize is invoked on scroll wheel even on the password size.
function onEditPagePasswordSize(page, sizeElem, event) {
  if (!event || !event.deltaY) {
    return;
  }
  let passSize = page.getElementsByClassName("edit-page-password-size")[0];
  let words = passSize.textContent.split(" ");
  let value = parseInt(words[0], 10);
  if (value < 32 && event.deltaY > 0) {
    value += 1;
  } else if (value > 3 && event.deltaY < 0) {
    value -= 1;
  }
  words[0] = value;
  passSize.textContent = words.join(" ");

  generateEditPagePassword(page);
  autoEditPageCopyButton(page);
  autoEditPageGenerateButton(page);
}

// onEditPagePasswordType is invoked when password type selection is changed.
function onEditPagePasswordType(page, typeSelect) {
  let disablePasswords = true;
  if (typeSelect.value == "") {
    disablePasswords = false;
  }

  // Toggle enabled/disabled flag on the password fields.
  var pass = page.getElementsByClassName("edit-page-password")[0];
  pass.disabled = disablePasswords;

  var repeat = page.getElementsByClassName("edit-page-repeat-password")[0];
  repeat.disabled = disablePasswords;

  if (disablePasswords) {
    generateEditPagePassword(page);
  } else {
    setEditPagePassword(page, "", "");
  }

  autoEditPageDoneButton(page);
  autoEditPageUndoButton(page);
  autoEditPageCopyButton(page);
  autoEditPagePasswordSize(page);
  autoEditPageGenerateButton(page);
}

function currentEditPagePasswordType(page, typeSelect) {
  if (!typeSelect) {
    typeSelect = page.getElementsByClassName("edit-page-password-type")[0];
  }
  return typeSelect.value
}

function currentEditPagePasswordSize(page) {
  let size = page.getElementsByClassName("edit-page-password-size")[0];
  let words = size.textContent.split(" ")
  return parseInt(words[0], 10);
}

function generateEditPagePassword(page) {
  let size = currentEditPagePasswordSize(page);
  let passType = currentEditPagePasswordType(page);
  if (passType == "") {
    autoEditPageCopyButton(page);
    autoEditPageGenerateButton(page);
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

  setEditPagePassword(page, password, password);
  autoEditPageCopyButton(page);
  autoEditPageGenerateButton(page);
}

function getEditPagePassword(page, pass, repeat) {
  if (!pass) {
    pass = page.getElementsByClassName("edit-page-password")[0];
  }
  if (!repeat) {
    repeat = page.getElementsByClassName("edit-page-repeat-password")[0];
  }
  if (pass.value != repeat.value) {
    return "";
  }
  return pass.value;
}

function setEditPagePassword(page, first, second) {
  let pass = page.getElementsByClassName("edit-page-password")[0];
  pass.value = first;

  let repeatPass = page.getElementsByClassName("edit-page-repeat-password")[0];
  repeatPass.value = second;
}

function pwgen(charset, length) {
  let result = "";
  for (var i = 0, n = charset.length; i < length; ++i) {
    result += charset.charAt(Math.floor(Math.random() * n));
  }
  return result;
}
