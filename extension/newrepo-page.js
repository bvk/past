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
      item.setAttribute("key-fingerprint", key.fingerprint);
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
      item.setAttribute("key-fingerprint", key.fingerprint);
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

function onNewrepoPageUndoButton(page, undoButton) {
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

function onNewrepoPageDoneButton(page, doneButton) {
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

function onNewrepoPageCreateRepoResponse(page, req, resp) {
  onNewrepoPageBackButton(page);
}

function autoNewrepoPageUndoButton(page, undoButton) {
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

function autoNewrepoPageDoneButton(page, doneButton) {
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
