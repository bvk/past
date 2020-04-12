'use strict';

function createNewkeyPage(params) {
  let newkeyPageTemplate = document.getElementById("newkey-page-template");
  let page = newkeyPageTemplate.cloneNode(true);

  page.setAttribute("page-params", "{}");
  if (params) {
    page.setAttribute("page-params", JSON.stringify(params));
  }

  let username = page.getElementsByClassName("newkey-page-username")[0];
  username.addEventListener("input", function() {
    onNewkeyPageUsernameChange(page, username);
  });

  let useremail = page.getElementsByClassName("newkey-page-useremail")[0];
  useremail.addEventListener("input", function() {
    onNewkeyPageUseremailChange(page, useremail);
  });

  let passphrase = page.getElementsByClassName("newkey-page-passphrase")[0];
  passphrase.addEventListener("input", function() {
    onNewkeyPagePassphraseChange(page, passphrase);
  });

  let backButton = page.getElementsByClassName("newkey-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onNewkeyPageBackButton(page, backButton);
  });

  let closeButton = page.getElementsByClassName("newkey-page-close-button")[0];
  closeButton.addEventListener("click", function() {
    onNewkeyPageCloseButton(page, closeButton);
  });

  let undoButton = page.getElementsByClassName("newkey-page-undo-button")[0];
  undoButton.addEventListener("click", function() {
    onNewkeyPageUndoButton(page, undoButton);
  });

  let doneButton = page.getElementsByClassName("newkey-page-done-button")[0];
  doneButton.addEventListener("click", function() {
    onNewkeyPageDoneButton(page, doneButton);
  });
  return page;
}

function onNewkeyPageDisplay(page) {
  let params;
  let pageParams = page.getAttribute("page-params");
  if (pageParams != "{}") {
    params = JSON.parse(pageParams);
  }

  // Update the elements with initial values.
}

function onNewkeyPageUsernameChange(page, username) {
  autoNewkeyPageUndoButton(page);
  autoNewkeyPageDoneButton(page);
}

function onNewkeyPageUseremailChange(page, useremail) {
  autoNewkeyPageUndoButton(page);
  autoNewkeyPageDoneButton(page);
}

function onNewkeyPagePassphraseChange(page, passphrase) {
  autoNewkeyPageUndoButton(page);
  autoNewkeyPageDoneButton(page);
}

function onNewkeyPageBackButton(page, backButton) {
  let req = {check_status:{}};
  backgroundPage.callBackend(req, function(resp) {
    let settingsPage = createSettingsPage(resp);
    showPage(settingsPage, "settings-page", onSettingsPageDisplay);
  });
}

function onNewkeyPageCloseButton(page, closeButton) {
  window.close();
}

function onNewkeyPageUndoButton(page, undoButton) {
  let username = page.getElementsByClassName("newkey-page-username")[0];
  let useremail = page.getElementsByClassName("newkey-page-useremail")[0];
  let passphrase = page.getElementsByClassName("newkey-page-passphrase")[0];
  username.value = "";
  useremail.value = "";
  passphrase.value = "";

  // TODO: add repeated passphrase field.

  if (!undoButton) {
    undoButton = page.getElementsByClassName("newkey-page-undo-button")[0];
  }
  undoButton.disabled = true;

  doneButton = page.getElementsByClassName("newkey-page-one-button")[0];
  doneButton.disabled = true;
}

function onNewkeyPageDoneButton(page, doneButton) {
  let username = page.getElementsByClassName("newkey-page-username")[0];
  let useremail = page.getElementsByClassName("newkey-page-useremail")[0];
  let passphrase = page.getElementsByClassName("newkey-page-passphrase")[0];
  // TODO: add repeated passphrase field.

  let req = {create_key:{name:username.value,email:useremail.value,passphrase:passphrase.value}};
  callBackend(req, function(req, resp) {
    onNewkeyPageCreateKeyResponse(page, req, resp);
  });
}

function onNewkeyPageCreateKeyResponse(page, req, resp) {
  onNewkeyPageBackButton(page);
}

function autoNewkeyPageUndoButton(page, undoButton) {
}

function autoNewkeyPageDoneButton(page, doneButton) {
  let username = page.getElementsByClassName("newkey-page-username")[0];
  let useremail = page.getElementsByClassName("newkey-page-useremail")[0];
  let passphrase = page.getElementsByClassName("newkey-page-passphrase")[0];
  // TODO: add repeated passphrase field.

  let disable = false
  if (username.value == "" || useremail.value == "" || passphrase.value == "") {
    disable = true;
  }

  if (!doneButton) {
    doneButton = page.getElementsByClassName("newkey-page-done-button")[0];
  }
  doneButton.disabled = disable;
}
