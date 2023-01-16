<template lang="pug">
  div
    .container(v-if="$themeConfig && $themeConfig.gutter")
      .wrapper
        .title(v-if="$themeConfig.gutter.title") {{$themeConfig.gutter.title}}
        .links
          a(:href="$themeConfig.gutter.chat.url" target="_blank" rel="noreferrer noopener" :style="{'--bg': $themeConfig.gutter.chat.bg}").links__item.links__item__chat
            tm-logo-chat(style="fill: white; width: 80px; height: 80px;").links__item__logo
            svg(width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg").links__item__icon
              path(d="M1.16962 13.7697C0.876728 14.0626 0.876728 14.5374 1.16962 14.8303C1.46252 15.1232 1.93739 15.1232 2.23028 14.8303L1.16962 13.7697ZM15.106 1.95457C15.3989 1.66168 15.3989 1.1868 15.106 0.893912C14.8131 0.601019 14.3382 0.601019 14.0454 0.893913L15.106 1.95457ZM14.9999 1H15.7499V0.25H14.9999V1ZM4.81811 0.25C4.4039 0.25 4.06811 0.585786 4.06811 1C4.06811 1.41421 4.4039 1.75 4.81811 1.75V0.25ZM14.2499 11.1818C14.2499 11.596 14.5857 11.9318 14.9999 11.9318C15.4141 11.9318 15.7499 11.596 15.7499 11.1818H14.2499ZM2.23028 14.8303L15.106 1.95457L14.0454 0.893913L1.16962 13.7697L2.23028 14.8303ZM14.9999 0.25H4.81811V1.75H14.9999V0.25ZM15.7499 11.1818V1H14.2499V11.1818H15.7499Z" fill="white")
            div(v-html="md($themeConfig.gutter.chat.title)").links__item__title
            div(v-html="md($themeConfig.gutter.chat.text)").links__item__text
          a(:href="$themeConfig.gutter.forum.url" target="_blank" rel="noreferrer noopener" :style="{'--bg': $themeConfig.gutter.forum.bg}").links__item.links__item__forum
            component(:is="`tm-logo-${$themeConfig.gutter.forum.logo}`" style="fill: white; width: 80px; height: 80px;").links__item__logo
            svg(width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg").links__item__icon
              path(d="M1.16962 13.7697C0.876728 14.0626 0.876728 14.5374 1.16962 14.8303C1.46252 15.1232 1.93739 15.1232 2.23028 14.8303L1.16962 13.7697ZM15.106 1.95457C15.3989 1.66168 15.3989 1.1868 15.106 0.893912C14.8131 0.601019 14.3382 0.601019 14.0454 0.893913L15.106 1.95457ZM14.9999 1H15.7499V0.25H14.9999V1ZM4.81811 0.25C4.4039 0.25 4.06811 0.585786 4.06811 1C4.06811 1.41421 4.4039 1.75 4.81811 1.75V0.25ZM14.2499 11.1818C14.2499 11.596 14.5857 11.9318 14.9999 11.9318C15.4141 11.9318 15.7499 11.596 15.7499 11.1818H14.2499ZM2.23028 14.8303L15.106 1.95457L14.0454 0.893913L1.16962 13.7697L2.23028 14.8303ZM14.9999 0.25H4.81811V1.75H14.9999V0.25ZM15.7499 11.1818V1H14.2499V11.1818H15.7499Z" fill="white")
            div(v-html="md($themeConfig.gutter.forum.title)").links__item__title
            div(v-html="md($themeConfig.gutter.forum.text)").links__item__text
          a(:href="editLink" target="_blank" rel="noreferrer noopener").links__item.links__item__regular
            tm-icon-paper-pen(fill="var(--color-link)" style="width: 80px; height: 80px;").links__item__logo
            svg(width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg").links__item__icon
              path(d="M1.16962 13.7697C0.876728 14.0626 0.876728 14.5374 1.16962 14.8303C1.46252 15.1232 1.93739 15.1232 2.23028 14.8303L1.16962 13.7697ZM15.106 1.95457C15.3989 1.66168 15.3989 1.1868 15.106 0.893912C14.8131 0.601019 14.3382 0.601019 14.0454 0.893913L15.106 1.95457ZM14.9999 1H15.7499V0.25H14.9999V1ZM4.81811 0.25C4.4039 0.25 4.06811 0.585786 4.06811 1C4.06811 1.41421 4.4039 1.75 4.81811 1.75V0.25ZM14.2499 11.1818C14.2499 11.596 14.5857 11.9318 14.9999 11.9318C15.4141 11.9318 15.7499 11.596 15.7499 11.1818H14.2499ZM2.23028 14.8303L15.106 1.95457L14.0454 0.893913L1.16962 13.7697L2.23028 14.8303ZM14.9999 0.25H4.81811V1.75H14.9999V0.25ZM15.7499 11.1818V1H14.2499V11.1818H15.7499Z" fill="black")
            div(v-html="md($themeConfig.gutter.github.title)").links__item__title
            div(v-html="md($themeConfig.gutter.github.text)" style="color: rgba(22, 25, 49, 0.65)").links__item__text
      .newsletter(v-if="$themeConfig.label == 'sdk'")
        .newsletter__image
          .newsletter__image__item(v-for="(item, index) in range(15)" :class="[`letter__${index}`]")
            image-letter
        .newsletter__form
          tm-newsletter-form
</template>

<style lang="stylus" scoped>
.newsletter
  box-shadow 0px 2px 4px rgba(22, 25, 49, 0.05), 0px 0px 1px rgba(22, 25, 49, 0.2), 0px 0.5px 0px rgba(22, 25, 49, 0.05)
  margin-top 4rem
  margin-bottom 1rem
  overflow hidden
  min-height 200px
  border-radius 0.5rem
  position relative
  display flex
  align-items center
  justify-content flex-end

  &:hover &__image__item
    opacity 0.25
    transform scale(0.98)

  &:hover &__image__item.letter__10
    transform scale(1.5) rotate(25deg)
    opacity 1

  &__image
    transform translate(-125px, -35px) rotate(-25deg)
    display grid
    width 300px
    gap 1rem
    grid-template-columns repeat(4, 1fr)
    position absolute
    left 0
    top 0

    &__item
      opacity 0.5
      transition all 0.5s

      &.letter__10
        opacity 1
        background white

  &__form
    margin-left 250px
    padding 30px
    width 100%

/deep/
  a[href]
    color var(--color-link)

  strong
    font-weight 700

strong
  font-weight 600

.container
  background var(--sidebar-bg)

.wrapper
  max-width calc(1400px - var(--sidebar-width))

.title
  font-size 2rem
  color var(--color-text, black)
  padding 1.5rem 0
  font-weight 700

.links
  display grid
  margin-top 1rem
  gap 2rem
  grid-template-columns repeat(auto-fit, minmax(250px, 1fr))

  &__item
    display flex
    flex-direction column
    align-items center
    box-shadow 0px 2px 4px rgba(22, 25, 49, 0.05), 0px 0px 1px rgba(22, 25, 49, 0.2), 0px 0.5px 0px rgba(22, 25, 49, 0.05)
    text-align center
    border-radius 0.5rem
    padding 2rem
    line-height 1.25rem
    background var(--bg)
    transition box-shadow 0.25s ease-out, transform 0.25s ease-out, opacity 0.4s ease-out
    position relative

    &__icon
      position absolute
      top 0
      right 0
      padding 1rem
      width 1rem
      height 1rem
      opacity .26
      transition opacity .15s

    &:hover &__icon
      opacity .36

    &__text
      font-size .875rem
      line-height 1.25rem

    &:hover:not(:active),
    &:focus:not(:active)
      box-shadow 0px 12px 24px rgba(22, 25, 49, 0.07), 0px 4px 8px rgba(22, 25, 49, 0.05), 0px 1px 0px rgba(22, 25, 49, 0.05)
      transform translateY(-2px)
      transition-duration 0.1s

    &:active
      opacity 0.7
      transition-duration 0s

    &__title
      margin-top 1.5rem
      margin-bottom 0.75rem
      font-size 1.25rem
      line-height 1.75rem
      font-weight 600

a.links__item
  color white

a.links__item__regular
  color var(--color-text, black)
  background rgba(176, 180, 207, 0.09)

@media screen and (max-width: 832px)
  .newsletter
    height initial

    &__form
      margin-top 250px
      margin-left initial
      grid-template-columns 1fr
</style>

<script>
import { range } from "lodash";

const endingSlashRE = /\/$/;
const outboundRE = /^[a-z]+:/i;

export default {
  computed: {
    editLink() {
      if (this.$page.frontmatter.editLink === false) {
        return;
      }
      const {
        repo,
        editLinks,
        docsDir = "",
        docsBranch = "master",
        docsRepo = repo
      } = this.$site.themeConfig;
      if (docsRepo && editLinks && this.$page.relativePath) {
        return this.createEditLink(
          repo,
          docsRepo,
          docsDir,
          docsBranch,
          this.$page.relativePath
        );
      }
    },
    editLinkText() {
      return (
        this.$themeLocaleConfig.editLinkText ||
        this.$site.themeConfig.editLinkText ||
        `Edit this page`
      );
    }
  },
  methods: {
    createEditLink(repo, docsRepo, docsDir, docsBranch, path) {
      const bitbucket = /bitbucket.org/;
      if (bitbucket.test(repo)) {
        const base = outboundRE.test(docsRepo) ? docsRepo : repo;
        return (
          base.replace(endingSlashRE, "") +
          `/src` +
          `/${docsBranch}/` +
          (docsDir ? docsDir.replace(endingSlashRE, "") + "/" : "") +
          path +
          `?mode=edit&spa=0&at=${docsBranch}&fileviewer=file-view-default`
        );
      }
      const base = outboundRE.test(docsRepo)
        ? docsRepo
        : `https://github.com/${docsRepo}`;
      return (
        base.replace(endingSlashRE, "") +
        `/edit` +
        `/${docsBranch}/` +
        (docsDir ? docsDir.replace(endingSlashRE, "") + "/" : "") +
        path
      );
    },
    range
  }
};
</script>
