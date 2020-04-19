'use strict';

function createNewrepoPage(params) {
  let newrepoPageTemplate = document.getElementById("newrepo-page-template");
  let page = newrepoPageTemplate.cloneNode(true);

  page.setAttribute("page-params", "{}");
  if (params) {
    page.setAttribute("page-params", JSON.stringify(params));
  }

  let backButton = page.getElementsByClassName("newrepo-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onNewrepoPageBackButton(page, backButton);
  });

  let closeButton = page.getElementsByClassName("newrepo-page-close-button")[0];
  closeButton.addEventListener("click", function() {
    onNewrepoPageCloseButton(page, closeButton);
  });

  let createButton = page.getElementsByClassName("newrepo-page-create-button")[0];
  createButton.addEventListener("click", function() {
    onNewrepoPageDisplayTab(page, "newrepo-page-create-button");
  });

  let importButton = page.getElementsByClassName("newrepo-page-import-button")[0];
  importButton.addEventListener("click", function() {
    onNewrepoPageDisplayTab(page, "newrepo-page-import-button");
  });

  let serverSelect = page.getElementsByClassName("newrepo-page-gitserver")[0];
  serverSelect.addEventListener("change", function() {
    onNewrepoPageGitServerChange(page, serverSelect);
  });

  let githost = page.getElementsByClassName("newrepo-page-githost")[0];
  githost.addEventListener("input", function() {
    autoNewrepoPageImportTabDoneButton(page);
    autoNewrepoPageImportTabUndoButton(page);
  });

  let gituser = page.getElementsByClassName("newrepo-page-gituser")[0];
  gituser.addEventListener("input", function() {
    autoNewrepoPageImportTabDoneButton(page);
    autoNewrepoPageImportTabUndoButton(page);
  });

  let gitpass = page.getElementsByClassName("newrepo-page-gitpass")[0];
  gitpass.addEventListener("input", function() {
    autoNewrepoPageImportTabDoneButton(page);
    autoNewrepoPageImportTabUndoButton(page);
  });

  let gitpassToggle = page.getElementsByClassName("newrepo-page-gitpass-toggle")[0];
  gitpassToggle.addEventListener("click", function() {
    onNewrepoPageImportTabToggleButton(page, gitpassToggle);
  });

  let gitpath = page.getElementsByClassName("newrepo-page-gitpath")[0];
  gitpath.addEventListener("input", function() {
    autoNewrepoPageImportTabDoneButton(page);
    autoNewrepoPageImportTabUndoButton(page);
  });

  let undoButton = page.getElementsByClassName("newrepo-page-undo-button")[0];
  undoButton.addEventListener("click", function() {
    onNewrepoPageUndoButton(page, undoButton);
  });

  let doneButton = page.getElementsByClassName("newrepo-page-done-button")[0];
  doneButton.addEventListener("click", function() {
    onNewrepoPageDoneButton(page, doneButton);
  });

  // Create local key list items dynamically.
  if (params && params.check_status && params.check_status.local_keys) {
    let keyItem = page.getElementsByClassName("newrepo-page-localkey-item")[0];
    for (let i = 0; i < params.check_status.local_keys.length; i++) {
      let key = params.check_status.local_keys[i];

      let item = keyItem.cloneNode(true);
      item.setAttribute("key-fingerprint", key.key_fingerprint);
      item.setAttribute("key-state", "off");
      item.getElementsByClassName("newrepo-page-key")[0].textContent = key.user_email;
      item.addEventListener("click", function() {
        onNewrepoPageKeyItem(page, item);
      });

      keyItem.parentNode.insertBefore(item, keyItem.nextSibling);
    }
    keyItem.remove();
  }

  // Create remote key list if at least one remote key exist. It is hidden by default.
  if (params && params.check_status && params.check_status.remote_keys) {
    let keyItem = page.getElementsByClassName("newrepo-page-remotekey-item")[0];
    for (let i = 0; i < params.check_status.local_keys.length; i++) {
      let key = params.check_status.local_keys[i];

      let item = keyItem.cloneNode(true);
      item.setAttribute("key-fingerprint", key.key_fingerprint);
      item.setAttribute("key-state", "off");
      item.getElementsByClassName("newrepo-page-key")[0].textContent = key.user_email;
      item.addEventListener("click", function() {
        onNewrepoPageKeyItem(page, item);
      });

      keyItem.parentNode.insertBefore(item, keyItem.nextSibling);
    }
    keyItem.remove();
    page.getElementsByClassName("newrepo-page-remotekeys-section")[0].style.display = "";
  }

  return page;
}

