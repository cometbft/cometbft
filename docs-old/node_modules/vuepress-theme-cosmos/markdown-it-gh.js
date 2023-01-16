var child_process = require("child_process");

function replaceUnsafeChar(ch) {
  return HTML_REPLACEMENTS[ch];
}

var HTML_REPLACEMENTS = {
  "&": "&amp;",
  "<": "&lt;",
  ">": "&gt;",
  '"': "&quot;",
};

function escapeHtml(str) {
  if (/[&<>"]/.test(str)) {
    return str.replace(/[&<>"]/g, replaceUnsafeChar);
  }
  return str;
}

module.exports = function asciiDiagram(md, options) {
  md.renderer.rules.text = function (tokens, idx, options, env) {
    const a = tokens[idx].content.split(" ");
    if (a[0] == "+++" && a[1]) {
      const isGithub = /github.com/.test(a[1]);
      const url = isGithub
        ? a[1]
            .replace("github.com", "raw.githubusercontent.com")
            .replace("/blob/", "/")
        : a[1];
      let data = child_process.spawnSync("curl", [url]).stdout.toString();
      const hasLines = a[1].match(/#L[0-9]+-L[0-9]+$/);
      const golang = !!a[1].replace(/#L.*$/, "").match(/\.go$/);
      if (data && hasLines) {
        const [lineStart, lineEnd] =
          hasLines &&
          hasLines[0].replace(/#/g, "").replace(/L/g, "").split("-");
        data = data
          .split("\n")
          .slice(+lineStart - 1, +lineEnd)
          .join("\n");
      }
      const lang = golang && "go";
      const base64 = Buffer.from(escapeHtml(data)).toString("base64");
      return `<tm-code-block class="codeblock" language="${lang}" base64="${base64}"></tm-code-block>`;
    }
    return escapeHtml(tokens[idx].content);
  };
};
