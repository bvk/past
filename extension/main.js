'use strict';

//
// Startup functions
//

let activeTab;

document.addEventListener('DOMContentLoaded', function () {
  chrome.tabs.getSelected(null, function(tab) {
    activeTab = tab;
    onActiveTabReady();
  });
});

let backgroundPage;

function onActiveTabReady() {
  chrome.runtime.getBackgroundPage(function(page) {
    if (page) {
      backgroundPage = page;
      onBackgroundPageReady();
    }
  });
}

function onBackgroundPageReady() {
  // Issue a backend request through background.js so it won't be canceled in
  // the middle when gpg askpass window makes the chrome popup to disappear.
  let req = { list_files: {} }
  backgroundPage.listFiles(req, function (resp) {
    onListFilesResponse(req, resp);
  });
}

let passwordFiles;

function onListFilesResponse(req, resp) {
  if (!resp) {
    setOperationStatus("Could not query for password file names.");
    return;
  }

  if (resp.status != "") {
    setOperationStatus("Password files query has failed (" + resp.status+").");
    return;
  }

  passwordFiles = [];
  if (resp.list_files && resp.list_files.files) {
    passwordFiles = resp.list_files.files;
  }

  backgroundPage.getLocalStorage(["persistentState"], function(result) {
    onGetPersistentState(result);
  });
}

let persistentState;

function onGetPersistentState(result) {
  persistentState = {};
  if (result.persistentState) {
    persistentState = result.persistentState;
  }
  if (!persistentState.fileCountMap) {
    persistentState.fileCountMap = {};
  }

  // Add new password files to the fileCountMap.
  let files = {}
  for (let i = 0; i < passwordFiles.length; i++) {
    let file = passwordFiles[i];
    files[file] = true;
    if (!(file in persistentState.fileCountMap)) {
      persistentState.fileCountMap[file] = 0;
    }
  }

  // Remove deleted password files from the fileCountMap.
  for (let key in persistentState.fileCountMap) {
    if (!(key in files)) {
      delete persistentState.fileCountMap[key];
    }
  }

  console.log("loaded state", persistentState, passwordFiles);

  let searchPage = createSearchPage();
  showPage(searchPage, "search-page", onSearchPageDisplay);
}

//
// Search Page functions
//

function onSearchPageDisplay(page) {
  setOperationStatus("Backend is ready.");
  setSearchPageRecentListItems(page, orderPasswordFiles(""));

  var searchBars = page.getElementsByClassName("search-bar");
  if (searchBars) {
    searchBars[0].focus();
  }
}

function onSearchPageSearchBar(page, searchInput) {
  let files = orderPasswordFiles(searchInput.value);
  setSearchPageRecentListItems(page, files);
}

function onSearchPageAddButton(page, addButton) {
  let addPage = createAddPage();
  showPage(addPage, "add-page", onAddPageDisplay);
}

function onSearchPageCopyButton(page, copyButton) {
  let name = copyButton.parentElement.childNodes[1].textContent;
  if (name) {
    let req = {view_file:{file:name}};
    backgroundPage.viewFile(req, function (resp) {
      onSearchPageViewFileResponse(page, req, resp);
    });
  }
}

function onSearchPageViewFileResponse(page, req, resp) {
  if (!resp) {
    setOperationStatus("Could not perform Copy password operation.");
    return;
  }
  if (resp.status != "") {
    setOperationStatus("Copy password operation has failed ("+resp.status+").");
    return;
  }

  let whenCleared = function() {
    setOperationStatus("Password is cleared from the clipboard.");
  };

  if (backgroundPage.copyPassword(resp.view_file.password, whenCleared)) {
    setOperationStatus("Password is copied to the clipboard.");
  } else {
    setOperationStatus("Cloud not copy the password to clipboard.");
  }

  persistentState.fileCountMap[req.view_file.file] += 1;

  console.log("saving state", persistentState);
  backgroundPage.setLocalStorage({"persistentState": persistentState});
}

function setSearchPageRecentListItems(page, names) {
  var elems = document.getElementsByClassName("search-page-password-name");
  for (let i = 0; i < elems.length; i++) {
    if (i < names.length) {
      elems[i].textContent = names[i];
      elems[i].nextElementSibling.disabled = false;
    } else {
      elems[i].textContent = "";
      elems[i].nextElementSibling.disabled = true;
    }
  }
}

