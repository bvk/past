'use strict';

function createViewkeyPage(params, backParams) {
  let template = document.getElementById("viewkey-page-template");
  let page = template.cloneNode(true);

  page.setAttribute("page-params", "{}");
  page.setAttribute("back-page-params", "{}");
  if (params) {
    page.setAttribute("page-params", JSON.stringify(params));
    page.setAttribute("back-page-params", JSON.stringify(backParams));
  }

  let backButton = page.getElementsByClassName("viewkey-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onViewkeyPageBackButton(page, backButton);
  });

  let closeButton = page.getElementsByClassName("viewkey-page-close-button")[0];
  closeButton.addEventListener("click", function() {
    window.close();
  });

  let copyButton = page.getElementsByClassName("viewkey-page-copy-button")[0];
  copyButton.addEventListener("click", function() {
    backgroundPage.copyString(params.key_fingerprint);
    setOperationStatus("Copied");
  });

  let toggleButton = page.getElementsByClassName("viewkey-page-trust-toggle")[0];
  toggleButton.addEventListener("click", function() {
    onViewkeyPageTrustToggleButton(page, toggleButton);
  });

  let exportButton = page.getElementsByClassName("viewkey-page-export-button")[0];
  exportButton.addEventListener("click", function() {
    onViewkeyPageExportButton(page, exportButton);
  });

  let deleteButton = page.getElementsByClassName("viewkey-page-delete-button")[0];
  deleteButton.addEventListener("click", function() {
    onViewkeyPageDeleteButton(page, deleteButton);
  });

  return page;
}

function onViewkeyPageDisplay(page) {
  let params = JSON.parse(page.getAttribute("page-params"));

  page.getElementsByClassName("viewkey-page-key-fingerprint")[0].textContent = params.key_fingerprint;
  page.getElementsByClassName("viewkey-page-key-username")[0].textContent = params.user_name;
  page.getElementsByClassName("viewkey-page-key-useremail")[0].textContent = params.user_email;

  let trusted = page.getElementsByClassName("viewkey-page-key-trusted")[0];
  if (params.is_trusted) {
    trusted.textContent = "Trusted";
  } else {
    trusted.textContent = "Not-Trusted";
  }

  let toggleButton = page.getElementsByClassName("viewkey-page-trust-toggle")[0];
  if (params.is_trusted) {
    toggleButton.textContent = "highlight_off";
  } else {
    toggleButton.textContent = "check_circle_outline";
  }

  let expired = page.getElementsByClassName("viewkey-page-key-expired")[0];
  if (params.is_expired) {
    expired.textContent = "Expired";
  } else if (params.days_to_expire == 0) {
    expired.textContent = "Never Expires";
  } else {
    expired.textContent = "Expires in "+params.days_to_expire+" Days";
  }
}

function onViewkeyPageBackButton(page, backButton) {
  let req = {check_status:{}};
  callBackend(req, function(req, resp) {
    let keyringPage = createKeyringPage(resp);
    showPage(keyringPage, "keyring-page", onKeyringPageDisplay);
  });
}

function onViewkeyPageTrustToggleButton(page, toggleButton) {
  let params = JSON.parse(page.getAttribute("page-params"));
  let req = {
    edit_key: {
      fingerprint: params.key_fingerprint,
      trust: !params.is_trusted,
    }
  };
  callBackend(req, function(req, resp) {
    page.setAttribute("page-params", JSON.stringify(resp.edit_key.key));
    onViewkeyPageDisplay(page);
  });
}

function onViewkeyPageExportButton(page, exportButton) {
  let params = JSON.parse(page.getAttribute("page-params"));
  let req = {export_key:{fingerprint: params.key_fingerprint}};
  callBackend(req, function(req, resp) {
    backgroundPage.copyString(resp.export_key.armor_key);
    setOperationStatus("Exported to Clipboard");
  });
}

function onViewkeyPageDeleteButton(page, deleteButton) {
  let params = JSON.parse(page.getAttribute("page-params"));
  let req = {delete_key:{fingerprint: params.key_fingerprint}};
  callBackend(req, function(req, resp) {
    let backButton = page.getElementsByClassName("viewkey-page-back-button")[0];
    onViewkeyPageBackButton(page, backButton);
  });
}
