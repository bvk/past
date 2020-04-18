'use strict';

function createNewremotePage(params) {
  let newremotePageTemplate = document.getElementById("newremote-page-template");
  let page = newremotePageTemplate.cloneNode(true);

  page.setAttribute("page-params", "{}");
  if (params) {
    page.setAttribute("page-params", JSON.stringify(params));
  }

  let backButton = page.getElementsByClassName("newremote-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onNewremotePageBackButton(page, backButton);
  });

  let closeButton = page.getElementsByClassName("newremote-page-close-button")[0];
  closeButton.addEventListener("click", function() {
    window.close();
  });

  let serverSelect = page.getElementsByClassName("newremote-page-gitserver")[0];
  serverSelect.addEventListener("change", function() {
    onNewremotePageGitServerChange(page, serverSelect);
  });

  let githost = page.getElementsByClassName("newremote-page-githost")[0];
  githost.addEventListener("input", function() {
    autoNewremotePageDoneButton(page);
    autoNewremotePageUndoButton(page);
  });

  let gituser = page.getElementsByClassName("newremote-page-gituser")[0];
  gituser.addEventListener("input", function() {
    autoNewremotePageDoneButton(page);
    autoNewremotePageUndoButton(page);
  });

  let gitpass = page.getElementsByClassName("newremote-page-gitpass")[0];
  gitpass.addEventListener("input", function() {
    autoNewremotePageDoneButton(page);
    autoNewremotePageUndoButton(page);
  });

  let gitpassToggle = page.getElementsByClassName("newremote-page-gitpass-toggle")[0];
  gitpassToggle.addEventListener("click", function() {
    onNewremotePageToggleButton(page, gitpassToggle);
  });

  let gitpath = page.getElementsByClassName("newremote-page-gitpath")[0];
  gitpath.addEventListener("input", function() {
    autoNewremotePageDoneButton(page);
    autoNewremotePageUndoButton(page);
  });

  let undoButton = page.getElementsByClassName("newremote-page-undo-button")[0];
  undoButton.addEventListener("click", function() {
    onNewremotePageUndoButton(page, undoButton);
  });

  let doneButton = page.getElementsByClassName("newremote-page-done-button")[0];
  doneButton.addEventListener("click", function() {
    onNewremotePageDoneButton(page, doneButton);
  });

  return page;
}

function onNewremotePageDisplay(params) {
}

function onNewremotePageBackButton(page, backButton) {
  let req = {check_status:{}};
  callBackend(req, function(req, resp) {
    let settingsPage = createSettingsPage(resp);
    showPage(settingsPage, "settings-page", onSettingsPageDisplay);
  });
}

function onNewremotePageGitServerChange(page, serverSelect) {
  let server = serverSelect.value;

  let githost = page.getElementsByClassName("newremote-page-githost")[0];
  if (server == "github-ssh" || server == "github-https") {
    githost.value = "github.com";
    githost.disabled = true;
  } else {
    githost.disabled = false;
  }

  let gituser = page.getElementsByClassName("newremote-page-gituser")[0];
  if (server == "github-ssh") {
    gituser.value = "git";
    gituser.disabled = true;
  } else if (server == "github-https" && gituser.value == "git") {
    gituser.value = "";
    gituser.disabled = false;
  } else {
    gituser.disabled = false;
  }

  let gitpass = page.getElementsByClassName("newremote-page-gitpass")[0];
  if (server == "ssh" || server == "github-ssh") {
    gitpass.disabled = false;
    gitpass.setAttribute("placeholder", "leave empty for password-less authentication")
  } else {
    gitpass.disabled = false;
    gitpass.setAttribute("placeholder", "password")
  }

  let gitpath = page.getElementsByClassName("newremote-page-gitpath")[0];
  if (server == "github-ssh" || server == "github-https") {
    gitpath.setAttribute("placeholder", "username/repository.git")
  } else {
    gitpath.setAttribute("placeholder", "path/to/repository.git")
  }
  gitpath.disabled = false;

  autoNewremotePageUndoButton(page);
  autoNewremotePageDoneButton(page);
}

