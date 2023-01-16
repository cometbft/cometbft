<template lang="pug">
  div
    .links
      .links__wrapper
        .links__container(v-if="$page.frontmatter.prev || (linkPrevNext && linkPrevNext.prev && linkPrevNext.prev.frontmatter && linkPrevNext.prev.frontmatter.order !== false)")
          router-link.links__item.links__item__left(:to="$page.frontmatter.prev || linkPrevNext.prev.regularPath")
            .links__item__icon
              svg(xmlns="http://www.w3.org/2000/svg" width="100%" height="100%" viewBox="0 0 64 64")
                title arrow-right
                g(stroke-linecap="square" stroke-linejoin="miter" stroke-width="2")
                  line(fill="none" stroke-miterlimit="10" x1="61" y1="32" x2="3" y2="32" stroke-linecap="butt")
                  polyline(fill="none" stroke-miterlimit="10" points="21,14 3,32 21,50 ")
            div
              .links__label Previous
              .links__item__title {{$page.frontmatter.prev || linkPrevNext.prev.title}}
              .links__item__desc(v-if="linkPrevNext.prev.frontmatter.description" v-html="shorten(linkPrevNext.prev.frontmatter.description)")
      .links__wrapper
        .links__container(v-if="$page.frontmatter.next || (linkPrevNext && linkPrevNext.next && linkPrevNext.next.frontmatter && linkPrevNext.next.frontmatter.order !== false)")
          router-link.links__item.links__item__right(:to="$page.frontmatter.next || linkPrevNext.next.regularPath")
            div
              .links__label Next
              .links__item__title {{$page.frontmatter.next || linkPrevNext.next.title}}
              .links__item__desc(v-if="linkPrevNext.next.frontmatter.description" v-html="shorten(linkPrevNext.next.frontmatter.description)")
            .links__item__icon
              svg(xmlns="http://www.w3.org/2000/svg" width="100%" height="100%" viewBox="0 0 64 64")
                title arrow-right
                g(stroke-linecap="square" stroke-linejoin="miter" stroke-width="2")
                  line(fill="none" stroke-miterlimit="10" x1="3" y1="32" x2="61" y2="32" stroke-linecap="butt")
                  polyline(fill="none" stroke-miterlimit="10" points="43,14 61,32 43,50 ")
</template>

<style lang="stylus" scoped>
.links
  display flex

  &__wrapper
    display flex
    width 100%
    margin-bottom 2rem

    &:first-child
      margin-right 2rem

  &__container
    width 100%
    align-items stretch
    display flex
    flex-direction column

  &__item
    margin-top 1rem
    padding 2rem
    box-shadow 0px 2px 4px rgba(22, 25, 49, 0.05), 0px 0px 1px rgba(22, 25, 49, 0.2), 0px 0.5px 0px rgba(22, 25, 49, 0.05)
    border-radius 0.5rem
    display grid
    grid-auto-flow column
    flex-grow 1
    align-items center
    gap 2rem
    overflow-x hidden
    transition box-shadow 0.25s ease-out, transform 0.25s ease-out, opacity 0.4s ease-out

    &:hover:not(:active)
      box-shadow 0px 12px 24px rgba(22, 25, 49, 0.07), 0px 4px 8px rgba(22, 25, 49, 0.05), 0px 1px 0px rgba(22, 25, 49, 0.05)
      transform translateY(-2px)
      transition-duration 0.1s

    &:active
      opacity 0.7
      transition-duration 0s

    &__left
      grid-template-columns 2.75rem auto

    &__right
      grid-template-columns auto 2.75rem

    &__icon
      display flex
      align-items center

      svg
        stroke #aaa
        transition fill .15s ease-out, transform .15s ease-out

    &:hover &__icon,
    &:focus &__icon
      svg
        stroke var(--color-link, #888)

    &__left:hover &__icon svg
      transform translateX(-0.25rem)

    &__right:hover &__icon svg
      transform translateX(0.25rem)

    &__title
      margin-top 5px
      font-weight 600
      font-size 1.25rem
      line-height 1.75rem

    &__desc
      color var(--color-text-dim, inherit)
      margin-top 0.5rem
      font-size 0.875rem
      line-height 1.25rem

  &__label
    color var(--color-text-dim, inherit)
    text-transform uppercase
    font-size 0.75rem
    line-height 1rem
    letter-spacing 0.2rem
    margin-bottom 0.5rem

@media screen and (max-width: 1280px)
  .links
    flex-direction column-reverse

@media screen and (max-width: 480px)
  .links__item
    display flex
    flex-direction column
    align-items stretch

    &__icon
      order -1
      width 2.5rem
      height 2.5rem
      align-self flex-end


</style>

<script>
import { findIndex, find } from "lodash";

export default {
  props: ["tree"],
  methods: {
    shorten(string) {
      let str = string.split(" ");
      str =
        str.length > 20 ? str.slice(0, 20).join(" ") + "..." : str.join(" ");
      return this.md(str);
    }
  },
  computed: {
    linkPrevNext() {
      if (!this.tree) return;
      let result = {};
      const search = tree => {
        return tree.forEach((item, i) => {
          const children = item.children;
          if (children) {
            const index = findIndex(children, ["regularPath", this.$page.path]);
            if (index >= 0 && children[index - 1]) {
              result.prev = children[index - 1];
            }
            if (index >= 0 && children[index + 1]) {
              result.next = children[index + 1];
            } else if (index >= 0 && tree[i + 1] && tree[i + 1].children) {
              result.next = find(tree[i + 1].children, x => {
                return x.frontmatter && x.frontmatter.order !== false;
              });
            }
            return search(item.children);
          }
        });
      };
      search(this.tree);
      return result;
    }
  }
};
</script>
