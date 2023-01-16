/* eslint-env es6 */
const md = require('markdown-it')();
const markdownItAttrs = require('./');

md.use(markdownItAttrs).use(require('../markdown-it-implicit-figures'));

let src = `header1 | header2
------- | -------
column1 | column2

{.special}
`;

let res = md.render(src);

console.log(res);  // eslint-disable-line
