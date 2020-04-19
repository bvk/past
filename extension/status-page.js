'use strict';

function createStatusPage(params) {
  let template = document.getElementById("status-page-template");
  let page = template.cloneNode(true);

  page.setAttribute("page-params", "{}");
  if (params) {
    page.setAttribute("page-params", JSON.stringify(params));
  }

  let backButton = page.getElementsByClassName("status-page-back-button")[0];
  backButton.addEventListener("click", function() {
    onStatusPageBackButton(page, backButton);
  });

  let closeButton = page.getElementsByClassName("status-page-close-button")[0];
  closeButton.addEventListener("click", function() {
    window.close();
  });

  return page;
}

function getStatusPageNextKeyIDType(idtype) {
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

function onStatusPageDisplay(page) {
  let params = JSON.parse(page.getAttribute("page-params"));

  if (params.scan_store.missing_key_file_count_map) {
    let template = page.getElementsByClassName("status-page-missing-key-template")[0];
    for (let next = template.nextElementSibling; next != null;) {
      let temp = next.nextElementSibling;
      next.remove();
      next = temp;
    }

    for (let key in params.scan_store.missing_key_file_count_map) {
      let count = params.scan_store.missing_key_file_count_map[key];

      let newkey = template.cloneNode(true);
      newkey.getElementsByClassName("status-page-missing-key")[0].textContent = key;
      newkey.getElementsByClassName("status-page-missing-filecount")[0].textContent = count;

      newkey.style.display = "";
      template.parentNode.insertBefore(newkey, template.nextSibling);
    }
  }

  if (params.scan_store.key_file_count_map) {
    let numDecrpt = getStatusPageNumReadableDecryptKeys(page, params);
    let template = page.getElementsByClassName("status-page-recipient-template")[0];
    for (let next = template.nextElementSibling; next != null;) {
      let temp = next.nextElementSibling;
      next.remove();
      next = temp;
    }

    for (let key in params.scan_store.key_file_count_map) {
      let count = params.scan_store.key_file_count_map[key];
      let pkey = params.scan_store.key_map[key];

      let newkey = template.cloneNode(true);
      let removeButton = newkey.getElementsByClassName("status-page-recipient-remove")[0];
      removeButton.addEventListener("click", function() {
        onStatusPageRemoveButton(page, pkey, removeButton);
      });
      removeButton.disabled = pkey.can_decrypt && numDecrpt == 1 && count > 0;

      let keyid = newkey.getElementsByClassName("status-page-recipient-keyid")[0];
      keyid.setAttribute("username", pkey.user_name);
      keyid.setAttribute("useremail", pkey.user_email);
      keyid.setAttribute("fingerprint", pkey.key_fingerprint);

      keyid.textContent = pkey.user_name;
      keyid.setAttribute("keyid_type", "username");
      keyid.addEventListener("click", function() {
        let idtype = getStatusPageNextKeyIDType(keyid.getAttribute("keyid_type"));
        keyid.textContent = keyid.getAttribute(idtype);
        keyid.setAttribute("keyid_type", idtype);
      });

      newkey.style.display = "";
      template.parentNode.insertBefore(newkey, template.nextSibling);
    }
  }

  if (params.scan_store.unused_key_map) {
    let template = page.getElementsByClassName("status-page-available-template")[0];
    for (let next = template.nextElementSibling; next != null;) {
      let temp = next.nextElementSibling;
      next.remove();
      next = temp;
    }

    for (let key in params.scan_store.unused_key_map) {
      let pkey = params.scan_store.unused_key_map[key];

      let newkey = template.cloneNode(true);
      let addButton = newkey.getElementsByClassName("status-page-available-add")[0];
      addButton.addEventListener("click", function() {
        onStatusPageAddButton(page, pkey, addButton);
      });
      addButton.disabled = !pkey.is_trusted;

      let keyid = newkey.getElementsByClassName("status-page-available-keyid")[0];
      keyid.setAttribute("username", pkey.user_name);
      keyid.setAttribute("useremail", pkey.user_email);
      keyid.setAttribute("fingerprint", pkey.key_fingerprint);

      keyid.textContent = pkey.user_name;
      keyid.setAttribute("keyid_type", "username");
      keyid.addEventListener("click", function() {
        let idtype = getStatusPageNextKeyIDType(keyid.getAttribute("keyid_type"));
        keyid.textContent = keyid.getAttribute(idtype);
        keyid.setAttribute("keyid_type", idtype);
      });

      newkey.style.display = "";
      template.parentNode.insertBefore(newkey, template.nextSibling);
    }
  }
}

function onStatusPageBackButton(page, backButton) {
  let req = {check_status:{}};
  callBackend(req, function(req, resp) {
    let settingsPage = createSettingsPage(resp);
    showPage(settingsPage, "settings-page", onSettingsPageDisplay);
  });
}

function onStatusPageRemoveButton(page, pkey, removeButton) {
  let params = JSON.parse(page.getAttribute("page-params"));
  let nreadable = getStatusPageNumReadable(page, params);

  let req = {
    remove_recipient: {
      fingerprint: pkey.key_fingerprint,
      num_skip: params.scan_store.num_files - nreadable,
    },
  };

  setOperationStatus("Removing...");
  callBackend(req, function(req, resp) {
    setOperationStatus("Removed");
    let params = {scan_store:resp.remove_recipient.scan_store};
    page.setAttribute("page-params", JSON.stringify(params));
    onStatusPageDisplay(page);
  });
}

function onStatusPageAddButton(page, pkey, addButton) {
  let params = JSON.parse(page.getAttribute("page-params"));
  let nreadable = getStatusPageNumReadable(page, params);

  let req = {
    add_recipient: {
      fingerprint: pkey.key_fingerprint,
      num_skip: params.scan_store.num_files - nreadable,
    },
  };

  setOperationStatus("Adding...");
  callBackend(req, function(req, resp) {
    setOperationStatus("Added");
    let params = {scan_store:resp.add_recipient.scan_store};
    page.setAttribute("page-params", JSON.stringify(params));
    onStatusPageDisplay(page);
  });
}

// getStatusPageNextKeyIDType returns number of files readable with the secret
// keys available locally.
function getStatusPageNumReadable(page, params) {
  let nreadable = 0;
  for (let key in params.scan_store.key_file_count_map) {
    let count = params.scan_store.key_file_count_map[key];
    let pkey = params.scan_store.key_map[key];
    if (pkey.can_decrypt && count > nreadable) {
      nreadable = count;
    }
  }
  return nreadable;
}

// getStatusPageNumReadableDecryptKeys returns number of decryption keys with
// readable secrets.
function getStatusPageNumReadableDecryptKeys(page, params) {
  let nkeys = 0;
  for (let key in params.scan_store.key_file_count_map) {
    let count = params.scan_store.key_file_count_map[key];
    let pkey = params.scan_store.key_map[key];
    if (pkey.can_decrypt && count > 0) {
      nkeys++;
    }
  }
  return nkeys;
}