function onNewremotePageUndoButton(page, undoButton) {
  let gitserver = page.getElementsByClassName("newremote-page-gitserver")[0];
  let githost = page.getElementsByClassName("newremote-page-githost")[0];
  let gituser = page.getElementsByClassName("newremote-page-gituser")[0];
  let gitpass = page.getElementsByClassName("newremote-page-gitpass")[0];
  let gitpath = page.getElementsByClassName("newremote-page-gitpath")[0];
  gitserver.value = "ssh";
  githost.value = "";
  githost.disabled = false;
  gituser.value = "";
  gituser.disabled = false;
  gitpass.value = "";
  gitpass.setAttribute("placeholder", "leave empty for password-less authentication");
  gitpath.value = "";

  if (!undoButton) {
    undoButton = page.getElementsByClassName("newremote-page-undo-button")[0];
  }
  undoButton.disabled = true;

  let doneButton = page.getElementsByClassName("newremote-page-done-button")[0];
  doneButton.disabled = true;
}

function onNewremotePageToggleButton(page, toggleButton) {
  let gitpass = page.getElementsByClassName("newremote-page-gitpass")[0];
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

function onNewremotePageDoneButton(page, doneButton) {
  let gitserver = page.getElementsByClassName("newremote-page-gitserver")[0];
  let githost = page.getElementsByClassName("newremote-page-githost")[0];
  let gituser = page.getElementsByClassName("newremote-page-gituser")[0];
  let gitpass = page.getElementsByClassName("newremote-page-gitpass")[0];
  let gitpath = page.getElementsByClassName("newremote-page-gitpath")[0];

  let protocol = "git";
  if (gitserver.value == "ssh" || gitserver.value == "github-ssh") {
    protocol = "ssh";
  } else if (gitserver.value == "https" || gitserver.value == "github-https") {
    protocol = "https";
  } else if (gitserver.value == "git") {
    protocol = "git";
  }

  let req = {
    add_remote: {
      protocol: protocol,
      username: gituser.value,
      password: gitpass.value,
      hostname: githost.value,
      path: gitpath.value,
    },
  }
  callBackend(req, function(req, resp) {
    let syncResp = {
      sync_remote: resp.add_remote.sync_remote,
    };
    let syncPage = createSyncPage(syncResp);
    showPage(syncPage, "sync-page", onSyncPageDisplay);
  });
}

function autoNewremotePageUndoButton(page, undoButton) {
  let gitserver = page.getElementsByClassName("newremote-page-gitserver")[0];
  let githost = page.getElementsByClassName("newremote-page-githost")[0];
  let gituser = page.getElementsByClassName("newremote-page-gituser")[0];
  let gitpass = page.getElementsByClassName("newremote-page-gitpass")[0];
  let gitpath = page.getElementsByClassName("newremote-page-gitpath")[0];

  var disable = true;
  if (gitserver.value != "ssh" || githost.value != "" || gituser.value != "" ||
      gitpass.value != "" || gitpath.value != "") {
    disable = false;
  }

  if (!undoButton) {
    undoButton = page.getElementsByClassName("newremote-page-undo-button")[0];
  }
  undoButton.disabled = disable;
}

function autoNewremotePageDoneButton(page, doneButton) {
  let gitserver = page.getElementsByClassName("newremote-page-gitserver")[0];
  let githost = page.getElementsByClassName("newremote-page-githost")[0];
  let gituser = page.getElementsByClassName("newremote-page-gituser")[0];
  let gitpass = page.getElementsByClassName("newremote-page-gitpass")[0];
  let gitpath = page.getElementsByClassName("newremote-page-gitpath")[0];

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
    doneButton = page.getElementsByClassName("newremote-page-done-button")[0];
  }
  console.log("donebutton.disabled = ", disable);
  doneButton.disabled = disable;
}