//
// Add Page functions
//

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

//
// Other functions
//

function sortPasswordFiles() {
  if (!persistentState || !persistentState.fileCountMap) {
    return passwordFiles.slice()
  }

  let fileCounts = [];
  for (let i = 0; i < passwordFiles.length; i++) {
    let count = 0;
    let file = passwordFiles[i];
    if (file in persistentState.fileCountMap) {
      count = persistentState.fileCountMap[file];
    }
    fileCounts.push([file, count]);
  }

  fileCounts.sort(function(a, b) {
    return b[1] - a[1];
  });

  let files = [];
  for (let i = 0; i < fileCounts.length; i++) {
    files.push(fileCounts[i][0]);
  }
  return files
}

function orderPasswordFiles(search) {
  let hosts = hostnameSuffixes();
  let sortedFiles = sortPasswordFiles();

  //
  // If search string is empty, we want to bring hostname-matches first
  // followed by all the rest in most-used order. Otherwise, i.e., if search
  // string is non-empty, we only want to show search-maches and
  // hostname-matches and nothing else.
  //

  let files = [];
  let skipped = [];
  for (let i = 0; i < sortedFiles.length; i++) {
    if (search != "" && sortedFiles[i].includes(search)) {
      files.push(sortedFiles[i]);
      continue;
    }

    let added = false;
    for (let j = 0; j < hosts.length; j++) {
      if (sortedFiles[i].includes(hosts[j])) {
        files.push(sortedFiles[i]);
        added = true;
        break
      }
    }

    if (!added) {
      skipped.push(sortedFiles[i]);
    }
  }

  let seen = {};
  let dedup = [];
  for (let i = 0; i < files.length; i++) {
    if (seen[files[i]]) {
      continue;
    }
    seen[files[i]] = true;
    dedup.push(files[i]);
  }

  if (search == "") {
    dedup = dedup.concat(skipped)
  }
  return dedup;
}

function hostnameSuffixes() {
  if (!activeTab || !activeTab.url) {
    return []
  }

  var a = document.createElement('a');
  a.href = activeTab.url;

  let hs = [a.hostname];
  if (a.hostname != a.host) {
    hs.push(a.host); // a.host could include the port number
  }

  if (isIP(a.hostname)) {
    console.log(hs);
    return hs;
  }

  // Split the hostnames into all suffixes with dot character.
  var suffixes = [];
  for (let i = 0; i < hs.length; i++) {
    let words = hs[i].split(".");
    for (let j = 0; j < words.length-1; j++) {
      let ws = words.slice(j);
      suffixes.push(ws.join("."));
    }
  }

  console.log(suffixes);
  return suffixes;
}

function isIP(hostname)
{
 if (/^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/.test(hostname))
  {
    return true
  }
  return false
}

//
// Status bar functions
//

function clearOperationStatus() {
  let elems = document.getElementsByClassName("footer-status");
  for (let i = 0; i < elems.length; i++) {
    elems[i].textContent = "";
  }
}

function setOperationStatus(message) {
  let elems = document.getElementsByClassName("footer-status");
  for (let i = 0; i < elems.length; i++) {
    elems[i].textContent = message;
  }
}

//
// Page functions
//

function showPage(page, id, callback) {
  page.id = id;
  page.style.display = "";
  document.body.replaceChild(page, document.body.firstElementChild);
  callback(page);
}

function createSearchPage() {
  let searchPageTemplate = document.getElementById("search-page-template");
  let page = searchPageTemplate.cloneNode(true);

  let copyButtons = page.getElementsByClassName("copy-button");
  for (let i = 0; i < copyButtons.length; i++) {
    let button = copyButtons[i];
    button.addEventListener("click", function() {
      onSearchPageCopyButton(page, button);
    });
  }

  let addButtons = page.getElementsByClassName("add-button");
  for (let i = 0; i < addButtons.length; i++) {
    let button = addButtons[i];
    button.addEventListener("click", function() {
      onSearchPageAddButton(page, button);
    });
  }

  let searchBar = page.getElementsByClassName("search-bar")[0];
  searchBar.addEventListener("input", function() {
    onSearchPageSearchBar(page, searchBar);
  });

  return page;
}

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
