'use strict';

//
// Startup functions in the initialization order.
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
// Other common functions
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

function showPage(page, id, callback) {
  page.id = id;
  page.style.display = "";
  document.body.replaceChild(page, document.body.firstElementChild);
  callback(page);
}
