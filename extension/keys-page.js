'use strict';

function createKeysPage(params) {
  let keysPageTemplate = document.getElementById("keys-page-template");
  let page = keysPageTemplate.cloneNode(true);

  page.setAttribute("page-params", "{}");
  if (params) {
    page.setAttribute("page-params", JSON.stringify(params));
  }

  let backButton = page.getElementsByClassName("keys-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onKeysPageBackButton(page, backButton);
  });

  let closeButton = page.getElementsByClassName("keys-page-close-button")[0];
  closeButton.addEventListener("click", function() {
    window.close();
  });

  let addButtons = page.getElementsByClassName("keys-page-add-button");
  for (let i = 0; i < addButtons.length; i++) {
    let addButton = addButtons[i];
    addButton.addEventListener("click", function() {
      onKeysPageAddButton(page, addButton);
    });
  }

  return page;
}

function onKeysPageDisplay(page) {
  let params = JSON.parse(page.getAttribute("page-params"))

  let lcount = page.getElementsByClassName("keys-page-localkey-count")[0];
  let rcount = page.getElementsByClassName("keys-page-remotekey-count")[0];
  let ecount = page.getElementsByClassName("keys-page-expiredkey-count")[0];
  let ucount = page.getElementsByClassName("keys-page-untrustedkey-count")[0];
  lcount.textContent = 0;
  rcount.textContent = 0;
  ecount.textContent = 0;
  ucount.textContent = 0;

  // FIXME: We should remove any existing visible entries if we want to call
  // onKeysPageDisplay with different result again.

  if (params.check_status.local_keys) {
    lcount = params.check_status.local_keys.length;
    let ltemplate = page.getElementsByClassName("keys-page-local-key-template")[0];
    // TODO: Remove all nextSiblings of ltemplate; they shall be overwritten.
    for (let i = 0; i < params.check_status.local_keys.length; i++) {
      let key = params.check_status.local_keys[i];

      let newkey = ltemplate.cloneNode(true);
      newkey.getElementsByClassName("keys-page-localkey-fingerprint")[0].textContent = key.fingerprint;
      newkey.getElementsByClassName("keys-page-localkey-username")[0].textContent = key.user_name;
      newkey.getElementsByClassName("keys-page-localkey-useremail")[0].textContent = key.user_email;

      newkey.style.display = "";
      ltemplate.parentNode.insertBefore(newkey, ltemplate.nextSibling);
    }
  }

  if (params.check_status.remote_keys) {
    rcount = params.check_status.remote_keys.length;
    let rtemplate = page.getElementsByClassName("keys-page-remote-key-template")[0];
    // TOOD: Remove all nextSiblings of rtemplate; they shall be overwritten.
    for (let i = 0; i < params.check_status.remote_keys.length; i++) {
      let key = params.check_status.remote_keys[i];

      let newkey = rtemplate.cloneNode(true);
      newkey.getElementsByClassName("keys-page-remotekey-fingerprint")[0].textContent = key.fingerprint;
      newkey.getElementsByClassName("keys-page-remotekey-username")[0].textContent = key.user_name;
      newkey.getElementsByClassName("keys-page-remotekey-useremail")[0].textContent = key.user_email;

      newkey.style.display = "";
      rtemplate.parentNode.insertBefore(newkey, rtemplate.nextSibling);
    }
  }

  if (params.check_status.expired_keys) {
    ecount = params.check_status.expired_keys.length;
    let template = page.getElementsByClassName("keys-page-expired-key-template")[0];
    // TOOD: Remove all nextSiblings of template; they shall be overwritten.
    for (let i = 0; i < params.check_status.expired_keys.length; i++) {
      let key = params.check_status.expired_keys[i];

      let newkey = template.cloneNode(true);
      newkey.getElementsByClassName("keys-page-expiredkey-fingerprint")[0].textContent = key.fingerprint;
      newkey.getElementsByClassName("keys-page-expiredkey-username")[0].textContent = key.user_name;
      newkey.getElementsByClassName("keys-page-expiredkey-useremail")[0].textContent = key.user_email;

      newkey.style.display = "";
      template.parentNode.insertBefore(newkey, template.nextSibling);
    }
  }

  if (params.check_status.untrusted_keys) {
    ucount = params.check_status.untrusted_keys.length;
    let template = page.getElementsByClassName("keys-page-untrusted-key-template")[0];
    // TOOD: Remove all nextSiblings of template; they shall be overwritten.
    for (let i = 0; i < params.check_status.untrusted_keys.length; i++) {
      let key = params.check_status.untrusted_keys[i];

      let newkey = template.cloneNode(true);
      newkey.getElementsByClassName("keys-page-untrustedkey-fingerprint")[0].textContent = key.fingerprint;
      newkey.getElementsByClassName("keys-page-untrustedkey-username")[0].textContent = key.user_name;
      newkey.getElementsByClassName("keys-page-untrustedkey-useremail")[0].textContent = key.user_email;

      newkey.style.display = "";
      template.parentNode.insertBefore(newkey, template.nextSibling);
    }
  }
}

function onKeysPageBackButton(page, backButton) {
  let req = {check_status:{}};
  backgroundPage.callBackend(req, function(resp) {
    let settingsPage = createSettingsPage(resp);
    showPage(settingsPage, "settings-page", onSettingsPageDisplay);
  });
}

function onKeysPageAddButton(page, addButton) {
  let addkeyPage = createAddkeyPage();
  showPage(addkeyPage, "addkey-page", onAddkeyPageDisplay);
}
