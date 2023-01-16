<template lang="pug">
  div
    .container
      .menu(@click="$emit('sidebar', true)")
        .menu__icon
          tm-icon-menu(stroke="#888")
      .logo
        router-link(to="/" tag="div").logo__wrapper
          .logo__image__custom(v-if="$themeConfig.logo && $themeConfig.logo.src")
            img(:src="$themeConfig.logo.src").logo__image__custom__img
          .logo__image(v-else)
            component(:is="`tm-logo-${$themeConfig.label || 'sdk'}`")
          .logo__text(v-if="!$themeConfig.logo") {{$site.title}}
      .toolbar(@click="$emit('search', true)")
        .toolbar__item
          .toolbar__item__icon
            icon-search(fill="#888")
</template>

<style lang="stylus" scoped>
.container
  display grid
  grid-template-columns auto 1fr auto
  align-items center

.menu
  padding .75rem

  &__icon
    height 1.5rem
    width 1.5rem
    cursor pointer

.logo

  &__wrapper
    display grid
    grid-auto-flow column
    justify-content center
    gap .5rem
    align-items center
    cursor pointer
    height 3.5rem

  &__image
    width 2rem
    height 2rem
    fill var(--color-primary, blue)

    &__custom
      height 2.5rem

      &__img
        max-width 100%
        max-height 100%

  &__text
    font-weight 600

.toolbar
  grid-column-start -2
  display flex
  justify-content flex-end

  &__item
    cursor pointer
    display flex
    flex-direction row
    align-items center
    padding .75rem

    &__icon
      height 1.5rem
      width 1.5rem
      margin-right .5rem

    &__text
      color var(--color-text-dim, inherit)

@media screen and (max-width: 752px)
  .menu

    &__icon
      visibility visible

  .toolbar__item__text__search
    display none

  .logo__wrapper
    visibility visible

  .toolbar
    visibility visible

    &__item__icon
      margin-right 0
</style>
