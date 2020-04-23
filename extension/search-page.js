'use strict';

let passwordFiles;

function createSearchPage() {
  let searchPageTemplate = document.getElementById("search-page-template");
  let page = searchPageTemplate.cloneNode(true);
  page.setAttribute("page-params", "{}");

  let searchBar = page.getElementsByClassName("search-page-search-bar")[0];
  searchBar.addEventListener("input", function() {
    onSearchPageSearchBar(page, searchBar);
  });

  let copyButtons = page.getElementsByClassName("search-page-copy-button");
  for (let i = 0; i < copyButtons.length; i++) {
    let button = copyButtons[i];
    button.addEventListener("click", function() {
      onSearchPageCopyButton(page, button);
    });
  }

  let viewButtons = page.getElementsByClassName("search-page-view-button");
  for (let i = 0; i < viewButtons.length; i++) {
    let button = viewButtons[i];
    button.addEventListener("click", function() {
      onSearchPageViewButton(page, button);
    });
  }

  let settingsButton = page.getElementsByClassName("search-page-settings-button")[0];
  settingsButton.addEventListener("click", function() {
    onSearchPageSettingsButton(page, settingsButton);
  });

  let addButton = page.getElementsByClassName("search-page-add-button")[0];
  addButton.addEventListener("click", function() {
    onSearchPageAddButton(page, addButton);
  });

  return page;
}

function onSearchPageDisplay(page) {
  let req = { list_files: {} };
  callBackend(req, function(req, resp) {
    onSearchPageListFilesResponse(page, req, resp);
  });
}

function onSearchPageListFilesResponse(page, req, resp) {
  if (resp && resp.list_files && resp.list_files.files && resp.list_files.files.length > 0) {
    passwordFiles = resp.list_files.files;
  } else {
    passwordFiles = [];
  }
  updatePersistentState(passwordFiles);

  setSearchPageRecentListItems(page, orderSearchPagePasswordFiles(""));

  var searchBars = page.getElementsByClassName("search-page-search-bar");
  if (searchBars) {
    searchBars[0].focus();
  }
}

function onSearchPageSearchBar(page, searchInput) {
  let files = orderSearchPagePasswordFiles(searchInput.value);
  setSearchPageRecentListItems(page, files);
}

function onSearchPageCopyButton(page, copyButton) {
  let name = copyButton.parentElement.firstElementChild.textContent;
  if (!name) {
    return;
  }

  let req = {view_file:{filename:name}};
  backgroundPage.callBackend(req, function(resp) {
    if (!checkResponse(resp)) {
      return;
    }
    onSearchPageViewFileResponseForCopy(page, req, resp);
  });
}

function onSearchPageViewButton(page, viewButton) {
  let name = viewButton.parentElement.firstElementChild.textContent;
  if (!name) {
    return;
  }

  let req = {view_file:{filename:name}};
  callBackend(req, function(req, resp) {
    onSearchPageViewFileResponseForViewPage(page, req, resp);
  });
}

function onSearchPageSettingsButton(page, settingsButton) {
  let settingsPage = createSettingsPage();
  showPage(settingsPage, "settings-page", function(page) {
    onSettingsPageDisplay(page);
    onSettingsPageCheckButton(page);
  });
}

function onSearchPageAddButton(page, addButton) {
  let editPage = createEditPage();
  showPage(editPage, "edit-page", onEditPageDisplay);
}


function onSearchPageViewFileResponseForCopy(page, req, resp) {
  let whenCleared = function() {
    setOperationStatus("Cleared.");
  };

  if (backgroundPage.copyString(resp.view_file.password, 10, whenCleared)) {
    setOperationStatus("Copied.");
  } else {
    setOperationStatus("Cloud not copy.");
  }

  persistentState.fileCountMap[req.view_file.file] += 1;
  backgroundPage.setLocalStorage({"persistentState": persistentState});
}

function onSearchPageViewFileResponseForViewPage(page, req, resp) {
  let viewPage = createViewPage(req, resp);
  showPage(viewPage, "view-page", onViewPageDisplay);
}

function setSearchPageRecentListItems(page, names) {
  var elems = document.getElementsByClassName("search-page-password-name");
  for (let i = 0; i < elems.length; i++) {
    let name = "";
    let disable = true;
    if (i < names.length) {
      name = names[i];
      disable = false;
    }
    elems[i].textContent = name;
    elems[i].nextElementSibling.disabled = disable;
    elems[i].nextElementSibling.nextElementSibling.disabled = disable;
  }
}

//
// Other functions
//

function sortSearchPagePasswordFiles() {
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

function orderSearchPagePasswordFiles(search) {
  let hosts = getSearchPageHostnameSuffixes();
  let sortedFiles = sortSearchPagePasswordFiles();

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

function getSearchPageHostnameSuffixes() {
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
