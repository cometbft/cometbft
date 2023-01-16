const matter = require("gray-matter");
const attrs = require("markdown-it-attrs");
const md = require("markdown-it")().use(attrs, {
  allowedAttributes: ["prereq", "hide", "synopsis"],
});
const cheerio = require("cheerio");

module.exports = (options = {}, context) => ({
  extendPageData($page) {
    let description = "";
    let frontmatter = {};
    try {
      const $ = cheerio.load(md.render($page._content));
      description = $("[synopsis]").text();
    } catch {
      console.log(
        `Error in processing description: $page.content is ${$page._content}`
      );
    }
    try {
      frontmatter = matter($page._content, { delims: ["<!--", "-->"] }).data;
    } catch {
      console.log(
        `Error in processing frontmatter: $page.content is ${$page._content}`
      );
    }
    $page.frontmatter = {
      description,
      ...$page.frontmatter,
      ...frontmatter,
    };
    try {
      const tokens = md.parse($page._content, {});
      tokens.forEach((t, i) => {
        if (t.type === "heading_open" && ["h1"].includes(t.tag)) {
          $page.title = tokens[i + 1].content;
          return;
        }
      });
    } catch {
      console.log(`Error in processing headings.`);
    }
  },
});
