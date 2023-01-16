<template lang="pug">
  div(style="width: 100%")
    .search__container
      .search(@click="$emit('search', true)")
        .search__icon
          icon-search
        .search__text Search
    .container
      slot
      tm-content-cards(v-if="$frontmatter.cards")
</template>

<style lang="stylus" scoped>
.search
  display flex
  align-items center
  color rgba(22, 25, 49, 0.65)
  padding-top 1rem
  width calc(var(--aside-width) - 6rem)
  cursor pointer
  position absolute
  top 1rem
  right 4rem
  justify-content flex-end
  transition color .15s ease-out

  &:hover
    color var(--color-text, black)

  &__container
    visibility hidden
    display flex
    justify-content flex-end
    margin-top 1rem
    margin-bottom 1rem

  &__icon
    width 1.5rem
    height 1.5rem
    fill #aaa
    margin-right 0.5rem
    transition fill .15s ease-out

  &:hover &__icon
    fill var(--color-text, black)

.footer__links
  padding-top 5rem
  padding-bottom 1rem
  border-top 1px solid rgba(176, 180, 207, 0.2)
  margin-top 5rem

.links
  display flex
  justify-content space-between
  margin-top 4rem

  a
    box-shadow none
    color var(--color-link, blue)

.container
  position relative
  width 100%
  max-width 45rem

.content
  padding-right var(--sidebar-width)
  width 100%
  position relative

  &.noAside
    padding-right 0

  &__container
    width 100%
    padding-left 4rem
    padding-right 2rem

    &.noAside
      max-width initial

