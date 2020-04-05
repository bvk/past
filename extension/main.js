'use strict';

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
  setRecentListItems(orderPasswordFiles(""));
  setOperationStatus("Backend is ready.");

  var copyButtons = document.getElementsByClassName("copy-button");
  for (let i = 0; i < copyButtons.length; i++) {
    let button = copyButtons[i];
    button.addEventListener("click", function() {
      onCopyButtonClick(button);
    });
  }

  var searchBars = document.getElementsByClassName("search-bar");
  for (let i = 0; i < searchBars.length; i++) {
    let bar = searchBars[i];
    bar.addEventListener("input", function() {
      onSearchBarChanged(bar);
    });
  }
  if (searchBars) {
    searchBars[0].focus();
  }
}

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

function onCopyButtonClick(elem) {
  let name = elem.parentElement.childNodes[1].textContent;
  if (name) {
    let req = {view_file:{file:name}};
    backgroundPage.viewFile(req, function (resp) {
      onViewFileResponse(req, resp);
    });
  }
}

function onViewFileResponse(req, resp) {
  if (!resp) {
    setOperationStatus("Could not perform Copy password operation.");
    return;
  }
  if (resp.status != "") {
    setOperationStatus("Copy password operation has failed ("+resp.status+").");
    return;
  }

  if (backgroundPage.copyPassword(resp.view_file.password)) {
    setOperationStatus("Password is copied to the clipboard.");
  } else {
    setOperationStatus("Cloud not copy Password to the clipboard.");
  }

  persistentState.fileCountMap[req.view_file.file] += 1;

  console.log("saving state", persistentState);
  backgroundPage.setLocalStorage({"persistentState": persistentState});
}

function onSearchBarChanged(elem) {
  let files = orderPasswordFiles(elem.value);
  setRecentListItems(files);
}

function clearOperationStatus() {
  let elems = document.getElementsByClassName("operation-status");
  for (let i = 0; i < elems.length; i++) {
    elems[i].textContent = "";
  }
}

function setOperationStatus(message) {
  let elems = document.getElementsByClassName("operation-status");
  for (let i = 0; i < elems.length; i++) {
    elems[i].textContent = message;
  }
}

function clearRecentListItems() {
  var elems = document.getElementsByClassName("recent-password-name");
  for (let i = 0; i < elems.length; i++) {
    elems[i].textContent = "";
    elems[i].nextElementSibling.disabled = true;
  }
}

function setRecentListItems(names) {
  var elems = document.getElementsByClassName("recent-password-name");
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
