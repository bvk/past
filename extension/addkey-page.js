'use strict';

function createAddkeyPage(params) {
  let addkeyTemplate = document.getElementById("addkey-page-template");
  let page = addkeyTemplate.cloneNode(true);

  page.setAttribute("page-params", "{}");
  if (params) {
    page.setAttribute("page-params", JSON.stringify(params));
  }

  let backButton = page.getElementsByClassName("addkey-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onAddkeyPageBackButton(page, backButton);
  });

  let closeButton = page.getElementsByClassName("addkey-page-close-button")[0];
  closeButton.addEventListener("click", function() {
    window.close();
  });

  let keydata = page.getElementsByClassName("addkey-page-keydata")[0];
  keydata.addEventListener("input", function() {
    onAddkeyPageKeyDataChange(page, keydata);
  });

  let doneButton = page.getElementsByClassName("addkey-page-done-button")[0];
  doneButton.addEventListener("click", function() {
    onAddkeyPageDoneButton(page, doneButton);
  });

  let undoButton = page.getElementsByClassName("addkey-page-undo-button")[0];
  undoButton.addEventListener("click", function() {
    onAddkeyPageUndoButton(page, undoButton);
  });

  return page;
}

function onAddkeyPageDisplay(page) {
  autoAddkeyPageUndoButton(page);
  autoAddkeyPageDoneButton(page);
}

function onAddkeyPageBackButton(page, backButton) {
  let req = {check_status:{}};
  callBackend(req, function(req, resp) {
    let keysPage = createKeysPage(resp);
    showPage(keysPage, "keys-page", onKeysPageDisplay);
  });
}

function onAddkeyPageKeyDataChange(page, keydata) {
  autoAddkeyPageUndoButton(page);
  autoAddkeyPageDoneButton(page);
}

function onAddkeyPageDoneButton(page, doneButton) {
  let keydata = page.getElementsByClassName("addkey-page-keydata")[0];
  let req = {import_key:{key:keydata.value}};
  callBackend(req, function(req, resp) {
    onAddkeyPageBackButton(page);
  });
}

function onAddkeyPageUndoButton(page, undoButton) {
  let keydata = page.getElementsByClassName("addkey-page-keydata")[0];
  keydata.value = "";

  if (!undoButton) {
    undoButton = page.getElementsByClassName("addkey-page-undo-button")[0];
  }
  undoButton.disabled = true;
  let doneButton = page.getElementsByClassName("addkey-page-done-button")[0];
  doneButton.disabled = true;
}

function autoAddkeyPageDoneButton(page, doneButton) {
  let disable = true;
  let keydata = page.getElementsByClassName("addkey-page-keydata")[0];
  if (keydata.value != "") {
    disable = false;
  }
  if (!doneButton) {
    doneButton = page.getElementsByClassName("addkey-page-done-button")[0];
  }
  doneButton.disabled = disable;
}

function autoAddkeyPageUndoButton(page, undoButton) {
  let disable = false;
  let keydata = page.getElementsByClassName("addkey-page-keydata")[0];
  if (keydata.value == "") {
    disable = true;
  }
  if (!undoButton) {
    undoButton = page.getElementsByClassName("addkey-page-undo-button")[0];
  }
  undoButton.disabled = disable;
}
