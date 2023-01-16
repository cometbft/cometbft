# Cosmos theme for VuePress

[![npm version](https://img.shields.io/npm/v/vuepress-theme-cosmos)](https://www.npmjs.com/package/vuepress-theme-cosmos)

## Install

```sh
# Remove previously installed version (optional)
rm -rf node_modules

# If there is no package.json file, initialize npm package
npm init

# Install or update the theme
npm install --save vuepress-theme-cosmos
```

## Usage

Minimal config in `.vuepress/config.js` to enable the theme:

```js
module.exports = {
  theme: "cosmos",
};
```

Modify `scripts` property in `package.json`:

```js
  "scripts": {
    "serve": "vuepress dev",
    "build": "vuepress build"
  },
```

### Run dev server

```sh
npm run serve
```

### Build the website

```
npm run build
```

## Configuration

Most of the configuration happens in the `.vuepress/config.js` file. All parameters all optional, except `theme`.

```js
module.exports = {
  // Enable the theme
  theme: "cosmos",
  // Configure default title
  title: "Default title",
  themeConfig: {
    // Logo in the top left corner, file in .vuepress/public/
    logo: "/logo.svg",
    // Configure the manual sidebar
    header: {
      img: {
        // Image in ./vuepress/public/logo.svg
        src: "/logo.svg",
        // Image width relative to the sidebar
        width: "75%",
      },
      title: "Documentation",
    },
    // algolia docsearch
    // https://docsearch.algolia.com/
    algolia: {
      id: "BH4D9OD16A",
      key: "ac317234e6a42074175369b2f42e9754",
      index: "cosmos-sdk"
    },
    // custom must be false, topbar.banner is true to enable
    topbar: {
      banner: false
    },
    sidebar: {
      // Auto-sidebar, true by default
      auto: false,
      children: [
        // Array of sections
        {
          title: "Section title",
          children: [
            {
              title: "External link",
              path: "https://example.org/",
            },
            {
              title: "Internal link",
              path: "/url/path/",
            },
            {
              title: "Directory",
              path: "/path/to/directory/",
              directory: true,
            },
            {
              title: "Link to ./vuepress/public/foo/index.html",
              path: "/foo/",
              static: true,
            },
          ],
        },
        // Configure Resources
        {
          title: "Resources",
          children: [
            {
              title: "Default resource 1",
              path: "https://github.com/cosmos/vuepress-theme-cosmos",
            },
            {
              title: "Default resource 2",
              path: "https://github.com/cosmos/vuepress-theme-cosmos",
            },
          ],
        },
      ],
    },
  },
};
```

### Header

`themeConfig.header` property is responsible for the sidebar header component.

If `header` is `undefined`, then a default image (hexagon, width 40px) is used along with a title "Documentation".

If `header` is a string, `header` is used as a path to the logo. For example, `"/logo.svg"` uses `.vuepress/public/logo.svg` in user's directory. Title string is hidden.

If `header` is an object and has a `logo` property. If `logo` is a string, it is used as a path to the logo with the width of 50% and title string is hidden unless `header.title` is defined. If `logo` is an object and has `src` property, `logo.src` is used as a path string with a width of 50% unless `logo.width` is defined.

Title string has a value of `header.title` if it is defined. If it is undefined and `header.logo` is defined, the value is "Documentation".

## File configuration

Markdown files can contain YAML frontmatter. Several properties (all of which are optional) are used by the theme:

```yaml
---
# title is displayed in the sidebar
title: Title of the file
# order specifies file's priority in the sidebar
order: 2
# parent is readme.md or index.md parent directory
parent:
  title: Directory title
  order: 1
---

```

Setting `order: false` removes the item (file or directory) from the sidebar. It is, however, remains accessible by means other than the sidebar. It is valid use a `readme.md` to set an order of a parent-directory and hide the file with `order: false`.

## Docs search

We're currently using [Algolia Docsearch](https://github.com/cosmos/vuepress-theme-cosmos/pull/48) to improve the search experience. You're required to [join the program](https://docsearch.algolia.com) to use Algolia Docssearch. Once you have acquired all the necessary Algolia config keys, you can modify the `$themeConfig.algolia` in the `config.js` as such:

```js
algolia: {
  id: "BH4D9OD16A",
  key: "ac317234e6a42074175369b2f42e9754",
  index: "cosmos-sdk"
},
```

## Syntax highlighter

`vuepress-theme-cosmos` uses [Prism](https://prismjs.com/) to highlight language syntax in Markdown code blocks. Modify the manually imported files in `TmCodeBlock.vue` to [support different languages](https://prismjs.com/#supported-languages).

## Versioning

### Prerequisites

0. Make sure the versioned branches exist.
1. `.vuepress/config.js` - to edit versions shown on version switcher
2. `docs/versions` - to loop through versions and generate versioned docs
3. `Makefile`/ `GitHub Actions` - to build versioned docs and docs with path prefixes.
    - You can use [Makefile](https://github.com/cosmos/cosmos-sdk/blob/master/Makefile#L195) directly or
    - [GitHub Actions](https://github.com/cosmos/cosmos-sdk/blob/master/.github/workflows/docs.yml#L24) to run `make build-docs` depending on your setup.

#### Steps
1. add versions to `.vuepress/config.js`

```
versions: [
  {
    "label": "v0.39",
    "key": "v0.39"
  },
  {
    "label": "master",
    "key": "master"
  }
],
```

2. edit `docs/versions` for Makefile to loop through and generate versioned docs.

```
master              master
launchpad/backports v0.39
```

> 🔔 There must be a newline at the end of `versions` file otherwise it will prevent `make build-docs` script from reading both lines (versions) of the file. cc: https://github.com/cosmos/cosmos-sdk/pull/7938

3. `Makefile`

https://github.com/cosmos/cosmos-sdk/blob/master/Makefile#L195 to build versioned docs and docs with path prefixes in the deployment.

## Used by

1. [Cosmos SDK Documentation](https://docs.cosmos.network) — [`github`](https://github.com/cosmos/cosmos-sdk) — [`.vuepress/config.js`](https://github.com/cosmos/cosmos-sdk/blob/master/docs/.vuepress/config.js)
2. [Cosmos SDK Tutorials](https://tutorials.cosmos.network) — [`github`](https://github.com/cosmos/sdk-tutorials) — [`.vuepress/config.js`](https://github.com/cosmos/sdk-tutorials/blob/master/.vuepress/config.js)
3. [Cosmos Hub](https://hub.cosmos.network) — [`github`](https://github.com/cosmos/gaia/tree/master/docs) — [`.vuepress/config.js`](https://github.com/cosmos/gaia/blob/master/docs/.vuepress/config.js)
4. [Tendermint Core Documentation](https://docs.tendermint.com) — [`github`](https://github.com/tendermint/tendermint/tree/master/docs) — [`.vuepress/config.js`](https://github.com/tendermint/tendermint/blob/master/docs/.vuepress/config.js)
5. [Kava Documentation](https://docs.kava.io) — [`github`](https://github.com/Kava-Labs/kava/tree/master/docs) — [`.vuepress/config.js`](https://github.com/Kava-Labs/kava/blob/master/docs/.vuepress/config.js)
6. [Ethermint Documentation](https://docs.ethermint.zone) — [`github`](https://github.com/ChainSafe/ethermint/tree/development/docs) — [`.vuepress/config.js`](https://github.com/ChainSafe/ethermint/blob/development/docs/.vuepress/config.js)
7. [Cosmwasm Documentation](https://docs.cosmwasm.com) — [`github`](https://github.com/CosmWasm/docs2) — [`.vuepress/config.js`](https://github.com/CosmWasm/docs2/blob/master/.vuepress/config.js)
8. [IBC-Go Documentation](https://ibc.cosmos.network) — [`github`](https://github.com/cosmos/ibc-go/tree/main/docs) — [`.vuepress/config.js`](https://github.com/cosmos/ibc-go/blob/main/docs/.vuepress/config.js)

## Contributing

```md
<!-- after cloning vuepress-theme-cosmos -->
$ git clone https://github.com/cosmos/vuepress-theme-cosmos.git

<!-- example: project using vuepress-cosmos-theme -->
$ git clone https://github.com/cosmos/cosmos-sdk.git
$ cd cosmos-sdk
$ cd docs
$ npm i
$ npm link vuepress-theme-cosmos
$ npm run serve
```

## License

vuepress-theme-cosmos is licensed under [Apache 2.0](./LICENSE).
