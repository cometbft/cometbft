# v-runtime-template

[![npm](https://img.shields.io/npm/v/v-runtime-template.svg)](https://www.npmjs.com/package/v-runtime-template)

A Vue.js components that makes easy compiling and interpreting a Vue.js template at runtime by using a `v-html` like API.

> Do you know **[VueDose](https://vuedose.tips)**? It's where you can learn tips about the Vue.js ecosystem in a concise format, perfect for busy devs! 🦄

*[See Demo on CodeSandbox](https://codesandbox.io/s/884v9kq790)*

## Motivation

This library solves the case where you get a vue-syntax template string on runtime, usually from a server. Think of a feature where you allow the user to create their own interfaces and structures. You save that as a vue template in your database, which your UI will request later. While components are pre-compiled at build time, this case isn't (since the template is received at runtime) and needs to be compiled at runtime.

v-runtime-template compiles that template and attaches it to the scope of the component that uses it, so it has access to its data, props, methods and computed properties.

Think of it as the `v-html` equivalent that also understands vue template syntax (while `v-html` is just for plain HTML).

## Getting Started

Install it:

```
npm install v-runtime-template
```

You must **use the with-compiler Vue.js version**. This is needed in order to compile on-the-fly Vue.js templates. For that, you can set a webpack alias for `vue` to the `vue/dist/vue.common` file.

For example, if you use the [Vue CLI](https://github.com/vuejs/vue-cli), create or modify the `vue.config.js` file adding the following alias:

```js
// vue.config.js
module.exports = {
  runtimeCompiler: true
};
```

And in [Nuxt](http://nuxtjs.org/), open the `nuxt.config.js` file and extend the webpack config by adding the following line to the `extend` key:

```js
// nuxt.config.js
{
  build: {
    extend(config, { isDev, isClient }) {
      config.resolve.alias["vue"] = "vue/dist/vue.common";
      // ...
```

## Usage

You just need to import the `v-runtime-template` component, and pass the template you want:

```html
<template>
	<div>
		<v-runtime-template :template="template"></v-runtime-template>
	</div>
</template>

<script>
import VRuntimeTemplate from "v-runtime-template";
import AppMessage from "./AppMessage";

export default {
  data: () => ({
    name: "Mellow",
    template: `
      <app-message>Hello {{ name }}!</app-message>
    `
  }),
  components: {
    AppMessage,
    VRuntimeTemplate
  }
};
</script>
```

The template you pass **have access to the parent component instance**. For example, in the last example we're using the `AppMessage` component and accessing the `{{ name }}` state variable.

But you can access computed properties and methods as well from the template:

```js
export default {
  data: () => ({
    name: "Mellow",
    template: `
      <div>
        <app-message>Hello {{ name }}!</app-message>
        <button @click="sayHi">Say Hi!</button>
        <p>{{ someComputed }}</p>
      </div>
		`,
  }),
  computed: {
    someComputed() {
      return "Wow, I'm computed";
    },
  },
  methods: {
    sayHi() {
      console.log("Hi");
    },
  },
};
```

## Limitations

Keep in mind that the template can only access the instance properties of the component who is using it. Read [this issue](https://github.com/alexjoverm/v-runtime-template/issues/9) for more information.

## Comparison

### v-runtime-template VS v-html

_TL;DR: If you need to interpret only HTML, use `v-html`. Use this library otherwise._

They both have the same goal: to interpret and attach a piece of structure to a scope at runtime. The difference is, `[v-html](https://vuejs.org/v2/api/#v-html)` doesn't understand vue template syntax, but only HTML. So, while this code works:

```html
<template>
	<div v-html="template"></div>
</template>

<script>
export default {
  data: () => ({
    template: `
      <a href="/mike-page">Go to Mike page</a>
    `
```

the following wouldn't since it uses the custom `router-link` component:

```html
<router-link to="mike-page">Go to Mike page</router-link>
```

But you can use v-runtime-template, which uses basically the same API than v-html:

```html
<template>
	<v-runtime-template :template="template"></v-runtime-template>
</template>

<script>
export default {
  data: () => ({
    template: `
      <router-link to="mike-page">Go to Mike page</router-link>
    `
```

### v-runtime-template VS dynamic components (`<component>`)

Dynamic components have somewhat different goal: to render a component dynamically by binding it to the `is` prop. Although, these components are usually pre-compiled. However, the goal of v-runtime-template can be achieved just by using the component options object form of dynamic components.

In fact, v-runtime-template uses that under the hood (in the render function form) along with other common tasks to achieve its goal.