/deep/
  .codeblock
    margin-top 2rem
    margin-bottom 2rem
    letter-spacing 0

  .custom-block
    &.danger
      margin-top 1.5rem
      margin-bottom 1.5rem

    &.danger, &.warning, &.tip
      padding 1rem 1.5rem 1rem 3.5rem
      border-radius 0.5rem
      position relative

      & :first-child
        margin-top 0

      & :last-child
        margin-bottom 0

      &:before
        content ''
        height 24px
        width 24px
        position absolute
        display block
        top 1rem
        left 1rem
        background-repeat no-repeat

    &.danger
      background #FFF6F9

      &:before
        background-image url("./images/icon-danger.svg")

    &.warning
      &:before
        background-image url("./images/icon-warning.svg")

    &.tip
      &:before
        background-image url("./images/icon-tip.svg")

  h2, h3, h4, h5, h6
    &:hover
      a.header-anchor
        opacity 1

  a.header-anchor
    opacity 0
    position absolute
    font-weight 400
    left -1.5em
    width 1.5em
    text-align center
    box-sizing border-box
    color rgba(0, 0, 0, 0.4)
    outline-color var(--color-link, blue)
    transition all 0.25s

    &:after
      transition all 0.25s
      border-radius 0.25rem
      content attr(data-header-anchor-text)
      max-width 4rem
      color white
      position absolute
      top -2.4em
      padding 7px 12px
      white-space nowrap
      left 50%
      transform translateX(-50%)
      font-size 0.8125rem
      line-height 1
      letter-spacing 0
      opacity 0
      box-shadow 0px 16px 32px rgba(22, 25, 49, 0.08), 0px 8px 12px rgba(22, 25, 49, 0.06), 0px 1px 0px rgba(22, 25, 49, 0.05)
      background var(--color-text, black)

    &:before
      transition all 0.25s
      content ''
      background-image url("data:image/svg+xml,  <svg xmlns='http://www.w3.org/2000/svg' width='100%' height='100%' viewBox='0 0 24 24'><path fill='rgb(22, 25, 49)' d='M12 21l-12-18h24z'/></svg>")
      position absolute
      width 8px
      height 8px
      top -0.7em
      left 50%
      font-size 0.5rem
      transform translateX(-50%)
      opacity 0

    &:focus,
    &:focus:before,
    &:hover:before
      opacity 1

    &:focus:after,
    &:hover:after
      opacity 1

  h1[id*='requisite'], h2[id*='requisite'], h3[id*='requisite'], h4[id*='requisite'], h5[id*='requisite'], h6[id*='requisite']
    display none
    align-items baseline
    cursor pointer

    &:before
      content ''
      width 1.5rem
      height 1.5rem
      display block
      flex none
      margin-right 0.5rem
      background url('./images/icon-chevron.svg')
      transition transform 0.2s ease-out

  h1[id*='requisite'].prereqTitleShow, h2[id*='requisite'].prereqTitleShow, h3[id*='requisite'].prereqTitleShow, h4[id*='requisite'].prereqTitleShow, h5[id*='requisite'].prereqTitleShow, h6[id*='requisite'].prereqTitleShow
    &:before
      transform rotate(90deg)

  h1[id*='requisite'] + ul, h2[id*='requisite'] + ul, h3[id*='requisite'] + ul, h4[id*='requisite'] + ul, h5[id*='requisite'] + ul, h6[id*='requisite'] + ul
    display none

  li[prereq]
    display none
    max-width 28rem
    margin-left 2rem

  li[prereq].prereqLinkShow
    display block

  li[prereq] a[href]
    box-shadow 0px 2px 4px rgba(22, 25, 49, 0.05), 0px 0px 1px rgba(22, 25, 49, 0.2), 0px 0.5px 0px rgba(22, 25, 49, 0.05)
    padding 1rem
    border-radius 0.5rem
    color var(--color-text, black)
    font-size 1rem
    font-weight 600
    line-height 1.25rem
    margin 1rem 0
    display block
    transition box-shadow 0.25s ease-out, transform 0.25s ease-out, opacity 0.4s ease-out

    &:hover:not(:active)
      color inherit
      text-decoration none
      box-shadow 0px 10px 20px rgba(0, 0, 0, 0.05), 0px 2px 6px rgba(0, 0, 0, 0.05), 0px 1px 0px rgba(0, 0, 0, 0.05)
      transform translateY(-2px)
      transition-duration 0.1s

    &:active
      opacity 0.7
      transition-duration 0s

  [synopsis]
    padding 1.5rem 2rem
    background-color rgba(176, 180, 207, 0.09)
    border-radius 0.5rem
    margin-top 3rem
    margin-bottom 3rem
    color rgba(22, 25, 49, 0.9)
    font-size 1rem
    line-height 1.625rem

    &:before
      content 'Synopsis'
      display block
      color rgba(22, 25, 49, 0.65)
      text-transform uppercase
      font-size 0.75rem
      margin-bottom 0.5rem
      letter-spacing 0.2em

  a[target='_blank']
    &:after
      content '↗'
      position absolute
      bottom 0.166em
      padding-left 0.1875em
      font-size 0.75em
      line-height 1
      word-break none
      transition transform 0.2s ease-out
    &:hover:after,
    &:focus:after
      transform translate(2px, -2px)

  .icon.outbound
    display none

  table
    display block
    width 100% // fallback
    width max-content
    max-width 100%
    overflow auto
    line-height 1.5rem
    margin-top 2rem
    margin-bottom 2rem
    box-shadow 0 0 0 1px rgba(140, 145, 177, 0.32)
    border-radius 0.5rem
    border-collapse collapse
    font-size 1rem

  th
    text-align left
    font-weight 700
    font-size 0.875rem

  td, th
    padding 0.75rem

  tr
    box-shadow 0 1px 0 0 rgba(140, 145, 177, 0.32)

  tr:only-child
    box-shadow none

  thead tr:only-child
    box-shadow 0 1px 0 0 rgba(140, 145, 177, 0.32)

  tr + tr:last-child
    box-shadow none

  tr:last-child td
    border-bottom none

  .code-block__container
    margin-top 2rem
    margin-bottom 2rem

  .content__default
    width 100%

  h1, h2, h3, h4
    font-weight 700

  h1 code, h2 code, h3 code
    font-weight normal

  .content__container
    img
      max-width 100%

  .term
    text-decoration underline

  img
    width 100%
    height auto
    display block
    margin-bottom 2rem
    margin-top 2rem

  .tooltip

    // temporary fixes for tooltips coming from cosmos-ui
    &__wrapper
      background white
      padding 1rem

    h1
      font-size 0.875rem
      line-height 1.25rem
      letter-spacing .01em
      font-weight 600
      margin-top 0
      margin-bottom 0

    p
      font-size 0.8125rem
      line-height 1.125rem
      margin-top 0.375rem
      margin-bottom 0

  hr
    border-width 1px
    border-style solid
    border-color rgba(0,0,0,0.1)
    margin-top 2.5rem
    margin-bottom 2.5rem

  strong
    font-weight 600
    letter-spacing .01em

  em
    font-style italic

  h1
    font-size 3rem
    margin-top 4rem
    margin-bottom 4rem
    line-height 3.5rem
    letter-spacing -0.02em

    &:first-child
      margin-top 0

  h2
    font-size 2rem
    margin-top 3.75rem
    margin-bottom 1.25rem
    line-height 2.5rem
    letter-spacing -0.01em

  h3
    font-size 1.5rem
    margin-top 2.5rem
    margin-bottom 1rem
    letter-spacing 0
    line-height 2rem

  h4
    font-size 1.25rem
    margin-top 2.25rem
    margin-bottom 0.875rem
    line-height 1.75rem
    letter-spacing .01em

  p,ul,ol
    font-size 1.125rem
    line-height 1.8125rem

  p
    margin-top 1em
    margin-bottom 1em

  ul, ol
    margin-top 1em
    margin-bottom 1.5em
    margin-left 0
    padding-left 0

  li
    padding-left 0
    margin-left 2rem
    margin-bottom 1rem
    position relative

  blockquote
    padding-left 2rem
    padding-right 2rem
    border-left 0.25rem solid rgba(0,0,0,0.1)
    color var(--color-text-dim, inherit)
    margin-top 1.75rem
    margin-bottom 1.75rem

  code
    background-color rgba(176, 180, 207, 0.175)
    border 1px solid rgba(176, 180, 207, 0.09)
    border-radius 0.25em
    padding-left 0.25em
    padding-right 0.25em
    font-size 0.8333em
    line-height 1.06666em
    letter-spacing 0
    color var(--color-code, inherit)
    margin-top 3rem
    overflow-wrap break-word
    word-wrap break-word
    -ms-word-break break-all
    word-break break-word

  h1, h2, h3, h4, h5, h6
    code
      font-size inherit

  h1, h2, h3, h4, h5, h6
    a
      color var(--color-link, blue)
      outline none
      position relative

    a[target='_blank']
      &:after
        position relative

  p, ul, ol
    a
      color var(--color-link, blue)
      outline-color var(--color-link, blue)
      border-radius 0.25rem
      position relative
      transition opacity 0.3s ease-out
      overflow-wrap break-word
      word-wrap break-word
      -ms-word-break break-all
      word-break break-word

    a[target='_blank']
      margin-right 0.888em

    a:hover
      text-decoration underline

    a:active
      opacity 0.65
      transition-duration 0s

    a code
      color inherit
      transition background-color 0.15s ease-out

    a:hover code,
    a:focus code
      background-color rgba(59, 66, 125, 0.12)

  td
    a
      color var(--color-link, blue)
      position relative
      transition opacity 0.3s ease-out
      overflow-wrap break-word
      word-wrap break-word
      -ms-word-break break-all
      word-break inherit
    a[target='_blank']
      &:after
        display none

