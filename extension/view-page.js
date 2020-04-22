'use strict';

function createViewPage(req, resp) {
  let viewPageTemplate = document.getElementById("view-page-template");
  let page = viewPageTemplate.cloneNode(true);
  page.setAttribute("page-params", JSON.stringify(resp));

  // Set the page title.
  let title = page.getElementsByClassName("view-page-filename")[0];
  title.textContent = req.view_file.filename;

  let password = page.getElementsByClassName("view-page-password")[0];
  password.value = resp.view_file.password;

  let username = page.getElementsByClassName("view-page-username")[0];
  username.value = resp.view_file.username;

  let userdata = page.getElementsByClassName("view-page-userdata")[0];
  userdata.value = resp.view_file.data;

  let backButton = page.getElementsByClassName("view-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onViewPageBackButton(page, backButton);
  });

  let closeButton = page.getElementsByClassName("view-page-close-button")[0];
  closeButton.addEventListener("click", function() {
    onViewPageCloseButton(page, closeButton);
  });

  let editButton = page.getElementsByClassName("view-page-edit-button")[0];
  editButton.addEventListener("click", function() {
    onViewPageEditButton(page, editButton);
  });

  let deleteButton = page.getElementsByClassName("view-page-delete-button")[0];
  deleteButton.addEventListener("click", function() {
    onViewPageDeleteButton(page, deleteButton);
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

function onViewPageCloseButton(page, closeButton) {
  window.close();
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

function onViewPageDeleteButton(page, deleteButton) {
  let title = page.getElementsByClassName("view-page-filename")[0]
  let file = title.textContent;
  let req = {delete_file:{file:file}};
  callBackend(req, function(req, resp) {
    onViewPageDeleteFileResponse(page, req, resp);
  });
}

function onViewPageDeleteFileResponse(page, req, resp) {
  setOperationStatus("File %q is removed successfully.");

  let searchPage = createSearchPage();
  showPage(searchPage, "search-page", onSearchPageDisplay);
}

function onViewPageEditButton(page, editButton) {
  let params = JSON.parse(page.getAttribute("page-params"));

  let entry = {
    username: params.view_file.username,
    password: params.view_file.password,
    sitename: params.view_file.sitename,
    filename: params.view_file.filename,
    data: params.view_file.data,
  };

  let editPage = createEditPage(entry);
  showPage(editPage, "edit-page", onEditPageDisplay);
}