function onNewrepoPageDisplay(page) {
  onNewrepoPageDisplayTab(page, "newrepo-page-import-button")
}

let newrepoPageTabs = {
  "newrepo-page-create-button":"newrepo-page-create-tab",
  "newrepo-page-import-button":"newrepo-page-import-tab",
};

function onNewrepoPageDisplayTab(page, tabButtonName) {
  if (!(tabButtonName in newrepoPageTabs)) {
    return;
  }
  for (let key in newrepoPageTabs) {
    let keyElem = page.getElementsByClassName(key)[0];
    let valueElem = page.getElementsByClassName(newrepoPageTabs[key])[0];
    if (key != tabButtonName) {
      keyElem.style.background = "transparent";
      valueElem.style.display = "none";
    } else {
      keyElem.style.background = "gray";
      valueElem.style.display = "";
    }
  }
}

function currentNewrepoPageTabButtonName(page) {
  for (let key in newrepoPageTabs) {
    let valueElem = page.getElementsByClassName(newrepoPageTabs[key])[0];
    if (valueElem.style.display == "") {
      return key;
    }
  }
}

function onNewrepoPageBackButton(page, backButton) {
  let req = {check_status:{}};
  backgroundPage.callBackend(req, function(resp) {
    let settingsPage = createSettingsPage(resp);
    showPage(settingsPage, "settings-page", onSettingsPageDisplay);
  });
}

function onNewrepoPageCloseButton(page, closeButton) {
  window.close();
}

function onNewrepoPageGitServerChange(page, serverSelect) {
  let server = serverSelect.value;

  let githost = page.getElementsByClassName("newrepo-page-githost")[0];
  if (server == "github-ssh" || server == "github-https") {
    githost.value = "github.com";
    githost.disabled = true;
  } else {
    githost.disabled = false;
  }

  let gituser = page.getElementsByClassName("newrepo-page-gituser")[0];
  if (server == "github-ssh") {
    gituser.value = "git";
    gituser.disabled = true;
  } else if (server == "github-https" && gituser.value == "git") {
    gituser.value = "";
    gituser.disabled = false;
  } else {
    gituser.disabled = false;
  }

  let gitpass = page.getElementsByClassName("newrepo-page-gitpass")[0];
  if (server == "ssh" || server == "github-ssh") {
    gitpass.disabled = false;
    gitpass.setAttribute("placeholder", "leave empty for password-less authentication")
  } else {
    gitpass.disabled = false;
    gitpass.setAttribute("placeholder", "password")
  }

  let gitpath = page.getElementsByClassName("newrepo-page-gitpath")[0];
  if (server == "github-ssh" || server == "github-https") {
    gitpath.setAttribute("placeholder", "username/repository.git")
  } else {
    gitpath.setAttribute("placeholder", "path/to/repository.git")
  }
  gitpath.disabled = false;

  autoNewrepoPageImportTabUndoButton(page);
  autoNewrepoPageImportTabDoneButton(page);
}

function onNewrepoPageKeyItem(page, keyItem) {
  let check = keyItem.getElementsByClassName("newrepo-page-checkbox")[0];

  let state = keyItem.getAttribute("key-state");
  if (state == "on") {
    keyItem.setAttribute("key-state", "off");
    check.textContent = "check_box_outline_blank";
  } else {
    keyItem.setAttribute("key-state", "on");
    check.textContent = "check_box";
  }

  autoNewrepoPageUndoButton(page);
  autoNewrepoPageDoneButton(page);
}

function onNewrepoPageCreateTabUndoButton(page, undoButton) {
  let localkeys = page.getElementsByClassName("newrepo-page-localkey-item");
  for (let i = 0; i < localkeys.length; i++) {
    let checkbox = localkeys[i].getElementsByClassName("newrepo-page-checkbox")[0];
    checkbox.textContent = "check_box_outline_blank";
    localkeys[i].setAttribute("key-state", "off");
  }

  let remotekeys = page.getElementsByClassName("newrepo-page-remotekey-item");
  for (let i = 0; i < remotekeys.length; i++) {
    let checkbox = remotekeys[i].getElementsByClassName("newrepo-page-checkbox")[0];
    checkbox.textContent = "check_box_outline_blank";
    remotekeys[i].setAttribute("key-state", "off");
  }

  if (!undoButton) {
    undoButton = page.getElementsByClassName("newrepo-page-undo-button")[0];
  }
  undoButton.disabled = true;

  let doneButton = page.getElementsByClassName("newrepo-page-done-button")[0];
  doneButton.disabled = true;
}

