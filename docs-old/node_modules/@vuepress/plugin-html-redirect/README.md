# @vuepress/plugin-html-redirect

[![NPM version](https://img.shields.io/npm/v/@vuepress/plugin-html-redirect.svg?style=flat)](https://npmjs.com/package/@vuepress/plugin-html-redirect) [![NPM downloads](https://img.shields.io/npm/dm/@vuepress/plugin-html-redirect.svg?style=flat)](https://npmjs.com/package/@vuepress/plugin-html-redirect) ![Node.js CI](https://github.com/vuepressjs/vuepress-plugin-html-redirect/workflows/Node.js%20CI/badge.svg)

## Feature

- Support virtual URLs as source URLs.
- Support `countdown`.
- Work with static `base` and [dynamic base](https://github.com/vuepressjs/vuepress-plugin-dynamic-base).

## Motivation

In the site development of vuepress, a small directory structure adjustment will invalidate some URLs, but these URLs may have been published. With this plugin, you can keep those disappeared URLs forever.

## Install

```bash
yarn add -D @vuepress/plugin-html-redirect
# OR npm install -D @vuepress/plugin-html-redirect
```

## Usage

- Write redirects:

The agreed file to write `redirects` config is `/path/to/.vuepress/redirects`, whose format is as follows:

```
[url] [redirect_url]
[url] [redirect_url]
[url] [redirect_url]
...
```

example:

```
/2020/03/27/webpack-5-module-federation/ /translations/2020/03/27/webpack-5-module-federation/
``` 

- Simple usage:

```js
// .vuepress/config.js
module.exports = {
  plugins: [
    '@vuepress/html-redirect' // OR full name: '@vuepress/plugin-html-redirect'
  ]
}
```

- Disable `countdown`:


```js
// .vuepress/config.js
module.exports = {
  plugins: [
    ['@vuepress/html-redirect', {
        duration: 0
    }]
  ]
}
```

It means that the publc path will be different acccording to the NEV you set, and the router base will be `'/'` when the host is `hostA`, and `'/blog/'` when the host is `hostB`.

## Options

### duration

- Type: `string`
- Description: Control how many seconds the page will be redirected.

## TODO

- Support directory redirects.

PR welcome!

## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :D

## Author

**@vuepress/plugin-html-redirect** © [ULIVZ](https://github.com/ulivz), Released under the [MIT](./LICENSE) License.<br>

> [github.com/ulivz](https://github.com/ulivz) · Twitter [@_ulivz](https://twitter.com/_ulivz)


