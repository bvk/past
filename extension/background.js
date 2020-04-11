'use strict';

function callBackend(req, callback) {
  chrome.runtime.sendNativeMessage('github.bvk.past', req, callback);
}

function setLocalStorage(state, callback) {
  chrome.storage.local.set(state, callback);
}

function getLocalStorage(keys, callback) {
  chrome.storage.local.get(keys, callback);
}

// copyString writes the password into the clipboard for timeout seconds and
// invokes the callback when password is cleared from the clipboard.
function copyString(password, timeout, callback) {
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

  // Schedule a callback to clear the password. FIXME: Scheduled callback
  // clears the clipboard content without checking if it was actually our
  // content.
  if (password != "*") {
    if (timeout && timeout > 0) {
      setTimeout(function() {
        copyString("*", 0, callback);
      }, timeout*1000);
    }
  } else {
    callback();
  }
  return true
}