function onNewrepoPageImportTabUndoButton(page, undoButton) {
  let gitserver = page.getElementsByClassName("newrepo-page-gitserver")[0];
  let githost = page.getElementsByClassName("newrepo-page-githost")[0];
  let gituser = page.getElementsByClassName("newrepo-page-gituser")[0];
  let gitpass = page.getElementsByClassName("newrepo-page-gitpass")[0];
  let gitpath = page.getElementsByClassName("newrepo-page-gitpath")[0];
  gitserver.value = "ssh";
  githost.value = "";
  githost.disabled = false;
  gituser.value = "";
  gituser.disabled = false;
  gitpass.value = "";
  gitpass.setAttribute("placeholder", "leave empty for password-less authentication");
  gitpath.value = "";

  if (!undoButton) {
    undoButton = page.getElementsByClassName("newrepo-page-undo-button")[0];
  }
  undoButton.disabled = true;

  let doneButton = page.getElementsByClassName("newrepo-page-done-button")[0];
  doneButton.disabled = true;
}

function onNewrepoPageUndoButton(page, undoButton) {
  let currentTab = currentNewrepoPageTabButtonName(page);
  if (currentTab == "newrepo-page-create-button") {
    onNewrepoPageCreateTabUndoButton(page, undoButton);
  } else if (currentTab == "newrepo-page-import-button") {
    onNewrepoPageImportTabUndoButton(page, undoButton);
  }
}

function onNewrepoPageCreateTabDoneButton(page, doneButton) {
  let fingerprints = [];
  let localkeys = page.getElementsByClassName("newrepo-page-localkey-item");
  for (let i = 0; i < localkeys.length; i++) {
    let state = localkeys[i].getAttribute("key-state");
    let fingerprint = localkeys[i].getAttribute("key-fingerprint");
    if (state == "on") {
      fingerprints.push(fingerprint);
    }
  }
  let remotekeys = page.getElementsByClassName("newrepo-page-remotekey-item");
  for (let i = 0; i < remotekeys.length; i++) {
    let state = remotekeys[i].getAttribute("key-state");
    let fingerprint = remotekeys[i].getAttribute("key-fingerprint");
    if (state == "on") {
      fingerprints.push(fingerprint);
    }
  }
  if (fingerprints.length == 0) {
    setOperationStatus("No keys are selected for the password store.");
    return;
  }
  let req = {create_repo:{fingerprints:fingerprints}};
  callBackend(req, function(req, resp) {
    onNewrepoPageCreateRepoResponse(page, req, resp);
  });
}

function onNewrepoPageImportTabDoneButton(page, doneButton) {
  let gitserver = page.getElementsByClassName("newrepo-page-gitserver")[0];
  let githost = page.getElementsByClassName("newrepo-page-githost")[0];
  let gituser = page.getElementsByClassName("newrepo-page-gituser")[0];
  let gitpass = page.getElementsByClassName("newrepo-page-gitpass")[0];
  let gitpath = page.getElementsByClassName("newrepo-page-gitpath")[0];

  let protocol = "git";
  if (gitserver.value == "ssh" || gitserver.value == "github-ssh") {
    protocol = "ssh";
  } else if (gitserver.value == "https" || gitserver.value == "github-https") {
    protocol = "https";
  } else if (gitserver.value == "git") {
    protocol = "git";
  }

  let req = {
    import_repo: {
      protocol: protocol,
      username: gituser.value,
      password: gitpass.value,
      hostname: githost.value,
      path: gitpath.value,
    },
  }
  callBackend(req, function(req, resp) {
    onNewrepoPageBackButton(page);
  });
}

function onNewrepoPageImportTabToggleButton(page, toggleButton) {
  let gitpass = page.getElementsByClassName("newrepo-page-gitpass")[0];
  if (gitpass.type == "text") {
    gitpass.type = "password";
  } else {
    gitpass.type = "text";
  }

  if (toggleButton.textContent == "visibility_off") {
    toggleButton.textContent = "visibility";
  } else {
    toggleButton.textContent = "visibility_off";
  }
}

function onNewrepoPageDoneButton(page, doneButton) {
  let currentTab = currentNewrepoPageTabButtonName(page);
  if (currentTab == "newrepo-page-create-button") {
    onNewrepoPageCreateTabDoneButton(page, doneButton);
  } else if (currentTab == "newrepo-page-import-button") {
    onNewrepoPageImportTabDoneButton(page, doneButton);
  }
}

