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

  // Create key list items dynamically.
  if (params && params.check_status && params.check_status.gpg_keys) {
    let keyItem = page.getElementsByClassName("newrepo-page-keylist-item")[0];
    for (let i = 0; i < params.check_status.gpg_keys.length; i++) {
      let key = params.check_status.gpg_keys[i];

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
  let items = page.getElementsByClassName("newrepo-page-keylist-item");
  for (let i = 0; i < items.length; i++) {
    let checkbox = items[i].getElementsByClassName("newrepo-page-checkbox")[0];
    checkbox.textContent = "check_box_outline_blank";
    items[i].setAttribute("key-state", "off");
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
  let items = page.getElementsByClassName("newrepo-page-keylist-item");
  for (let i = 0; i < items.length; i++) {
    let state = items[i].getAttribute("key-state");
    let fingerprint = items[i].getAttribute("key-fingerprint");
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
  let items = page.getElementsByClassName("newrepo-page-keylist-item");
  for (let i = 0; i < items.length; i++) {
    let state = items[i].getAttribute("key-state");
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
  let items = page.getElementsByClassName("newrepo-page-keylist-item");
  for (let i = 0; i < items.length; i++) {
    let state = items[i].getAttribute("key-state");
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
