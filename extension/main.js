'use strict';

document.addEventListener('DOMContentLoaded', function () {
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
	  go.run(result.instance);
  });
});