@media screen and (max-width: 1136px)
  >>> h2, >>> h3, >>> h4, >>> h5, >>> h6
    padding-right 1.75rem

  >>> a.header-anchor
    left initial
    right 0
    text-align right
    opacity 1

    &:after
      transform none
      left initial
      right -5px

  >>> h1 a.header-anchor
    display none

@media screen and (max-width: 1024px)
  .content
    padding-right 0

    &__container
      padding-left 2rem

@media screen and (max-width: 1136px) and (min-width: 833px)
  .search__container
    visibility visible

@media screen and (max-width: 1136px)
  >>> h1[id*='requisite'], >>> h2[id*='requisite'], >>> h3[id*='requisite'], >>> h4[id*='requisite'], >>> h5[id*='requisite'], >>> h6[id*='requisite']
    display flex

  >>> h1[id*='requisite'] + ul, >>> h2[id*='requisite'] + ul, >>> h3[id*='requisite'] + ul, >>> h4[id*='requisite'] + ul, >>> h5[id*='requisite'] + ul, >>> h6[id*='requisite'] + ul
    display block

@media screen and (max-width: 480px)
  >>> h1
    font-size 2.5rem
    margin-bottom 3rem
    line-height 3rem

  >>> h2
    font-size 1.75rem
    margin-top 3.5rem
    margin-bottom 1rem
    line-height 2.25rem

  >>> h3
    font-size 1.25rem
    margin-top 2.25rem
    margin-bottom 0.875rem
    line-height 1.75rem

  >>> h4
    font-size 1.125rem
    margin-top 2rem
    margin-bottom 0.75rem
    line-height 1.5rem

  >>> p, >>> ul, >>> ol
    font-size 1rem
    line-height 1.625rem

  >>> [synopsis]
    padding 1rem
    font-size 0.875rem
    line-height 1.25rem

