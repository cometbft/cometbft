var ghmd = require("./markdown-it-gh.js")
var fcb = require("./markdown-it-fcb.js")

function replaceUnsafeChar(ch) {
  return HTML_REPLACEMENTS[ch];
}

var HTML_REPLACEMENTS = {
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;',
  '"': '&quot;'
};

function escapeHtml(str) {
  if (/[&<>"]/.test(str)) {
    return str.replace(/[&<>"]/g, replaceUnsafeChar);
  }
  return str;
}

module.exports = (opts, ctx) => {
  return {
    plugins: [
      require('./plugin-frontmatter.js'),
      ...["warning", "tip", "danger"].map(type => [
        "container",
        { type, defaultTitle: false }
      ]),
    ],
    extendMarkdown: md => {
      md.use(ghmd)
      md.use(fcb)
      md.use(require('markdown-it-attrs'), {
        allowedAttributes: ['prereq', 'hide', 'synopsis']
      })
    }
  }
}