function onNewrepoPageCreateRepoResponse(page, req, resp) {
  onNewrepoPageBackButton(page);
}

function autoNewrepoPageCreateTabUndoButton(page, undoButton) {
  let disable = true;
  let localkeys = page.getElementsByClassName("newrepo-page-localkey-item");
  for (let i = 0; i < localkeys.length; i++) {
    let state = localkeys[i].getAttribute("key-state");
    if (state == "on") {
      disable = false;
      break;
    }
  }
  let remotekeys = page.getElementsByClassName("newrepo-page-remotekey-item");
  for (let i = 0; disable && i < remotekeys.length; i++) {
    let state = remotekeys[i].getAttribute("key-state");
    if (state == "on") {
      disable = false;
      break;
    }
  }
  if (!undoButton) {
    undoButton = page.getElementsByClassName("newrepo-page-undo-button")[0];
  }
  undoButton.disabled = disable;
}

function autoNewrepoPageImportTabUndoButton(page, undoButton) {
  let gitserver = page.getElementsByClassName("newrepo-page-gitserver")[0];
  let githost = page.getElementsByClassName("newrepo-page-githost")[0];
  let gituser = page.getElementsByClassName("newrepo-page-gituser")[0];
  let gitpass = page.getElementsByClassName("newrepo-page-gitpass")[0];
  let gitpath = page.getElementsByClassName("newrepo-page-gitpath")[0];

  var disable = true;
  if (gitserver.value != "ssh" || githost.value != "" || gituser.value != "" ||
      gitpass.value != "" || gitpath.value != "") {
    disable = false;
  }

  if (!undoButton) {
    undoButton = page.getElementsByClassName("newrepo-page-undo-button")[0];
  }
  undoButton.disabled = disable;
}

function autoNewrepoPageUndoButton(page, undoButton) {
  if (!undoButton) {
    undoButton = page.getElementsByClassName("newrepo-page-undo-button")[0];
  }
  let currentTab = currentNewrepoPageTabButtonName(page);
  if (currentTab == "newrepo-page-create-button") {
    autoNewrepoPageCreateTabUndoButton(page, undoButton);
  } else if (currentTab == "newrepo-page-import-button") {
    autoNewrepoPageImportTabUndoButton(page, undoButton);
  }
}

function autoNewrepoPageCreateTabDoneButton(page, doneButton) {
  let disable = true;

  let localkeys = page.getElementsByClassName("newrepo-page-localkey-item");
  for (let i = 0; i < localkeys.length; i++) {
    let state = localkeys[i].getAttribute("key-state");
    if (state == "on") {
      disable = false;
      break;
    }
  }
  if (!doneButton) {
    doneButton = page.getElementsByClassName("newrepo-page-done-button")[0];
  }
  doneButton.disabled = disable;
}

function autoNewrepoPageImportTabDoneButton(page, doneButton) {
  let gitserver = page.getElementsByClassName("newrepo-page-gitserver")[0];
  let githost = page.getElementsByClassName("newrepo-page-githost")[0];
  let gituser = page.getElementsByClassName("newrepo-page-gituser")[0];
  let gitpass = page.getElementsByClassName("newrepo-page-gitpass")[0];
  let gitpath = page.getElementsByClassName("newrepo-page-gitpath")[0];

  let allowEmptyPassword = false;
  if (gitserver.value == "github-ssh" || gitserver.value == "ssh") {
    allowEmptyPassword = true;
  }

  let disable = false;
  if (githost.value == "" || gituser.value == "" || gitpath.value == "") {
    disable = true;
  }

  if (disable == false && allowEmptyPassword == false && gitpass.value == "") {
    disable = true;
  }
  if (!doneButton) {
    doneButton = page.getElementsByClassName("newrepo-page-done-button")[0];
  }
  console.log("donebutton.disabled = ", disable);
  doneButton.disabled = disable;
}

function autoNewrepoPageDoneButton(page, doneButton) {
  if (!doneButton) {
    doneButton = page.getElementsByClassName("newrepo-page-done-button")[0];
  }
  let currentTab = currentNewrepoPageTabButtonName(page);
  if (currentTab == "newrepo-page-create-button") {
    autoNewrepoPageCreateTabDoneButton(page, doneButton);
  } else if (currentTab == "newrepo-page-import-button") {
    autoNewrepoPageImportTabDoneButton(page, doneButton);
  }
}
