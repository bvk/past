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

  let keylength = page.getElementsByClassName("newkey-page-key-length")[0];
  keylength.addEventListener("change", function() {
    autoNewkeyPageUndoButton(page);
  });

  let keyyears = page.getElementsByClassName("newkey-page-key-years")[0];
  keyyears.addEventListener("change", function() {
    autoNewkeyPageUndoButton(page);
  });

  let passphrase = page.getElementsByClassName("newkey-page-passphrase")[0];
  passphrase.addEventListener("input", function() {
    onNewkeyPagePassphraseChange(page, passphrase);
  });

  let repeatPassphrase = page.getElementsByClassName("newkey-page-repeat-passphrase")[0];
  repeatPassphrase.addEventListener("input", function() {
    onNewkeyPageRepeatPassphraseChange(page, repeatPassphrase);
  });

  let backButton = page.getElementsByClassName("newkey-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onNewkeyPageBackButton(page, backButton);
  });

  let closeButton = page.getElementsByClassName("newkey-page-close-button")[0];
  closeButton.addEventListener("click", function() {
    onNewkeyPageCloseButton(page, closeButton);
  });

  let toggleButton = page.getElementsByClassName("newkey-page-toggle-button")[0];
  toggleButton.addEventListener("click", function() {
    onNewkeyPageToggleButton(page, toggleButton);
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

  let username = page.getElementsByClassName("newkey-page-username")[0];
  username.focus();
}

function toggleNewkeyPageButtonsDisabled(page, state, overrides) {
  let backButton = page.getElementsByClassName("newkey-page-back-button")[0];
  let closeButton = page.getElementsByClassName("newkey-page-close-button")[0];
  let toggleButton = page.getElementsByClassName("newkey-page-toggle-button")[0];
  let undoButton = page.getElementsByClassName("newkey-page-undo-button")[0];
  let doneButton = page.getElementsByClassName("newkey-page-done-button")[0];

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

function onNewkeyPageRepeatPassphraseChange(page, repeatPassphrase) {
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

function onNewkeyPageToggleButton(page, toggleButton) {
  let pass = page.getElementsByClassName("newkey-page-passphrase")[0];
  if (pass.type == "text") {
    pass.type = "password";
  } else {
    pass.type = "text";
  }

  let repeat = page.getElementsByClassName("newkey-page-repeat-passphrase")[0];
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

function onNewkeyPageUndoButton(page, undoButton) {
  let username = page.getElementsByClassName("newkey-page-username")[0];
  let useremail = page.getElementsByClassName("newkey-page-useremail")[0];
  let passphrase = page.getElementsByClassName("newkey-page-passphrase")[0];
  let repeatPassphrase = page.getElementsByClassName("newkey-page-repeat-passphrase")[0];
  let keylength = page.getElementsByClassName("newkey-page-key-length")[0];
  let keyyears = page.getElementsByClassName("newkey-page-key-years")[0];
  username.value = "";
  useremail.value = "";
  passphrase.value = "";
  repeatPassphrase.value = "";
  keylength.value = "4096";
  keyyears.value = "0";

  if (!undoButton) {
    undoButton = page.getElementsByClassName("newkey-page-undo-button")[0];
  }
  undoButton.disabled = true;

  let doneButton = page.getElementsByClassName("newkey-page-done-button")[0];
  doneButton.disabled = true;
}

function onNewkeyPageDoneButton(page, doneButton) {
  setOperationStatus("Please wait...");
  let stateMap = toggleNewkeyPageButtonsDisabled(page, true);

  let username = page.getElementsByClassName("newkey-page-username")[0];
  let useremail = page.getElementsByClassName("newkey-page-useremail")[0];
  let keylength = page.getElementsByClassName("newkey-page-key-length")[0];
  let keyyears = page.getElementsByClassName("newkey-page-key-years")[0];
  let passphrase = page.getElementsByClassName("newkey-page-passphrase")[0];

  let req = {
    create_key:{
      name: username.value,
      email: useremail.value,
      passphrase: passphrase.value,
      key_length: keylength.value,
      key_years: keyyears.value,
    },
  };
  backgroundPage.callBackend(req, function(resp) {
    clearOperationStatus();
    toggleNewkeyPageButtonsDisabled(page, false, stateMap);
    onNewkeyPageCreateKeyResponse(page, req, resp);
  });
}

function onNewkeyPageCreateKeyResponse(page, req, resp) {
  if (!resp) {
    setOperationStatus("Could not perform backend operation.");
    return;
  }
  if (resp.status != "") {
    setOperationStatus("Backend operation has failed ("+resp.status+").");
    return;
  }
  onNewkeyPageBackButton(page);
}

function autoNewkeyPageUndoButton(page, undoButton) {
  let username = page.getElementsByClassName("newkey-page-username")[0];
  let useremail = page.getElementsByClassName("newkey-page-useremail")[0];
  let passphrase = page.getElementsByClassName("newkey-page-passphrase")[0];
  let repeatPassphrase = page.getElementsByClassName("newkey-page-repeat-passphrase")[0];
  let keylength = page.getElementsByClassName("newkey-page-key-length")[0];
  let keyyears = page.getElementsByClassName("newkey-page-key-years")[0];

  let disable = false
  if (username.value == "" && useremail.value == "" && passphrase.value == "" &&
      passphrase.value == repeatPassphrase.value &&
      keylength.value == "4096" && keyyears.value == "0") {
    disable = true;
  }

  if (!undoButton) {
    undoButton = page.getElementsByClassName("newkey-page-undo-button")[0];
  }
  undoButton.disabled = disable;
}

function autoNewkeyPageDoneButton(page, doneButton) {
  let username = page.getElementsByClassName("newkey-page-username")[0];
  let useremail = page.getElementsByClassName("newkey-page-useremail")[0];
  let passphrase = page.getElementsByClassName("newkey-page-passphrase")[0];
  let repeatPassphrase = page.getElementsByClassName("newkey-page-repeat-passphrase")[0];

  let disable = false
  if (username.value == "" || useremail.value == "" || passphrase.value == "" ||
      passphrase.value != repeatPassphrase.value) {
    disable = true;
  }

  if (!doneButton) {
    doneButton = page.getElementsByClassName("newkey-page-done-button")[0];
  }
  doneButton.disabled = disable;
}
