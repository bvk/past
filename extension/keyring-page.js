'use strict';

function createKeyringPage(params) {
  let keyringPageTemplate = document.getElementById("keyring-page-template");
  let page = keyringPageTemplate.cloneNode(true);

  page.setAttribute("page-params", "{}");
  if (params) {
    page.setAttribute("page-params", JSON.stringify(params));
  }

  let backButton = page.getElementsByClassName("keyring-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onKeyringPageBackButton(page, backButton);
  });

  let closeButton = page.getElementsByClassName("keyring-page-close-button")[0];
  closeButton.addEventListener("click", function() {
    window.close();
  });

  let addButtons = page.getElementsByClassName("keyring-page-add-button");
  for (let i = 0; i < addButtons.length; i++) {
    let addButton = addButtons[i];
    addButton.addEventListener("click", function() {
      onKeyringPageAddButton(page, addButton);
    });
  }

  return page;
}

function getKeyringPageNextKeyIDType(idtype) {
  if (idtype == "username") {
    return "useremail";
  }
  if (idtype == "useremail") {
    return "fingerprint";
  }
  if (idtype == "fingerprint") {
    return "username";
  }
  return idtype;
}

function onKeyringPageDisplay(page) {
  let params = JSON.parse(page.getAttribute("page-params"))
  console.log(params);

  // FIXME: We should remove any existing visible entries if we want to call
  // onKeyringPageDisplay with different result again.

  if (params.check_status.local_keys) {
    let template = page.getElementsByClassName("keyring-page-local-key-template")[0];
    // TODO: Remove all nextSiblings of template; they shall be overwritten.
    for (let i = 0; i < params.check_status.local_keys.length; i++) {
      let key = params.check_status.local_keys[i];

      let newkey = template.cloneNode(true);
      let viewButton = newkey.getElementsByClassName("keyring-page-localkey-view")[0];
      viewButton.addEventListener("click", function() {
        let p = createViewkeyPage(key, params);
        showPage(p, "viewkey-page", onViewkeyPageDisplay);
      });

      let keyid = newkey.getElementsByClassName("keyring-page-localkey-keyid")[0];
      keyid.setAttribute("username", key.user_name);
      keyid.setAttribute("useremail", key.user_email);
      keyid.setAttribute("fingerprint", key.key_fingerprint);

      keyid.textContent = key.user_name;
      keyid.setAttribute("keyid_type", "username");
      keyid.addEventListener("click", function() {
        let idtype = getKeyringPageNextKeyIDType(keyid.getAttribute("keyid_type"));
        keyid.textContent = keyid.getAttribute(idtype);
        keyid.setAttribute("keyid_type", idtype);
      });

      newkey.style.display = "";
      template.parentNode.insertBefore(newkey, template.nextSibling);
    }
  }

  if (params.check_status.remote_keys) {
    let template = page.getElementsByClassName("keyring-page-remote-key-template")[0];
    // TOOD: Remove all nextSiblings of template; they shall be overwritten.
    for (let i = 0; i < params.check_status.remote_keys.length; i++) {
      let key = params.check_status.remote_keys[i];

      let newkey = template.cloneNode(true);
      let viewButton = newkey.getElementsByClassName("keyring-page-remotekey-view")[0];
      viewButton.addEventListener("click", function() {
        let p = createViewkeyPage(key, params);
        showPage(p, "viewkey-page", onViewkeyPageDisplay);
      });

      let keyid = newkey.getElementsByClassName("keyring-page-remotekey-keyid")[0];
      keyid.setAttribute("username", key.user_name);
      keyid.setAttribute("useremail", key.user_email);
      keyid.setAttribute("fingerprint", key.key_fingerprint);

      keyid.textContent = key.user_name;
      keyid.setAttribute("keyid_type", "username");
      keyid.addEventListener("click", function() {
        let idtype = getKeyringPageNextKeyIDType(keyid.getAttribute("keyid_type"));
        keyid.textContent = keyid.getAttribute(idtype);
        keyid.setAttribute("keyid_type", idtype);
      });

      newkey.style.display = "";
      template.parentNode.insertBefore(newkey, template.nextSibling);
    }
  }
}

function onKeyringPageBackButton(page, backButton) {
  let req = {check_status:{}};
  backgroundPage.callBackend(req, function(resp) {
    let settingsPage = createSettingsPage(resp);
    showPage(settingsPage, "settings-page", onSettingsPageDisplay);
  });
}

function onKeyringPageAddButton(page, addButton) {
  let addkeyPage = createAddkeyPage();
  showPage(addkeyPage, "addkey-page", onAddkeyPageDisplay);
}
