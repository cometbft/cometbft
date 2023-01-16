<template lang="pug">
  form(action="https://app.mailerlite.com/webforms/submit/u7o3l7" data-code="u7o3l7" method="POST" target="_blank" rel="noreferrer noopener").newsletter__form__component__internal
    .newsletter__desc
      .newsletter__desc__h1
        label(for="newsletter-email-input") Get Cosmos updates
      .newsletter__desc__p You’ll be notified with Cosmos updates, news and announcements.
    .newsletter__input
      .newsletter__input__input
        input(type="email" required name="fields[email]" id="newsletter-email-input" placeholder="Your email").newsletter__input__input__el
        button(type="submit").newsletter__input__input__button Sign up
        input(type="hidden" name="ml-submit" value="1")

      .newsletter__input__p You can unsubscribe at any time.
</template>

<style lang="stylus" scoped>
.newsletter__form__component__internal
  display grid
  grid-template-columns repeat(auto-fit, minmax(250px, 1fr))
  width 100%
  align-items flex-end
  background-color white

.newsletter

  &__desc
    padding-right 1.5rem

    &__h1
      color var(--color-text, black)
      font-size 1.5rem
      line-height 2rem
      font-weight 600
      margin-bottom 0.5rem

    &__p
      font-size .875rem
      line-height 1.25rem
      color rgba(22, 25, 49, 0.9)

  &__input
    padding-top 1rem

    &__input
      border 1px solid rgba(140, 145, 177, 0.32)
      display grid
      grid-template-columns 1fr auto
      align-items center
      padding .5rem
      border-radius .25rem

      &:focus-within
        border-color var(--color-primary, blue)

      &__button
        padding .5rem
        text-transform uppercase
        color var(--color-link)
        font-weight 600
        cursor pointer
        background none
        border none

      &__el
        border none
        width 100%
        outline none
        padding .5rem
        background none

    &__p
      margin-top .5rem
      color rgba(22, 25, 49, 0.65)
      font-size .8125rem

@media screen and (max-width: 832px)
  .newsletter__form__component__internal
    display grid
    grid-template-columns repeat(auto-fit, minmax(300px, 1fr))

  .newsletter

    &__desc
      padding-right 0
      padding-bottom 1rem

@media screen and (max-width: 480px)
  .newsletter__form__component__internal
      grid-template-columns repeat(auto-fit, minmax(200px, 1fr))
</style>
