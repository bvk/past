'use strict';

function createSettingsPage(params) {
  let settingsPageTemplate = document.getElementById("settings-page-template");
  let page = settingsPageTemplate.cloneNode(true);

  page.setAttribute("page-params", "{}");
  if (params) {
    page.setAttribute("page-params", JSON.stringify(params));
  }

  let backButton = page.getElementsByClassName("settings-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onSettingsPageBackButton(page, backButton);
  });

  let closeButton = page.getElementsByClassName("settings-page-close-button")[0];
  closeButton.addEventListener("click", function() {
    onSettingsPageCloseButton(page, closeButton);
  });

  let checkButton = page.getElementsByClassName("settings-page-check-button")[0];
  checkButton.addEventListener("click", function() {
    onSettingsPageCheckButton(page, checkButton);
  });

  let repoButton = page.getElementsByClassName("settings-page-repo-button")[0];
  repoButton.addEventListener("click", function() {
    onSettingsPageRepoButton(page, repoButton);
  });

  let keysButton = page.getElementsByClassName("settings-page-keys-button")[0];
  keysButton.addEventListener("click", function() {
    onSettingsPageKeysButton(page, keysButton);
  });

  let remoteButton = page.getElementsByClassName("settings-page-remote-button")[0];
  remoteButton.addEventListener("click", function() {
    onSettingsPageRemoteButton(page, remoteButton);
  });

  return page;
}

function onSettingsPageDisplay(page) {
  let pageParams = page.getAttribute("page-params");
  let params = JSON.parse(pageParams);

  let messagingReady = false;
  let toolsReady = false;
  let keysReady = false;
  let repoReady = false;
  let remoteReady = false;

  if (params.status == "" && params.check_status) {
    let messagingCheck = page.getElementsByClassName("settings-page-messaging-check")[0];
    messagingCheck.textContent = "done";
    messagingReady = true;

    let toolsCheck = page.getElementsByClassName("settings-page-tools-check")[0];
    if (!messagingReady || params.check_status.git_path == "" || params.check_status.gpg_path == "") {
      toolsCheck.textContent = "clear";
    } else {
      toolsCheck.textContent = "done";
      toolsReady = true;
    }

    let keysCheck = page.getElementsByClassName("settings-page-keys-check")[0];
    let keysButton = page.getElementsByClassName("settings-page-keys-button")[0];
    if (!toolsReady || !params.check_status.local_keys || params.check_status.local_keys.length == 0) {
      keysCheck.textContent = "clear";
      keysButton.disabled = !toolsReady;
    } else {
      keysReady = true;
      keysCheck.textContent = "done";
      keysButton.disabled = false;
      keysButton.textContent = "navigate_next";
    }

    let repoCheck = page.getElementsByClassName("settings-page-repo-check")[0];
    let repoButton = page.getElementsByClassName("settings-page-repo-button")[0];
    if (!keysReady || !params.check_status.password_store_keys || params.check_status.password_store_keys.length == 0) {
      repoCheck.textContent = "clear";
      repoButton.disabled = !keysReady;
    } else {
      repoReady = true;
      repoCheck.textContent = "done";
      repoButton.style.display = "none";
    }

    let remoteCheck = page.getElementsByClassName("settings-page-remote-check")[0];
    let remoteButton = page.getElementsByClassName("settings-page-remote-button")[0];
    if (!repoReady || !params.check_status.remote || params.check_status.remote == "") {
      remoteCheck.textContent = "clear";
      remoteButton.textContent = "create_new_folder";
      remoteButton.disabled = !repoReady;
    } else {
      remoteReady = true;
      remoteButton.disabled = false;
      remoteCheck.textContent = "done";
      remoteButton.textContent = "navigate_next";
    }
  }

  if (messagingReady && toolsReady && keysReady && repoReady) {
    setOperationStatus("Backend is now ready.");
  }
  autoSettingsPageBackButton(page);
}

function onSettingsPageBackButton(page, backButton) {
  let searchPage = createSearchPage();
  showPage(searchPage, "search-page", onSearchPageDisplay);
}

function onSettingsPageCloseButton(page, closeButton) {
  window.close();
}

function onSettingsPageCheckButton(page, checkButton) {
  let req = {check_status:{}};
  backgroundPage.callBackend(req, function(resp) {
    page.setAttribute("page-params", JSON.stringify(resp));
    onSettingsPageDisplay(page);
  });
}

function onSettingsPageRepoButton(page, repoButton) {
  let req = {check_status:{}};
  callBackend(req, function(req, resp) {
    let newrepoPage = createNewrepoPage(resp);
    showPage(newrepoPage, "newrepo-page", onNewrepoPageDisplay);
  });
}

function onSettingsPageKeysButton(page, keysButton) {
  let params = JSON.parse(page.getAttribute("page-params"));
  if (!params) {
    console.log("unexpected: params must exist when keys button is enabled");
    return
  }

  if (!params.check_status.local_keys || params.check_status.local_keys.length == 0) {
    let newkeyPage = createNewkeyPage();
    showPage(newkeyPage, "newkey-page", onNewkeyPageDisplay);
    return;
  }

  let keysPage = createKeysPage(params);
  showPage(keysPage, "keys-page", onKeysPageDisplay);
  return;
}

function onSettingsPageRemoteButton(page, remoteButton) {
  let params = JSON.parse(page.getAttribute("page-params"));
  if (!params) {
    console.log("unexpected: params must exist when remote button is enabled");
    return
  }

  if (!params.check_status.remote || params.check_status.remote == "") {
    let newremotePage = createNewremotePage();
    showPage(newremotePage, "newremote-page", onNewremotePageDisplay);
    return;
  }

  let req = {
    sync_remote: {
    },
  }
  callBackend(req, function(req, resp) {
    clearOperationStatus();
    let syncPage = createSyncPage(resp);
    showPage(syncPage, "sync-page", onSyncPageDisplay);
  });
}

function onSettingsPageCheckReponse(page, req, resp) {
  page.setAttribute("page-params", JSON.stringify(resp));
  onSettingsPageDisplay(page);
}

function onSettingsPageCreateRepoResponse(page, req, resp) {
  // Redirect to the check button.
  checkButton = page.getElementsByClassName("settings-page-check-button")[0];
  onSettingsPageCheckButton(page, checkButton)
}

function autoSettingsPageBackButton(page, backButton) {
  let pageParams = page.getAttribute("page-params");
  let params = JSON.parse(pageParams);

  let showSettings = true;
  if (params &&
      params.status == "" &&
      params.check_status.git_path != "" &&
      params.check_status.gpg_path != "" &&
      params.check_status.local_keys && params.check_status.local_keys.length > 0 &&
      params.check_status.password_store_keys && params.check_status.password_store_keys.length > 0) {
    showSettings = false;
  }

  if (!backButton) {
    backButton = page.getElementsByClassName("settings-page-back-button")[0];
  }
  backButton.disabled = showSettings;
}
