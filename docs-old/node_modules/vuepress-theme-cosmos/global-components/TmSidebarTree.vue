<template lang="pug">
  div
    div(v-for="item in value")
      // todo: add event.preventDefault() somehow
      component(
        v-if="!hide(item)"
        :style="{'--vline': level < 1 ? 'hidden' : 'visible', '--vline-color': (iconActive(item) || iconExpanded(item)) && !iconExpanded(item) ? 'var(--color-primary)' : 'rgba(0,0,0,0.1)' }"
        :is="componentName(item)"
        :to="item.path"
        :target="outboundLink(item.path) && '_blank'"
        :rel="outboundLink(item.path) && 'noreferrer noopener'"
        :href="(outboundLink(item.path) || item.static) ? item.path : '#'"
        :class="[level > 0 && 'item__child',{'item__dir': !item.path}]"
        tag="a"
        :role="!item.path && 'button'"
        @keydown.enter="handleEnter(item)"
        @click="!outboundLink(item.path) && revealChild(item.title)"
      ).item
        tm-icon-hex(v-if="iconExpanded(item) && level < 1" style="--icon-color: var(--color-primary, blue)").item__icon.item__icon__expanded
        tm-icon-hex(v-if="iconCollapsed(item) && level < 1" style="--icon-color: rgba(0,0,0,0.18)").item__icon.item__icon__collapsed
        tm-icon-hex(v-else-if="!outboundLink(item.path) && level < 1 && !iconExpanded(item)").item__icon.item__icon__internal
        tm-icon-outbound(v-else-if="outboundLink(item.path) || item.static").item__icon.item__icon__outbound
        div(:style="{'padding-left': `${1*level}rem`}" :class="{'item__selected': iconActive(item) || iconExpanded(item), 'item__selected__dir': iconCollapsed(item), 'item__selected__alt': iconExpanded(item)}" v-html="titleFormatted(titleText(item))")
      div(v-if="item.children || directoryChildren(item) || []")
        transition(name="reveal" v-on:enter="setHeight" v-on:leave="setHeight")
          tm-sidebar-tree(:level="level+1" :value="item.children || directoryChildren(item) || []" v-show="item.title == show" v-if="!hide(item)" :title="item.title" @active="revealParent($event)")
</template>

<style lang="stylus" scoped>
.item
  position relative
  padding-left 1.5rem
  display block
  padding-top .375rem
  padding-bottom .375rem
  cursor pointer
  font-size .875rem
  line-height 1.25rem
  outline-color var(--color-primary, blue)
  color rgba(0,0,0,0.8)
  transition color .15s ease-out

  &:hover,
  &:hover &__icon
    transition-duration 0s

  &:hover, &:focus
    color var(--color-text, black)

  &__child:not(.router-link-active)
    &:hover:after,
    &:focus:after
      background rgba(0,0,0,0.2)

  &:hover, &:focus
    .item__icon.item__icon__outbound,
    .item__icon.item__icon__collapsed
      fill var(--color-primary, blue)
      stroke none

  &:hover
    .item__icon.item__icon__internal
      visibility unset
      stroke var(--color-primary, blue)

  &:focus
    .item__icon.item__icon__internal
      stroke transparent
      visibility unset
      fill var(--color-primary, blue)

  &:after
    content ''
    width 2px
    height 100%
    visibility var(--vline)
    background var(--vline-color, black)
    position absolute
    top 0
    left 5px
    transition background-color .15s ease-out

  &__selected
    font-weight 600
    color var(--color-link, blue)

    &__dir
      font-weight 400

    &__alt
      color initial
      font-weight 600

  &__dir
    font-weight 600

  &__icon
    position absolute
    top .65rem
    left 0
    width 12px
    height 12px
    fill var(--icon-color)
    transition fill .15s ease-out, height .15s ease-out

    &__expanded
      stroke none
      fill none
      background var(--color-primary, blue)
      height 1px
      padding-top 1px
      margin-top 4px

    &__internal
      stroke rgba(0,0,0,0.2)
      stroke-width 1.5px
      fill none

    &__outbound
      fill rgba(0,0,0,0.2)