</style>

<script>
import { findIndex, sortBy } from "lodash";
import copy from "clipboard-copy";

export default {
  props: {
    aside: {
      type: Boolean,
      default: true
    },
    tree: {
      type: Array
    }
  },
  mounted() {
    this.emitPrereqLinks();
    const headerAnchorClick = event => {
      event.target.setAttribute("data-header-anchor-text", "Copied!");
      copy(event.target.href);
      setTimeout(() => {
        event.target.setAttribute("data-header-anchor-text", "Copy link");
      }, 4000);
      event.preventDefault();
    };
    document
      .querySelectorAll(
        'h1[id*="requisite"], h2[id*="requisite"], h3[id*="requisite"], h4[id*="requisite"], h5[id*="requisite"], h6[id*="requisite"]'
      )
      .forEach(node => {
        node.addEventListener("click", this.prereqToggle);
      });
    document
      .querySelectorAll(".content__default a.header-anchor")
      .forEach(node => {
        node.setAttribute("data-header-anchor-text", "Copy link");
        node.addEventListener("click", headerAnchorClick);
      });
    if (window.location.hash) {
      const elementId = document.querySelector(window.location.hash);
      if (elementId) elementId.scrollIntoView();
    }
  },
  methods: {
    emitPrereqLinks() {
      const prereq = [...document.querySelectorAll("[prereq]")].map(item => {
        const link = item.querySelector("[href]");
        return {
          href: link.getAttribute("href"),
          text: link.innerText
        };
      });
      this.$emit("prereq", prereq);
    },
    prereqToggle(e) {
      if (e.target.classList.contains('header-anchor')) return
      e.target.classList.toggle("prereqTitleShow");
      document.querySelectorAll("[prereq]").forEach(node => {
        node.classList.toggle("prereqLinkShow");
      });
    }
  },
  computed: {
    noAside() {
      return !this.aside;
    },
    linkPrevNext() {
      if (!this.tree) return;
      let result = {};
      const search = tree => {
        return tree.forEach(item => {
          const children = item.children;
          if (children) {
            const index = findIndex(children, ["regularPath", this.$page.path]);
            if (index >= 0 && children[index - 1]) {
              result.prev = children[index - 1];
            }
            if (index >= 0 && children[index + 1]) {
              result.next = children[index + 1];
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
