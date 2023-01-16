# pug-plain-loader

A loader that simply compiles pug templates into HTML.

## Installation

Note `pug` is a peer dependency, so make sure to install both:

``` sh
npm install -D pug-plain-loader pug
```

## Usage

This loader is mostly intended to be used alongside `vue-loader` v15+, since it now requires using webpack loaders to handle template preprocessors. There's also [`pug-html-loader`](https://github.com/willyelm/pug-html-loader) which unfortunately is out-of-date and not actively maintained.

If you are only using this loader for templating in single-file Vue components, simply configure it with:

``` js
{
  module: {
    rules: [
      {
        test: /\.pug$/,
        loader: 'pug-plain-loader'
      }
    ]
  }
}
```

This will apply this loader to all `<template lang="pug">` blocks in your Vue components.

If you also intend to use it to import `.pug` files as HTML strings in JavaScript, you will need to chain `raw-loader` after this loader. Note however adding `raw-loader` would break the output for Vue components, so you need to have two rules, one of them excluding Vue components:

``` js
{
  module: {
    rules: [
      {
        test: /\.pug$/,
        oneOf: [
          // this applies to pug imports inside JavaScript
          {
            exclude: /\.vue$/,
            use: ['raw-loader', 'pug-plain-loader']
          },
          // this applies to <template lang="pug"> in Vue components
          {
            use: ['pug-plain-loader']
          }
        ]
      }
    ]
  }
}
```

## Options

See [Pug compiler options](https://pugjs.org/api/reference.html#options).

The `doctype` option is set to `html` by default, since most Vue templates are HTML fragments without explicit doctype.

An additional option `data` can be used to pass locals for the template, although this is typically not recommended when using in Vue components.
