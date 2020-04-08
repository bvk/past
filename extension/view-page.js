'use strict';

function createViewPage(req, resp) {
  let viewPageTemplate = document.getElementById("view-page-template");
  let page = viewPageTemplate.cloneNode(true);

  // Set the page title.
  let title = page.getElementsByClassName("view-page-file-name")[0]
  title.textContent = req.view_file.file;

  let password = page.getElementsByClassName("view-page-password")[0]
  password.value = resp.view_file.password;

  let username = page.getElementsByClassName("view-page-username")[0]
  for (let i = 0; i < resp.view_file.values.length; i++) {
    let key = resp.view_file.values[i][0];
    if (key == "username" || key == "user" || key == "login") {
      username.value = resp.view_file.values[i][1];
      break
    }
  }

  // Dynamically added key-value rows, one for each key-value in the response.
  for (let i = 0; i < resp.view_file.values.length; i++) {
    let key = resp.view_file.values[i][0];
    let value = resp.view_file.values[i][1];
    appendViewPageKeyValueRow(page, key, value);
  }
  if (resp.view_file.values.length > 0) {
    // Hide the empty key-value row cause one or more kv entries are present.
    let kv = page.getElementsByClassName("view-page-key-value")[0]
    kv.style.display = "none";
  }

  let backButton = page.getElementsByClassName("view-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onViewPageBackButton(page, backButton);
  });

  let editButton = page.getElementsByClassName("view-page-edit-button")[0];
  editButton.addEventListener("click", function() {
    onViewPageEditButton(page, editButton);
  });

  let userCopyButton = page.getElementsByClassName("view-page-username-copy")[0];
  userCopyButton.addEventListener("click", function() {
    onViewPageUserCopyButton(page, userCopyButton);
  });

  let passCopyButton = page.getElementsByClassName("view-page-password-copy")[0];
  passCopyButton.addEventListener("click", function() {
    onViewPagePassCopyButton(page, passCopyButton);
  });

  let toggleButton = page.getElementsByClassName("view-page-password-toggle")[0];
  toggleButton.addEventListener("click", function() {
    onViewPageToggleButton(page, toggleButton);
  });

  let valueCopyButtons = page.getElementsByClassName("view-page-value-copy");
  for (let i = 0; i < valueCopyButtons.length; i++) {
    let button = valueCopyButtons[i];
    button.addEventListener("click", function() {
      onViewPageValueCopyButton(page, button);
    });
  }

  return page;
}

function onViewPageDisplay(page) {
  // Nothing to do.
}

function onViewPageBackButton(page, backButton) {
  let searchPage = createSearchPage();
  showPage(searchPage, "search-page", onSearchPageDisplay);
}

function onViewPageToggleButton(page, toggleButton) {
  let pass = page.getElementsByClassName("view-page-password")[0];
  if (pass.type == "text") {
    pass.type = "password";
  } else {
    pass.type = "text";
  }

  if (toggleButton.textContent == "visibility_off") {
    toggleButton.textContent = "visibility";
  } else {
    toggleButton.textContent = "visibility_off";
  }
}

function onViewPagePassCopyButton(page, copyButton) {
  let whenCleared = function() {
    setOperationStatus("Password is cleared from the clipboard.");
  };

  let password = page.getElementsByClassName("view-page-password")[0]
  if (backgroundPage.copyString(password.value, 10, whenCleared)) {
    setOperationStatus("Password is copied to the clipboard.");
  } else {
    setOperationStatus("Cloud not copy the password to clipboard.");
  }
}

function onViewPageUserCopyButton(page, copyButton) {
  let username = page.getElementsByClassName("view-page-username")[0]
  if (backgroundPage.copyString(username.value)) {
    setOperationStatus("Username is copied to the clipboard.");
  } else {
    setOperationStatus("Cloud not copy the username to clipboard.");
  }
}

function onViewPageValueCopyButton(page, copyButton) {
  let value = copyButton.previousElementSibling;
  if (backgroundPage.copyString(value.value)) {
    setOperationStatus("Value is copied to the clipboard.");
  } else {
    setOperationStatus("Cloud not copy the value to clipboard.");
  }
}

function onViewPageEditButton(page, editButton) {
  setOperationStatus("Edit operation is not implemented yet.");
}

function appendViewPageKeyValueRow(page, key, value) {
  let kvs = page.getElementsByClassName("view-page-key-value");
  let kv = kvs[kvs.length-1];
  let row = kv.cloneNode(true);
  row.getElementsByClassName("view-page-key")[0].value = key;
  row.getElementsByClassName("view-page-value")[0].value = value;
  if (value) {
    let copy = row.getElementsByClassName("view-page-value-copy")[0];
    copy.disabled = false;
  }
  kv.parentNode.insertBefore(row, kv.nextSibling);
  return row;
}
