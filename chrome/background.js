'use strict';

function listFiles(req, callback) {
  chrome.runtime.sendNativeMessage('github.bvk.past', req, function(resp) {
    callback(resp);
  });
}

function viewFile(req, callback) {
  chrome.runtime.sendNativeMessage('github.bvk.past', req, function(resp) {
    callback(resp);
  });
}

function setLocalStorage(state, callback) {
  chrome.storage.local.set(state, callback);
}

function getLocalStorage(keys, callback) {
  chrome.storage.local.get(keys, callback);
}

function copyPassword(password) {
  // See https://htmldom.dev/copy-text-to-the-clipboard

  // Create a fake textarea
  const textAreaEle = document.createElement('textarea');

  // Reset styles
  textAreaEle.style.border = '0';
  textAreaEle.style.padding = '0';
  textAreaEle.style.margin = '0';

  // Set the absolute position
  // User won't see the element
  textAreaEle.style.position = 'absolute';
  textAreaEle.style.left = '-9999px';
  textAreaEle.style.top = `0px`;

  // Set the value
  textAreaEle.value = password;

  // Append the textarea to body
  document.body.appendChild(textAreaEle);

  // Focus and select the text
  textAreaEle.focus();
  textAreaEle.select();

  // Execute the "copy" command
  try {
    document.execCommand('copy');
  } catch (err) {
    return false
  } finally {
    // Remove the textarea
    document.body.removeChild(textAreaEle);
  }

  // Schedule a callback to clear the password.
  if (password != "*") {
    setTimeout(function() {
      copyPassword("*");
    }, 10*1000);
  }
  return true
}
