local config = require("lapis.config").get()

local blobs = {}

blobs.main_css = [[
body {
  font: 15px/1.3 Arial, sans-serif;
}

table {
  margin-top: 0.3em;
  border-collapse: collapse;
}

th,td {
  font-weight: normal;
  padding: 0.5em;
}

h1, h2 {
  margin-top: 5px;
  margin-bottom: 5px;
}
]]

blobs.highlight_default = [[
.hljs{display:block;overflow-x:auto;padding:0.5em}.hljs,.hljs-subst{color:#444}.hljs-keyword,.hljs-attribute,.hljs-selector-tag,.hljs-meta-keyword,.hljs-doctag,.hljs-name{font-weight:bold}.hljs-built_in,.hljs-literal,.hljs-bullet,.hljs-code,.hljs-addition{color:#1F811F}.hljs-regexp,.hljs-symbol,.hljs-variable,.hljs-template-variable,.hljs-link,.hljs-selector-attr,.hljs-selector-pseudo{color:#BC6060}.hljs-type,.hljs-string,.hljs-number,.hljs-selector-id,.hljs-selector-class,.hljs-quote,.hljs-template-tag,.hljs-deletion{color:#880000}.hljs-title,.hljs-section{color:#880000;font-weight:bold}.hljs-comment{color:#888888}.hljs-meta{color:#2B6EA1}.hljs-emphasis{font-style:italic}.hljs-strong{font-weight:bold}]]

-- https://stackoverflow.com/a/17928466
blobs.apply_highlightjs = ([[
(function() {
  var s = document.createElement('script');
  s.type = 'text/javascript';
  s.src = %q;
  s.async = true;
  s.onload = function() {
    hljs.initHighlightingOnLoad();
  };
  document.body.appendChild(s);
})();
]]):format(config.highlight_js_path .. 'highlight-custom.pack.js')

return blobs
