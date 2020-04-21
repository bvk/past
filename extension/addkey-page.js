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

  let createButton = page.getElementsByClassName("addkey-page-create-button")[0];
  createButton.addEventListener("click", function() {
    onAddkeyPageDisplayTab(page, "addkey-page-create-button");
  });

  let importButton = page.getElementsByClassName("addkey-page-import-button")[0];
  importButton.addEventListener("click", function() {
    onAddkeyPageDisplayTab(page, "addkey-page-import-button");
  });

  let username = page.getElementsByClassName("addkey-page-username")[0];
  username.addEventListener("input", function() {
    autoAddkeyPageCreateTabDoneButton(page);
    autoAddkeyPageCreateTabUndoButton(page);
  });

  let useremail = page.getElementsByClassName("addkey-page-useremail")[0];
  useremail.addEventListener("input", function() {
    autoAddkeyPageCreateTabDoneButton(page);
    autoAddkeyPageCreateTabUndoButton(page);
  });

  let keylength = page.getElementsByClassName("addkey-page-key-length")[0];
  keylength.addEventListener("change", function() {
    autoAddkeyPageCreateTabDoneButton(page);
    autoAddkeyPageCreateTabUndoButton(page);
  });

  let keyyears = page.getElementsByClassName("addkey-page-key-years")[0];
  keyyears.addEventListener("change", function() {
    autoAddkeyPageCreateTabDoneButton(page);
    autoAddkeyPageCreateTabUndoButton(page);
  });

  let passphrase = page.getElementsByClassName("addkey-page-passphrase")[0];
  passphrase.addEventListener("input", function() {
    autoAddkeyPageCreateTabDoneButton(page);
    autoAddkeyPageCreateTabUndoButton(page);
  });

  let repeatPassphrase = page.getElementsByClassName("addkey-page-repeat-passphrase")[0];
  repeatPassphrase.addEventListener("input", function() {
    autoAddkeyPageCreateTabDoneButton(page);
    autoAddkeyPageCreateTabUndoButton(page);
  });

  let toggleButton = page.getElementsByClassName("addkey-page-toggle-button")[0];
  toggleButton.addEventListener("click", function() {
    onAddkeyPageToggleButton(page, toggleButton);
  });

  let keydata = page.getElementsByClassName("addkey-page-keydata")[0];
  keydata.addEventListener("input", function() {
    autoAddkeyPageImportTabUndoButton(page);
    autoAddkeyPageImportTabDoneButton(page);
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
  onAddkeyPageDisplayTab(page, "addkey-page-create-button");
}

let addkeyPageTabs = {
  "addkey-page-create-button": "addkey-page-create-tab",
  "addkey-page-import-button": "addkey-page-import-tab",
};

function currentAddkeyPageTabButtonName(page) {
  for (let key in addkeyPageTabs) {
    let valueElem = page.getElementsByClassName(addkeyPageTabs[key])[0];
    if (valueElem.style.display == "") {
      return key;
    }
  }
}

function onAddkeyPageDisplayTab(page, tabButtonName) {
  if (!(tabButtonName in addkeyPageTabs)) {
    return;
  }
  for (let key in addkeyPageTabs) {
    let keyElem = page.getElementsByClassName(key)[0];
    let valueElem = page.getElementsByClassName(addkeyPageTabs[key])[0];
    if (key != tabButtonName) {
      keyElem.style.background = "transparent";
      valueElem.style.display = "none";
    } else {
      keyElem.style.background = "gray";
      valueElem.style.display = "";
    }
  }
  autoAddkeyPageUndoButton(page);
  autoAddkeyPageDoneButton(page);
}

function toggleAddkeyPageButtonsDisabled(page, state, overrides) {
  let backButton = page.getElementsByClassName("addkey-page-back-button")[0];
  let closeButton = page.getElementsByClassName("addkey-page-close-button")[0];
  let toggleButton = page.getElementsByClassName("addkey-page-toggle-button")[0];
  let undoButton = page.getElementsByClassName("addkey-page-undo-button")[0];
  let doneButton = page.getElementsByClassName("addkey-page-done-button")[0];

  let oldStateMap = {
    backButton: backButton.disabled,
    closeButton: closeButton.disabled,
    toggleButton: toggleButton.disabled,
    undoButton: undoButton.disabled,
    doneButton: doneButton.disabled,
  }

  backButton.disabled = state;
  closeButton.disabled = state;
  toggleButton.disabled = state;
  undoButton.disabled = state;
  doneButton.disabled = state;

  if (overrides) {
    if (overrides.backButton !== undefined) {
      backButton.disabled = overrides.backButton;
    }
    if (overrides.closeButton !== undefined) {
      closeButton.disabled = overrides.closeButton;
    }
    if (overrides.toggleButton !== undefined) {
      toggleButton.disabled = overrides.toggleButton;
    }
    if (overrides.undoButton !== undefined) {
      undoButton.disabled = overrides.undoButton;
    }
    if (overrides.doneButton !== undefined) {
      doneButton.disabled = overrides.doneButton;
    }
  }

  return oldStateMap;
}

function onAddkeyPageBackButton(page, backButton) {
  let req = {check_status:{}};
  callBackend(req, function(req, resp) {
    if (!resp.check_status.local_keys || resp.check_status.local_keys.length == 0) {
      let settingsPage = createSettingsPage(resp);
      showPage(settingsPage, "settings-page", onSettingsPageDisplay);
      return;
    }
    let keyringPage = createKeyringPage(resp);
    showPage(keyringPage, "keyring-page", onKeyringPageDisplay);
  });
}

function onAddkeyPageToggleButton(page, toggleButton) {
  let pass = page.getElementsByClassName("addkey-page-passphrase")[0];
  if (pass.type == "text") {
    pass.type = "password";
  } else {
    pass.type = "text";
  }

  let repeat = page.getElementsByClassName("addkey-page-repeat-passphrase")[0];
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

function onAddkeyPageCreateTabUndoButton(page, undoButton) {
  let username = page.getElementsByClassName("addkey-page-username")[0];
  let useremail = page.getElementsByClassName("addkey-page-useremail")[0];
  let passphrase = page.getElementsByClassName("addkey-page-passphrase")[0];
  let repeatPassphrase = page.getElementsByClassName("addkey-page-repeat-passphrase")[0];
  let keylength = page.getElementsByClassName("addkey-page-key-length")[0];
  let keyyears = page.getElementsByClassName("addkey-page-key-years")[0];
  username.value = "";
  useremail.value = "";
  passphrase.value = "";
  repeatPassphrase.value = "";
  keylength.value = "4096";
  keyyears.value = "0";

  if (!undoButton) {
    undoButton = page.getElementsByClassName("addkey-page-undo-button")[0];
  }
  undoButton.disabled = true;

  let doneButton = page.getElementsByClassName("addkey-page-done-button")[0];
  doneButton.disabled = true;
}

function onAddkeyPageImportTabUndoButton(page, undoButton) {
  let keydata = page.getElementsByClassName("addkey-page-keydata")[0];
  keydata.value = "";

  if (!undoButton) {
    undoButton = page.getElementsByClassName("addkey-page-undo-button")[0];
  }
  undoButton.disabled = true;
  let doneButton = page.getElementsByClassName("addkey-page-done-button")[0];
  doneButton.disabled = true;
}

function onAddkeyPageUndoButton(page, undoButton) {
  let currentTab = currentAddkeyPageTabButtonName(page);
  if (currentTab == "addkey-page-create-button") {
    onAddkeyPageCreateTabUndoButton(page, undoButton);
  } else if (currentTab == "addkey-page-import-button") {
    onAddkeyPageImportTabUndoButton(page, undoButton);
  }
}

function onAddkeyPageImportTabDoneButton(page, doneButton) {
  let keydata = page.getElementsByClassName("addkey-page-keydata")[0];
  let req = {import_key:{key:keydata.value}};
  callBackend(req, function(req, resp) {
    onAddkeyPageBackButton(page);
  });
}

function onAddkeyPageCreateTabDoneButton(page, doneButton) {
  let stateMap = toggleAddkeyPageButtonsDisabled(page, true);

  let username = page.getElementsByClassName("addkey-page-username")[0];
  let useremail = page.getElementsByClassName("addkey-page-useremail")[0];
  let keylength = page.getElementsByClassName("addkey-page-key-length")[0];
  let keyyears = page.getElementsByClassName("addkey-page-key-years")[0];
  let passphrase = page.getElementsByClassName("addkey-page-passphrase")[0];

  let req = {
    create_key:{
      name: username.value,
      email: useremail.value,
      passphrase: passphrase.value,
      key_length: keylength.value,
      key_years: keyyears.value,
    },
  };

  setOperationStatus("Please wait...");
  backgroundPage.callBackend(req, function(resp) {
    clearOperationStatus();
    toggleAddkeyPageButtonsDisabled(page, false, stateMap);
    onAddkeyPageCreateKeyResponse(page, req, resp);
  });
}

function onAddkeyPageCreateKeyResponse(page, req, resp) {
  if (!resp) {
    setOperationStatus("Could not perform backend operation.");
    return;
  }
  if (resp.status != "") {
    setOperationStatus("Backend operation has failed ("+resp.status+").");
    return;
  }
  let checkReq = {check_status:{}};
  callBackend(checkReq, function(req, resp) {
    let keyringPage = createKeyringPage(resp);
    showPage(keyringPage, "keyring-page", onKeyringPageDisplay);
  });
}

function onAddkeyPageDoneButton(page, doneButton) {
  let currentTab = currentAddkeyPageTabButtonName(page);
  if (currentTab == "addkey-page-create-button") {
    onAddkeyPageCreateTabDoneButton(page, doneButton);
  } else if (currentTab == "addkey-page-import-button") {
    onAddkeyPageImportTabDoneButton(page, doneButton);
  }
}

function onAddkeyPageImportTabUndoButton(page, undoButton) {
  let keydata = page.getElementsByClassName("addkey-page-keydata")[0];
  keydata.value = "";

  if (!undoButton) {
    undoButton = page.getElementsByClassName("addkey-page-undo-button")[0];
  }
  undoButton.disabled = true;
  let doneButton = page.getElementsByClassName("addkey-page-done-button")[0];
  doneButton.disabled = true;
}

function autoAddkeyPageImportTabDoneButton(page, doneButton) {
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

function autoAddkeyPageCreateTabDoneButton(page, doneButton) {
  let username = page.getElementsByClassName("addkey-page-username")[0];
  let useremail = page.getElementsByClassName("addkey-page-useremail")[0];
  let passphrase = page.getElementsByClassName("addkey-page-passphrase")[0];
  let repeatPassphrase = page.getElementsByClassName("addkey-page-repeat-passphrase")[0];

  let disable = false
  if (username.value == "" || useremail.value == "" || passphrase.value == "" ||
      passphrase.value != repeatPassphrase.value) {
    disable = true;
  }

  if (!doneButton) {
    doneButton = page.getElementsByClassName("addkey-page-done-button")[0];
  }
  doneButton.disabled = disable;
}

function autoAddkeyPageDoneButton(page, doneButton) {
  let currentTab = currentAddkeyPageTabButtonName(page);
  if (currentTab == "addkey-page-create-button") {
    autoAddkeyPageCreateTabDoneButton(page, doneButton);
  } else if (currentTab == "addkey-page-import-button") {
    autoAddkeyPageImportTabDoneButton(page, doneButton);
  }
}

function autoAddkeyPageImportTabUndoButton(page, undoButton) {
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

function autoAddkeyPageCreateTabUndoButton(page, undoButton) {
  let username = page.getElementsByClassName("addkey-page-username")[0];
  let useremail = page.getElementsByClassName("addkey-page-useremail")[0];
  let passphrase = page.getElementsByClassName("addkey-page-passphrase")[0];
  let repeatPassphrase = page.getElementsByClassName("addkey-page-repeat-passphrase")[0];
  let keylength = page.getElementsByClassName("addkey-page-key-length")[0];
  let keyyears = page.getElementsByClassName("addkey-page-key-years")[0];

  let disable = false
  if (username.value == "" && useremail.value == "" && passphrase.value == "" &&
      passphrase.value == repeatPassphrase.value &&
      keylength.value == "4096" && keyyears.value == "0") {
    disable = true;
  }

  if (!undoButton) {
    undoButton = page.getElementsByClassName("addkey-page-undo-button")[0];
  }
  undoButton.disabled = disable;
}

function autoAddkeyPageUndoButton(page, undoButton) {
  let currentTab = currentAddkeyPageTabButtonName(page);
  if (currentTab == "addkey-page-create-button") {
    autoAddkeyPageCreateTabUndoButton(page, undoButton);
  } else if (currentTab == "addkey-page-import-button") {
    autoAddkeyPageImportTabUndoButton(page, undoButton);
  }
}