.reveal-enter-active, .reveal-leave-active
  transition all 0.25s
  overflow hidden

.reveal-enter, .reveal-leave-to
  max-height 0
  opacity 0

.reveal-enter-to, .reveal-leave
  max-height var(--max-height)
  opacity 1
</style>

<script>
import { sortBy, find } from "lodash";
import MarkdownIt from "markdown-it";

export default {
  name: "tm-sidebar-tree",
  props: ["value", "title", "tree", "level"],
  data: function() {
    return {
      show: null,
    };
  },
  mounted() {
    const active = find(this.value, ["key", this.$page.key]);
    if (active) {
      this.$emit("active", this.title);
    }
  },
  watch: {
    $route(to, from) {
      const found = find(this.value, ["key", to.name]);
      if (found) {
        this.revealParent(this.title);
      }
    },
  },
  methods: {
    hide(item) {
      const index = this.indexFile(item);
      const fileHide = item.frontmatter && item.frontmatter.order === false;
      const dirHide =
        index &&
        index.frontmatter &&
        index.frontmatter.parent &&
        index.frontmatter.parent.order === false;
      return dirHide || fileHide;
    },
    iconCollapsed(item) {
      if (item.directory && !this.iconExpanded(item)) return true;
      return (
        !item.path &&
        this.show != item.title &&
        (item.children || item.directory)
      );
    },
    iconExpanded(item) {
      return this.show == item.title && !item.key;
    },
    iconActive(item) {
      if (this.$page.path === "") return false; // Workaround for 404 page
      if (item.path == this.$page.path) return true;
      return item.key == this.$page.key;
    },
    outboundLink(path) {
      return /^[a-z]+:/i.test(path);
    },
    isInternalLink(item) {
      return (
        item.path &&
        !item.directory &&
        !item.static &&
        !this.outboundLink(item.path)
      );
    },
    isOutboundLink(item) {
      return (item.path && this.outboundLink(item.path)) || item.static;
    },
    handleEnter(item) {
      console.log("enter");
      this.revealChild(item.title);
    },
    componentName(item) {
      if (this.isInternalLink(item)) return "router-link";
      if (this.isOutboundLink(item)) return "a";
      return "a";
    },
    indexFile(item) {
      if (!item.children) return false;
      return find(item.children, (page) => {
        const path = page.relativePath;
        if (!path) return false;
        return (
          path.toLowerCase().match(/index.md$/i) ||
          path.toLowerCase().match(/readme.md$/i)
        );
      });
    },
    setHeight(el) {
      el.style.setProperty("--max-height", el.scrollHeight + "px");
    },
    titleFormatted(string) {
      const md = new MarkdownIt({ html: true, linkify: true });
      return `<div>${md.renderInline(string)}</div>`;
    },
    titleText(item) {
      const index = this.indexFile(item);
      if (item.frontmatter) {
        return item.frontmatter.title || item.title;
      }
      if (index) {
        if (
          index.frontmatter &&
          index.frontmatter.parent &&
          index.frontmatter.parent.title
        )
          return index.frontmatter.parent.title;
        if (index.title.match(/readme\.md/i) || index.title.match(/index\.md/i))
          return item.title;
        return index.title;
      }
      return item.title;
    },
    revealChild(title) {
      this.show = this.show == title ? null : title;
    },
    revealParent(title) {
      this.show = title;
      this.$emit("active", this.title);
    },
    directoryChildren(item) {
      if (item.directory === true) {
        let result = item.path && item.path.split("/").filter((i) => i != "");
        result = result.reduce((acc, cur) => {
          return find(acc.children || acc, ["title", cur]);
        }, this.tree);
        return result.children || [];
      }
      return [];
    },
  },
};
</script>
