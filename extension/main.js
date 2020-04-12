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

  // let settingsPage = createSettingsPage();
  // showPage(settingsPage, "settings-page",  onSettingsPageDisplay);

  let req = {check_status:{}};
  backgroundPage.callBackend(req, function(resp) {
    onStatusResponse(req, resp);
  });
}

function onStatusResponse(req, resp) {
  let showSettings = true;
  if (resp &&
      resp.status == "" &&
      resp.check_status.git_path != "" &&
      resp.check_status.gpg_path != "" &&
      resp.check_status.gpg_keys && resp.check_status.gpg_keys.length > 0 &&
      resp.check_status.password_store_keys && resp.check_status.password_store_keys.length > 0) {
    showSettings = false;
  }

  if (showSettings) {
    let settingsPage = createSettingsPage(resp);
    showPage(settingsPage, "settings-page",  onSettingsPageDisplay);
    return
  }

  let searchPage = createSearchPage();
  showPage(searchPage, "search-page", onSearchPageDisplay);
}

function updatePersistentState(passFiles) {
  // Add new password files to the fileCountMap.
  let files = {}
  for (let i = 0; i < passFiles.length; i++) {
    let file = passFiles[i];
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

function callBackend(req, callback) {
  backgroundPage.callBackend(req, function(resp) {
    if (!resp) {
      console.log("request", req, "response", resp);
      setOperationStatus("Could not perform backend operation.");
      return;
    }
    if (resp.status != "") {
      console.log("request", req, "response", resp);
      setOperationStatus("Backend operation has failed ("+resp.status+").");
      return;
    }
    callback(req, resp);
  });
}
