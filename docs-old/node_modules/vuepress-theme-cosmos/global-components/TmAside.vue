<template lang="pug">
  div
    .container
      .search__container
        tm-select-version(v-if="$themeConfig.versions")
        .search(@click="$emit('search', true)" v-if="$themeConfig.algolia")
          .search__icon
            icon-search
          .search__text Search
      .banners(v-if="asideBanners && !$themeConfig.custom")
        .banners__item(v-for="banner in asideBanners")
          a(:href="banner.href" target="_blank" rel="noreferrer noopener")
            img(:src="`${asideBannersUrl}/${banner.src}`" :alt="banner.alt" @error="$emit('bannerError', true)").aside__image
      div(v-if="prereq && prereq.length > 0")
        .aside__title Pre-requisite reading
        a(v-for="item in prereq" :href="item.href").prereq__item {{item.text}}
      div(v-if="$page.headers && $page.headers.length > 0")
        .aside__title On this page
        .aside__link(v-for="link in headersFiltered" :class="[`aside__link__active__${headerCurrent && headerCurrent.slug === link.slug}`]" :ref="link.slug")
          a(:href="`#${link.slug}`" :class="{selected: link.slug == selected}").aside__link__href.header-anchor {{link.title}}
</template>

<style lang="stylus" scoped>
.container
  padding 1rem 1rem 1rem 2rem

.search__container
  display flex
  justify-content space-between
  padding-top 0.5rem
  padding-bottom 3.5rem

.banners
  margin-bottom 3rem

  &__item
    margin-bottom 0.5rem

    a
      display block
      transition transform 150ms ease-out, opacity 150ms ease-out, box-shadow 150ms ease-out
      outline-color var(--color-primary, blue)

      &:hover:not(:active), &:focus:not(:active)
        transform translateY(-2px)
        opacity 0.85
        box-shadow 0px 10px 20px rgba(0, 0, 0, 0.05), 0px 2px 6px rgba(0, 0, 0, 0.05), 0px 1px 0px rgba(0, 0, 0, 0.05)

      &:active
        transition none

.search
  cursor pointer
  display flex
  justify-content flex-end
  align-items center

  &__icon
    height 1.5rem
    width 1.5rem
    margin-right 0.5rem
    fill #aaa
    transition fill .15s ease-out

  &__text
    color rgba(22, 25, 49, 0.65)
    transition color .15s ease-out

  &:hover &__icon
    fill var(--color-text, black)

  &:hover &__text
    color var(--color-text, black)

.selected
  font-weight 700

.aside
  &__image
    width 100%
    border-radius 0.25rem
    display block

  &__title
    font-size 0.75rem
    text-transform uppercase
    letter-spacing 0.2em
    color var(--color-text-dim, inherit)
    margin-top 3rem
    margin-bottom 0.75rem

  &__link
    color var(--color-text-dim, inherit)
    padding-top 0.375rem
    padding-bottom 0.375rem
    font-size 0.875rem
    line-height 1.125rem
    word-break break-word

    &__href:hover
      color var(--color-text, black)

    &__active__true
      color var(--color-text, black)
      font-weight bold

.prereq__item
  box-shadow 0px 2px 4px rgba(22, 25, 49, 0.05), 0px 0px 1px rgba(22, 25, 49, 0.2), 0px 0.5px 0px rgba(22, 25, 49, 0.05)
  padding 1rem
  border-radius 0.5rem
  color var(--color-text, black)
  font-size 0.875rem
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

</style>

<script>
export default {
  props: ["selected", "asideBanners", "asideBannersUrl", "prereq"],
  data: function() {
    return {
      headerCurrent: null,
    };
  },
  async mounted() {
    window.addEventListener("scroll", this.headerActive);
    window.addEventListener("hashchange", this.headerActive);
  },
  beforeDestroy() {
    window.removeEventListener("scroll", this.headerActive);
    window.removeEventListener("hashchange", this.headerActive);
  },
  computed: {
    headersFiltered() {
      return this.$page.headers.filter((e) => {
        const notHidden = !e.title.match(/{hide}/);
        const notPrereq = !e.title.match(/pre-requisite/i);
        return notHidden && notPrereq;
      });
    },
  },
  methods: {
    headerActive(e) {
      const middleY = window.scrollY + 50;
      if (!this.$page.headers) return;
      const headers = this.$page.headers
        .map((h) => ({
          ...h,
          y: document.getElementById(h.slug).getBoundingClientRect().top,
        }))
        .filter((h) => !h.title.match(/\{hide\}/))
        .map((h) => ({
          ...h,
          y: h.y + window.scrollY,
        }));
      headers.forEach((h, i) => {
        const curr = headers[i];
        const next = headers[i + 1];
        if (curr && next) {
          if (middleY >= curr.y && middleY < next.y) {
            return (this.headerCurrent = { ...curr });
          }
        } else {
          if (middleY >= curr.y) {
            return (this.headerCurrent = { ...curr });
          }
        }
      });
    },
  },
};
</script>
