<template lang="pug">
  div
    .container
      .title Contents
      router-link(v-for="card in cards" :to="card.path" tag="a").card
        .card__title {{card.title}}
        .card__description(v-if="card.frontmatter.description") {{card.frontmatter.description}}
        svg(width="12" height="20" viewBox="0 0 12 20" fill="none" xmlns="http://www.w3.org/2000/svg").card__icon
          path(d="M1.5 1.75L9.75 10L1.5 18.25" stroke-width="2" stroke-linecap="round")
</template>

<style lang="stylus" scoped>
.container
  margin-top 3rem

.title
  font-size 1.5rem
  line-height 2rem
  font-weight 700
  margin-bottom 2rem

.card
  box-shadow 0px 2px 4px rgba(22, 25, 49, 0.05), 0px 0px 1px rgba(22, 25, 49, 0.2), 0px 0.5px 0px rgba(22, 25, 49, 0.05)
  border-radius 0.5rem
  padding 1.5rem 2rem
  position relative
  display block
  color inherit
  max-width 40em
  outline-color var(--color-primary, blue)
  transition box-shadow 0.25s ease-out, transform 0.25s ease-out, opacity 0.4s ease-out
  margin-bottom 1.5rem

  &,
  + *
    margin-top 2rem

  & + &
    margin-top 1rem

  &:hover:not(:active)
    box-shadow 0px 12px 24px rgba(22, 25, 49, 0.07), 0px 4px 8px rgba(22, 25, 49, 0.05), 0px 1px 0px rgba(22, 25, 49, 0.05)
    transform translateY(-2px)
    transition-duration 0.1s

  &:active
    opacity 0.7
    transition-duration 0s

  &__title
    color var(--color-text, black)
    font-size 1.25rem
    line-height 1.5rem
    font-weight 700
    margin-right 1.5rem

  &__description
    margin-top .75rem
    font-size .875rem
    line-height 1.25rem
    color var(--color-text-dim, black)

  &__icon
    top 1.5rem
    right 1.5rem
    position absolute
    stroke rgba(0,0,0,0.2)

  &:hover &__icon
    stroke rgba(0,0,0,0.4)

@media screen and (max-width 768px)
  .card
    padding 1.25rem 1.5rem

    &__icon
      top 1.25rem
</style>

<script>
export default {
  computed: {
    cards() {
      return this.$site.pages
        .filter(page => page.path.match(this.$page.path))
        .filter(page => page.path != this.$page.path);
    }
  }
};
</script>
