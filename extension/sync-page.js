'use strict';

function createSyncPage(params) {
  let syncPageTemplate = document.getElementById("sync-page-template");
  let page = syncPageTemplate.cloneNode(true);

  if (!params) {
    return null;
  }
  page.setAttribute("page-params", JSON.stringify(params));

  let backButton = page.getElementsByClassName("sync-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onSyncPageBackButton(page, backButton);
  });

  let closeButton = page.getElementsByClassName("sync-page-close-button")[0];
  closeButton.addEventListener("click", function() {
    window.close();
  });

  let fetchButton = page.getElementsByClassName("sync-page-fetch-button")[0];
  fetchButton.addEventListener("click", function() {
    onSyncPageFetchButton(page, fetchButton);
  });

  let pullButton = page.getElementsByClassName("sync-page-pull-button")[0];
  pullButton.addEventListener("click", function() {
    onSyncPagePullButton(page, pullButton);
  });

  let pushButton = page.getElementsByClassName("sync-page-push-button")[0];
  pushButton.addEventListener("click", function() {
    onSyncPagePushButton(page, pushButton);
  })

  return page;
}

function onSyncPageDisplay(page) {
  let params = JSON.parse(page.getAttribute("page-params"));

  let lcommit = page.getElementsByClassName("sync-page-local-commit")[0];
  lcommit.textContent = params.sync_remote.head.commit;

  let lauthor = page.getElementsByClassName("sync-page-local-author")[0];
  lauthor.textContent = params.sync_remote.head.author;

  let lauthordate = page.getElementsByClassName("sync-page-local-authordate")[0];
  lauthordate.textContent = params.sync_remote.head.author_date;

  let ltitle = page.getElementsByClassName("sync-page-local-title")[0];
  ltitle.textContent = params.sync_remote.head.title;

  let rcommit = page.getElementsByClassName("sync-page-remote-commit")[0];
  rcommit.textContent = params.sync_remote.remote.commit;

  let rauthor = page.getElementsByClassName("sync-page-remote-author")[0];
  rauthor.textContent = params.sync_remote.remote.author;

  let rauthordate = page.getElementsByClassName("sync-page-remote-authordate")[0];
  rauthordate.textContent = params.sync_remote.remote.author_date;

  let rtitle = page.getElementsByClassName("sync-page-remote-title")[0];
  rtitle.textContent = params.sync_remote.remote.title;

  let pullButton = page.getElementsByClassName("sync-page-pull-button")[0];
  let pushButton = page.getElementsByClassName("sync-page-push-button")[0];
  if (params.sync_remote.head.commit == params.sync_remote.remote.commit) {
    pullButton.disabled = true;
    pushButton.disabled = true;
    setOperationStatus("Synced.");
    return;
  }

  if (params.sync_remote.newer_commit == params.sync_remote.remote.commit) {
    pullButton.disabled = false;
    return;
  }

  if (params.sync_remote.newer_commit == params.sync_remote.head.commit) {
    pushButton.disabled = false;
    return;
  }

  pushButton.disabled = false;
  pullButton.disabled = false;
  pushButton.textContent = "publish";
  pullButton.textContent = "get_app";
  setOperationStatus("Diverged. Syncing will overwrite.");
}

function onSyncPageBackButton(page, backButton) {
  let req = {check_status:{}};
  callBackend(req, function(req, resp) {
    let settingsPage = createSettingsPage(resp);
    showPage(settingsPage, "settings-page", onSettingsPageDisplay);
  });
}

function onSyncPageFetchButton(page, fetchButton) {
  let req = {sync_remote:{fetch:true}};
  callBackend(req, function(req, resp) {
    page.setAttribute("page-params", JSON.stringify(resp));
    onSyncPageDisplay(page);
  });
}

function onSyncPagePushButton(page, pushButton) {
  let req = {sync_remote:{push:true}};
  callBackend(req, function(req, resp) {
    page.setAttribute("page-params", JSON.stringify(resp));
    onSyncPageDisplay(page);
  });
}

function onSyncPagePullButton(page, pullButton) {
  let req = {sync_remote:{pull:true}};
  callBackend(req, function(req, resp) {
    page.setAttribute("page-params", JSON.stringify(resp));
    onSyncPageDisplay(page);
  });
}
